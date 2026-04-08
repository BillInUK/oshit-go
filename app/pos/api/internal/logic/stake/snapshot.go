package stake

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
)

type StakeSnapShotLogic struct {
	prefix            string
	decimals          uint8
	ctx               context.Context
	srvCtx            *svc.ServiceContext
	db                *gorm.DB
	rd                redis.UniversalClient
	rs                redsync.Redsync
	baseClient        *posrpc.BaseClient
	rpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
}

func NewStakeSnapShotLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeSnapShotLogic {
	return &StakeSnapShotLogic{
		prefix:            "TakeToken业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
	}
}

// GetConfig 获取奖励规则配置
func (l *StakeSnapShotLogic) GetConfig() (interface{}, error) {
	return nil, nil
}
