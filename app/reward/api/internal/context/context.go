package context

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oshit-go/app/reward/api/internal/config"
	rewardrpc "oshit-go/app/reward/api/internal/rpc"
	"oshit-go/common/pkg/dal/model"
	"sync"
)

type CoreContext struct {
	Config  *config.Config
	DB      *gorm.DB
	Redis   redis.UniversalClient
	RedSync redsync.Redsync
	Ctx     context.Context

	NacosConfigClient config_client.IConfigClient
	ConfigMu          sync.RWMutex

	// 全局变量
	RpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
	TokenDecimal      float64
	KafkaProducer     interface{} // *kafka.Writer，在kafka.go中定义
	KafkaConsumer     interface{} // *kafka.Reader，在kafka.go中定义
	BaseClient        *rewardrpc.BaseClient

	// 配置表数据
	SystemConfig *model.SystemConfig
	ChainConfig  *model.ChainConfig
	TokenConfig  *model.TokenConfig
	FeeTolerance *model.FeeTolerance
}
