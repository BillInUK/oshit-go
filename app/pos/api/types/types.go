package types

const (
	StakeFixed          = 0 // 质押每日固定利息
	StakeInvite         = 1 // 质押邀请奖励
	StakeStarIndividual = 2 // 质押激励奖励 -  个人奖励
	StakeStar           = 3 // 质押激励奖励 - 星级奖励
	StakeStarGroup      = 4 // 质押激励奖励 - 团队奖励
	StakeAreaLeader     = 5 // 质押激励奖励 - 区域领导奖励
)

type GetByTxIdReq struct {
	TxId string `json:"txId"`
}

type StakeSnapShotDetail struct {
	NativeAccount string
	SnapBase      float64
	SnapTotal     float64
}

type StakeInviteNode struct {
	Inviter   string  `gorm:"column:inviter"`
	Invitee   string  `gorm:"column:invitee"`
	Level     int     `gorm:"column:level"`
	Amount    float64 `gorm:"column:amount"`
	StarLevel int     `gorm:"column:star_level"`
	Rate      float64 `gorm:"column:rate"`
	GroupId   string  `gorm:"column:group_id"`
	Base      float64 `gorm:"column:base"`
}

type StakeRewardStat struct {
	StarLevel           int32   `json:"starLevel"`           // 星级
	StakeAmount         float64 `json:"stakeAmount"`         // 个人质押量
	AccumulatedInterest float64 `json:"accumulatedInterest"` // 质押累计固定利息
	InviteReward        float64 `json:"inviteReward"`        // 质押邀请奖励
	StarReward          float64 `json:"starReward"`          // 质押激励奖励星级部分
	GroupReward         float64 `json:"groupReward"`         // 质押激励奖励团队部分
	TotalReward         float64 `json:"totalReward"`         // 质押当日总奖励
	GroupTotalStake     float64 `json:"groupTotalStake"`     // 团队总质押量
}

type CommitStakeRewardTxReq struct {
	EncodedTx string `json:"encodedTx"`
}

type ClaimStakeRewardTxInfo struct {
	RewardAccount string  `json:"rewardAccount"`
	Mint          string  `json:"mint"`
	Decimals      int32   `json:"decimals"`
	TotalReward   float64 `json:"totalReward"`
	QuoteAmount   float64 `json:"quoteAmount"`
	CostAccount   string  `json:"costAccount"`
	CostFeeRate   int32   `json:"costFeeRate"`
	CostFee       float64 `json:"costFee"`
}
