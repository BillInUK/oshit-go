package stake

import "oshit-go/common/pkg/entity"

func (l *StakeRewardLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	return nil
}

func (l *StakeRewardLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	return nil
}
