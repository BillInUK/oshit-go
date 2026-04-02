package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// AcquireDistributedRateLimit 分布式限流函数
func AcquireDistributedRateLimit(rdb redis.UniversalClient, key string, limit int) bool {
	now := time.Now().Unix() // 当前时间戳（秒级）

	// 计算 Redis key（每秒一个）
	rateKey := fmt.Sprintf("%s:%d", key, now)

	// 使用 Redis INCR 递增请求数
	count, err := rdb.Incr(context.Background(), rateKey).Result()
	if err != nil {
		fmt.Println("Redis 连接失败:", err)
		return false
	}

	// 设置 1 秒后过期，防止 key 无限制增长
	if count == 1 {
		rdb.Expire(context.Background(), rateKey, time.Second)
	}

	// 如果请求数超过限制，则返回 false
	return count <= int64(limit)
}

// IsRpcRateLimitedError 判断是否为Rpc请求限制错误
func IsRpcRateLimitedError(err error) bool {
	if err == nil {
		return false
	}

	// 如果错误包含 "request limit reached" 字符串，说明是请求限制错误
	if strings.Contains(err.Error(), "request limit reached") {
		return true
	}

	return false
}
