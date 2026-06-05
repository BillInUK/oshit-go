package svc

import (
	"context"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math"
	"oshit-go/app/reward/api/internal/config"
	core_context "oshit-go/app/reward/api/internal/context"
	rewardrpc "oshit-go/app/reward/api/internal/rpc"
	"oshit-go/app/reward/api/internal/task"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"strconv"
)

type ServiceContext struct {
	core_context.CoreContext
	LevelDist           *model.LevelDist
	LevelRatio          []model.LevelRatio
	LevelRatioMap       map[int32]model.LevelRatio
	DiscountRate        *model.DiscountRate
	TakeTokenConfig     *model.TakeTokenConfig
	GiveTokenConfig     *model.GiveTokenConfig
	LotteryConfig       *model.LotteryConfig
	CampaignQuoteConfig *model.CampaignQuoteConfig
	RewardCodeConfig    *model.RewardCodeConfig
	//RewardKeyMap        map[string]solana.PrivateKey
	LightHouseAddress solana.PublicKey
	TaskMgr           *task.TaskManager
	CampaignClientV1  *rewardrpc.CampaignClient
}

const (
	decryptAlgo = "PBEWithHMACSHA512AndAES_256"
	decryptPwd  = "fktYimwMl3OfUF3m"
)

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

	// 初始化 Redis
	redisClient, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}

	pool := goredis.NewPool(*redisClient)
	redSync := redsync.New(pool)

	// 创建 ServiceContext
	srvCtx := &ServiceContext{
		CoreContext: core_context.CoreContext{
			Config:  cfg,
			DB:      db,
			Redis:   *redisClient,
			RedSync: *redSync,
			Ctx:     context.Background(),
		},
	}

	// 初始化数据库配置
	if err := srvCtx.initDatabaseConfigs(); err != nil {
		return nil, err
	}

	if err := srvCtx.initNacosConfigClient(); err != nil {
		fmt.Printf("Init nacos config client error: %v\n", err)
	} else if err := srvCtx.initNacosRuntimeConfigs(); err != nil {
		fmt.Printf("Init nacos runtime configs error: %v\n", err)
	}

	// 初始化Solana RPC客户端
	srvCtx.initSolanaRPC()

	// 初始化Kafka生产者
	if err := srvCtx.initKafkaProducer(); err != nil {
		fmt.Printf("Init kafka producer error: %v\n", err)
	}

	// 初始化Kafka消费者
	if err := srvCtx.initKafkaConsumer(); err != nil {
		fmt.Printf("Init kafka consumer error: %v\n", err)
	}

	// 初始化 Base 模块 RPC 客户端
	if err := srvCtx.initBaseClient(); err != nil {
		fmt.Printf("Init base client error: %v\n", err)
	}

	// 初始化 Campaign 模块 RPC 客户端
	if err := srvCtx.initCampaignClient(); err != nil {
		fmt.Printf("Init campaign client error: %v\n", err)
	}

	// 初始化 dtoken 管理器（JWT 鉴权）
	utils.InitDTokenManager()

	// 初始化任务管理器
	srvCtx.startTasks()

	if err := srvCtx.listenNacosConfigs(); err != nil {
		fmt.Printf("Listen nacos configs error: %v\n", err)
	}

	return srvCtx, nil
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
	s.FeeTolerance = &feeTolerance

	// 初始化LightHouse地址
	lighthouseAddr, err := solana.PublicKeyFromBase58("L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95")
	if err != nil {
		return fmt.Errorf("invalid lighthouse address: %v", err)
	}
	s.LightHouseAddress = lighthouseAddr

	// 初始化奖励层级和每个层级的抽取费用配置
	var levelDist model.LevelDist
	if err := s.DB.First(&levelDist).Error; err != nil {
		return fmt.Errorf("can not load reward level dist from database: %v", err)
	}
	s.LevelDist = &levelDist

	var levelRatio []model.LevelRatio
	if err := s.DB.Find(&levelRatio).Error; err != nil {
		return fmt.Errorf("can not load reward level ratio from database: %v", err)
	}
	s.LevelRatio = levelRatio

	s.LevelRatioMap = make(map[int32]model.LevelRatio)
	for _, ratio := range s.LevelRatio {
		s.LevelRatioMap[ratio.DistLevel] = ratio
	}

	// 加载折扣率配置
	var discountRate model.DiscountRate
	if err := s.DB.First(&discountRate).Error; err != nil {
		return fmt.Errorf("can not load discount rate from database: %v", err)
	}
	s.DiscountRate = &discountRate

	// 加载 take token 业务配置
	var takeTokenConfig model.TakeTokenConfig
	if err := s.DB.First(&takeTokenConfig).Error; err != nil {
		return fmt.Errorf("can not find take token config from database: %v", err)
	}
	s.TakeTokenConfig = &takeTokenConfig

	// 加载 give token 业务配置
	var giveTokenConfig model.GiveTokenConfig
	if err := s.DB.First(&giveTokenConfig).Error; err != nil {
		return fmt.Errorf("can not find take give config from database: %v", err)
	}
	s.GiveTokenConfig = &giveTokenConfig

	// 加载 lottery 业务配置
	var lotteryConfig model.LotteryConfig
	if err := s.DB.First(&lotteryConfig).Error; err != nil {
		return fmt.Errorf("can not find lottery config from database: %v", err)
	}
	s.LotteryConfig = &lotteryConfig

	// 加载 campaign 业务配置
	var campaignQuoteConfig model.CampaignQuoteConfig
	if err := s.DB.First(&campaignQuoteConfig).Error; err != nil {
		return fmt.Errorf("can not find campaign exchange config from database: %v", err)
	}
	s.CampaignQuoteConfig = &campaignQuoteConfig

	// 加载 reward code 业务配置
	var rewardCodeConfig model.RewardCodeConfig
	if err := s.DB.First(&rewardCodeConfig).Error; err != nil {
		return fmt.Errorf("can not find reward code config from database: %v", err)
	}
	s.RewardCodeConfig = &rewardCodeConfig

	// 加载私钥
	//s.RewardKeyMap = make(map[string]solana.PrivateKey)
	// 初始化所有业务的发送奖励私钥
	//table := s.DB.Table(model.TableNameRewardKeyConfig)
	//var rewardKeyConfigs []model.RewardKeyConfig
	//if err := table.Find(&rewardKeyConfigs).Error; err != nil {
	//	return fmt.Errorf("can not load encrypted reward key config from database")
	//}
	//if len(rewardKeyConfigs) == 0 {
	//	panic("can not load any reward key from database")
	//}
	//for _, keyConfig := range rewardKeyConfigs {
	//	privateKey, err := utils.JasyptDecrypt(keyConfig.EncryptedKey, decryptPwd, decryptAlgo)
	//	if err != nil {
	//		return fmt.Errorf("decrypt service %s private key from database error: %v", keyConfig.Service, err)
	//	}
	//	if decryptedKey, err := solana.PrivateKeyFromBase58(privateKey); err != nil {
	//		return fmt.Errorf("malformed service %s private key error: %v", keyConfig.Service, err)
	//	} else {
	//		s.RewardKeyMap[keyConfig.Service] = decryptedKey
	//	}
	//}
	return nil
}

func (s *ServiceContext) initSolanaRPC() {
	// RPC clients are now initialized from Nacos config or via Dubbo client
}

func (s *ServiceContext) startTasks() {
	taskCtx := &task.TaskContext{
		CoreContext: s.CoreContext,
	}
	s.TaskMgr = task.NewTaskManager(taskCtx)
}

func (s *ServiceContext) initBaseClient() error {
	nacosCfg := s.Config.Nacos
	host := nacosCfg.Host
	port := nacosCfg.Port
	if len(nacosCfg.ServerConfig) > 0 {
		host = nacosCfg.ServerConfig[0].Host
		port = nacosCfg.ServerConfig[0].Port
	}
	if port == 0 {
		port = 8848
	}
	nacosAddr := fmt.Sprintf("%s:%d", host, port)

	username := nacosCfg.ClientConfig.Username
	password := nacosCfg.ClientConfig.Password
	if username == "" {
		username = nacosCfg.Username
		password = nacosCfg.Password
	}

	cli, err := rewardrpc.NewBaseClient(
		nacosAddr,
		s.Config.App.Name,
		nacosCfg.ClientConfig.NamespaceId,
		username,
		password,
	)
	if err != nil {
		return fmt.Errorf("init base client error: %w", err)
	}
	s.BaseClient = cli
	return nil
}

func (s *ServiceContext) initCampaignClient() error {
	if s.SystemConfig.Env == 0 {
		s.CampaignClientV1 = rewardrpc.NewCampaignClient("http://172.31.48.20:4000/internal/api/v1", map[string]string{})
	} else {
		s.CampaignClientV1 = rewardrpc.NewCampaignClient("http://172.31.48.157:80/internal/api/v1", map[string]string{})
	}
	return nil
}

func (s *ServiceContext) Close() error {

	// 关闭Kafka消费者
	if err := s.closeKafkaConsumer(); err != nil {
		fmt.Printf("Close kafka consumer error: %v\n", err)
	}

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
