package types

import "oshit-go/common/pkg/dal/model"

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
	RewardNativeAccount string  `json:"rewardNativeAccount"`
	RewardTokenAccount  string  `json:"rewardTokenAccount"`
	TokenMintAccount    string  `json:"tokenMintAccount"`
	DexNativeAccount    string  `json:"dexAccount"`
	DexFeeRate          float64 `json:"dexFeeRate"`
	MaxDexFee           float64 `json:"maxDexFee"`
	Decimals            int32   `json:"decimals"`
	QuoteSOLPrice       float64 `json:"quoteSOLPrice"`

	TotalRewardAmount float64            `json:"totalRewardAmount"`
	QuotedSOLAmount   float64            `json:"quotedSOLAmount"`
	Claims            []model.LevelRatio `json:"claims"`
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
	RewardNativeAccount string  `json:"rewardNativeAccount"`
	RewardTokenAccount  string  `json:"rewardTokenAccount"`
	TokenMintAccount    string  `json:"tokenMintAccount"`
	DexNativeAccount    string  `json:"dexAccount"`
	DexFeeRate          float64 `json:"dexFeeRate"`
	MaxDexFee           float64 `json:"maxDexFee"`
	Decimals            int32   `json:"decimals"`
	InviteCode          string  `json:"inviteCode"`
	InviteCodeValid     bool    `json:"inviteCodeValid"`
	InviteDetermine     bool    `json:"inviteDetermine"`
	QuoteSOLPrice       float64 `json:"quoteSOLPrice"`

	TotalRewardAmount float64            `json:"totalRewardAmount"`
	QuotedSOLAmount   float64            `json:"quotedSOLAmount"`
	Claims            []model.LevelRatio `json:"claims"`
	RewardInfo        RewardTokenItem    `json:"rewardInfo"`
	RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"`
}

type RewardTokenItem struct {
	Index         int    `json:"index"`
	NativeAccount string `json:"nativeAccount"`
	TokenAccount  string `json:"tokenAccount"`
	Amount        uint64 `json:"amount"`
}
