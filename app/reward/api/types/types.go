package types

import "oshit-go/common/pkg/dal/model"

type RewardTokenItem struct {
	Index          int    `json:"index"`
	ReceiptAccount string `json:"receiptAccount"`
	Amount         uint64 `json:"amount"`
}

type GetByTxIdReq struct {
	TxId string `json:"txId"`
}

type GetGiveTokenTxInfoReq struct {
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type CommitGiveTokenTxInfoReq struct {
	EncodedTx string `json:"encodedTx"`
	To        string `json:"to"`
}

type GiveTokenTxInfo struct {
	RewardAccount string  `json:"rewardAccount"`
	Mint          string  `json:"mint"`
	CostAccount   string  `json:"costAccount"`
	DexFeeRate    float64 `json:"dexFeeRate"`
	MaxDexFee     float64 `json:"maxDexFee"`
	Decimals      int32   `json:"decimals"`
	QuoteSOLPrice float64 `json:"quoteSOLPrice"`

	TotalReward       float64            `json:"totalReward"`
	QuotedSOLAmount   float64            `json:"quotedSOLAmount"`
	Claims            []model.LevelRatio `json:"claims"`
	GiveInfo          RewardTokenItem    `json:"giveInfo"`
	RewardInfo        RewardTokenItem    `json:"rewardInfo"`
	RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"`
}

type GetTakeTokenTxInfoReq struct {
	InviteCode     string  `json:"inviteCode"`
	Custom         bool    `json:"custom"`
	CustomAmount   float64 `json:"customAmount"`
	ReceiptAccount string  `json:"receiptAccount"`
}

type CommitTakeTokenTxInfoReq struct {
	EncodedTx  string `json:"encodedTx"`
	InviteCode string `json:"inviteCode"`
}

type TakeTokenTxInfo struct {
	RewardAccount   string  `json:"rewardAccount"`
	Mint            string  `json:"mint"`
	CostAccount     string  `json:"costAccount"`
	DexFeeRate      float64 `json:"dexFeeRate"`
	MaxDexFee       float64 `json:"maxDexFee"`
	Decimals        int32   `json:"decimals"`
	InviteCode      string  `json:"inviteCode"`
	InviteCodeValid bool    `json:"inviteCodeValid"`
	Invited         bool    `json:"invited"`
	QuoteSOLPrice   float64 `json:"quoteSOLPrice"`

	TotalReward       float64            `json:"totalReward"`
	QuotedSOLAmount   float64            `json:"quotedSOLAmount"`
	Claims            []model.LevelRatio `json:"claims"`
	RewardInfo        RewardTokenItem    `json:"rewardInfo"`
	RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"`
}

type ClaimLotteryTxInfo struct {
	RecordId      string  `json:"recordId"`      // 抽奖的记录id
	RewardAccount string  `json:"rewardAccount"` // 发放抽奖记录的token
	Mint          string  `json:"mint"`          // token地址
	CostAccount   string  `json:"costAccount"`   // 成本费
	Decimals      int32   `json:"decimals"`      // 币种精度
	LotteryAmount float64 `json:"lotteryAmount"` // 奖励金额
	CostFee       float64 `json:"CostFee"`       // 成本费
}
