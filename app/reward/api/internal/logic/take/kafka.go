package take

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"runtime/debug"
	"strconv"
	"time"
)

// updateDailyClaimStats 在 dbTx 事务内更新每日领取统计，并在 take_count 达到阈值时设置 need_lottery=true
func (l *TakeTokenLogic) updateDailyClaimStats(dbTx *gorm.DB, nativeAccount string, takeAmountUI float64) error {
	today := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	// 使用 raw SQL upsert，RETURNING take_count 获取更新后的值
	type result struct {
		TakeCount int32 `gorm:"column:take_count"`
	}
	var res result
	rawSQL := `
		INSERT INTO t_daily_claim_stats (native_account, take_date, take_count, need_lottery, last_take_time, total_take)
		VALUES (?, ?, 1, false, ?, ?)
		ON CONFLICT (native_account, take_date)
		DO UPDATE SET take_count = t_daily_claim_stats.take_count + 1, last_take_time = ?, total_take = t_daily_claim_stats.total_take + ?
		RETURNING take_count
	`
	if err := dbTx.Raw(rawSQL, nativeAccount, today, now, takeAmountUI, now, takeAmountUI).Scan(&res).Error; err != nil {
		log.Errorf("TakeToken - 更新每日领取统计错误: %v", err)
		return err
	}

	// 如果 take_count 达到阈值，设置 need_lottery=true
	if res.TakeCount == 5 || res.TakeCount == 10 || res.TakeCount == 20 {
		if err := dbTx.Table(model.TableNameDailyClaimStats).
			Where("native_account = ? AND take_date = ?", nativeAccount, today).
			Updates(map[string]interface{}{"need_lottery": true, "updated_at": now}).Error; err != nil {
			log.Errorf("TakeToken - 设置 need_lottery=true 错误: %v", err)
			return err
		}
		log.Infof("TakeToken - 地址 %v 的 take_count 达到 %d，设置 need_lottery=true", nativeAccount, res.TakeCount)
	}

	return nil
}

// HandleScannedTx 处理 base 模块推送的 TakeToken 链上已确认交易
func (l *TakeTokenLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	txId := msg.TxSig.Signature.String()
	prefix := fmt.Sprintf("%s 处理扫描到的交易 %s -", l.prefix, txId)

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Errorf("%s 处理消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			dbTx.Rollback()
		}
	}()
	var err error
	var takeTokenRecord model.TakeTokenRecord

	// 根据交易 Id 找到记录，SELECT FOR UPDATE 防止并发重复处理
	if err = dbTx.Table(model.TableNameTakeTokenRecord).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&takeTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找业务记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if takeTokenRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, takeTokenRecord.TxState)
		dbTx.Rollback()
		// 无论如何都清理 Redis 锁，防止 90s 超时写 Failed 后锁残留
		l.rd.Del(context.Background(), "take-token:process:commit-tx:"+takeTokenRecord.ReceiptAccount)
		return nil
	}

	if msg.TxSig.Err != nil {
		if err = dbTx.Table(model.TableNameTakeTokenRecord).
			Where("record_id = ?", takeTokenRecord.RecordID).Update("tx_state", constants.TxStateFailed).Error; err != nil {
			log.Errorf("%s 更新领取交易状态为失败，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
	} else {
		if err = dbTx.Table(model.TableNameTakeTokenRecord).
			Where("record_id = ?", takeTokenRecord.RecordID).Update("tx_state", constants.TxStateSuccess).Error; err != nil {
			log.Errorf("%s 更新领取交易状态为成功，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
		log.Infof("%s 更新记录为成功，邀请码 %v，确定邀请关系 %v", prefix, takeTokenRecord.InviteCode, takeTokenRecord.Invited)

		// 更新每日领取统计（存 raw 金额，前端 /1000 显示）
		if err = l.updateDailyClaimStats(dbTx, takeTokenRecord.ReceiptAccount, takeTokenRecord.Amount); err != nil {
			log.Errorf("%s 更新每日领取统计错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}

		// 如果需要确定邀请层级关系
		if takeTokenRecord.Invited {
			inviterInfo, err := l.inviteLogic.GetAccountByInviteCode(takeTokenRecord.InviteCode)
			if err != nil {
				log.Errorf("%s 确定邀请关系错误，无法根据邀请码找到邀请人", prefix)
				dbTx.Rollback()
				return err
			}
			if inviterInfo != nil {
				_, err = l.inviteLogic.RecordDetermineInvitationHierarchy(
					inviterInfo.NativeAccount,
					takeTokenRecord.ReceiptAccount,
					takeTokenRecord.TxID,
					"InviteCode",
				)
				if err != nil {
					log.Errorf("%s 记录邀请层级关系错误: %v", prefix, err)
					dbTx.Rollback()
					return err
				}
			}
		}
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("官网领取奖励 - 提交事务失败: %v", err)
		dbTx.Rollback()
		return err
	}

	// 清理 ProcessCommitTx 持有的分布式锁（服务重启导致 defer 未执行时锁会残留长达 1 小时）
	l.rd.Del(context.Background(), "take-token:process:commit-tx:"+takeTokenRecord.ReceiptAccount)

	// 通过 Redis Pub/Sub 通知 ProcessCommitTx（支持多实例部署）
	finalState := constants.TxStateSuccess
	if msg.TxSig.Err != nil {
		finalState = constants.TxStateFailed
	}
	if err := l.rd.Publish(context.Background(), "tx:notify:"+txId, strconv.Itoa(int(finalState))).Err(); err != nil {
		log.Warnf("%s Redis 通知发布失败: %v", prefix, err)
	}

	return nil
}

// HandleExpiredTx 处理 base 模块推送的 TakeToken 已超时交易
func (l *TakeTokenLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	txId := msg.TxID
	prefix := fmt.Sprintf("%s 处理超时交易 %s -", l.prefix, txId)

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Errorf("%s 处理消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			dbTx.Rollback()
		}
	}()
	var err error
	var takeTokenRecord model.TakeTokenRecord

	// 根据交易 Id 找到记录，SELECT FOR UPDATE 防止并发重复处理
	if err = dbTx.Table(model.TableNameTakeTokenRecord).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).First(&takeTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找领取记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 防重复处理：只有 TxStateInit 状态才需要处理
	if takeTokenRecord.TxState != int32(constants.TxStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, takeTokenRecord.TxState)
		dbTx.Rollback()
		return nil
	}

	if err = dbTx.Model(&model.TakeTokenRecord{}).Where("tx_id = ?", txId).Update("tx_state", constants.TxStateFailed).Error; err != nil {
		log.Errorf("%s 更新领取交易状态为失败，错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}
	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	// 通过 Redis Pub/Sub 通知 ProcessCommitTx（交易已过期）
	if err := l.rd.Publish(context.Background(), "tx:notify:"+txId, strconv.Itoa(int(constants.TxStateExpired))).Err(); err != nil {
		log.Warnf("%s Redis 通知发布失败: %v", prefix, err)
	}

	return nil
}

