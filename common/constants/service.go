package constants

// ServiceName is a typed service name enum (slug is unexported, preventing forgery).
type ServiceName struct {
	slug string
}

func (s ServiceName) String() string { return s.slug }

var (
	ServiceReward = ServiceName{"reward"}
	ServicePos    = ServiceName{"pos"}
)

// SubServiceName is a typed sub-service name enum.
type SubServiceName struct {
	slug string
}

func (s SubServiceName) String() string { return s.slug }

var (
	SubServiceTakeToken         = SubServiceName{"take token"}
	SubServiceGiveToken         = SubServiceName{"give token"}
	SubServiceLottery           = SubServiceName{"lottery"}
	SubServiceRewardCode        = SubServiceName{"reward code"}
	SubServicePosReward         = SubServiceName{"pos reward"}
	SubServiceStakeToken        = SubServiceName{"stake token"}
	SubServiceStakeReward       = SubServiceName{"stake reward"}
	SubServiceStakeLeaderReward = SubServiceName{"stake leader reward"}
	SubServiceMarketBuyToken    = SubServiceName{"market buy token"}
	SubServiceCampaignQuote     = SubServiceName{"campaign quote"}
)
