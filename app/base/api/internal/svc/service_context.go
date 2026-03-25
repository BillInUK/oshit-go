package svc

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math"
	"oshit-go/app/base/api/internal/config"
	core_context "oshit-go/app/base/api/internal/context"
	"oshit-go/app/base/api/internal/task"
	"oshit-go/app/base/dal/model"
	"oshit-go/common/utils"
	"strconv"
)

type ServiceContext struct {
	core_context.CoreContext
}

func NewServiceContext() (*ServiceContext, error) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
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

	// 初始化Solana RPC客户端
	svcCtx.initSolanaRPC()

	// 初始化Kafka生产者
	if err := svcCtx.initKafkaProducer(); err != nil {
		fmt.Printf("Init kafka producer error: %v\n", err)
	}

	// 初始化任务管理器
	svcCtx.startTasks()

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

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

const rsaPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCym6SwEHnkHqpVcS9sxP4I2D4b
aSxPflUNtEqE0dmLfbA8kZw7Rs8eGkUj4kOEMSZA4y4jtp1wn0QJJF31Obop60j1
9j3KtTuSLBY9xuJoGNMxzYZCybzxcp+h2olUsp0SrjEfs/Z6ePY0k+5+0umwbvM4
+7CsfFwcASSNYuCbrwIDAQAB
-----END PUBLIC KEY-----`

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
	if err := s.DB.Where("chain = ?", "SOL").First(&chainConfig).Error; err != nil {
		return fmt.Errorf("can not load solana chain configure of chain SOL from database: %v", err)
	}
	s.ChainConfig = &chainConfig

	// 初始化用户钱包RPC配置 - SOL链
	var userWalletRPCConfig model.UserWalletRpcConfig
	if err := s.DB.Where("chain = ?", "SOL").First(&userWalletRPCConfig).Error; err != nil {
		return fmt.Errorf("can not load solana user wallet rpc configure of chain SOL from database: %v", err)
	}
	s.UserWalletRPCConfig = &userWalletRPCConfig

	// 初始化Token配置
	var tokenConfig model.TokenConfig
	if err := s.DB.First(&tokenConfig).Error; err != nil {
		return fmt.Errorf("can not load solana token configure from database: %v", err)
	}
	s.TokenConfig = &tokenConfig

	// 计算TokenDecimal
	s.TokenDecimal = math.Pow(10, float64(tokenConfig.Decimals))

	// 初始化AWS配置
	var awsConfig model.AwsConfig
	if err := s.DB.First(&awsConfig).Error; err != nil {
		return fmt.Errorf("can not load aws config from database: %v", err)
	}
	s.AwsConfig = &awsConfig

	// 初始化LightHouse地址
	lighthouseAddr, err := solana.PublicKeyFromBase58("L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95")
	if err != nil {
		return fmt.Errorf("invalid lighthouse address: %v", err)
	}
	s.LightHouseAddress = lighthouseAddr

	return nil
}

func (s *ServiceContext) initSolanaRPC() {
	// 初始化Solana RPC客户端
	if s.ChainConfig != nil && s.ChainConfig.RPCURL != "" {
		s.RpcClient = rpc.New(s.ChainConfig.RPCURL)
	}

	if s.UserWalletRPCConfig != nil && s.UserWalletRPCConfig.RPCURL != "" {
		s.UserWalletRpcClient = rpc.New(s.UserWalletRPCConfig.RPCURL)
	}
}

func (s *ServiceContext) startTasks() {
	taskCtx := &task.TaskContext{
		CoreContext: s.CoreContext,
	}
	taskMgr := task.NewTaskManager(taskCtx)
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
