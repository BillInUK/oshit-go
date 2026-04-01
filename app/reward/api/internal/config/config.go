package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Nacos    NacosConfig    `mapstructure:"nacos"`
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
	Brokers  []string            `mapstructure:"brokers"`
	Consumer KafkaConsumerConfig `mapstructure:"consumer"`
}

type KafkaConsumerConfig struct {
	GroupID string   `mapstructure:"group_id"`
	Topics  []string `mapstructure:"topics"`
}

// NacosServerConfig Nacos服务端配置结构体（匹配yaml中的server_config）
type NacosServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint64 `mapstructure:"port"`      // 注意：Nacos SDK的Port是uint64类型
	GrpcPort uint64 `mapstructure:"grpc_port"` // 注意：Nacos SDK的Port是uint64类型
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

// NacosSubscribeConfig Nacos订阅配置结构体（匹配yaml中的subscribe_config）
type NacosSubscribeConfig struct {
	DataId string `mapstructure:"data_id"`
	Group  string `mapstructure:"group"`
}

type NacosConfig struct {
	ServerConfig    []NacosServerConfig  `mapstructure:"server_config"`
	ClientConfig    NacosClientConfig    `mapstructure:"client_config"`
	SubscribeConfig NacosSubscribeConfig `mapstructure:"subscribe_config"`
}

// NacosOrderCfg 原有业务配置结构体
type NacosOrderCfg struct {
	AccountRpcTimeout string `json:"accountRpcTimeout"`
	AppName           string `json:"appName"`
	LogLevel          string `json:"logLevel"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("./etc/reward.yaml")
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
