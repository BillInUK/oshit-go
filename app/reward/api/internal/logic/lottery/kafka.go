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

		// 记录流水
		if err := l.recordFundFlow(dbTx, txId, &reward, &msg.DecodedTx); err != nil {
			log.Errorf("%s 记录流水失败: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return err
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

	return nil
}

// recordFundFlow 记录抽奖的资金流水（在 dbTx 事务内）
// 使用 decodedTx 的 TransferInstructions[0] 作为 SOL cost，TransferCheckedInstructions[0] 作为 token reward
func (l *LotteryLogic) recordFundFlow(dbTx *gorm.DB, txId string, reward *model.LotteryReward, decodedTx *entity.DecodedSolanaTransaction) error {
	var fundFlows []model.FundFlow

	// 1. SOL 成本入账流水
	if len(decodedTx.TransferInstructions) > 0 {
		solInst := decodedTx.TransferInstructions[0]
		log.Infof("记录抽奖流水 - 记录SOL成本入账 from %v to %v amount %v",
			solInst.FromNativeAccount, solInst.ToNativeAccount, solInst.Amount)
		fundFlows = append(fundFlows, model.FundFlow{
			IsToken:     false,
			FromAccount: solInst.FromNativeAccount.String(),
			ToAccount:   solInst.ToNativeAccount.String(),
			TxID:        txId,
			Direction:   constants.FlowInput.String(),
			ServiceType: constants.SubServiceLottery.String(),
			FlowType:    constants.FlowCost.String(),
			Decimals:    9,
			Amount:      float64(solInst.Amount),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	// 2. Token 奖励出账流水
	if len(decodedTx.TransferCheckedInstructions) > 0 {
		tokenInst := decodedTx.TransferCheckedInstructions[0]
		log.Infof("记录抽奖流水 - 记录token奖励出账 from %v to %v amount %v",
			tokenInst.FromNativeAccount, tokenInst.ToNativeAccount, tokenInst.Amount)
		fundFlows = append(fundFlows, model.FundFlow{
			IsToken:     true,
			FromAccount: tokenInst.FromNativeAccount.String(),
			ToAccount:   tokenInst.ToNativeAccount.String(),
			TxID:        txId,
			Direction:   constants.FlowOutput.String(),
			ServiceType: constants.SubServiceLottery.String(),
			FlowType:    constants.FlowReceipt.String(),
			Decimals:    int16(tokenInst.Decimals),
			Amount:      float64(tokenInst.Amount),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if len(fundFlows) == 0 {
		return nil
	}

	batchSize := 100
	if err := dbTx.Table(model.TableNameFundFlow).Omit("record_id").Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tx_id"},
			{Name: "to_account"},
			{Name: "flow_type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"updated_at": time.Now(),
		}),
	}).CreateInBatches(&fundFlows, batchSize).Error; err != nil {
		log.Errorf("Lottery - 记录流水错误: %v", err)
		return err
	}

	return nil
}
