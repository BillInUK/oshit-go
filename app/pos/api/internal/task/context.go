package task

import (
	"context"
	"oshit-go/common/pkg/dal/model"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	core_context "oshit-go/app/pos/api/internal/context"
	"oshit-go/common/pkg/entity"
)

// SnapShotHandler 处理快照消息
type SnapShotHandler func(ctx context.Context, tx entity.KafkaNewSnapShotMsg) error

// ScannedTxHandler 处理链上已确认交易的回调函数类型
type ScannedTxHandler func(ctx context.Context, tx entity.NewScannedTx) error

// ExpiredTxHandler 处理已超时交易的回调函数类型
type ExpiredTxHandler func(ctx context.Context, tx entity.NewExpiredTx) error

type TaskContext struct {
	core_context.CoreContext
	SnapShotHandlers map[string]SnapShotHandler
	// ScannedHandlers 按 SubService 注册的已确认交易处理器，key 为 SubService 名称
	ScannedHandlers map[string]ScannedTxHandler
	// ExpiredHandlers 按 SubService 注册的超时交易处理器，key 为 SubService 名称
	ExpiredHandlers       map[string]ExpiredTxHandler
	RewardConfig          *model.StakeRewardConfig
	SnapShotKafkaConsumer interface{} // *kafka.Reader，消费 PosTopic + StakeTopic
}

// runPeriodic 启动时立即执行一次，之后按固定间隔周期执行。
// 每次执行前尝试获取分布式锁，抢不到则跳过本轮（当前已有其他实例在执行）。
//
// 适用于执行时间可预期、远小于 lockTTL 的常规周期任务。
// lockKey: Redis 锁 key，格式: {service}:{chain}:{module}:{name}:lock
func runPeriodic(rs *redsync.Redsync, interval time.Duration, lockKey string, lockTTL time.Duration, fn func()) {
	execute := func() {
		mutex := rs.NewMutex(lockKey, redsync.WithTries(1), redsync.WithExpiry(lockTTL))
		if err := mutex.Lock(); err != nil {
			return
		}
		fn()
		mutex.Unlock()
	}
	execute()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		execute()
	}
}

// runPeriodicWithWatchdog 启动时立即执行一次，之后按固定间隔周期执行。
// 每次执行使用 watchdog 自动续期锁。
//
// 适用于执行时间不可预测的长任务（如需要启动浏览器、等待外部接口等）。
// lockKey: Redis 锁 key，格式: {service}:{chain}:{module}:{name}:lock
func runPeriodicWithWatchdog(rs *redsync.Redsync, interval time.Duration, lockKey string, initTTL time.Duration, fn func()) {
	runWithWatchdog(rs, lockKey, initTTL, fn)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		runWithWatchdog(rs, lockKey, initTTL, fn)
	}
}

// runWithWatchdog 获取分布式锁并启动 watchdog goroutine 自动续期
//
// 适用于执行时间不可预测的长任务（如需要启动浏览器、等待外部接口等），
// 防止锁在任务执行期间过期导致其他实例重复执行。
//
// lockKey: Redis 锁 key，格式: {service}:{chain}:{module}:{name}:lock
// initTTL: 锁初始有效期，watchdog 会在到期前自动续期
// fn:      需要在锁保护下执行的任务逻辑
func runWithWatchdog(rs *redsync.Redsync, lockKey string, initTTL time.Duration, fn func()) {
	mutex := rs.NewMutex(lockKey,
		redsync.WithTries(1),
		redsync.WithExpiry(initTTL),
	)
	if err := mutex.Lock(); err != nil {
		return // 获取不到锁，静默跳过（当前已有其他实例在执行）
	}
	defer mutex.Unlock()

	// Watchdog：每 initTTL/3 续期一次，确保任务执行期间锁不会提前过期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		ticker := time.NewTicker(initTTL / 3)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := mutex.Extend(); err != nil {
					log.Errorf("watchdog 续期锁失败 [%s]: %v", lockKey, err)
				}
			}
		}
	}()

	fn()
}
