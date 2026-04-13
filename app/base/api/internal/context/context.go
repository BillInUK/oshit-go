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
	"oshit-go/common/pkg/dal/model"
)

// ServiceKey 二维索引: ServiceKeyMap[service][subService] = privateKey
type ServiceKey = map[string]map[string]solana.PrivateKey

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
	MainnetRPCConfig    *model.MainnetRpcConfig
	TokenConfig         *model.TokenConfig
	FeeTolerance        model.FeeTolerance
	AwsConfig           *model.AwsConfig

	// 服务签名私钥: ServiceKeyMap[service][subService]
	ServiceKeyMap ServiceKey

	// 服务配置: ServiceInfoMap[service][subService] = ServiceInfo
	ServiceInfoMap map[string]map[string]model.ServiceInfo
}
