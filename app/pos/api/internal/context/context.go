package context

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oshit-go/app/pos/api/internal/config"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/common/pkg/dal/model"
)

type CoreContext struct {
	Config  *config.Config
	DB      *gorm.DB
	Redis   redis.UniversalClient
	RedSync redsync.Redsync
	Ctx     context.Context

	// 全局变量
	RpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
	TokenDecimal      float64
	KafkaProducer     interface{} // *kafka.Writer，在kafka.go中定义
	KafkaConsumer     interface{} // *kafka.Reader，在kafka.go中定义
	BaseClient        *posrpc.BaseClient

	// 配置表数据
	SystemConfig *model.SystemConfig
	ChainConfig  *model.ChainConfig
	TokenConfig  *model.TokenConfig
	FeeTolerance *model.FeeTolerance
}
