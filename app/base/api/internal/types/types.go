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
	InviteCode string `json:"invite_code"`
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
	CreatedAt string `json:"created_at"`
}

type GetFeeToleranceRsp struct {
	MaxLessRate float64 `json:"max_less_rate"`
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

type FindInviteRelationByAccountReq struct {
	NativeAccount string `json:"native_account"`
}

type RecursiveQueryReq struct {
	Depth         int    `json:"depth"`
	NativeAccount string `json:"native_account"`
}

type InviteRelation struct {
	RecordID             string `json:"record_id"`
	Inviter string `json:"inviter"`
	Invitee string `json:"invitee"`
	Channel              string `json:"channel"`
	Level                int32  `json:"level"`
	TxID                 string `json:"tx_id"`
	CreatedAt            string `json:"created_at"`
}

type RecursiveQueryRsp struct {
	Records []InviteRelation `json:"records"`
}

// Tx Types
type SendTransactionReq struct {
	EncodedTx  string `json:"encoded_tx"`
	Service    string `json:"service"`
	SubService string `json:"sub_service"`
}

type SendTransactionRsp struct {
	RecordID string `json:"record_id"`
	TxID     string `json:"tx_id"`
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
