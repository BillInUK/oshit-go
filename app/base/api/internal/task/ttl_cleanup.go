package task

import (
	"time"

	"github.com/go-redsync/redsync/v4"
	log "github.com/sirupsen/logrus"
)

const ttlCleanupBatchSize = 5000

// TTLCleanupTask 统一管理所有 TTL 表的 DELETE + VACUUM FULL 清理
// 每天 SGT 04:00 执行，凌晨低流量时段
type TTLCleanupTask struct {
	taskCtx *TaskContext
}

func NewTTLCleanupTask(taskCtx *TaskContext) *TTLCleanupTask {
	return &TTLCleanupTask{taskCtx: taskCtx}
}

func (t *TTLCleanupTask) Start() {
	go func() {
		for {
			now := time.Now().In(time.FixedZone("SGT", 8*3600))
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 4, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
			t.run()
		}
	}()
}

// ttlTable 描述一张需要 TTL 清理的表
type ttlTable struct {
	name  string
	where string        // DELETE WHERE 条件，? 占位符
	ttl   time.Duration // 数据保留时长
}

var ttlTables = []ttlTable{
	// ── TTL 2天 ──
	{name: "t_qn_fee", where: "created_at < ?", ttl: 2 * 24 * time.Hour},
	{name: "t_fee_statistics", where: "created_at < ?", ttl: 2 * 24 * time.Hour},
	{name: "t_stake_buy_token", where: "created_at < ?", ttl: 2 * 24 * time.Hour},

	// ── TTL 7天 ──
	{name: "t_service_tx", where: "created_at < ?", ttl: 7 * 24 * time.Hour},

	// ── TTL 1月 ──
	{name: "t_take_token_record", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_give_token_record", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_lottery_claim", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_daily_claim_stats", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_lottery_reward", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_campaign_quote_record", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_pos_snap_shot", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_pos_reward", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_pos_reward_claim", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_stake_snap_shot", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_stake_reward_claim", where: "created_at < ?", ttl: 30 * 24 * time.Hour},
	{name: "t_reward_code", where: "created_at < ?", ttl: 30 * 24 * time.Hour},

	// ── TTL 1月 (特殊条件: 保留未领取的固定利息) ──
	{name: "t_stake_reward", where: "created_at < ? AND NOT (reward_type = 0 AND reward_state = 0)", ttl: 30 * 24 * time.Hour},
}

func (t *TTLCleanupTask) run() {
	rs := &t.taskCtx.RedSync
	mutex := rs.NewMutex("lock:ttl_cleanup", redsync.WithExpiry(30*time.Minute), redsync.WithTries(1))
	if err := mutex.Lock(); err != nil {
		return
	}
	defer mutex.Unlock()

	log.Info("[TTLCleanup] 开始清理")

	for _, tbl := range ttlTables {
		cutoff := time.Now().Add(-tbl.ttl)
		t.batchDelete(tbl.name, tbl.where, cutoff)
		t.vacuumFull(tbl.name)
	}

	log.Info("[TTLCleanup] 清理完成")
}

func (t *TTLCleanupTask) batchDelete(tableName, where string, args ...interface{}) {
	totalDeleted := int64(0)
	for {
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
