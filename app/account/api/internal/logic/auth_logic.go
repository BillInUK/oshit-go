package logic

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"oshit-go/app/account/api/internal/svc"
	"oshit-go/app/account/api/internal/types"
)

type AuthLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
	db     *gorm.DB
	rd     *redis.Client
}

func NewAuthLogic(ctx context.Context, srvCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
		db:     srvCtx.DB,
		rd:     srvCtx.Redis,
	}
}

func (l *AuthLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	// 业务逻辑实现
	// 1. 验证用户名密码
	// 2. 生成token
	// 3. 返回结果

	return &types.LoginResponse{
		UserID:  1,
		Token:   "generated_token",
		Message: "Login successful",
	}, nil
}
