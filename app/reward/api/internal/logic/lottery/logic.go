package lottery

import (
	"context"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	rewardrpc "oshit-go/app/reward/api/internal/rpc"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/common/pkg/dal/model"
)

type LotteryLogic struct {
	prefix          string
	ctx             context.Context
	srvCtx          *svc.ServiceContext
	db              *gorm.DB
	rd              redis.UniversalClient
	rs              redsync.Redsync
	rpcClient       *rpc.Client
	baseClient      *rewardrpc.BaseClient
	takeTokenConfig *model.TakeTokenConfig
}

func NewLotteryLogic(ctx context.Context, srvCtx *svc.ServiceContext) *LotteryLogic {
	return &LotteryLogic{
		prefix:          "Lottery",
		ctx:             ctx,
		srvCtx:          srvCtx,
		db:              srvCtx.DB,
		rd:              srvCtx.Redis,
		rs:              srvCtx.RedSync,
		rpcClient:       srvCtx.RpcClient,
		baseClient:      srvCtx.BaseClient,
		takeTokenConfig: srvCtx.TakeTokenConfig,
	}
}
