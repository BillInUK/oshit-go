package rpc

import (
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/registry/nacos" // 必须匿名导入 Nacos 驱动
	"fmt"
	"os"
)

// ！！！替换为你的 Account RPC 服务接口定义（示例，按需修改）
// 示例：import "your-project-path/apis/account"
// var AccountRPCService = new(account.AccountService) // 声明RPC服务客户端实例

// ！！！配置中心必选：替换为你Nacos控制台的实际DataId（与业务配置的DataId一致即可）
const NacosConfigDataId = "order-service-config"

// InitDubboClient 初始化dubbo-go消费端（彻底解决所有空指针+DataId为空+消除Config center日志）
// 调用时机：main函数中 Nacos 业务配置初始化后、Gofiber 启动前
func InitDubboClient() error {
	// 指定自定义配置文件路径（与你的dubbo.yaml一致，修正文件名）
	os.Setenv("CONF_DUBBO_GO_CONFIG_PATH", "./etc/dubbo.yaml")

	// 1. 构建 Nacos 注册中心配置（与dubbo.yaml中registry参数完全一致）
	registryConfig := &config.RegistryConfig{
		Protocol:  "nacos",
		Address:   "127.0.0.1:8848",
		Username:  "nacos",
		Password:  "lJPQwjjO9k",
		Namespace: "public",
		Group:     "DEFAULT_GROUP",
	}

	// 2. 构建 Nacos 配置中心配置（核心修正：添加必选的DataId，消除dataId为空警告）
	configCenter := config.NewConfigCenterConfigBuilder().
		SetProtocol("nacos").
		SetAddress("127.0.0.1:8848").
		SetUserName("nacos").
		SetPassword("lJPQwjjO9k").
		SetNamespace("public").
		SetGroup("DEFAULT_GROUP").
		SetDataID(NacosConfigDataId). // 关键必选：设置Nacos配置的DataId，与控制台一致
		Build()

	// 3. 构建消费端配置（官方Builder模式，与dubbo.yaml一致）
	consumerConfig := config.NewConsumerConfigBuilder().
		SetRequestTimeout("3s").
		SetCheck(false).
		SetRegistryIDs("nacos-registry").
		Build()

	// 4. 构建Logger配置（官方Builder，彻底解决Logger空指针）
	loggerConfig := config.NewLoggerConfigBuilder().
		SetDriver("zap").               // 日志驱动，默认zap
		SetLevel("info").               // 日志级别，默认info
		SetFormat("json").              // 日志格式，默认text
		SetAppender("console,file").    // 输出方式，同时控制台+文件，默认console
		SetFileName("order-dubbo.log"). // 日志文件名，默认dubbo.log
		SetFileMaxSize(200).            // 单个日志文件最大大小，默认100Mb
		SetFileCompress(true).          // 日志文件压缩，默认true
		Build()

	// 5. 构建顶层RootConfig（值类型，核心修正：初始化Logger+CustomConfig双字段，解决所有空指针）
	otelEnable := false
	metricEnable := false
	rootConfig := config.RootConfig{
		Registries: map[string]*config.RegistryConfig{
			"nacos-registry": registryConfig,
		},
		ConfigCenter: configCenter,   // 注入配置中心，消除刷屏日志
		Consumer:     consumerConfig, // 注入消费端配置
		Application: &config.ApplicationConfig{
			Name: "order-service", // 与dubbo.yaml、Nacos完全一致
		},
		Logger: loggerConfig,           // 已初始化，非nil
		Custom: &config.CustomConfig{}, // 核心新增：初始化CustomConfig，解决本次panic
		Otel: &config.OtelConfig{
			TraceConfig: &config.OtelTraceConfig{
				Enable: &otelEnable,
			},
		}, // 核心新增：初始化Otel空结构体，关闭opentelemetry，解决otel config is nil
		Metrics:  &config.MetricsConfig{Enable: &metricEnable},
		Shutdown: &config.ShutdownConfig{},
		Provider: &config.ProviderConfig{ // 核心新增：初始化Provider，适配纯消费端
			Services: make(map[string]*config.ServiceConfig), // 关键：赋空map，避免遍历nil触发panic
		},
	}

	// 6. 设置全局根配置+加载（严格按你的SetRootConfig值类型传参，无错误）
	config.SetRootConfig(rootConfig)
	if err := config.Load(); err != nil {
		return fmt.Errorf("dubbo-go 配置加载失败: %v", err)
	}

	// ！！！可选：引用你的Account RPC服务（按需打开，无语法错误）
	// if err := config.GetConsumer().Refer(AccountRPCService); err != nil {
	// 	return fmt.Errorf("Account RPC 服务引用失败: %v", err)
	// }

	fmt.Println("✅ dubbo-go消费端初始化成功（Nacos注册+配置中心集成完成，所有日志/空指针问题解决）")
	return nil
}

// ！！！可选：Account RPC服务引用配置（按需使用）
// func buildAccountReferenceConfig() *config.ReferenceConfig {
// 	return config.NewReferenceConfigBuilder()
// 		.SetInterface("com.xxx.AccountService")
// 		.SetProtocol("dubbo")
// 		.SetRequestTimeout("3s")
// 		.SetCheck(false)
// 		.Build()
// }
