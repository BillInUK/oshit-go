package campaign

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
	"strconv"
	"time"
)

func (l *CampaignLogic) HandleScannedTx(msg entity.NewScannedTx) {
	txId := msg.TxSig.Signature.String()
	prefix := fmt.Sprintf("%s 处理扫描到的交易 %s -", l.prefix, txId)

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			// 打印堆栈信息
			stack := debug.Stack()
			// 打印错误日志和堆栈信息
			log.Errorf("%s - 处理Kafka消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			// 回滚事务
			dbTx.Rollback()
		}
	}()
	var err error
	var exchangeRecord model.CampaignQuoteRecord

	// 根据交易 Id 找到记录，SELECT FOR UPDATE 防止并发重复处理
	if err = dbTx.Table(model.TableNameCampaignQuoteRecord).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).
		First(&exchangeRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return
		}
		log.Errorf("%s - 处理Kafka消息，根据交易 Id %s 查找领取记录错误: %v", prefix, txId, err)
		dbTx.Rollback()
		return
	}

	// 防重复处理：只有 QuoteStateInit 状态才需要处理
	if exchangeRecord.QuoteState != int32(constants.QuoteStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, exchangeRecord.QuoteState)
		dbTx.Rollback()
		return
	}

	// 从记录里取下单时的场次和用户SGT自然日，保证跨日回调时恢复到正确的记录
	globalQuotaDate := exchangeRecord.CreatedAt.UTC().Truncate(24 * time.Hour)
	recordSession := exchangeRecord.Session
	userQuotaDate := exchangeRecord.UserQuotaDate

	if msg.TxSig.Err != nil {
		if err = dbTx.Table(model.TableNameCampaignQuoteRecord).
			Where("tx_id = ?", txId).
			Update("quote_state", constants.QuoteStateFailed).Error; err != nil {
			log.Errorf("%s - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为失败，错误: %v", prefix, txId, err)
			dbTx.Rollback()
			return
		}
		// 解冻积分
		userId, _ := strconv.ParseUint(exchangeRecord.UserID, 10, 64)
		flowId := uint64(exchangeRecord.ScoreFlowID)
		if _, err := l.srvCtx.CampaignClientV1.UnfreezeScore(userId, txId, flowId, ExchangeSys, ExchangeBiz, ReasonUnFreeze); err != nil {
			dbTx.Rollback()
			log.Errorf("%s - 处理Kafka消息，交易 Id %s 失败，解冻积分错误: %v", prefix, txId, err)
			return
		}
		// 解冻用户兑换额度（使用下单时的SGT自然日定位用户额度记录）
		if err = dbTx.Table(model.TableNameUserDailyQuota).
			Where("user_id = ? AND quota_date = ?", exchangeRecord.UserID, userQuotaDate).
			Updates(map[string]interface{}{
				"frozen_quota":    gorm.Expr("frozen_quota - ?", exchangeRecord.Amount),
				"available_quota": gorm.Expr("available_quota + ?", exchangeRecord.Amount),
				"updated_at":      time.Now(),
			}).Error; err != nil {
			dbTx.Rollback()
			log.Errorf("%s - 处理Kafka消息，交易 Id %s 失败，解冻用户兑换额度错误: %v", prefix, txId, err)
			return
		}
		// 恢复全局兑换额度（使用下单时的 quota_date + session 定位对应场次记录）
		if err = dbTx.Table(model.TableNameCampaignQuoteLimit).
			Where("quota_date = ? AND session = ?", globalQuotaDate, recordSession).
			Updates(map[string]interface{}{
				"daily_limit": gorm.Expr("daily_limit + ?", exchangeRecord.Amount),
				"updated_at":  time.Now(),
			}).Error; err != nil {
			dbTx.Rollback()
			log.Errorf("%s - 处理Kafka消息，交易 Id %s 失败，恢复全局兑换额度错误: %v", prefix, txId, err)
			return
		}
	} else {
		if err = dbTx.Table(model.TableNameCampaignQuoteRecord).
			Where("tx_id = ?", txId).
			Update("quote_state", constants.QuoteStateSuccess).Error; err != nil {
			log.Errorf("%s - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为成功，错误: %v", prefix, txId, err)
			dbTx.Rollback()
			return
		}
		// 扣减积分
		userId, _ := strconv.ParseUint(exchangeRecord.UserID, 10, 64)
		flowId := exchangeRecord.ScoreFlowID
		if _, err := l.srvCtx.CampaignClientV1.ConsumeFrozenScore(userId, txId, uint64(flowId), ExchangeSys, ExchangeBiz, ReasonConsumeFreeze); err != nil {
			dbTx.Rollback()
			log.Errorf("%s - 处理Kafka消息，交易 Id %s 成功，消耗解冻积分错误: %v", prefix, txId, err)
			return
		}
		// 扣除用户冻结额度（使用下单时的SGT自然日定位用户额度记录）
		if err = dbTx.Table(model.TableNameUserDailyQuota).
			Where("user_id = ? AND quota_date = ?", exchangeRecord.UserID, userQuotaDate).
			Updates(map[string]interface{}{
				"frozen_quota": gorm.Expr("frozen_quota - ?", exchangeRecord.Amount),
				"updated_at":   time.Now(),
			}).Error; err != nil {
			dbTx.Rollback()
			log.Errorf("%s - 处理Kafka消息，交易 Id %s 成功，扣除用户冻结额度错误: %v", prefix, txId, err)
			return
		}
		log.Infof("%s - 更新记录为成功，交易ID %s，社交媒体 %v，用户Id %v", prefix, txId, exchangeRecord.Provider, exchangeRecord.UserID)
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s - 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return
	}
}

func (l *CampaignLogic) HandleExpiredTx(msg entity.NewExpiredTx) {
	txId := msg.TxID
	prefix := fmt.Sprintf("%s 处理超时交易 %s -", l.prefix, txId)

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			// 打印堆栈信息
			stack := debug.Stack()
			// 打印错误日志和堆栈信息
			log.Errorf("%s - 处理Kafka消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			// 回滚事务
			dbTx.Rollback()
		}
	}()
	var err error
	var exchangeRecord model.CampaignQuoteRecord

	// 根据交易 Id 找到记录，SELECT FOR UPDATE 防止并发重复处理
	if err = dbTx.Table(model.TableNameCampaignQuoteRecord).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tx_id = ?", txId).
		First(&exchangeRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return
		}
		log.Errorf("%s - 处理Kafka消息，根据交易 Id %s 查找领取记录错误: %v", prefix, txId, err)
		dbTx.Rollback()
		return
	}

	// 防重复处理：只有 QuoteStateInit 状态才需要处理
	if exchangeRecord.QuoteState != int32(constants.QuoteStateInit) {
		log.Infof("%s 交易状态已为 %d，跳过重复处理", prefix, exchangeRecord.QuoteState)
		dbTx.Rollback()
		return
	}

	if err = dbTx.Table(model.TableNameCampaignQuoteRecord).
		Where("tx_id = ?", txId).
		Update("quote_state", constants.QuoteStateFailed).Error; err != nil {
		log.Errorf("%s - 处理Kafka消息，根据交易 Id %s 更新领取交易状态为失败，错误: %v", prefix, txId, err)
		dbTx.Rollback()
		return
	}

	// 解冻积分
	userId, _ := strconv.ParseUint(exchangeRecord.UserID, 10, 64)
	flowId := uint64(exchangeRecord.ScoreFlowID)
	if _, err := l.srvCtx.CampaignClientV1.UnfreezeScore(userId, txId, flowId, ExchangeSys, ExchangeBiz, ReasonUnFreeze); err != nil {
		dbTx.Rollback()
		log.Errorf("%s - 处理Kafka消息，交易 Id %s 超时，解冻积分错误: %v", prefix, txId, err)
		return
	}

	// 解冻用户兑换额度（使用下单时的SGT自然日定位用户额度记录）
	globalQuotaDate := exchangeRecord.CreatedAt.UTC().Truncate(24 * time.Hour)
	if err = dbTx.Table(model.TableNameUserDailyQuota).
		Where("user_id = ? AND quota_date = ?", exchangeRecord.UserID, exchangeRecord.UserQuotaDate).
		Updates(map[string]interface{}{
			"frozen_quota":    gorm.Expr("frozen_quota - ?", exchangeRecord.Amount),
			"available_quota": gorm.Expr("available_quota + ?", exchangeRecord.Amount),
			"updated_at":      time.Now(),
		}).Error; err != nil {
		dbTx.Rollback()
		log.Errorf("%s - 处理Kafka消息，交易 Id %s 超时，解冻用户兑换额度错误: %v", prefix, txId, err)
		return
	}

	// 恢复全局兑换额度（使用下单时的 quota_date + session 定位对应场次记录）
	if err = dbTx.Table(model.TableNameCampaignQuoteLimit).
		Where("quota_date = ? AND session = ?", globalQuotaDate, exchangeRecord.Session).
		Updates(map[string]interface{}{
			"daily_limit": gorm.Expr("daily_limit + ?", exchangeRecord.Amount),
			"updated_at":  time.Now(),
		}).Error; err != nil {
		dbTx.Rollback()
		log.Errorf("%s - 处理Kafka消息，交易 Id %s 超时，恢复全局兑换额度错误: %v", prefix, txId, err)
		return
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s - 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return
	}
}

// recordExchangeRecord 记录官网领取token记录
func (l *CampaignLogic) recordExchangeRecord(decodedServiceTx *entity.DecodedServiceTransaction, userId uint64, opResult *entity.ScoreOperationResult, tokenAmount float64, userQuota *model.UserDailyQuota, globalLimit *model.CampaignQuoteLimit) error {
	// 开启事务
	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("兑换Campaign积分为token - 开启事务错误: %v", dbTx.Error)
		return dbTx.Error
	}

	// 记录领取奖励记录
	score := float64(opResult.OperationResult.Log.DeltaScore)
	scoreTxId := opResult.OperationResult.Log.TransactionID
	scoreFlowId := int32(opResult.OperationResult.Log.LogID)
	amount := decodedServiceTx.TransferTokenInst.Amount
	exchangeTokenRecord := model.CampaignQuoteRecord{
		RewardAccount:  decodedServiceTx.RewardInst.FromNativeAccount,
		ReceiptAccount: decodedServiceTx.RewardInst.ToNativeAccount,
		Provider:       "campaign",
		UserID:         strconv.FormatUint(userId, 10),
		ScoreFlowID:    scoreFlowId,
		ScoreTxID:      scoreTxId,
		TxID:           decodedServiceTx.TxID,
		Amount:         amount,
		Score:          score,
		QuoteState:     int32(constants.QuoteStateInit),
		Session:        globalLimit.Session,    // 记录下单时的全局场次，Kafka回调时恢复对应session额度
		UserQuotaDate:  userQuota.QuotaDate,    // 记录下单时的SGT自然日，Kafka回调时恢复用户额度
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := dbTx.Table(model.TableNameCampaignQuoteRecord).Create(&exchangeTokenRecord).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("兑换Campaign积分为token - 插入领取记录表错误: %v", err)
		return err
	}

	// 冻结用户兑换额度
	userQuota.FrozenQuota += tokenAmount
	userQuota.AvailableQuota -= tokenAmount
	userQuota.UpdatedAt = time.Now()
	if err := dbTx.Table(model.TableNameUserDailyQuota).Where("id = ?", userQuota.ID).Updates(map[string]interface{}{
		"frozen_quota":    userQuota.FrozenQuota,
		"available_quota": userQuota.AvailableQuota,
		"updated_at":      userQuota.UpdatedAt,
	}).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("兑换Campaign积分为token - 更新用户冻结额度错误: %v", err)
		return err
	}

	// 更新全局兑换额度（按 id 精确定位本场次记录）
	globalLimit.DailyLimit -= tokenAmount
	globalLimit.UpdatedAt = time.Now()
	if err := dbTx.Table(model.TableNameCampaignQuoteLimit).Where("id = ?", globalLimit.ID).Updates(map[string]interface{}{
		"daily_limit": globalLimit.DailyLimit,
		"updated_at":  globalLimit.UpdatedAt,
	}).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("兑换Campaign积分为token - 更新全局兑换额度错误: %v", err)
		return err
	}


	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("兑换Campaign积分为token - 提交事务错误: %v", err)
		return err
	}

	return nil
}
