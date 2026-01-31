package svc

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"oshit-go/app/account/rpc/internal/config"
	"strconv"
)

type ServiceContext struct {
	Ctx    context.Context
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

func NewServiceContext() (*ServiceContext, error) {
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

	log.Printf("Service context initialized for %s", cfg.App.Name)
	return &ServiceContext{
		Ctx:    context.Background(),
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
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
