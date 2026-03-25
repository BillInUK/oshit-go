package logic

import (
	"context"
	"gorm.io/gorm"
	base_pb "oshit-go/app/base/rpc/pb"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
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
	user, err := cli.GetUser(l.ctx, &base_pb.GetUserRequest{
		UserId: req.UserID,
	})
	if err != nil {
		return nil, err
	}
	return &types.UserOrdersResp{
		UserID: user.Id,
		Orders: []types.OrderInfo{},
	}, nil
}
