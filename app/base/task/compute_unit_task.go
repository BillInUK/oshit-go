package task

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

// ComputeUnitTask 计算单元消耗任务
type ComputeUnitTask struct {
	*BaseTask
	redis *redis.Client
}

// NewComputeUnitTask 创建计算单元任务
func NewComputeUnitTask(redis *redis.Client) *ComputeUnitTask {
	base := NewBaseTask("compute_unit_task", 1*time.Hour)
	return &ComputeUnitTask{
		BaseTask: base,
		redis:    redis,
	}
}

// Start 启动任务
func (t *ComputeUnitTask) Start(ctx context.Context) error {
	go t.RunWithInterval(ctx, t.estimateComputeUnitConsumed)
	return nil
}

// estimateComputeUnitConsumed 估计计算单元消耗
func (t *ComputeUnitTask) estimateComputeUnitConsumed(ctx context.Context) error {
	// 简化实现：设置固定的计算单元消耗值
	// 实际实现应该通过模拟交易来估计消耗

	miniRent := "890880"        // 最小账户租金
	associatedAccount := "6472" // 创建关联token账户消耗
	memo := "300"               // memo指令消耗

	// 更新Redis缓存
	t.redis.Set(ctx, "COMPUTE-UNIT-ASSOCIATED-ACCOUNT-MINI-RENT", miniRent, 2*time.Hour)
	t.redis.Set(ctx, "COMPUTE-UNIT-ASSOCIATED-ACCOUNT", associatedAccount, 2*time.Hour)
	t.redis.Set(ctx, "COMPUTE-UNIT-MEMO", memo, 2*time.Hour)

	fmt.Printf("ComputeUnitTask updated Redis cache at %s\n", time.Now().Format(time.RFC3339))
	return nil
}

// GetComputeUnitConsumed 获取计算单元消耗（供handler调用）
func (t *ComputeUnitTask) GetComputeUnitConsumed(ctx context.Context) (map[string]uint64, error) {
	result := make(map[string]uint64)

	// 从Redis获取数据
	keys := []string{
		"COMPUTE-UNIT-ASSOCIATED-ACCOUNT-MINI-RENT",
		"COMPUTE-UNIT-ASSOCIATED-ACCOUNT",
		"COMPUTE-UNIT-MEMO",
	}

	for _, key := range keys {
		val, err := t.redis.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("get %s error: %v", key, err)
		}

		uintVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse %s error: %v", key, err)
		}

		// 转换key为小写蛇形命名
		simpleKey := ""
		switch key {
		case "COMPUTE-UNIT-ASSOCIATED-ACCOUNT-MINI-RENT":
			simpleKey = "mini_rent"
		case "COMPUTE-UNIT-ASSOCIATED-ACCOUNT":
			simpleKey = "associated_account"
		case "COMPUTE-UNIT-MEMO":
			simpleKey = "memo"
		}

		result[simpleKey] = uintVal
	}

	// 添加固定值
	result["transfer_checked"] = 6472

	return result, nil
}
