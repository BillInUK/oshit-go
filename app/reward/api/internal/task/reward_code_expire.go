package task

import (
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"time"
)

// RewardCodeExpireTask 定期将已过期但未使用的奖励码标记为超时（tx_state=-2）
type RewardCodeExpireTask struct {
	taskCtx *TaskContext
	db      *gorm.DB
}

func NewRewardCodeExpireTask(taskCtx *TaskContext) *RewardCodeExpireTask {
	return &RewardCodeExpireTask{
		taskCtx: taskCtx,
		db:      taskCtx.DB,
	}
}

func (t *RewardCodeExpireTask) Start() {
	go runPeriodic(
		&t.taskCtx.RedSync,
		5*time.Minute,
		"reward:sol:reward-code:expire:lock",
		6*time.Minute,
		t.run,
	)
}

func (t *RewardCodeExpireTask) run() {
	result := t.db.Table(model.TableNameRewardCode).
		Where("tx_state = ? AND expired_at < ?", constants.TxStateInit, time.Now()).
		Update("tx_state", constants.TxStateExpired)

	if result.Error != nil {
		log.Errorf("RewardCodeExpireTask: 更新过期奖励码状态失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Infof("RewardCodeExpireTask: 标记 %d 条奖励码为超时 (tx_state=-2)", result.RowsAffected)
	}
}
