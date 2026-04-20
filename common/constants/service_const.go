package constants

const (
	TxFetchInit    = 0  // 交易获取状态
	TxFetchFailed  = -1 //交易获取状态 - 失败
	TxFetchSuccess = 1  // 交易获取状态 - 成功
)

// TxState represents transaction confirmation state.
type TxState int32

const (
	TxStateExpired TxState = -2
	TxStateFailed  TxState = -1
	TxStateInit    TxState = 0
	TxStateSuccess TxState = 1
)

// RewardState represents reward claim state.
type RewardState int32

const (
	RewardStateExpired RewardState = -2
	RewardStateFailed  RewardState = -1
	RewardStateInit    RewardState = 0
	RewardStateClaimed RewardState = 1
)

// QuoteState represents campaign quote state.
type QuoteState int32

const (
	QuoteStateFailed  QuoteState = -1
	QuoteStateInit    QuoteState = 0
	QuoteStateSuccess QuoteState = 1
)

// FlowDirection represents fund flow direction (input or output).
type FlowDirection struct {
	slug string
}

func (d FlowDirection) String() string { return d.slug }

var (
	FlowInput  = FlowDirection{"input"}
	FlowOutput = FlowDirection{"output"}
)

// FundFlowType represents the specific fund flow type within a service.
type FundFlowType struct {
	slug string
}

func (f FundFlowType) String() string { return f.slug }

var (
	FlowCost    = FundFlowType{"cost"}
	FlowReceipt = FundFlowType{"receipt"}
	FlowInviter = FundFlowType{"inviter"}
)
