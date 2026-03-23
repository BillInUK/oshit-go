package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"time"
)

// FeeTask 手续费统计任务
type FeeTask struct {
	*BaseTask
	db    *gorm.DB
	redis *redis.Client
	rpc   *rpc.Client
}

// FeeDetail 手续费详情
type FeeDetail struct {
	P50  uint64 `json:"p50"`
	P90  uint64 `json:"p90"`
	P95  uint64 `json:"p95"`
	P99  uint64 `json:"p99"`
	Mean uint64 `json:"mean"`
}

// PriorityFee 优先手续费
type PriorityFee struct {
	PerComputeUnit FeeDetail `json:"per_compute_unit"`
	PerTransaction FeeDetail `json:"per_transaction"`
}

// NewFeeTask 创建手续费任务
func NewFeeTask(db *gorm.DB, redis *redis.Client, rpc *rpc.Client) *FeeTask {
	base := NewBaseTask("fee_task", 30*time.Second)
	return &FeeTask{
		BaseTask: base,
		db:       db,
		redis:    redis,
		rpc:      rpc,
	}
}

// Start 启动任务
func (t *FeeTask) Start(ctx context.Context) error {
	go t.RunWithInterval(ctx, t.statistPriorityFeeOnBlock)
	return nil
}

// statistPriorityFeeOnBlock 统计链上手续费
func (t *FeeTask) statistPriorityFeeOnBlock(ctx context.Context) error {
	// 简化实现：从RPC获取当前手续费数据并更新Redis
	// 实际实现应该从链上获取最近区块的手续费数据

	// 模拟数据
	perComputeUnitDetail := FeeDetail{
		P50:  1000,
		P90:  2000,
		P95:  3000,
		P99:  5000,
		Mean: 1500,
	}

	perTransactionDetail := FeeDetail{
		P50:  5000,
		P90:  10000,
		P95:  15000,
		P99:  25000,
		Mean: 7500,
	}

	// 更新Redis缓存
	t.updateRedisCache(ctx, perComputeUnitDetail, perTransactionDetail)

	fmt.Printf("FeeTask updated Redis cache at %s\n", time.Now().Format(time.RFC3339))
	return nil
}

// getRecentSlots 获取最近的slot列表
func (t *FeeTask) getRecentSlots(currentSlot uint64, count int) []uint64 {
	slots := make([]uint64, 0, count)
	for i := 0; i < count; i++ {
		if currentSlot > uint64(i) {
			slots = append(slots, currentSlot-uint64(i))
		}
	}
	return slots
}

// getFeeDataForSlot 获取指定slot的手续费数据
func (t *FeeTask) getFeeDataForSlot(ctx context.Context, slot uint64) (*struct {
	PerComputeUnit uint64
	PerTransaction uint64
}, error) {
	// TODO: 实现从RPC获取手续费数据
	// 这里简化实现，返回模拟数据
	return &struct {
		PerComputeUnit uint64
		PerTransaction uint64
	}{
		PerComputeUnit: 1000,
		PerTransaction: 5000,
	}, nil
}

// calculateFeeDetail 计算手续费详情
func (t *FeeTask) calculateFeeDetail(fees []uint64) FeeDetail {
	if len(fees) == 0 {
		return FeeDetail{}
	}

	// 简化实现：返回固定值
	// 实际实现应该计算百分位数和平均值
	return FeeDetail{
		P50:  1000,
		P90:  2000,
		P95:  3000,
		P99:  5000,
		Mean: 1500,
	}
}

// updateRedisCache 更新Redis缓存
func (t *FeeTask) updateRedisCache(ctx context.Context, perComputeUnit, perTransaction FeeDetail) {
	// 更新每计算单元手续费
	perComputeUnitJSON, _ := json.Marshal(perComputeUnit)
	t.redis.Set(ctx, "SOL-PRIORITY-FEE-PER-COMPUTE-UNIT-ON-BLOCKCHAIN", perComputeUnitJSON, 0)

	// 更新每交易手续费
	perTransactionJSON, _ := json.Marshal(perTransaction)
	t.redis.Set(ctx, "SOL-PRIORITY-FEE-PER-TRANSACTION-ON-BLOCKCHAIN", perTransactionJSON, 0)
}
