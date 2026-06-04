package context

import (
	"context"
	"crypto/rsa"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oshit-go/app/base/api/internal/config"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"sync"
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
	AppGlobalPublicKey *rsa.PublicKey
	RpcPool            *utils.RPCPool // 跟随环境的多 RPC 轮询
	RpcClient          *rpc.Client    // RpcPool.First() 的快捷引用
	MainnetRpcClient   *rpc.Client    // 固定主网 RPC
	MainnetRpcURL      string         // 固定主网 RPC URL（给需要 URL 的调用方用）
	HeliusAPIKey       string         // Helius API Key（给 HeliusParseMarketBuyTx 用）
	LightHouseAddress  solana.PublicKey
	TokenDecimal       float64
	KafkaProducer      interface{} // kafka.Producer类型，在kafka.go中定义
	NacosConfigClient  config_client.IConfigClient

	// 配置表数据
	ConfigMu     sync.RWMutex
	SystemConfig *model.SystemConfig
	ChainConfig  *model.ChainConfig
	TokenConfig  *model.TokenConfig
	FeeTolerance model.FeeTolerance
	AwsConfig    *model.AwsConfig

	// 服务签名私钥: ServiceKeyMap[service][subService]
	ServiceKeyMap ServiceKey

	// 服务配置: ServiceInfoMap[service][subService] = ServiceInfo
	ServiceInfoMap map[string]map[string]model.ServiceInfo
}
