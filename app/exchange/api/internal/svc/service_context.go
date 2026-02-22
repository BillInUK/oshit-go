package svc

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/client"
	_ "dubbo.apache.org/dubbo-go/v3/client"
	dubbo_config "dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/config_center/nacos"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"dubbo.apache.org/dubbo-go/v3/registry"
	_ "dubbo.apache.org/dubbo-go/v3/registry"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	nacos_client "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	nacos_const "github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	account_pb "oshit-go/app/account/rpc/pb"
	"oshit-go/app/exchange/api/internal/config"
	"strconv"
	"sync"
)

type ServiceContext struct {
	Config      *config.Config
	DB          *gorm.DB
	Redis       *redis.Client
	Ctx         context.Context
	AccountCli  account_pb.AccountService
	NacosCfgCli nacos_client.IConfigClient
}

var (
	configRWMutex sync.RWMutex         // 并发安全读写锁
	nacosOrderCfg config.NacosOrderCfg // nacos配置的全局参数
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

	// 初始化Redis
	redisCli, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}

	// 初始化配置中心
	nacosCfgCli, err := initNacosCfgCli(cfg.Nacos)
	if err != nil {
		return nil, err
	}

	// 初始化dubbo client
	err = initDubboCli()
	if err != nil {
		return nil, err
	}

	// 初始化account的rpc客户端
	accountCli, err := initAccountCli(cfg.Nacos)
	if err != nil {
		return nil, err
	}

	// 返回最终的service context
	return &ServiceContext{
		Ctx:         context.Background(),
		Config:      cfg,
		DB:          db,
		Redis:       redisCli,
		AccountCli:  accountCli,
		NacosCfgCli: nacosCfgCli,
	}, nil
}

// GetNacosOrderCfg 原有安全获取配置方法
func (s *ServiceContext) GetNacosOrderCfg() config.NacosOrderCfg {
	configRWMutex.RLock()
	defer configRWMutex.RUnlock()
	return nacosOrderCfg
}

// Close 程序结束时关闭资源
func (s *ServiceContext) Close() error {
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

// initDatabase 初始化数据库
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

// initRedis 初始化redis
func initRedis(cfg config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// initNacosCfgCli 初始化Nacos配置客户端
func initNacosCfgCli(cfg config.NacosConfig) (nacos_client.IConfigClient, error) {

	// 步骤1：转换为Nacos SDK要求的ServerConfig格式（替代硬编码）
	var serverConfigs []nacos_const.ServerConfig
	for _, s := range cfg.ServerConfig {
		serverConfigs = append(serverConfigs, nacos_const.ServerConfig{
			IpAddr: s.Host,
			Port:   s.Port,
		})
	}

	// 步骤2：转换为Nacos SDK要求的ClientConfig格式（替代硬编码）
	clientConfig := nacos_const.ClientConfig{
		NamespaceId:         cfg.ClientConfig.NamespaceId,
		TimeoutMs:           cfg.ClientConfig.TimeoutMs,
		NotLoadCacheAtStart: cfg.ClientConfig.NotLoadCacheAtStart,
		LogDir:              cfg.ClientConfig.LogDir,
		CacheDir:            cfg.ClientConfig.CacheDir,
		LogLevel:            cfg.ClientConfig.LogLevel,
		Username:            cfg.ClientConfig.Username,
		Password:            cfg.ClientConfig.Password,
	}

	// 步骤3：创建Nacos配置客户端（逻辑不变，参数来源改为配置文件）
	cfgClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("创建Nacos配置客户端失败: %v", err)
	}

	// 步骤4：监听业务配置变更（改用配置文件中的DataId/Group）
	if err := listenConfigChange(cfg, cfgClient); err != nil {
		return nil, fmt.Errorf("监听业务配置变更失败: %v", err)
	}

	fmt.Println("Nacos配置中心初始化成功，初始业务配置：", nacosOrderCfg)
	return cfgClient, nil
}

// listenConfigChange 监听Nacos配置变更
func listenConfigChange(cfg config.NacosConfig, cli nacos_client.IConfigClient) error {
	// 从配置文件中获取订阅的DataId和Group，替代硬编码
	subscribeCfg := cfg.SubscribeConfig
	return cli.ListenConfig(vo.ConfigParam{
		DataId: subscribeCfg.DataId,
		Group:  subscribeCfg.Group,
		OnChange: func(namespace, group, dataId, content string) {
			fmt.Printf("Nacos业务配置已变更（DataId：%s），新配置：%s\n", dataId, content)
			// 重新解析配置（逻辑不变）
			configRWMutex.Lock()
			defer configRWMutex.Unlock()
			if err := json.Unmarshal([]byte(content), &nacosOrderCfg); err != nil {
				fmt.Printf("解析变更后的Nacos业务配置失败: %v\n", err)
				return
			}
			fmt.Println("全局业务配置已更新：", nacosOrderCfg)
		},
	})
}

// initDubboCli 初始化dubbo client
func initDubboCli() error {
	if err := dubbo_config.Load(dubbo_config.WithPath("./etc/dubbo.yaml")); err != nil {
		return fmt.Errorf("dubbo-go 加载配置文件失败: %v", err)
	}
	fmt.Println("✅ dubbo-go消费端初始化成功（配置文件驱动，Nacos注册+配置中心集成完成）")
	return nil
}

// initAccountCli 初始化AccountService的Rpc客户端
func initAccountCli(cfg config.NacosConfig) (account_pb.AccountService, error) {

	userName := cfg.ClientConfig.Username
	password := cfg.ClientConfig.Password
	host := cfg.ServerConfig[0].Host
	port := cfg.ServerConfig[0].Port
	grpcPort := cfg.ServerConfig[0].GrpcPort

	fmt.Printf("Account RPC客户端 通过注册Nacos host %s port %d grpc-port %d\n", host, port, grpcPort)

	grpcPortStr := strconv.FormatUint(grpcPort, 10)
	cli, err := client.NewClient(
		client.WithClientRegistry(
			registry.WithNacos(),
			registry.WithUsername(userName),
			registry.WithPassword(password),
			registry.WithAddress(fmt.Sprintf("%s:%d", host, port)),
			registry.WithParams(map[string]string{
				"grpc-port":     grpcPortStr,
				"metadata-type": "local",
			}),
		),
	)
	if err != nil {
		fmt.Printf("Account RPC客户端 链接Nacos失败: %v\n", err)
		return nil, err
	}
	accountClient, err := account_pb.NewAccountService(cli)
	if err != nil {
		return nil, fmt.Errorf("创建AccountService代理失败: %w", err)
	}

	fmt.Println("Account RPC客户端 (通过Nacos服务发现) 初始化成功")
	return accountClient, nil
}
