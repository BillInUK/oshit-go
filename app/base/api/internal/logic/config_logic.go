package logic

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/app/base/dal/model"
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

// GetTokenInfo 获取token信息
func (l *ConfigLogic) GetTokenInfo(req *types.GetTokenInfoReq) (*types.GetTokenInfoRsp, error) {
	var tokenConfig model.TokenConfig
	if err := l.db.Model(&tokenConfig).
		Where("name = ? AND symbol = ?", req.Brand, req.Symbol).
		First(&tokenConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token config not found")
		}
		return nil, errors.New("query token info error")
	}

	return &types.GetTokenInfoRsp{
		Name:      tokenConfig.Name,
		Symbol:    tokenConfig.Symbol,
		Decimal:   tokenConfig.Decimal,
		Mint:      tokenConfig.Mint,
		CreatedAt: tokenConfig.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
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
