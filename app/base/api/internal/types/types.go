package types

import (
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
)

// Auth Types
type LoginReq struct {
	Brand      string `json:"brand"`
	Symbol     string `json:"symbol"`
	Account    string `json:"account"`
	Sign       string `json:"sign"`
	Nonce      uint64 `json:"nonce"`
	InviteCode string `json:"invite_code"`
}

type LoginRsp struct {
	Token utils.Tokens `json:"token"`
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
	CreatedAt string `json:"created_at"`
}

type GetFeeToleranceRsp struct {
	MinFee uint64 `json:"min_fee"`
	MaxFee uint64 `json:"max_fee"`
}

type PriorityFeeRsp struct {
	PerComputeUnit entity.FeeDetail `json:"per_compute_unit"`
	PerTransaction entity.FeeDetail `json:"per_transaction"`
}

type ComputeUnitConsumedRsp struct {
	MiniRent          uint64 `json:"mini_rent"`
	AssociatedAccount uint64 `json:"associated_account"`
	TransferChecked   uint64 `json:"transfer_checked"`
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
	InviteCode string `json:"invite_code"`
}

type GetAccountByInviteCodeRsp struct {
	RecordID      string `json:"record_id"`
	NativeAccount string `json:"native_account"`
	TokenAccount  string `json:"token_account"`
	InviteCode    string `json:"invite_code"`
	CreatedAt     string `json:"created_at"`
}

type CheckInviteRecordReq struct {
	NativeAccount string `json:"native_account"`
}

type CheckInviteRecordRsp struct {
	Exists bool `json:"exists"`
}

type RecursiveQueryReq struct {
	Depth         int    `json:"depth"`
	NativeAccount string `json:"native_account"`
}

type InviteRelation struct {
	RecordID  string `json:"record_id"`
	Inviter   string `json:"inviter"`
	Invitee   string `json:"invitee"`
	Channel   string `json:"channel"`
	Level     int32  `json:"level"`
	TxID      string `json:"tx_id"`
	CreatedAt string `json:"created_at"`
}

type RecursiveQueryRsp struct {
	Records []InviteRelation `json:"records"`
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
	HoldersNumber int `json:"holders_number"`
}
