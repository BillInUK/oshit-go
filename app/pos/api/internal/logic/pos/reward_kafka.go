package pos

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"time"
)

// HandleScannedTx 处理 base 模块推送的 TakeToken 链上已确认交易
func (l *PosRewardLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	var err error
	var claimRecord model.PosRewardClaim
	txId := msg.TxSig.Signature.String()

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
			log.Errorf("Pos业务 - 处理Kafka消息时发生错误: %v", r)
		}
	}()

	// 根据交易 Id 找到记录，SELECT FOR UPDATE 防止并发重复处理
	if err = dbTx.Table(model.TableNamePosRewardClaim).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&claimRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 查找领取记录错误: %v", txId, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if claimRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("Pos业务 - 交易 %s 状态已为 %d，跳过重复处理", txId, claimRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	rewardIds := utils.ParseDbArray(claimRecord.RewardIds)

	if msg.TxSig.Err != nil {
		if err = dbTx.Table(model.TableNamePosRewardClaim).
			Where("record_id = ?", claimRecord.RecordID).Update("tx_state", constants.TxStateFailed).Error; err != nil {
			log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为失败，错误: %v", txId, err)
			dbTx.Rollback()
			return err
		}
		// 将奖励记录的状态设置处理完成
		rewardTable := dbTx.Table(model.TableNamePosReward)
		if err = rewardTable.Where("record_id IN ? ", claimRecord.RewardIds).Updates(map[string]interface{}{"reward_state": constants.RewardStateInit, "pending": false}).Error; err != nil {
			log.Errorf("pos业务 - 更新奖励记录为处理中，错误: %v", err)
			return err
		}
		if len(rewardIds) > 0 {
			if err = rewardTable.Where("record_id IN ? ", rewardIds).Updates(map[string]interface{}{
				"reward_state": constants.RewardStateInit,
				"pending":      false,
				"updated_at":   time.Now(),
			}).Error; err != nil {
				log.Errorf("pos业务 - 更新奖励记录为处理中，错误: %v", err)
				return err
			}
		}
	} else {
		if err = dbTx.Table(model.TableNamePosRewardClaim).
			Where("record_id = ?", claimRecord.RecordID).Update("tx_state", constants.TxStateSuccess).Error; err != nil {
			log.Errorf("pos业务 - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为成功，错误: %v", txId, err)
			dbTx.Rollback()
			return err
		}
		// 将奖励记录的状态设置为已经处理
		// 更新当中的记录为
		rewardTable := dbTx.Table(model.TableNamePosReward)
		if len(rewardIds) > 0 {
			if err = rewardTable.Where("record_id IN ? ", rewardIds).Updates(map[string]interface{}{
				"reward_state": constants.RewardStateClaimed,
				"pending":      false,
				"updated_at":   time.Now(),
			}).Error; err != nil {
				log.Errorf("pos业务 - 更新奖励记录为处理中，错误: %v", err)
				return err
			}
		}
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("pos业务 - 提交事务失败: %v", err)
		dbTx.Rollback()
		return err
	}
	return nil
}

// HandleExpiredTx 处理 base 模块推送的 TakeToken 链上已确认交易
func (l *PosRewardLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	var err error
	var claimRecord model.PosRewardClaim
	txId := msg.TxID

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
			log.Errorf("Pos业务 - 处理Kafka消息时发生错误: %v", r)
		}
	}()

	// 根据交易 Id 找到记录，SELECT FOR UPDATE 防止并发重复处理
	if err = dbTx.Table(model.TableNamePosRewardClaim).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&claimRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 查找领取记录错误: %v", txId, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if claimRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("Pos业务 - 超时交易 %s 状态已为 %d，跳过重复处理", txId, claimRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	if err = dbTx.Table(model.TableNamePosRewardClaim).Where("tx_id = ?", txId).Update("tx_state", constants.TxStateFailed).Error; err != nil {
		log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为失败，错误: %v", txId, err)
		dbTx.Rollback()
		return err
	}
	// 将奖励记录的状态重置为可领取
	rewardIds := utils.ParseDbArray(claimRecord.RewardIds)
	if err = dbTx.Table(model.TableNamePosReward).Where("record_id IN ? ", rewardIds).Updates(map[string]interface{}{"reward_state": constants.RewardStateInit, "pending": false}).Error; err != nil {
		log.Errorf("pos业务 - 更新奖励记录为处理中，错误: %v", err)
		dbTx.Rollback()
		return err
	}
	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("pos业务 - 提交事务失败: %v", err)
		dbTx.Rollback()
		return err
	}
	return nil
}

