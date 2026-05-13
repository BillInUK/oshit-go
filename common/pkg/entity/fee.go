package entity

type Context struct {
	Slot uint64 `json:"slot"`
}

type FeeDetail struct {
	Extreme uint64 `json:"extreme"`
	High    uint64 `json:"high"`
	Low     uint64 `json:"low"`
	Medium  uint64 `json:"medium"`
}

type QnSOLPriorityFee struct {
	Context        Context   `json:"context"`
	PerComputeUnit FeeDetail `json:"perComputeUnit"`
	PerTransaction FeeDetail `json:"perTransaction"`
}

type SOLWeightAvgFee struct {
	PriceGroup       uint64  `json:"price_group"`
	WeightedAvgPrice float64 `json:"weighted_avg_price"`
}

type ComputeUnitDetailResp struct {
	AssociatedAccount string `json:"associatedAccount"`
	MiniRent          string `json:"miniRent"`
	TransferChecked   string `json:"transferChecked"`
}

type ComputeUnitDetail struct {
	AssociatedAccount uint64 `json:"associatedAccount"`
	MiniRent          uint64 `json:"miniRent"`
	TransferChecked   uint64 `json:"transferChecked"`
	Memo              uint64 `json:"memo"`
}

type PriorityFee struct {
	PerComputeUnit FeeDetail `json:"perComputeUnit"`
	PerTransaction FeeDetail `json:"perTransaction"`
}

//type PriorityFeeData struct {
//	ID      int               `json:"id"`
//	JSONRPC string            `json:"jsonrpc"`
//	Result  PriorityFeeResult `json:"result"`
//}
//
//type PriorityFeeResponse struct {
//	Code  int             `json:"code"`
//	Count int             `json:"count"`
//	Data  PriorityFeeData `json:"data"`
//	Msg   string          `json:"msg"`
//}

type RaydiumPriceRes struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Version string `json:"version"`
	Data    struct {
		SwapType             string  `json:"swapType"`
		InputMint            string  `json:"inputMint"`
		InputAmount          string  `json:"inputAmount"`
		OutputMint           string  `json:"outputMint"`
		OutputAmount         string  `json:"outputAmount"`
		OtherAmountThreshold string  `json:"otherAmountThreshold"`
		SlippageBps          int     `json:"slippageBps"`
		PriceImpactPct       float64 `json:"priceImpactPct"`
		ReferrerAmount       string  `json:"referrerAmount"`
		RoutePlan            []struct {
			PoolID            string   `json:"poolId"`
			InputMint         string   `json:"inputMint"`
			OutputMint        string   `json:"outputMint"`
			FeeMint           string   `json:"feeMint"`
			FeeRate           int      `json:"feeRate"`
			FeeAmount         string   `json:"feeAmount"`
			RemainingAccounts []string `json:"remainingAccounts"`
			LastPoolPriceX64  string   `json:"lastPoolPriceX64,omitempty"`
		} `json:"routePlan"`
	} `json:"data"`
}
