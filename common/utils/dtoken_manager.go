package utils

import (
	"github.com/Zany2/dtoken-go/core/builder"
	"github.com/Zany2/dtoken-go/dtoken"
	"time"
)

// InitDTokenManager 初始化 dtoken 管理器（JWT 模式 + 内存存储）。
// 三个服务使用相同的 JWT_SECRET_KEY，token 由 base 签发，pos/reward 直接解析验证，无需共享存储。
func InitDTokenManager() {
	mgr := builder.NewBuilder().
		JwtSecret(JWT_SECRET_KEY).
		TimeoutDuration(JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT * time.Minute).
		IsPrintBanner(false).
		Build()
	dtoken.SetManager(mgr)
}
