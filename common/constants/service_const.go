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

// FundFlowServiceType represents the business service type for fund flows.
type FundFlowServiceType struct {
	slug string
}

func (s FundFlowServiceType) String() string { return s.slug }

var (
	FundFlowServiceAirDrop   = FundFlowServiceType{"AirDrop"}
	FundFlowServiceTakeToken = FundFlowServiceType{"TakeToken"}
	FundFlowServiceGiveToken = FundFlowServiceType{"GiveToken"}
	FundFlowServicePosDaily  = FundFlowServiceType{"PosDaily"}
	FundFlowServiceStake     = FundFlowServiceType{"Stake"}
	FundFlowServiceLottery   = FundFlowServiceType{"Lottery"}
)

// FundFlowType represents the specific fund flow type within a service.
type FundFlowType struct {
	slug string
}

func (f FundFlowType) String() string { return f.slug }

var (
	// TakeToken flows
	FlowTakeTokenCost    = FundFlowType{"take_token_cost"}
	FlowTakeTokenReceipt = FundFlowType{"take_token_receipt"}
	FlowTakeTokenInviter = FundFlowType{"take_token_inviter"}

	// GiveToken flows
	FlowGiveTokenCost    = FundFlowType{"give_token_cost"}
	FlowGiveTokenReceipt = FundFlowType{"give_token_receipt"}
	FlowGiveTokenInviter = FundFlowType{"give_token_inviter"}

	// POS flows
	FlowPosCost    = FundFlowType{"pos_cost"}
	FlowPosReceipt = FundFlowType{"pos_receipt"}

	// Stake flows
	FlowStakeCost    = FundFlowType{"stake_cost"}
	FlowStakeReceipt = FundFlowType{"stake_receipt"}

	// Lottery flows
	FlowLotteryCost    = FundFlowType{"lottery_cost"}
	FlowLotteryReceipt = FundFlowType{"lottery_receipt"}
)
