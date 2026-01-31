package config

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// -------------------------- 1. 定义业务配置结构体 --------------------------
// 存储从Nacos读取的配置（按需扩展字段）
type OrderServiceConfig struct {
	AccountRpcTimeout string `json:"accountRpcTimeout"` // 调用account rpc的超时时间
	AppName           string `json:"appName"`           // 应用名称
	LogLevel          string `json:"logLevel"`          // 日志级别
}

// -------------------------- 2. 全局配置（加读写锁保证并发安全） --------------------------
var (
	GlobalOrderConfig OrderServiceConfig          // 全局配置实例
	configRWMutex     sync.RWMutex                // 读写锁（防止配置读写冲突）
	nacosConfigClient config_client.IConfigClient // Nacos配置客户端
)

// -------------------------- 3. 初始化Nacos配置客户端 --------------------------
// InitNacosConfig 初始化Nacos配置中心（启动时调用）
func InitNacosConfig() error {
	// ① Nacos服务端配置（集群模式可添加多个ServerConfig）
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "127.0.0.1", // Nacos服务端IP（替换为你的实际地址）
			Port:   8848,        // Nacos服务端端口
		},
	}

	// ② Nacos客户端配置（匹配你的Nacos服务端配置）
	clientConfig := constant.ClientConfig{
		NamespaceId:         "public",        // 命名空间ID（public填""或"public"）
		TimeoutMs:           5000,            // 超时时间
		NotLoadCacheAtStart: true,            // 启动时不加载本地缓存
		LogDir:              "./nacos/log",   // 日志目录
		CacheDir:            "./nacos/cache", // 缓存目录
		LogLevel:            "info",          // 日志级别
		// 若Nacos开启认证，添加用户名密码
		Username: "nacos",
		Password: "lJPQwjjO9k",
	}

	// ③ 创建Nacos配置客户端
	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return fmt.Errorf("创建Nacos配置客户端失败: %v", err)
	}
	nacosConfigClient = client

	// ④ 加载初始配置
	if err := loadNacosConfig(); err != nil {
		return fmt.Errorf("加载初始配置失败: %v", err)
	}

	// ⑤ 监听配置变更（热更新）
	if err := listenConfigChange(); err != nil {
		return fmt.Errorf("监听配置变更失败: %v", err)
	}

	fmt.Println("Nacos配置中心初始化成功，初始配置：", GetGlobalOrderConfig())
	return nil
}

// -------------------------- 4. 读取Nacos配置并更新全局实例 --------------------------
func loadNacosConfig() error {
	// 从Nacos读取配置（DataId/Group需和Nacos控制台一致）
	content, err := nacosConfigClient.GetConfig(vo.ConfigParam{
		DataId: "order-service-config", // 配置ID（自定义）
		Group:  "DEFAULT_GROUP",        // 配置分组（默认）
	})
	if err != nil {
		return err
	}

	// 解析JSON配置到全局结构体（加写锁）
	configRWMutex.Lock()
	defer configRWMutex.Unlock()
	if err := json.Unmarshal([]byte(content), &GlobalOrderConfig); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}
	return nil
}

// -------------------------- 5. 监听配置变更（热更新） --------------------------
func listenConfigChange() error {
	return nacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: "order-service-config",
		Group:  "DEFAULT_GROUP",
		// 配置变更时的回调函数
		OnChange: func(namespace, group, dataId, content string) {
			fmt.Printf("Nacos配置已变更，新配置：%s\n", content)
			// 重新解析配置并更新全局实例
			configRWMutex.Lock()
			defer configRWMutex.Unlock()
			if err := json.Unmarshal([]byte(content), &GlobalOrderConfig); err != nil {
				fmt.Printf("解析变更后的配置失败: %v\n", err)
				return
			}
			fmt.Println("全局配置已更新：", GlobalOrderConfig)
		},
	})
}

// -------------------------- 6. 安全获取全局配置（加读锁） --------------------------
func GetGlobalOrderConfig() OrderServiceConfig {
	configRWMutex.RLock()
	defer configRWMutex.RUnlock()
	return GlobalOrderConfig
}
