package logic

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/dal/query"
)

type InfoLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
	db     *gorm.DB
}

func NewInfoLogic(ctx context.Context, srvCtx *svc.ServiceContext) *InfoLogic {
	return &InfoLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
		db:     srvCtx.DB,
	}
}

// GetTokenInfo 获取token信息
func (l *InfoLogic) GetTokenInfo() (*types.GetTokenInfoRsp, error) {
	var tokenConfig model.TokenConfig
	if err := l.db.Model(&tokenConfig).
		First(&tokenConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token config not found")
		}
		return nil, errors.New("query token info error")
	}

	return &types.GetTokenInfoRsp{
		Name:     tokenConfig.Name,
		Symbol:   tokenConfig.Symbol,
		Decimals: tokenConfig.Decimals,
		Mint:     tokenConfig.Mint,
	}, nil
}

func (l *InfoLogic) GetTokenHolders() (*types.TokenHoldersRsp, error) {
	q := query.Use(l.srvCtx.DB)
	naInfo := q.NativeAccountInfo

	count, err := naInfo.WithContext(l.ctx).Count()
	if err != nil {
		return nil, fmt.Errorf("get token holders error: %v", err)
	}

	return &types.TokenHoldersRsp{
		HoldersNumber: int(count),
	}, nil
}
