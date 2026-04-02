package give

import (
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/common/pkg/entity"
)

// HandleScannedTx 处理 base 模块推送的 GiveToken 链上已确认交易
func (l *GiveTokenLogic) HandleScannedTx(tx entity.NewScannedTx) error {
	// TODO: 根据链上交易信息更新 GiveToken 相关记录状态
	log.Infof("%s - 处理扫描到的交易: %v", l.prefix, tx.TxSig.Signature)
	return nil
}

// HandleExpiredTx 处理 base 模块推送的 GiveToken 已超时交易
func (l *GiveTokenLogic) HandleExpiredTx(tx entity.NewExpiredTx) error {
	// TODO: 将对应 GiveToken 记录标记为失败
	log.Infof("%s - 处理超时的交易: %v", l.prefix, tx.TxID)
	return nil
}
