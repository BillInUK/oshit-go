package config

import (
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Dubbo    DubboConfig    `mapstructure:"dubbo"`
	Nacos    NacosConfig    `mapstructure:"nacos"`
}

type DubboConfig struct {
	Protocol ProtocolConfig `mapstructure:"protocol"`
	Nacos    NacosConfig    `mapstructure:"nacos"`
}

type ProtocolConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type NacosConfig struct {
	Host             string                `mapstructure:"host"`
	Port             int                   `mapstructure:"port"`
	GrpcPort         int                   `mapstructure:"grpc-port"`
	Namespace        string                `mapstructure:"namespace"`
	Username         string                `mapstructure:"username"`
	Password         string                `mapstructure:"password"`
	ServerConfig     []NacosServerConfig   `mapstructure:"server_config"`
	ClientConfig     NacosClientConfig     `mapstructure:"client_config"`
	SubscribeConfigs NacosSubscribeConfigs `mapstructure:"subscribe_configs"`
}

// NacosServerConfig Nacos服务端配置结构体（匹配yaml中的server_config）
type NacosServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint64 `mapstructure:"port"`
	GrpcPort uint64 `mapstructure:"grpc_port"`
}

// NacosClientConfig Nacos客户端配置结构体（匹配yaml中的client_config）
type NacosClientConfig struct {
	NamespaceId         string `mapstructure:"namespace_id"`
	TimeoutMs           uint64 `mapstructure:"timeout_ms"`
	NotLoadCacheAtStart bool   `mapstructure:"not_load_cache_at_start"`
	LogDir              string `mapstructure:"log_dir"`
	CacheDir            string `mapstructure:"cache_dir"`
	LogLevel            string `mapstructure:"log_level"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
}

type NacosSubscribeConfig struct {
	DataId string `mapstructure:"data_id"`
	Group  string `mapstructure:"group"`
}

type NacosSubscribeConfigs struct {
	BaseRuntime     NacosSubscribeConfig `mapstructure:"base_runtime"`
	ServiceRegistry NacosSubscribeConfig `mapstructure:"service_registry"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	MasterName string   `mapstructure:"master-name"`
	Hosts      []string `mapstructure:"hosts"`
	Password   string   `mapstructure:"password"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

func LoadConfig() (*Config, error) {
	path := os.Getenv("APP_CONF")
	if path == "" {
		path = "./etc/application.yaml"
	}
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
