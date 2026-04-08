package pos

import "oshit-go/common/pkg/entity"

// HandleScannedTx 处理 base 模块推送的 TakeToken 链上已确认交易
func (l *PosRewardLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	return nil
}

// HandleExpiredTx 处理 base 模块推送的 TakeToken 链上已确认交易
func (l *PosRewardLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	return nil
}
