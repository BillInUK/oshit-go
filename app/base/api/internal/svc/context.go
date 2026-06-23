package svc

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math"
	"oshit-go/app/base/api/internal/config"
	core_context "oshit-go/app/base/api/internal/context"
	"oshit-go/app/base/api/internal/task"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"strconv"
)

const (
	serviceKeyDecryptAlgo = "PBEWithHMACSHA512AndAES_256"
	rsaPublicKey          = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCym6SwEHnkHqpVcS9sxP4I2D4b
aSxPflUNtEqE0dmLfbA8kZw7Rs8eGkUj4kOEMSZA4y4jtp1wn0QJJF31Obop60j1
9j3KtTuSLBY9xuJoGNMxzYZCybzxcp+h2olUsp0SrjEfs/Z6ePY0k+5+0umwbvM4
+7CsfFwcASSNYuCbrwIDAQAB
-----END PUBLIC KEY-----`
)

type ServiceContext struct {
	core_context.CoreContext
	TaskMgr              *task.TaskManager
	serviceKeyDecryptPwd string // AES key for config fields (DB/Redis/Nacos passwords, API keys)
	privKeyDecryptPwd    string // AES key for Solana private keys (t_service_key / Nacos encrypted_key)
}

func NewServiceContext() (*ServiceContext, error) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// 3.1 加载配置解密密钥（用于 DB/Redis/Nacos 等配置密码）
	decryptKey, err := utils.LoadConfigDecryptKey()
	if err != nil {
		return nil, fmt.Errorf("load config decrypt key: %w", err)
	}

	// 3.2 加载私钥解密密钥（用于 t_service_key / Nacos 中的 Solana 私钥）
	privKeyDecryptKey, err := utils.LoadPrivKeyDecryptKey()
	if err != nil {
		return nil, fmt.Errorf("load priv key decrypt key: %w", err)
	}

	// 解密 application.yaml 中的 ENC~ 字段（DB密码、Redis密码、Nacos凭据等）
	if err := utils.JasyptDecode(cfg, decryptKey, serviceKeyDecryptAlgo); err != nil {
		return nil, fmt.Errorf("decrypt application config: %w", err)
	}

	// 初始化数据库
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}

	// 初始化Redis
	redisClient, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}

	pool := goredis.NewPool(*redisClient)
	redSync := redsync.New(pool)

	// 创建ServiceContext
	svcCtx := &ServiceContext{
		serviceKeyDecryptPwd: decryptKey,
		privKeyDecryptPwd:    privKeyDecryptKey,
		CoreContext: core_context.CoreContext{
			Config:  cfg,
			DB:      db,
			Redis:   *redisClient,
			RedSync: *redSync,
			Ctx:     context.Background(),
		},
	}

	// 初始化RSA公钥
	if err := svcCtx.initRSAPublicKey(); err != nil {
		return nil, err
	}

	// 初始化数据库配置
	if err := svcCtx.initDatabaseConfigs(); err != nil {
		return nil, err
	}

	// 初始化服务签名私钥
	if err := svcCtx.initServiceKeys(); err != nil {
		return nil, err
	}

	// 初始化 Nacos 配置中心，并用 Nacos 配置覆盖 DB 启动配置（本地未配置时回退 DB）
	if err := svcCtx.initNacosConfigClient(); err != nil {
		fmt.Printf("Init nacos config client error, fallback to database configs: %v\n", err)
	}
	initialScanConfigs, err := svcCtx.initNacosRuntimeAndRegistry()
	if err != nil {
		return nil, err
	}

	// 初始化Solana RPC客户端
	svcCtx.initSolanaRPC()

	// 初始化Kafka生产者
	if err := svcCtx.initKafkaProducer(); err != nil {
		fmt.Printf("Init kafka producer error: %v\n", err)
	}

	// 初始化 dtoken 管理器（JWT 鉴权）
	utils.InitDTokenManager()

	// 初始化任务管理器
	svcCtx.startTasks()
	if initialScanConfigs != nil && svcCtx.TaskMgr != nil {
		if err := svcCtx.TaskMgr.ReconcileScanConfigs(initialScanConfigs); err != nil {
			return nil, err
		}
	}
	if err := svcCtx.listenNacosConfigs(); err != nil {
		fmt.Printf("Listen nacos configs error: %v\n", err)
	}

	return svcCtx, nil
}

func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	portStr := strconv.Itoa(cfg.Port)
	dsn := "host=" + cfg.Host +
		" user=" + cfg.User +
		" password=" + cfg.Password +
		" dbname=" + cfg.DBName +
		" port=" + portStr +
		" sslmode=" + cfg.SSLMode

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
}

func initRedis(cfg config.RedisConfig) (*redis.UniversalClient, error) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		MasterName: cfg.MasterName,
		Addrs:      cfg.Hosts,
		Password:   cfg.Password,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &client, nil
}

func (s *ServiceContext) initRSAPublicKey() error {
	pubKey, err := utils.ParseRsaPublicKeyFromPemStr(rsaPublicKey)
	if err != nil {
		return fmt.Errorf("read app global pub key error: %v", err)
	}
	s.AppGlobalPublicKey = pubKey
	return nil
}

func (s *ServiceContext) initDatabaseConfigs() error {
	// 初始化系统配置
	var systemConfig model.SystemConfig
	if err := s.DB.First(&systemConfig).Error; err != nil {
		return fmt.Errorf("can not load system config: %v", err)
	}
	s.SystemConfig = &systemConfig

	// 初始化链配置 - SOL链
	var chainConfig model.ChainConfig
	if err := s.DB.Where("chain_name = ?", "solana").First(&chainConfig).Error; err != nil {
		return fmt.Errorf("can not load solana chain configure of chain SOL from database: %v", err)
	}
	s.ChainConfig = &chainConfig

	// Load RPC endpoints from DB
	var rpcEndpoints []model.RpcEndpoint
	if err := s.DB.Where("weight > 0").Find(&rpcEndpoints).Error; err != nil {
		fmt.Printf("Warning: cannot load rpc endpoints from database: %v\n", err)
	}
	// 解密 ENC~ 字段（api_key, wss_api_key）
	for i := range rpcEndpoints {
		if err := utils.JasyptDecode(&rpcEndpoints[i], s.serviceKeyDecryptPwd, serviceKeyDecryptAlgo); err != nil {
			return fmt.Errorf("decrypt rpc endpoint: %w", err)
		}
	}
	var envEndpoints []utils.RPCEndpointConfig
	for _, ep := range rpcEndpoints {
		cfg := utils.RPCEndpointConfig{
			Provider:    ep.Provider,
			Endpoint:    ep.Endpoint,
			APIKey:      ep.APIKey,
			WssEndpoint: ep.WssEndpoint,
			WssAPIKey:   ep.WssAPIKey,
			Weight:      int(ep.Weight),
		}
		if ep.Scope == "env" {
			envEndpoints = append(envEndpoints, cfg)
		} else if ep.Scope == "mainnet" {
			mainnetURL := utils.BuildRPCURL(ep.Provider, ep.Endpoint, ep.APIKey)
			s.MainnetRpcClient = rpc.New(mainnetURL)
			s.MainnetRpcURL = mainnetURL
			if ep.Provider == "helius" && ep.APIKey != "" {
				s.HeliusAPIKey = ep.APIKey
			}
		}
	}
	if len(envEndpoints) > 0 {
		pool, err := utils.NewRPCPool(envEndpoints)
		if err != nil {
			return fmt.Errorf("create rpc pool from db: %v", err)
		}
		s.RpcPool = pool
		s.RpcClient = pool.First()
	}

	// 初始化Token配置
	var tokenConfig model.TokenConfig
	if err := s.DB.First(&tokenConfig).Error; err != nil {
		return fmt.Errorf("can not load solana token configure from database: %v", err)
	}
	s.TokenConfig = &tokenConfig

	// 计算TokenDecimal
	s.TokenDecimal = math.Pow(10, float64(tokenConfig.Decimals))

	// 初始化手续费容错配置
	var feeTolerance model.FeeTolerance
	if err := s.DB.First(&feeTolerance).Error; err != nil {
		return fmt.Errorf("can not load solana fee tolerance from database: %v", err)
	}
	s.FeeTolerance = feeTolerance

	// 初始化AWS配置
	var awsConfig model.AwsConfig
	if err := s.DB.First(&awsConfig).Error; err != nil {
		return fmt.Errorf("can not load aws config from database: %v", err)
	}
	// 解密 ENC~ 字段（access_key_id, secret_access_key）
	if err := utils.JasyptDecode(&awsConfig, s.serviceKeyDecryptPwd, serviceKeyDecryptAlgo); err != nil {
		return fmt.Errorf("decrypt aws config: %w", err)
	}
	s.AwsConfig = &awsConfig

	// 初始化LightHouse地址
	lighthouseAddr, err := solana.PublicKeyFromBase58("L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95")
	if err != nil {
		return fmt.Errorf("invalid lighthouse address: %v", err)
	}
	s.LightHouseAddress = lighthouseAddr

	// 初始化服务配置
	var serviceInfos []model.ServiceInfo
	if err := s.DB.Find(&serviceInfos).Error; err != nil {
		return fmt.Errorf("can not load service info from database: %v", err)
	}
	s.ServiceInfoMap = make(map[string]map[string]model.ServiceInfo)
	for _, si := range serviceInfos {
		if s.ServiceInfoMap[si.Service] == nil {
			s.ServiceInfoMap[si.Service] = make(map[string]model.ServiceInfo)
		}
		s.ServiceInfoMap[si.Service][si.SubService] = si
	}

	return nil
}

func (s *ServiceContext) initServiceKeys() error {
	var keys []model.ServiceKey
	if err := s.DB.Find(&keys).Error; err != nil {
		return fmt.Errorf("can not load service keys from database: %v", err)
	}
	if len(keys) == 0 {
		return fmt.Errorf("no service keys found in t_service_key")
	}

	s.ServiceKeyMap = make(core_context.ServiceKey)
	for _, k := range keys {
		plainKey, err := utils.JasyptDecrypt(k.EncryptedKey, s.privKeyDecryptPwd, serviceKeyDecryptAlgo)
		if err != nil {
			return fmt.Errorf("decrypt service key [%s/%s] error: %v", k.Service, k.SubService, err)
		}
		privateKey, err := solana.PrivateKeyFromBase58(plainKey)
		if err != nil {
			return fmt.Errorf("malformed service key [%s/%s]: %v", k.Service, k.SubService, err)
		}
		if s.ServiceKeyMap[k.Service] == nil {
			s.ServiceKeyMap[k.Service] = make(map[string]solana.PrivateKey)
		}
		s.ServiceKeyMap[k.Service][k.SubService] = privateKey
	}
	return nil
}

func (s *ServiceContext) initSolanaRPC() {
	// RPC clients are now initialized from t_rpc_endpoint in initDatabaseConfigs
	// and may be overridden by Nacos config in applyRuntimeConfigContent
}

func (s *ServiceContext) startTasks() {
	taskCtx := &task.TaskContext{
		CoreContext: s.CoreContext,
	}
	taskMgr := task.NewTaskManager(taskCtx)
	s.TaskMgr = taskMgr
	taskMgr.StartAllTasks()
}

func (s *ServiceContext) Close() error {

	// 关闭Kafka生产者
	if err := s.closeKafkaProducer(); err != nil {
		fmt.Printf("Close kafka producer error: %v\n", err)
	}

	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			return err
		}
	}

	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
