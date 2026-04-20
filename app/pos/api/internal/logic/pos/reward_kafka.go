package pos

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	app_utils "oshit-go/app/utils"
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

	// 根据交易 Id 找到记录
	table := dbTx.Table(model.TableNamePosRewardClaim)
	if err = table.Where("tx_id = ?", txId).First(&claimRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 查找领取记录错误: %v", txId, err)
		dbTx.Rollback()
		return err
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
		table = dbTx.Table(model.TableNamePosReward)
		if err = table.Where("record_id IN ? ", claimRecord.RewardIds).Updates(map[string]interface{}{"reward_state": constants.RewardStateInit, "pending": false}).Error; err != nil {
			log.Errorf("pos业务 - 更新奖励记录为处理中，错误: %v", err)
			return err
		}
		if len(rewardIds) > 0 {
			if err = table.Where("record_id IN ? ", rewardIds).Updates(map[string]interface{}{
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
		table = dbTx.Table(model.TableNamePosReward)
		if len(rewardIds) > 0 {
			if err = table.Where("record_id IN ? ", rewardIds).Updates(map[string]interface{}{
				"reward_state": constants.RewardStateClaimed,
				"pending":      false,
				"updated_at":   time.Now(),
			}).Error; err != nil {
				log.Errorf("pos业务 - 更新奖励记录为处理中，错误: %v", err)
				return err
			}
		}
		// 记录流水
		decodedServiceTx, err := app_utils.DecodeServiceTransaction(&msg.DecodedTx)
		if err != nil {
			log.Errorf("pos业务 - 解码服务交易失败: %v", err)
			dbTx.Rollback()
			return err
		}
		err = l.recordClaimRewardFlow(decodedServiceTx)
		if err != nil {
			log.Errorf("pos业务 - 记录流水失败: %v", err)
			dbTx.Rollback()
			return err
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

	// 根据交易 Id 找到记录
	table := dbTx.Table(model.TableNamePosRewardClaim)
	if err = table.Where("tx_id = ?", txId).First(&claimRecord).Error; err != nil {
		log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 查找领取记录错误: %v", txId, err)
		dbTx.Rollback()
		return err
	}
	if err = table.Where("tx_id = ?", txId).Update("tx_state", constants.TxStateFailed).Error; err != nil {
		log.Errorf("Pos业务 - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为失败，错误: %v", txId, err)
		dbTx.Rollback()
		return err
	}
	// 将奖励记录的状态设置处理完成
	rewardIds := utils.ParseDbArray(claimRecord.RewardIds)
	table = dbTx.Table(model.TableNamePosReward)
	if err = table.Where("record_id IN ? ", rewardIds).Updates(map[string]interface{}{"tx_state": constants.TxStateInit, "pending": false}).Error; err != nil {
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

// recordClaimRewardFlow 记录POS领取token记录
func (l *PosRewardLogic) recordClaimRewardFlow(decodedServiceTx *entity.DecodedServiceTransaction) error {
	var err error
	var fundFlows []model.FundFlow

	table := l.db.Table(model.TableNameFundFlow)
	// 1.记录dex入账sol流水
	dexInputFlow := model.FundFlow{
		IsToken:     false,
		FromAccount: decodedServiceTx.FromNativeAccount,
		ToAccount:   decodedServiceTx.ToDexInst.ToNativeAccount,
		TxID:        decodedServiceTx.TxID,
		Direction:   constants.FlowInput.String(),
		ServiceType: constants.SubServicePosReward.String(),
		FlowType:    constants.FlowCost.String(),
		Decimals:    9,
		Amount:      float64(decodedServiceTx.ToDexInst.Amount),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	fundFlows = append(fundFlows, dexInputFlow)
	// 2.记录奖励转账人出账流水
	rewardOutputFlow := model.FundFlow{
		IsToken:     true,
		FromAccount: decodedServiceTx.RewardInst.FromNativeAccount,
		ToAccount:   decodedServiceTx.RewardInst.ToNativeAccount,
		TxID:        decodedServiceTx.TxID,
		Direction:   constants.FlowOutput.String(),
		ServiceType: constants.SubServicePosReward.String(),
		FlowType:    constants.FlowReceipt.String(),
		Decimals:    int16(decodedServiceTx.RewardInst.Decimals),
		Amount:      decodedServiceTx.RewardInst.Amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	fundFlows = append(fundFlows, rewardOutputFlow)

	// 批量插入，当唯一键冲突时更新UpdateTime
	batchSize := 100

	// 使用ON CONFLICT DO UPDATE SET (推荐)
	if err = table.Omit("record_id").Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tx_id"},
			{Name: "to_account"},
			{Name: "flow_type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"updated_at": time.Now(),
		}),
	}).CreateInBatches(&fundFlows, batchSize).Error; err != nil {
		log.Errorf("%s - 记录流水错误: %v", l.prefix, err)
		return err
	}

	return nil
}
