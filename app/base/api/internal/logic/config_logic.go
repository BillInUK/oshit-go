package logic

import (
	"context"
	"gorm.io/gorm"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
)

type ConfigLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
	db     *gorm.DB
}

func NewConfigLogic(ctx context.Context, srvCtx *svc.ServiceContext) *ConfigLogic {
	return &ConfigLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
		db:     srvCtx.DB,
	}
}

// GetFeeTolerance 获取手续费容错
func (l *ConfigLogic) GetFeeTolerance() (*types.GetFeeToleranceRsp, error) {
	// 简化实现：返回固定值
	// 实际应该从数据库查询
	return &types.GetFeeToleranceRsp{
		MinFee: 1000,
		MaxFee: 10000,
	}, nil
}
