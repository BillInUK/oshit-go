package svc

import (
	"context"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math"
	"oshit-go/app/pos/api/internal/config"
	core_context "oshit-go/app/pos/api/internal/context"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/task"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"strconv"
)

type ServiceContext struct {
	core_context.CoreContext
	LightHouseAddress solana.PublicKey
	TaskMgr           *task.TaskManager

	// pos业务配置
	PosStarLevelRule map[int32]model.PosStarLevelRule
	PosRewardConfig  *model.PosRewardConfig
	PosWhiteListMap  map[string]model.PosStarWhitelist

	// stake业务配置
	StakeAmmConfig     *model.StakeAmmConfig
	StakeRewardConfig  *model.StakeRewardConfig
	LeaderRewardConfig *model.StakeLeaderRewardConfig
	StakeFixConfig     map[int32]model.StakeFixRateConfig
	StakeInviteRate    map[int32]model.StakeInviteRate
	StakeStarLevelRule map[int32]model.StakeStarLevelRule
	TotalAreaLeaders   []model.StakeTotalLeader
	StakeTokenPoolMap  map[string]model.StakeTokenPool
	StakeDistLevel     int32
	StakeStarWhitelist map[string]model.StakeStarWhitelist
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

	// 初始化快照 Kafka 消费者（PosTopic + StakeTopic）
	if err := srvCtx.initSnapShotKafkaConsumer(); err != nil {
		fmt.Printf("Init snapshot kafka consumer error: %v\n", err)
	}

	// 初始化 Base 模块 RPC 客户端
	if err := srvCtx.initBaseClient(); err != nil {
		fmt.Printf("Init base client error: %v\n", err)
	}

	// 初始化pos配置
	//if err := srvCtx.initPosConfig(); err != nil {
	//	fmt.Printf("Init pos config error: %v", err)
	//}

	// 初始化stake配置
	if err := srvCtx.initStakeConfig(); err != nil {
		return nil, err
	}

	// 初始化 dtoken 管理器（JWT 鉴权）
	utils.InitDTokenManager()

	// 初始化任务管理器
	srvCtx.startTasks()

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
	if err := s.DB.Where("chain = ?", "SOL").First(&chainConfig).Error; err != nil {
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

	return nil
}

func (s *ServiceContext) initSolanaRPC() {
	// 初始化Solana RPC客户端
	if s.ChainConfig != nil && s.ChainConfig.RPCURL != "" {
		s.RpcClient = rpc.New(s.ChainConfig.RPCURL)
	}
}

func (s *ServiceContext) initBaseClient() error {
	nacosServers := s.Config.Nacos.ServerConfig
	if len(nacosServers) == 0 {
		return fmt.Errorf("nacos server config is empty")
	}
	nacosAddr := fmt.Sprintf("%s:%d", nacosServers[0].Host, nacosServers[0].Port)
	cli, err := posrpc.NewBaseClient(nacosAddr, s.Config.App.Name)
	if err != nil {
		return fmt.Errorf("init base client error: %w", err)
	}
	s.BaseClient = cli
	return nil
}

func (s *ServiceContext) initPosConfig() error {
	// 初始化pos星级白名单
	// TODO: 可以放到配置中心
	var whitelists []model.PosStarWhitelist
	table := s.DB.Table(model.TableNamePosStarWhitelist)
	if err := table.Find(&whitelists).Order("star_level asc").Error; err != nil {
		return errors.New("can not load any pos star level config from database")
	}
	s.PosWhiteListMap = make(map[string]model.PosStarWhitelist)
	for _, c := range whitelists {
		s.PosWhiteListMap[c.NativeAccount] = c
	}

	// 初始化pos星级评定规则
	var starLevelRules []model.PosStarLevelRule
	table = s.DB.Table(model.TableNamePosStarLevelRule)
	if err := table.Find(&starLevelRules).Order("star_level asc").Error; err != nil {
		return errors.New("can not load any pos star level rule from database")
	}
	for _, r := range s.PosStarLevelRule {
		s.PosStarLevelRule[r.StarLevel] = r
	}

	// 初始化pos奖励配置
	var rewardConfig model.PosRewardConfig
	table = s.DB.Table(model.TableNamePosRewardConfig)
	if err := table.Find(&s.PosRewardConfig).First(&rewardConfig).Error; err != nil {
		return errors.New("can not load any pos reward config from database")
	}
	s.PosRewardConfig = &rewardConfig

	return nil
}

func (s *ServiceContext) initStakeConfig() error {
	var dist model.StakeInviteDist
	var fixConfig []model.StakeFixRateConfig
	var inviteRates []model.StakeInviteRate
	var stakeTokenPools []model.StakeTokenPool
	s.StakeFixConfig = make(map[int32]model.StakeFixRateConfig)
	s.StakeInviteRate = make(map[int32]model.StakeInviteRate)
	s.StakeStarLevelRule = make(map[int32]model.StakeStarLevelRule)
	s.StakeTokenPoolMap = make(map[string]model.StakeTokenPool)

	table := s.DB.Table(model.TableNameStakeInviteDist)
	if err := table.Find(&dist).Error; err != nil {
		panic("can not find any stake invite dist config")
	}
	s.StakeDistLevel = dist.DistLevel

	table = s.DB.Table(model.TableNameStakeInviteRate)
	if err := table.Find(&inviteRates).Error; err != nil || len(inviteRates) == 0 {
		return errors.New("can not find any stake invite rate config")
	}
	for _, rate := range inviteRates {
		s.StakeInviteRate[rate.DistLevel] = rate
	}
	table = s.DB.Table(model.TableNameStakeFixRateConfig)
	if err := table.Find(&fixConfig).Error; err != nil || len(fixConfig) == 0 {
		return errors.New("can not find any stake fix config")
	}
	for _, c := range fixConfig {
		s.StakeFixConfig[c.StakeType] = c
	}

	// 初始化stake星级配置
	var startLevelConfigs []model.StakeStarWhitelist
	table = s.DB.Table(model.TableNameStakeStarWhitelist)
	if err := table.Find(&startLevelConfigs).Order("star_level asc").Error; err != nil {
		return errors.New("can not load any pos star level config from database")
	}
	s.StakeStarWhitelist = make(map[string]model.StakeStarWhitelist)
	for _, c := range startLevelConfigs {
		s.StakeStarWhitelist[c.NativeAccount] = c
	}
	var starLevelRules []model.StakeStarLevelRule
	table = s.DB.Table(model.TableNameStakeStarLevelRule)
	if err := table.Find(&starLevelRules).Order("star_level asc").Error; err != nil {
		return errors.New("can not load any pos star level rule from database")
	}
	for _, r := range s.StakeStarLevelRule {
		s.StakeStarLevelRule[r.StarLevel] = r
	}
	var rewardConfig model.StakeRewardConfig
	table = s.DB.Table(model.TableNameStakeRewardConfig)
	if err := table.First(&rewardConfig).Error; err != nil {
		return errors.New("can not load any pos reward rule from database")
	}
	s.StakeRewardConfig = &rewardConfig
	table = s.DB.Table(model.TableNameStakeLeaderRewardConfig)
	if err := table.First(&s.LeaderRewardConfig).Error; err != nil {
		return errors.New("can not load any pos reward rule from database")
	}
	table = s.DB.Table(model.TableNameStakeAmmConfig)
	if err := table.First(&s.StakeAmmConfig).Error; err != nil {
		return errors.New("can not load any stake amm config from database")
	}
	table = s.DB.Table(model.TableNameStakeTotalLeader)
	if err := table.Find(&s.TotalAreaLeaders).Error; err != nil {
		return errors.New("can not find total area leaders from database")
	}
	if len(s.TotalAreaLeaders) != 2 {
		panic("expected exactly 2 total area leaders in database")
	}
	table = s.DB.Table(model.TableNameStakeTokenPool)
	if err := table.Find(&stakeTokenPools).Error; err != nil {
		return errors.New("can not load any stake pool config from database")
	}
	for _, tokenPool := range stakeTokenPools {
		s.StakeTokenPoolMap[tokenPool.FromTokenAccount] = tokenPool
	}
	return nil
}

func (s *ServiceContext) startTasks() {
	taskCtx := &task.TaskContext{
		CoreContext:           s.CoreContext,
		RewardConfig:          s.StakeRewardConfig,
		SnapShotKafkaConsumer: s.SnapShotKafkaConsumer,
	}
	s.TaskMgr = task.NewTaskManager(taskCtx)
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
