package context

import (
	"context"
	"crypto/rsa"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oshit-go/app/base/api/internal/config"
	"oshit-go/app/base/dal/model"
)

type CoreContext struct {
	Config  *config.Config
	DB      *gorm.DB
	Redis   redis.UniversalClient
	RedSync redsync.Redsync
	Ctx     context.Context

	// 全局变量
	AppGlobalPublicKey  *rsa.PublicKey
	RpcClient           *rpc.Client
	UserWalletRpcClient *rpc.Client
	LightHouseAddress   solana.PublicKey
	TokenDecimal        float64
	KafkaProducer       interface{} // kafka.Producer类型，在kafka.go中定义

	// 配置表数据
	SystemConfig        *model.SystemConfig
	ChainConfig         *model.ChainConfig
	UserWalletRPCConfig *model.UserWalletRpcConfig
	TokenConfig         *model.TokenConfig
	AwsConfig           *model.AwsConfig
}
