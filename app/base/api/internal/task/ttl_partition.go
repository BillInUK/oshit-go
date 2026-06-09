package task

import (
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	log "github.com/sirupsen/logrus"
)

// partitionedTable 描述一张需要分区管理的表
type partitionedTable struct {
	TableName    string        // 父表名
	PartitionKey string        // 分区键列名
	KeyType      string        // "timestamp" 或 "date"
	TTL          time.Duration // 数据保留时长
	Granularity  string        // "day" 或 "week"
}

var partitionedTables = []partitionedTable{
	// TTL 2天, 按天分区
	{TableName: "t_qn_fee", PartitionKey: "created_at", KeyType: "timestamp", TTL: 2 * 24 * time.Hour, Granularity: "day"},
	{TableName: "t_fee_statistics", PartitionKey: "created_at", KeyType: "timestamp", TTL: 2 * 24 * time.Hour, Granularity: "day"},
	{TableName: "t_stake_buy_token", PartitionKey: "created_at", KeyType: "timestamp", TTL: 2 * 24 * time.Hour, Granularity: "day"},

	// TTL 7天, 按天分区
	{TableName: "t_service_tx", PartitionKey: "created_at", KeyType: "timestamp", TTL: 7 * 24 * time.Hour, Granularity: "day"},

	// TTL 1月, 按周分区 (分区键: created_at)
	{TableName: "t_take_token_record", PartitionKey: "created_at", KeyType: "timestamp", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_give_token_record", PartitionKey: "created_at", KeyType: "timestamp", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_lottery_claim", PartitionKey: "created_at", KeyType: "timestamp", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_campaign_quote_record", PartitionKey: "created_at", KeyType: "timestamp", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_pos_reward_claim", PartitionKey: "created_at", KeyType: "timestamp", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_reward_code", PartitionKey: "created_at", KeyType: "timestamp", TTL: 30 * 24 * time.Hour, Granularity: "week"},

	// TTL 1月, 按周分区 (分区键: 业务日期字段)
	{TableName: "t_daily_claim_stats", PartitionKey: "take_date", KeyType: "date", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_lottery_reward", PartitionKey: "reward_day", KeyType: "date", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_pos_snap_shot", PartitionKey: "snap_day", KeyType: "date", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_pos_reward", PartitionKey: "snap_day", KeyType: "date", TTL: 30 * 24 * time.Hour, Granularity: "week"},
	{TableName: "t_stake_snap_shot", PartitionKey: "snap_day", KeyType: "date", TTL: 30 * 24 * time.Hour, Granularity: "week"},
}

type TTLPartitionTask struct {
	taskCtx *TaskContext
}

func NewTTLPartitionTask(taskCtx *TaskContext) *TTLPartitionTask {
	return &TTLPartitionTask{taskCtx: taskCtx}
}

func (t *TTLPartitionTask) Start() {
	go func() {
		// 启动时立即执行一次，确保分区存在
		t.run()

		// 之后每天 SGT 04:00 执行
		for {
			now := time.Now().In(time.FixedZone("SGT", 8*3600))
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 4, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
			t.run()
		}
	}()
}

func (t *TTLPartitionTask) run() {
	rs := &t.taskCtx.RedSync
	mutex := rs.NewMutex("lock:ttl_partition", redsync.WithExpiry(10*time.Minute), redsync.WithTries(1))
	if err := mutex.Lock(); err != nil {
		return
	}
	defer mutex.Unlock()

	log.Info("[TTLPartition] 开始分区管理")

	for _, pt := range partitionedTables {
		t.ensureFuturePartitions(pt)
		t.dropExpiredPartitions(pt)
	}

	log.Info("[TTLPartition] 分区管理完成")
}

// ensureFuturePartitions 预创建未来的分区
func (t *TTLPartitionTask) ensureFuturePartitions(pt partitionedTable) {
	now := time.Now()

	var periods []struct{ start, end time.Time }

	switch pt.Granularity {
	case "day":
		// 创建今天 + 未来2天的分区
		for i := 0; i < 3; i++ {
			day := now.AddDate(0, 0, i)
			start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 0, 1)
			periods = append(periods, struct{ start, end time.Time }{start, end})
		}
	case "week":
		// 创建本周 + 未来2周的分区
		// 找到本周一
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		for i := 0; i < 3; i++ {
			start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
			start = start.AddDate(0, 0, 7*i)
			end := start.AddDate(0, 0, 7)
			periods = append(periods, struct{ start, end time.Time }{start, end})
		}
	}

	for _, p := range periods {
		partName := t.partitionName(pt.TableName, pt.Granularity, p.start)
		fromStr := p.start.Format("2006-01-02")
		toStr := p.end.Format("2006-01-02")

		sql := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			partName, pt.TableName, fromStr, toStr,
		)
		if err := t.taskCtx.DB.Exec(sql).Error; err != nil {
			log.Errorf("[TTLPartition] 创建分区 %s 失败: %v", partName, err)
		}
	}
}

// dropExpiredPartitions 删除过期分区
func (t *TTLPartitionTask) dropExpiredPartitions(pt partitionedTable) {
	cutoff := time.Now().Add(-pt.TTL).Format("2006-01-02")

	// 查询所有子分区
	rows, err := t.taskCtx.DB.Raw(`
		SELECT c.relname, pg_get_expr(c.relpartbound, c.oid)
		FROM pg_class c
		JOIN pg_inherits i ON c.oid = i.inhrelid
		JOIN pg_class p ON i.inhparent = p.oid
		WHERE p.relname = ? AND c.relname != ?
		ORDER BY c.relname
	`, pt.TableName, pt.TableName+"_default").Rows()
	if err != nil {
		log.Errorf("[TTLPartition] 查询 %s 分区列表失败: %v", pt.TableName, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var partName, partBound string
		if err := rows.Scan(&partName, &partBound); err != nil {
			continue
		}

		// 解析分区上界，格式: FOR VALUES FROM ('2025-06-01') TO ('2025-06-08')
		endDate := t.parsePartitionEnd(partBound)
		if endDate == "" {
			continue
		}

		// 分区上界 <= cutoff 说明整个分区的数据都过期了
		if endDate <= cutoff {
			sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", partName)
			if err := t.taskCtx.DB.Exec(sql).Error; err != nil {
				log.Errorf("[TTLPartition] 删除分区 %s 失败: %v", partName, err)
			} else {
				log.Infof("[TTLPartition] 已删除过期分区 %s (上界 %s <= %s)", partName, endDate, cutoff)
			}
		}
	}
}

// partitionName 生成分区名称
func (t *TTLPartitionTask) partitionName(tableName, granularity string, start time.Time) string {
	switch granularity {
	case "day":
		return fmt.Sprintf("%s_p%s", tableName, start.Format("20060102"))
	case "week":
		year, week := start.ISOWeek()
		return fmt.Sprintf("%s_p%dw%02d", tableName, year, week)
	default:
		return fmt.Sprintf("%s_p%s", tableName, start.Format("20060102"))
	}
}

// parsePartitionEnd 从分区边界表达式中提取上界日期
func (t *TTLPartitionTask) parsePartitionEnd(partBound string) string {
	// 格式: FOR VALUES FROM ('2025-06-01') TO ('2025-06-08')
	// 或:   FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-06-08 00:00:00')
	const toMarker = "TO ('"
	idx := 0
	for i := len(partBound) - 1; i >= 0; i-- {
		if i+len(toMarker) <= len(partBound) && partBound[i:i+len(toMarker)] == toMarker {
			idx = i + len(toMarker)
			break
		}
	}
	if idx == 0 {
		return ""
	}
	// 提取日期部分 (前10个字符 = YYYY-MM-DD)
	end := partBound[idx:]
	if len(end) >= 10 {
		return end[:10]
	}
	return ""
}
