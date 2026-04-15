package give

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"runtime/debug"
	"time"
)

// HandleScannedTx 处理 base 模块推送的 GiveToken 链上已确认交易
func (l *GiveTokenLogic) HandleScannedTx(msg entity.NewScannedTx) error {
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

	var giveTokenRecord model.GiveTokenRecord
	table := dbTx.Table(model.TableNameGiveTokenRecord)
	if err := table.Where("tx_id = ?", txId).First(&giveTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找业务记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	if msg.TxSig.Err != nil {
		if err := dbTx.Table(model.TableNameGiveTokenRecord).
			Where("record_id = ?", giveTokenRecord.RecordID).Update("tx_state", -1).Error; err != nil {
			log.Errorf("%s 更新转账交易状态为失败，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
	} else {
		if err := dbTx.Table(model.TableNameGiveTokenRecord).
			Where("record_id = ?", giveTokenRecord.RecordID).Update("tx_state", 1).Error; err != nil {
			log.Errorf("%s 更新转账交易状态为成功，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}

		decodedServiceTx, err := app_utils.DecodeServiceTransaction(&msg.DecodedTx)
		if err != nil {
			log.Errorf("%s 解码服务交易失败: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
		if err = l.recordFundFlow(decodedServiceTx); err != nil {
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

// HandleExpiredTx 处理 base 模块推送的 GiveToken 已超时交易
func (l *GiveTokenLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
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

	var giveTokenRecord model.GiveTokenRecord
	table := dbTx.Table(model.TableNameGiveTokenRecord)
	if err := table.Where("tx_id = ?", txId).First(&giveTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找转账记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	if err := dbTx.Model(&model.GiveTokenRecord{}).Where("tx_id = ?", txId).Update("tx_state", -1).Error; err != nil {
		log.Errorf("%s 更新转账交易状态为失败，错误: %v", prefix, err)
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

// recordFundFlow 记录 give token 的资金流水
func (l *GiveTokenLogic) recordFundFlow(decodedServiceTx *entity.DecodedServiceTransaction) error {
	var fundFlows []model.FundFlow
	table := l.db.Table(model.TableNameFundFlow)

	// 1. 记录 dex 入账 sol 流水
	log.Infof("记录官方转账流水 - 记录dex入账sol流水 from %v to %v", decodedServiceTx.FromNativeAccount, decodedServiceTx.ToDexInst.ToNativeAccount)
	fundFlows = append(fundFlows, model.FundFlow{
		IsToken:     false,
		FromAccount: decodedServiceTx.FromNativeAccount,
		ToAccount:   decodedServiceTx.ToDexInst.ToNativeAccount,
		TxID:        decodedServiceTx.TxID,
		Direction:   constants.FlowInput,
		ServiceType: constants.ServiceGiveToken,
		FlowType:    constants.FlowGiveTokenCost,
		Decimals:    9,
		Amount:      decodedServiceTx.ToDexInst.Amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	// 2. 记录平台奖励转账人出账流水
	log.Infof("记录官方转账流水 - 记录奖励转账人出账流水 from %v to %v amount %v",
		decodedServiceTx.RewardInst.FromNativeAccount, decodedServiceTx.RewardInst.ToNativeAccount,
		decodedServiceTx.RewardInst.Amount)
	fundFlows = append(fundFlows, model.FundFlow{
		IsToken:     true,
		FromAccount: decodedServiceTx.RewardInst.FromNativeAccount,
		ToAccount:   decodedServiceTx.RewardInst.ToNativeAccount,
		TxID:        decodedServiceTx.TxID,
		Direction:   constants.FlowOutput,
		ServiceType: constants.ServiceGiveToken,
		FlowType:    constants.FlowGiveTokenReceipt,
		Decimals:    int16(decodedServiceTx.RewardInst.Decimals),
		Amount:      decodedServiceTx.RewardInst.Amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	// 3. 记录平台奖励上级邀请人出账流水
	for _, inst := range decodedServiceTx.RewardInviterInst {
		log.Infof("记录官方转账流水 - 记录奖励邀请人出账流水 from %v to %v amount %v",
			inst.FromNativeAccount, inst.ToNativeAccount, inst.Amount)
		fundFlows = append(fundFlows, model.FundFlow{
			IsToken:     true,
			FromAccount: inst.FromNativeAccount,
			ToAccount:   inst.ToNativeAccount,
			TxID:        decodedServiceTx.TxID,
			Direction:   constants.FlowOutput,
			ServiceType: constants.ServiceGiveToken,
			FlowType:    constants.FlowGiveTokenInviter,
			Decimals:    int16(inst.Decimals),
			Amount:      inst.Amount,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	batchSize := 100
	if err := table.Omit("record_id").Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tx_id"},
			{Name: "to_account"},
			{Name: "flow_type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"updated_at": time.Now(),
		}),
	}).CreateInBatches(&fundFlows, batchSize).Error; err != nil {
		log.Errorf("GiveToken - 记录流水错误: %v", err)
		return err
	}

	return nil
}
