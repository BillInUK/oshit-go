package logic

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	account_pb "oshit-go/app/account/rpc/pb"
	"oshit-go/app/exchange/api/internal/svc"
	"oshit-go/app/exchange/api/types"
)

type OrderLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
	db     *gorm.DB
	rd     *redis.Client
}

func NewOrderLogic(ctx context.Context, srvCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
		db:     srvCtx.DB,
		rd:     srvCtx.Redis,
	}
}

func (l *OrderLogic) FetchUserOrders(req *types.UserOrdersReq) (*types.UserOrdersResp, error) {
	cli := l.srvCtx.AccountCli
	user, err := cli.GetUser(l.ctx, &account_pb.GetUserRequest{
		UserId: int64(1),
	})
	if err != nil {
		return nil, err
	}
	return &types.UserOrdersResp{
		UserID: user.Id,
		Orders: []types.OrderInfo{},
	}, nil
}
