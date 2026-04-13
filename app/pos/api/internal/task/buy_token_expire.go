package task

import (
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
	"oshit-go/common/pkg/dal/model"
	"time"
)

const stakeBuyTokenExpireHours = 26

// BuyTokenExpireTask 定期将超过 26 小时未使用的购买记录标记为过期
type BuyTokenExpireTask struct {
	taskCtx *TaskContext
	db      *gorm.DB
}

func NewStakeBuyTokenExpireTask(taskCtx *TaskContext) *BuyTokenExpireTask {
	return &BuyTokenExpireTask{
		taskCtx: taskCtx,
		db:      taskCtx.DB,
	}
}

func (t *BuyTokenExpireTask) Start() {
	go runPeriodic(
		&t.taskCtx.RedSync,
		10*time.Minute,
		"pos:sol:stake-buy-token:expire:lock",
		12*time.Minute,
		t.run,
	)
}

func (t *BuyTokenExpireTask) run() {
	expireThreshold := time.Now().Add(-stakeBuyTokenExpireHours * time.Hour)
	now := time.Now()

	result := t.db.Table(model.TableNameStakeBuyToken).
		Where("expired = ? AND created_at < ?", false, expireThreshold).
		Updates(map[string]interface{}{
			"expired":          true,
			"remaining_amount": 0,
			"expired_at":       now,
		})

	if result.Error != nil {
		log.Errorf("BuyTokenExpireTask: 更新过期购买记录失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Infof("BuyTokenExpireTask: 标记 %d 条购买记录为过期", result.RowsAffected)
	}
}
