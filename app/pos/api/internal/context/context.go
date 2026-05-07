package context

import (
	"context"
	"sync"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"oshit-go/app/pos/api/internal/config"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/common/pkg/dal/model"
)

type CoreContext struct {
	Config  *config.Config        // 从配置文件读取的基础配置
	DB      *gorm.DB              // postgres 数据库句柄
	Redis   redis.UniversalClient // redis 句柄
	RedSync redsync.Redsync       // redlock 句柄
	Ctx     context.Context       // 上下文句柄

	NacosConfigClient config_client.IConfigClient // Nacos 配置中心客户端
	ConfigMu          sync.RWMutex               // 配置读写锁

	// 全局变量
	RpcClient             *rpc.Client        // solana rpc 客户端
	LightHouseAddress     solana.PublicKey    // light house 指令
	TokenDecimal          float64            // token 精度基数
	KafkaProducer         interface{}        // *kafka.Writer，在kafka.go中定义
	KafkaConsumer         interface{}        // *kafka.Reader，消费 ServiceTransaction（base 模块）
	SnapShotKafkaConsumer interface{}        // *kafka.Reader，消费 PosTopic + StakeTopic
	BaseClient            *posrpc.BaseClient // base 模块dubbo 客户端

	// 配置表数据
	SystemConfig *model.SystemConfig // 系统配置
	ChainConfig  *model.ChainConfig  // 链配置
	TokenConfig  *model.TokenConfig  // token配置
	FeeTolerance *model.FeeTolerance // 手续费容错配置
}
