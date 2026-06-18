package lottery

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"runtime/debug"
	"strings"
	"time"
)

// HandleScannedTx 处理 base 模块推送的 Lottery 链上已确认交易
func (l *LotteryLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	txId := msg.TxSig.Signature.String()
	prefix := fmt.Sprintf("%s 处理扫描到的交易 %s -", l.prefix, txId)

	dbTx := l.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Errorf("%s 处理消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			dbTx.Rollback()
		}
	}()

	// 1. 找到 t_lottery_claim_record，SELECT FOR UPDATE 防止并发重复处理
	var claimRecord model.LotteryClaim
	if err := dbTx.Table(model.TableNameLotteryClaim).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&claimRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找领取记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if claimRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, claimRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	rewardId := strings.Trim(claimRecord.RewardIds, "{}")

	if msg.TxSig.Err != nil {
		// 交易失败
		// 更新 claim_record.state = -1
		if err := dbTx.Table(model.TableNameLotteryClaim).
			Where("record_id = ?", claimRecord.RecordID).
			Updates(map[string]interface{}{"tx_state": constants.TxStateFailed, "updated_at": time.Now()}).Error; err != nil {
			log.Errorf("%s 更新领取记录状态为失败错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
		// 更新 t_lottery_reward.state = -1, pending = false
		if err := dbTx.Table(model.TableNameLotteryReward).
			Where("record_id = ?", rewardId).
			Updates(map[string]interface{}{"reward_state": constants.TxStateFailed, "pending": false, "updated_at": time.Now()}).Error; err != nil {
			log.Errorf("%s 更新抽奖奖励状态为失败错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
	} else {
		// 交易成功
		// 更新 claim_record.state = 1
		if err := dbTx.Table(model.TableNameLotteryClaim).
			Where("record_id = ?", claimRecord.RecordID).
			Updates(map[string]interface{}{"tx_state": constants.TxStateSuccess, "updated_at": time.Now()}).Error; err != nil {
			log.Errorf("%s 更新领取记录状态为成功错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}

		// 更新 t_lottery_reward.state = 1, pending = false
		var reward model.LotteryReward
		if err := dbTx.Table(model.TableNameLotteryReward).
			Where("record_id = ?", rewardId).First(&reward).Error; err != nil {
			log.Errorf("%s 查找抽奖奖励记录错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
		if err := dbTx.Table(model.TableNameLotteryReward).
			Where("record_id = ?", rewardId).
			Updates(map[string]interface{}{"reward_state": constants.RewardStateClaimed, "pending": false, "updated_at": time.Now()}).Error; err != nil {
			log.Errorf("%s 更新抽奖奖励状态为成功错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}

		// 更新 t_daily_claim_stats.need_lottery = false
		today := time.Now().Truncate(24 * time.Hour)
		if err := dbTx.Table(model.TableNameDailyClaimStats).
			Where("native_account = ? AND take_date = ?", reward.NativeAccount, today).
			Updates(map[string]interface{}{"need_lottery": false, "updated_at": time.Now()}).Error; err != nil {
			log.Errorf("%s 更新每日统计need_lottery为false错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}

	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 通知 ProcessCommitTx 中等待确认的 channel
	finalState := constants.TxStateSuccess
	if msg.TxSig.Err != nil {
		finalState = constants.TxStateFailed
	}
	if ch, ok := lotteryPendingMap.LoadAndDelete(txId); ok {
		if c, ok := ch.(chan int32); ok {
			c <- int32(finalState)
		}
	}

	return nil
}

// HandleExpiredTx 处理 base 模块推送的 Lottery 已超时交易
func (l *LotteryLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	txId := msg.TxID
	prefix := fmt.Sprintf("%s 处理超时交易 %s -", l.prefix, txId)

	dbTx := l.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Errorf("%s 处理消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			dbTx.Rollback()
		}
	}()

	// 找到 t_lottery_claim_record，SELECT FOR UPDATE 防止并发重复处理
	var claimRecord model.LotteryClaim
	if err := dbTx.Table(model.TableNameLotteryClaim).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&claimRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找领取记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if claimRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, claimRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	rewardId := strings.Trim(claimRecord.RewardIds, "{}")

	// 更新 claim_record.state = -1
	if err := dbTx.Table(model.TableNameLotteryClaim).
		Where("record_id = ?", claimRecord.RecordID).
		Updates(map[string]interface{}{"tx_state": constants.TxStateFailed, "updated_at": time.Now()}).Error; err != nil {
		log.Errorf("%s 更新领取记录状态为失败错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 更新 t_lottery_reward.state = -1, pending = false
	if err := dbTx.Table(model.TableNameLotteryReward).
		Where("record_id = ?", rewardId).
		Updates(map[string]interface{}{"reward_state": constants.RewardStateFailed, "pending": false, "updated_at": time.Now()}).Error; err != nil {
		log.Errorf("%s 更新抽奖奖励状态为失败错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 通知 ProcessCommitTx 中等待确认的 channel（交易已过期）
	if ch, ok := lotteryPendingMap.LoadAndDelete(txId); ok {
		if c, ok := ch.(chan int32); ok {
			c <- int32(constants.TxStateExpired)
		}
	}

	return nil
}

