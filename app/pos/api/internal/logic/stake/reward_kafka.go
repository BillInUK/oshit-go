package stake

import (
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm/clause"
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

	// 查询记录并加行锁，防重复处理
	var stakeReward model.StakeReward
	if err = dbTx.Table(model.TableNameStakeReward).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&stakeReward).Error; err != nil {
		log.Errorf("%s 处理Kafka消息 - 根据交易 Id %s 查找奖励记录错误: %v", l.prefix, txId, err)
		return err
	}

	// 只有待领取状态才处理
	if stakeReward.RewardState != int32(constants.RewardStateInit) {
		log.Infof("%s 交易 %s 奖励状态已为 %d，跳过重复处理", l.prefix, txId, stakeReward.RewardState)
		return nil
	}

	// 如果交易执行失败
	if msg.TxSig.Err != nil {
		// 重置奖励领取状态
		table := dbTx.Table(model.TableNameStakeReward)
		if err = table.Where("tx_id = ? ", txId).Updates(map[string]interface{}{
			"tx_id":        nil,
			"reward_state": constants.RewardStateInit,
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
			"tx_state":   constants.TxStateSuccess,
			"updated_at": time.Now(),
		}).Error; err != nil {
			log.Errorf("%s 更新claim record状态失败，错误: %v", l.prefix, err)
			return err
		}
		// 将关联的 t_stake_reward 记录标记为已领取
		table = dbTx.Table(model.TableNameStakeReward)
		if err = table.Where("tx_id = ?", txId).Updates(map[string]interface{}{
			"reward_state": constants.RewardStateClaimed,
			"pending":      false,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			log.Errorf("%s 更新stake reward状态失败，错误: %v", l.prefix, err)
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

	// 查询记录并加行锁，防重复处理
	var stakeReward model.StakeReward
	if err = dbTx.Table(model.TableNameStakeReward).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&stakeReward).Error; err != nil {
		log.Errorf("%s 处理超时交易 - 根据交易 Id %s 查找奖励记录错误: %v", l.prefix, txId, err)
		return err
	}

	// 只有待领取状态才处理
	if stakeReward.RewardState != int32(constants.RewardStateInit) {
		log.Infof("%s 超时交易 %s 奖励状态已为 %d，跳过重复处理", l.prefix, txId, stakeReward.RewardState)
		return nil
	}

	// 将 t_stake_reward_claim_record 标记为失败
	claimTable := dbTx.Table(model.TableNameStakeRewardClaim)
	if err = claimTable.Where("tx_id = ?", txId).Updates(map[string]interface{}{
		"tx_state":   constants.TxStateFailed,
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
		"reward_state": constants.RewardStateInit,
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

// HandleLeaderScannedTx 处理区域经理领取奖励交易上链确认的 Kafka 消息
func (l *StakeRewardLogic) HandleLeaderScannedTx(msg entity.NewScannedTx) error {
	txId := msg.TxSig.Signature.String()

	mutex := l.rs.NewMutex("stake:leader-reward:process-tx:" + txId)
	if err := mutex.Lock(); err != nil {
		return err
	}
	defer mutex.Unlock()

	dbTx := l.db.Begin()
	defer dbTx.Rollback()

	// 查询记录并加行锁，防重复处理
	var leaderReward model.StakeLeaderReward
	if err := dbTx.Table(model.TableNameStakeLeaderReward).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&leaderReward).Error; err != nil {
		log.Errorf("%s HandleLeaderScannedTx - 查找记录失败 txId=%s: %v", l.prefix, txId, err)
		return err
	}

	// 只有待领取状态才处理
	if leaderReward.RewardState != int32(constants.RewardStateInit) {
		log.Infof("%s HandleLeaderScannedTx - 交易 %s 奖励状态已为 %d，跳过重复处理", l.prefix, txId, leaderReward.RewardState)
		return nil
	}

	if msg.TxSig.Err != nil {
		// 链上执行失败：重置奖励领取状态
		if err := dbTx.Table(model.TableNameStakeLeaderReward).
			Where("tx_id = ?", txId).
			Updates(map[string]interface{}{
				"tx_id":        nil,
				"reward_state": constants.RewardStateInit,
				"pending":      false,
				"updated_at":   time.Now(),
			}).Error; err != nil {
			log.Errorf("%s HandleLeaderScannedTx - 重置奖励状态失败 txId=%s: %v", l.prefix, txId, err)
			return err
		}
		if err := dbTx.Table(model.TableNameStakeLeaderRewardClaim).
			Where("tx_id = ?", txId).
			Updates(map[string]interface{}{
				"tx_state":   constants.TxStateFailed,
				"updated_at": time.Now(),
			}).Error; err != nil {
			log.Errorf("%s HandleLeaderScannedTx - 更新claim record失败 txId=%s: %v", l.prefix, txId, err)
			return err
		}
	} else {
		// 链上执行成功
		if err := dbTx.Table(model.TableNameStakeLeaderRewardClaim).
			Where("tx_id = ?", txId).
			Updates(map[string]interface{}{
				"tx_state":   constants.TxStateSuccess,
				"updated_at": time.Now(),
			}).Error; err != nil {
			log.Errorf("%s HandleLeaderScannedTx - 更新claim record状态失败 txId=%s: %v", l.prefix, txId, err)
			return err
		}
		if err := dbTx.Table(model.TableNameStakeLeaderReward).
			Where("tx_id = ?", txId).
			Updates(map[string]interface{}{
				"reward_state": constants.RewardStateClaimed,
				"pending":      false,
				"updated_at":   time.Now(),
			}).Error; err != nil {
			log.Errorf("%s HandleLeaderScannedTx - 更新leader reward状态失败 txId=%s: %v", l.prefix, txId, err)
			return err
		}
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s HandleLeaderScannedTx - 提交事务失败 txId=%s: %v", l.prefix, txId, err)
		return err
	}
	return nil
}

// HandleLeaderExpiredTx 处理区域经理领取奖励交易超时的 Kafka 消息
func (l *StakeRewardLogic) HandleLeaderExpiredTx(msg entity.NewExpiredTx) error {
	txId := msg.TxID

	mutex := l.rs.NewMutex("stake:leader-reward:process-tx:" + txId)
	if err := mutex.Lock(); err != nil {
		return err
	}
	defer mutex.Unlock()

	dbTx := l.db.Begin()
	defer dbTx.Rollback()

	// 查询记录并加行锁，防重复处理
	var leaderReward model.StakeLeaderReward
	if err := dbTx.Table(model.TableNameStakeLeaderReward).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&leaderReward).Error; err != nil {
		log.Errorf("%s HandleLeaderExpiredTx - 查找记录失败 txId=%s: %v", l.prefix, txId, err)
		return err
	}

	// 只有待领取状态才处理
	if leaderReward.RewardState != int32(constants.RewardStateInit) {
		log.Infof("%s HandleLeaderExpiredTx - 交易 %s 奖励状态已为 %d，跳过重复处理", l.prefix, txId, leaderReward.RewardState)
		return nil
	}

	if err := dbTx.Table(model.TableNameStakeLeaderRewardClaim).
		Where("tx_id = ?", txId).
		Updates(map[string]interface{}{
			"tx_state":   constants.TxStateFailed,
			"updated_at": time.Now(),
		}).Error; err != nil {
		log.Errorf("%s HandleLeaderExpiredTx - 更新claim record失败 txId=%s: %v", l.prefix, txId, err)
		return err
	}
	if err := dbTx.Table(model.TableNameStakeLeaderReward).
		Where("tx_id = ?", txId).
		Updates(map[string]interface{}{
			"tx_id":        nil,
			"reward_state": constants.RewardStateInit,
			"pending":      false,
			"updated_at":   time.Now(),
		}).Error; err != nil {
		log.Errorf("%s HandleLeaderExpiredTx - 重置leader reward状态失败 txId=%s: %v", l.prefix, txId, err)
		return err
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s HandleLeaderExpiredTx - 提交事务失败 txId=%s: %v", l.prefix, txId, err)
		return err
	}
	return nil
}

