package give

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
	if err := dbTx.Table(model.TableNameGiveTokenRecord).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&giveTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找业务记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if giveTokenRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, giveTokenRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	if msg.TxSig.Err != nil {
		if err := dbTx.Table(model.TableNameGiveTokenRecord).
			Where("record_id = ?", giveTokenRecord.RecordID).Update("tx_state", constants.TxStateFailed).Error; err != nil {
			log.Errorf("%s 更新转账交易状态为失败，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
	} else {
		if err := dbTx.Table(model.TableNameGiveTokenRecord).
			Where("record_id = ?", giveTokenRecord.RecordID).Update("tx_state", constants.TxStateSuccess).Error; err != nil {
			log.Errorf("%s 更新转账交易状态为成功，错误: %v", prefix, err)
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
	if err := dbTx.Table(model.TableNameGiveTokenRecord).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&giveTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找转账记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if giveTokenRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, giveTokenRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	if err := dbTx.Model(&model.GiveTokenRecord{}).Where("tx_id = ?", txId).Update("tx_state", constants.TxStateFailed).Error; err != nil {
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

