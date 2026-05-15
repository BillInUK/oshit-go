package rewardcode

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

// HandleScannedTx 处理 base 模块推送的 RewardCode 链上已确认交易
func (l *RewardCodeLogic) HandleScannedTx(msg entity.NewScannedTx) error {
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

	// 根据 tx_id 找到奖励码记录，SELECT FOR UPDATE 防止并发重复处理
	var rc model.RewardCode
	if err := dbTx.Table(model.TableNameRewardCode).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).
		First(&rc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			log.Warnf("%s 找不到对应的奖励码记录", prefix)
			return err
		}
		log.Errorf("%s 查询奖励码记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if rc.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, rc.TxState)
		dbTx.Rollback()
		return nil
	}

	var newState constants.TxState
	if msg.TxSig.Err != nil {
		newState = constants.TxStateFailed
		log.Infof("%s 交易失败，设置 tx_state=-1", prefix)
	} else {
		newState = constants.TxStateSuccess
		log.Infof("%s 交易成功，设置 tx_state=1", prefix)
	}

	if err := dbTx.Table(model.TableNameRewardCode).
		Where("record_id = ?", rc.RecordID).
		Update("tx_state", newState).Error; err != nil {
		log.Errorf("%s 更新奖励码状态错误: %v", prefix, err)
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

// HandleExpiredTx 处理 base 模块推送的 RewardCode 已超时交易
func (l *RewardCodeLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
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

	// 根据 tx_id 找到奖励码记录，SELECT FOR UPDATE 防止并发重复处理
	var rc model.RewardCode
	if err := dbTx.Table(model.TableNameRewardCode).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).
		First(&rc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查询奖励码记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if rc.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, rc.TxState)
		dbTx.Rollback()
		return nil
	}

	if err := dbTx.Table(model.TableNameRewardCode).
		Where("record_id = ?", rc.RecordID).
		Update("tx_state", constants.TxStateFailed).Error; err != nil {
		log.Errorf("%s 更新奖励码状态为失败错误: %v", prefix, err)
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
