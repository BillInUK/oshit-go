package task

import (
	"time"

	"github.com/go-redsync/redsync/v4"
	log "github.com/sirupsen/logrus"
)

const ttlCleanupBatchSize = 5000

// TTLCleanupTask 处理无法用分区表 DROP 的 TTL 表（需条件保留的表）
// 目前包含: t_stake_reward, t_stake_reward_claim
type TTLCleanupTask struct {
	taskCtx *TaskContext
}

func NewTTLCleanupTask(taskCtx *TaskContext) *TTLCleanupTask {
	return &TTLCleanupTask{taskCtx: taskCtx}
}

func (t *TTLCleanupTask) Start() {
	go func() {
		// 启动时不立即执行，等分区任务先跑完
		for {
			now := time.Now().In(time.FixedZone("SGT", 8*3600))
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 4, 5, 0, 0, now.Location())
			time.Sleep(time.Until(next))
			t.run()
		}
	}()
}

func (t *TTLCleanupTask) run() {
	rs := &t.taskCtx.RedSync
	mutex := rs.NewMutex("lock:ttl_cleanup", redsync.WithExpiry(30*time.Minute), redsync.WithTries(1))
	if err := mutex.Lock(); err != nil {
		return
	}
	defer mutex.Unlock()

	log.Info("[TTLCleanup] 开始清理 t_stake_reward / t_stake_reward_claim")

	cutoff := time.Now().AddDate(0, -1, 0)

	// 1. 清理 t_stake_reward: 删除1月前的数据，但保留 reward_type=0 AND reward_state=0 (未领取的固定利息)
	t.batchDelete(
		"t_stake_reward",
		"created_at < ? AND NOT (reward_type = 0 AND reward_state = 0)",
		cutoff,
	)

	// 2. 清理 t_stake_reward_claim: 删除1月前的所有数据
	t.batchDelete(
		"t_stake_reward_claim",
		"created_at < ?",
		cutoff,
	)

	// 3. VACUUM FULL 回收磁盘空间
	t.vacuumFull("t_stake_reward")
	t.vacuumFull("t_stake_reward_claim")

	log.Info("[TTLCleanup] 清理完成")
}

func (t *TTLCleanupTask) batchDelete(tableName, where string, args ...interface{}) {
	totalDeleted := int64(0)
	for {
		// ctid 子查询限制每次删除批量大小，避免长事务
		sql := "DELETE FROM " + tableName + " WHERE ctid IN (SELECT ctid FROM " + tableName + " WHERE " + where + " LIMIT ?)"
		allArgs := append(args, ttlCleanupBatchSize)
		result := t.taskCtx.DB.Exec(sql, allArgs...)
		if result.Error != nil {
			log.Errorf("[TTLCleanup] 删除 %s 失败: %v", tableName, result.Error)
			break
		}
		totalDeleted += result.RowsAffected
		if result.RowsAffected < ttlCleanupBatchSize {
			break
		}
	}
	if totalDeleted > 0 {
		log.Infof("[TTLCleanup] %s 删除 %d 行", tableName, totalDeleted)
	}
}

func (t *TTLCleanupTask) vacuumFull(tableName string) {
	// VACUUM FULL 不能在事务内执行，需要用原生连接
	sqlDB, err := t.taskCtx.DB.DB()
	if err != nil {
		log.Errorf("[TTLCleanup] 获取 DB 连接失败: %v", err)
		return
	}
	if _, err := sqlDB.Exec("VACUUM FULL " + tableName); err != nil {
		log.Errorf("[TTLCleanup] VACUUM FULL %s 失败: %v", tableName, err)
	} else {
		log.Infof("[TTLCleanup] VACUUM FULL %s 完成", tableName)
	}
}
