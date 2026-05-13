package types

import (
	"oshit-go/common/pkg/entity"
)

// Auth Types
type LoginReq struct {
	Brand      string `json:"brand"`
	Symbol     string `json:"symbol"`
	Account    string `json:"account"`
	Sign       string `json:"sign"`
	Nonce      uint64 `json:"nonce"`
	InviteCode string `json:"inviteCode"`
}

type LoginRsp struct {
	Token string `json:"token"`
}

// Config Types
type GetTokenInfoReq struct {
	Brand  string `json:"brand"`
	Symbol string `json:"symbol"`
}

type GetTokenInfoRsp struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Decimals  int32  `json:"decimal"`
	Mint      string `json:"mint"`
	CreatedAt string `json:"createdAt"`
}

type GetFeeToleranceRsp struct {
	MaxLessRate float64 `json:"maxLessRate"`
}

type PriorityFeeRsp struct {
	PerComputeUnit entity.FeeDetail `json:"perComputeUnit"`
	PerTransaction entity.FeeDetail `json:"perTransaction"`
}

type ComputeUnitConsumedRsp struct {
	MiniRent          uint64 `json:"miniRent"`
	AssociatedAccount uint64 `json:"associatedAccount"`
	TransferChecked   uint64 `json:"transferChecked"`
	Memo              uint64 `json:"memo"`
}

// Price Types
type GetBirdEyePriceReq struct {
	Interval string `json:"interval"` // 1D, 1W, 1M
}

type PriceQuoteRsp struct {
	Price float64 `json:"price"`
}

type BirdEyePriceRsp struct {
	Data interface{} `json:"data"`
}

// Invite Types
type GetAccountByInviteCodeReq struct {
	InviteCode string `json:"inviteCode"`
}

type GetAccountByInviteCodeRsp struct {
	RecordID      string `json:"recordId"`
	NativeAccount string `json:"nativeAccount"`
	TokenAccount  string `json:"tokenAccount"`
	InviteCode    string `json:"inviteCode"`
	CreatedAt     string `json:"createdAt"`
}

type CheckInviteRecordReq struct {
	NativeAccount string `json:"nativeAccount"`
}

type CheckInviteRecordRsp struct {
	Exists bool `json:"exists"`
}

type FindInviteRelationByAccountReq struct {
	NativeAccount string `json:"nativeAccount"`
}

type RecursiveQueryReq struct {
	Depth         int    `json:"depth"`
	NativeAccount string `json:"nativeAccount"`
}

type InviteRelation struct {
	RecordID string `json:"recordId"`
	Inviter  string `json:"inviter"`
	Invitee  string `json:"invitee"`
	Channel  string `json:"channel"`
	Level    int32  `json:"level"`
	TxID     string `json:"txId"`
	CreatedAt string `json:"createdAt"`
}

type RecursiveQueryRsp struct {
	Records []InviteRelation `json:"records"`
}

// Tx Types
type SendTransactionReq struct {
	EncodedTx  string `json:"encodedTx"`
	Service    string `json:"service"`
	SubService string `json:"subService"`
}

type SendTransactionRsp struct {
	RecordID string `json:"recordId"`
	TxID     string `json:"txId"`
}

type RewardDistributionRsp struct {
	Level int32 `json:"level"`
	// 其他字段根据实际表结构添加
}

type RewardClaimRsp struct {
	Level int32   `json:"level"`
	Rate  float64 `json:"rate"`
	// 其他字段根据实际表结构添加
}

type TokenHoldersRsp struct {
	HoldersNumber int `json:"holdersNumber"`
}
