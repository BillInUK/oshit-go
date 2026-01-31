package svc

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/client"
	_ "dubbo.apache.org/dubbo-go/v3/client"
	"dubbo.apache.org/dubbo-go/v3/config_center"
	_ "dubbo.apache.org/dubbo-go/v3/config_center/nacos"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"dubbo.apache.org/dubbo-go/v3/registry"
	_ "dubbo.apache.org/dubbo-go/v3/registry"
	"fmt"
	"github.com/dubbogo/gost/log/logger"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	account_pb "oshit-go/app/account/rpc/pb"
	"oshit-go/app/order/api/internal/config"
	"strconv"
)

type ServiceContext struct {
	Config        *config.Config
	DB            *gorm.DB
	Redis         *redis.Client
	Ctx           context.Context
	AccountClient account_pb.AccountService
}

// 自定义配置监听器（可选，用于监听配置变更）
type configListener struct{}

func (l *configListener) Process(event *config_center.ConfigChangeEvent) {
	logger.Infof("配置中心监听器: Key=%s, Value=%s", event.Key, event.Value)
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

	accountClient, err := initAccountRpcClient(cfg.Nacos)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Ctx:           context.Background(),
		Config:        cfg,
		DB:            db,
		Redis:         redisClient,
		AccountClient: accountClient,
	}, nil
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

//func initAccountRpcClient(cfg config.NacosConfig) (account_pb.AccountService, error) {
//	fmt.Println("正在连接Account RPC服务...")
//	fmt.Println("地址: tri://127.0.0.1:20880")
//
//	cli, err := client.NewClient(
//		client.WithClientURL("tri://127.0.0.1:20880"),
//	)
//	if err != nil {
//		fmt.Printf("创建客户端失败: %v\n", err)
//		return nil, err
//	}
//
//	srv, err := account_pb.NewAccountService(cli)
//	if err != nil {
//		fmt.Printf("创建服务代理失败: %v\n", err)
//		return nil, err
//	}
//
//	fmt.Println("Account RPC客户端初始化成功")
//	return srv, nil
//}

//func initDynamicConfig(cfg config.NacosConfig) {
//	rootConfig := dubbo_config.NewRootConfigBuilder().
//		SetApplication(dubbo_config.NewApplicationConfigBuilder().
//			SetName("account-consumer").
//			Build()).
//		AddRegistry("nacos", dubbo_config.NewRegistryConfigBuilder().
//			SetAddress(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)).
//			SetUsername("nacos").
//			SetPassword("lJPQwjjO9k").
//			AddParam("grpc-port", strconv.Itoa(cfg.GrpcPort)).
//			AddParam("metadata-type", "local").
//			Build()).
//		SetConfigCenter(dubbo_config.NewConfigCenterConfigBuilder().
//			SetProtocol("nacos").
//			SetAddress(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)).
//			SetUserName("nacos").
//			SetPassword("lJPQwjjO9k").
//			SetGroup("DEFAULT_GROUP").
//			SetDataID("account-service-config").
//			Build()).
//		AddReference("AccountService", dubbo_config.NewReferenceConfigBuilder().
//			SetProtocol("grpc").
//			SetInterface("com.example.account.service.AccountService").
//			Build()).
//		Build()
//
//	dubbo_config.SetRootConfig(rootConfig)
//}

func initAccountRpcClient(cfg config.NacosConfig) (account_pb.AccountService, error) {
	fmt.Printf("Account RPC客户端 通过注册Nacos host %s port %d grpc-port %d\n",
		cfg.Host, cfg.Port, cfg.GrpcPort)

	// 方法1：直接指定gRPC端口（推荐）
	cli, err := client.NewClient(
		client.WithClientRegistry(
			registry.WithNacos(),
			registry.WithUsername("nacos"),
			registry.WithPassword("lJPQwjjO9k"),
			registry.WithAddress(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
			registry.WithParams(map[string]string{
				"grpc-port":     strconv.Itoa(cfg.GrpcPort),
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

//func initAccountRpcClient(cfg config.NacosConfig) (account_pb.AccountService, error) {
//	// 1. 构建并加载完整的Dubbo-go根配置（核心）
//	rootConfig := dubbo_config.NewRootConfigBuilder().
//		// 1.1 设置注册中心（用于服务发现）
//		AddRegistry("nacos",
//			dubbo_config.NewRegistryConfigBuilder().
//				SetProtocol("nacos").
//				SetAddress(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)).
//				SetParams(map[string]string{
//					"username":      "nacos",
//					"password":      cfg.Password,
//					"grpc-port":     strconv.Itoa(cfg.GrpcPort),
//					"metadata-type": "local", // 避免metadata报告错误
//				}).
//				SetNamespace(cfg.Namespace).
//				SetGroup(cfg.Group).
//				Build(),
//		).
//		// 1.2 设置配置中心（指向同一个Nacos，用于管理动态配置）
//		SetConfigCenter(
//			dubbo_config.NewConfigCenterConfigBuilder().
//				SetProtocol("nacos").
//				SetAddress(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)).
//				SetDataID("dubbo-order-config"). // 你的order服务专用配置DataID
//				SetGroup("DUBBO_GROUP").
//				SetNamespace(cfg.Namespace).
//				Build(),
//		).
//		// 1.3 设置消费者引用 (关键修正：使用SetRegistryIDs)
//		SetConsumer(
//			dubbo_config.NewConsumerConfigBuilder().
//				AddReference("AccountService",
//					dubbo_config.NewReferenceConfigBuilder().
//						SetInterface("dubbo"). // 必须与account服务注册名完全一致
//						SetProtocol("tri").
//						SetRegistryIDs("nacos"). // 修正：关联到上面AddRegistry的ID "nacos"
//						SetVersion("1.0.0").
//						SetGroup(cfg.Group).
//						// 可选：设置超时、重试等参数
//						// SetTimeout("5s").
//						// SetRetries("3").
//						Build(),
//				).
//				Build(),
//		).
//		Build()
//
//	// 2. 初始化框架（这会启动配置中心、注册中心等所有组件）
//	if err := rootConfig.Init(); err != nil {
//		return nil, fmt.Errorf("Dubbo-go框架初始化失败: %w", err)
//	}
//
//	// 3. 现在可以安全地创建Dubbo客户端
//	// 注意：这里创建的是与全局配置关联的客户端
//	cli, err := client.NewClient()
//	if err != nil {
//		return nil, fmt.Errorf("创建Dubbo客户端失败: %w", err)
//	}
//
//	// 4. 使用客户端创建AccountService代理
//	accountClient, err := account_pb.NewAccountService(cli)
//	if err != nil {
//		return nil, fmt.Errorf("创建AccountService代理失败: %w", err)
//	}
//
//	logger.Info("Account RPC客户端初始化成功 (通过Nacos配置中心与服务发现)")
//	return accountClient, nil
//}

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
