package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Dubbo    DubboConfig    `mapstructure:"dubbo"`
	Nacos    NacosConfig    `mapstructure:"nacos"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
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
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type DubboConfig struct {
	Protocol ProtocolConfig `mapstructure:"protocol"`
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
	Group     string `mapstructure:"group"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("./etc/base.yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
