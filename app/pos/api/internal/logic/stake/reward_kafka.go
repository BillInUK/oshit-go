package stake

import (
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm/clause"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"time"
)

func (l *StakeRewardLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	var err error
	txId := msg.TxSig.Signature.String()

	// 分布式锁
	mutex := l.rs.NewMutex("stake:reward:process-tx:" + txId)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer mutex.Unlock()

	// 开启数据库事务
	dbTx := l.db.Begin()
	defer dbTx.Rollback()

	// 如果交易执行失败
	if msg.TxSig.Err != nil {
		// 重置奖励领取状态
		table := dbTx.Table(model.TableNameStakeReward)
		if err = table.Where("tx_id = ? ", txId).Updates(map[string]interface{}{
			"tx_id":        nil,
			"reward_state": 0,
			"pending":      false,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			log.Errorf("%s 处理Kafka消息 - 根据交易 Id %s执行错误 - 重置奖励状态为初始化，错误: %v", l.prefix, txId, err)
			return err
		}
	} else {
		// 更新 t_stake_reward_claim_record 的交易状态为已成功
		table := dbTx.Table(model.TableNameStakeRewardClaim)
		if err = table.Where("tx_id = ?", txId).Updates(map[string]interface{}{
			"tx_state":   1,
			"updated_at": time.Now(),
		}).Error; err != nil {
			log.Errorf("%s 更新claim record状态失败，错误: %v", l.prefix, err)
			return err
		}
		// 将关联的 t_stake_reward 记录标记为已领取
		table = dbTx.Table(model.TableNameStakeReward)
		if err = table.Where("tx_id = ?", txId).Updates(map[string]interface{}{
			"reward_state": 1,
			"pending":      false,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			log.Errorf("%s 更新stake reward状态失败，错误: %v", l.prefix, err)
			return err
		}
		// 记录流水
		decodedServiceTx, err := app_utils.DecodeServiceTransaction(&msg.DecodedTx)
		if err != nil {
			log.Errorf("%s 处理Kafka消息 - 解码服务交易失败: %v", l.prefix, err)
			dbTx.Rollback()
			return err
		}
		err = l.recordClaimRewardFlow(decodedServiceTx)
		if err != nil {
			log.Errorf("%s 记录流水失败: %v", l.prefix, err)
			dbTx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", l.prefix, err)
		dbTx.Rollback()
		return nil
	}
	return nil
}

func (l *StakeRewardLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	var err error
	txId := msg.TxID

	// 分布式锁
	mutex := l.rs.NewMutex("stake:reward:process-tx:" + txId)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer mutex.Unlock()

	// 开启数据库事务
	dbTx := l.db.Begin()
	defer dbTx.Rollback()

	// 将 t_stake_reward_claim_record 标记为失败
	claimTable := dbTx.Table(model.TableNameStakeRewardClaim)
	if err = claimTable.Where("tx_id = ?", txId).Updates(map[string]interface{}{
		"tx_state":   -1,
		"updated_at": time.Now(),
	}).Error; err != nil {
		log.Errorf("%s 更新claim record为失败，错误: %v", l.prefix, err)
		dbTx.Rollback()
		return err
	}
	// 将关联的 t_stake_reward 重置为可领取状态
	table := dbTx.Table(model.TableNameStakeReward)
	if err = table.Where("tx_id = ?", txId).Updates(map[string]interface{}{
		"tx_id":        nil,
		"reward_state": 0,
		"pending":      false,
		"updated_at":   time.Now(),
	}).Error; err != nil {
		log.Errorf("%s 重置stake reward状态失败，错误: %v", l.prefix, err)
		dbTx.Rollback()
		return err
	}
	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", l.prefix, err)
		dbTx.Rollback()
		return err
	}
	return nil
}

// recordClaimRewardFlow 记录Stake领取token记录
func (l *StakeRewardLogic) recordClaimRewardFlow(decodedServiceTx *entity.DecodedServiceTransaction) error {
	var err error
	var fundFlows []model.FundFlow

	db := l.db.Table(model.TableNameFundFlow)
	// 1.记录dex入账sol流水
	dexInputFlow := model.FundFlow{
		IsToken:     false,
		FromAccount: decodedServiceTx.FromNativeAccount,
		ToAccount:   decodedServiceTx.ToDexInst.ToNativeAccount,
		TxID:        decodedServiceTx.TxID,
		Direction:   constants.FlowInput,
		ServiceType: constants.ServiceStake,
		FlowType:    constants.FlowStakeCost,
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
		Direction:   constants.FlowOutput,
		ServiceType: constants.ServiceStake,
		FlowType:    constants.FlowStakeReceipt,
		Decimals:    int16(decodedServiceTx.RewardInst.Decimals),
		Amount:      decodedServiceTx.RewardInst.Amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	fundFlows = append(fundFlows, rewardOutputFlow)
	// 批量插入，当唯一键冲突时更新UpdateTime
	batchSize := 100

	// 使用ON CONFLICT DO UPDATE SET (推荐)
	if err = db.Omit("record_id").Clauses(clause.OnConflict{
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
