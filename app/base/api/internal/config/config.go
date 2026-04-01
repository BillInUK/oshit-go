package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Dubbo    DubboConfig    `mapstructure:"dubbo"`
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
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	GrpcPort  int    `mapstructure:"grpc-port"`
	Namespace string `mapstructure:"namespace"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
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
	viper.SetConfigFile("./etc/base.yaml")
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
