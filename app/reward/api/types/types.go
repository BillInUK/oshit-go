package types

import "oshit-go/common/pkg/dal/model"

type RewardTokenItem struct {
	Index          int    `json:"index"`          // 奖励项索引，防止返回的回复里面json数组乱序，前端可以依据该字段进行重新排列奖励项
	ReceiptAccount string `json:"receiptAccount"` // 接收奖励的 solana 地址
	Amount         uint64 `json:"amount"`         // 奖励 token的金额
}

type GetByTxIdReq struct {
	TxId string `json:"txId"`
}

type GetGiveTokenTxInfoReq struct {
	To     string  `json:"to"`     // 接收 solana 地址
	Amount float64 `json:"amount"` // 用户转出 token的数量
}

type CommitGiveTokenTxInfoReq struct {
	EncodedTx string `json:"encodedTx"`
	To        string `json:"to"`
}

type GiveTokenTxInfo struct {
	RewardAccount string  `json:"rewardAccount"` // 服务端下发奖励的地址
	Mint          string  `json:"mint"`          // token 地址
	CostAccount   string  `json:"costAccount"`   // 接收成本费的地址
	CostFeeRate   float64 `json:"costFeeRate"`   // 成本费费率
	MaxCostFee    float64 `json:"maxCostFee"`    // 最大成本费
	Decimals      int32   `json:"decimals"`      // token金额的精度
	QuoteSOLPrice float64 `json:"quoteSOLPrice"` // token兑换solana的价格

	TotalReward     float64 `json:"totalReward"`     // 整笔交易下发的token奖励总额
	QuotedSOLAmount float64 `json:"quotedSOLAmount"` // token兑换solana的金额
	CostFee         uint64  `json:"costFee"`         // SOL 成本费，单位 lamports

	Claims            []model.LevelRatio `json:"claims"`            // 向上奖励邀请人的级别以及每个级别的奖励费率
	GiveInfo          RewardTokenItem    `json:"giveInfo"`          // give token 业务发起人 转给 to 地址的奖励项
	RewardInfo        RewardTokenItem    `json:"rewardInfo"`        // 奖励 give token 业务发起人的奖励项
	RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"` // 奖励 give token 业务发起人的邀请人的奖励项
}

type GetTakeTokenTxInfoReq struct {
	InviteCode     string  `json:"inviteCode"`     // 可选，邀请码
	Custom         bool    `json:"custom"`         // 是否使用自定义金额（暂未使用）
	CustomAmount   float64 `json:"customAmount"`   // 自定义金额（暂未使用）
	ReceiptAccount string  `json:"receiptAccount"` // 领取奖励的 native account
}

type CommitTakeTokenTxInfoReq struct {
	EncodedTx  string `json:"encodedTx"`
	InviteCode string `json:"inviteCode"`
}

type TakeTokenTxInfo struct {
	RewardAccount   string  `json:"rewardAccount"`   // 下发奖励的 solana 地址
	Mint            string  `json:"mint"`            // token 地址
	CostAccount     string  `json:"costAccount"`     // 接收solana成本费的地址
	CostFeeRate     float64 `json:"costFeeRate"`     // 成本费率
	MaxCostFee      float64 `json:"maxCostFee"`      // 最大成本费
	Decimals        int32   `json:"decimals"`        // token 金额的精度
	InviteCode      string  `json:"inviteCode"`      // 邀请码
	InviteCodeValid bool    `json:"inviteCodeValid"` // 邀请码是否有效
	Invited         bool    `json:"invited"`         // take token 后是否确定邀请关系
	QuoteSOLPrice   float64 `json:"quoteSOLPrice"`   // token 兑换solana的价格

	TotalReward     float64 `json:"totalReward"`     // 整笔交易奖励的token总额
	QuotedSOLAmount float64 `json:"quotedSOLAmount"` // token 兑换 solana 的金额
	CostFee         uint64  `json:"costFee"`         // SOL 成本费，单位 lamports

	Claims            []model.LevelRatio `json:"claims"`            // 向上奖励邀请人的级别以及每个级别的奖励费率
	RewardInfo        RewardTokenItem    `json:"rewardInfo"`        // 发起 take token 流程地址的奖励项
	RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"` // 发起 take token 流程地址的邀请人的奖励项
}

type GetLotteryStatusReq struct {
	NativeAccount string `json:"nativeAccount"`
}

type ExecuteLotteryReq struct {
	// native account comes from JWT
}

type GetUnclaimedLotteryReq struct {
	NativeAccount string `json:"nativeAccount"`
}

type GetLotteryTxInfoReq struct {
	RecordId string `json:"recordId"`
}

type CommitLotteryTxReq struct {
	EncodedTx string `json:"encodedTx"`
	RewardId  string `json:"rewardId"` // t_lottery_reward.record_id
}

type ClaimLotteryTxInfo struct {
	RecordId      string  `json:"recordId"`      // 抽奖的记录id
	RewardAccount string  `json:"rewardAccount"` // 发放抽奖记录的token
	Mint          string  `json:"mint"`          // token地址
	Decimals      int32   `json:"decimals"`      // 币种精度
	CostAccount   string  `json:"costAccount"`   // 成本费
	QuoteSOLPrice float64 `json:"quoteSOLPrice"` // 兑换solana的价格
	LotteryAmount float64 `json:"lotteryAmount"` // 奖励金额
	CostFee       float64 `json:"costFee"`       // 成本费
}

type CampaignQuoteTxInfoReq struct {
	Score uint64 `json:"score"`
}

type CampaignQuoteReq struct {
	EncodedTx string `json:"encodedTx"`
	XAcJwt    string `json:"x-ac-jwt"`
	Score     uint64 `json:"score"`
}

type CampaignQuoteTxInfo struct {
	RewardAccount string  `json:"rewardAccount"` // 发放抽奖记录的地址
	Mint          string  `json:"mint"`          // token地址
	Decimals      int32   `json:"decimals"`      // 币种精度
	CostAccount   string  `json:"costAccount"`   // 成本费
	TokenAmount   float64 `json:"tokenAmount"`   // 兑换出来token的金额
	CostFee       float64 `json:"costFee"`       // solana成本费
}

type GetRewardCodeTxInfoReq struct {
	RewardCode string `json:"rewardCode"`
}

type CommitRewardCodeTxReq struct {
	EncodedTx  string `json:"encodedTx"`
	RewardCode string `json:"rewardCode"`
}

type RewardCodeTxInfo struct {
	RewardAccount string  `json:"rewardAccount"` // 发放token的地址
	Mint          string  `json:"mint"`          // token地址
	Decimals      int32   `json:"decimals"`      // 币种精度
	CostAccount   string  `json:"costAccount"`   // 收取成本费的地址
	RewardAmount  float64 `json:"rewardAmount"`  // 奖励的token数量(raw)
	CostFee       float64 `json:"costFee"`       // 需要支付的SOL成本费(lamports)
}
