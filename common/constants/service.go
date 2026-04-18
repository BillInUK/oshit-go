package constants

// ServiceName is a typed service name enum (slug is unexported, preventing forgery).
type ServiceName struct {
	slug string
}

func (s ServiceName) String() string { return s.slug }

var (
	ServiceReward   = ServiceName{"Reward"}
	ServicePos      = ServiceName{"Pos"}
	ServiceCampaign = ServiceName{"Campaign"}
)

// SubServiceName is a typed sub-service name enum.
type SubServiceName struct {
	slug string
}

func (s SubServiceName) String() string { return s.slug }

var (
	SubServiceTakeToken         = SubServiceName{"TakeToken"}
	SubServiceGiveToken         = SubServiceName{"GiveToken"}
	SubServiceLottery           = SubServiceName{"Lottery"}
	SubServiceRewardCode        = SubServiceName{"RewardCode"}
	SubServicePosReward         = SubServiceName{"PosReward"}
	SubServiceStakeToken        = SubServiceName{"StakeToken"}
	SubServiceStakeReward       = SubServiceName{"StakeReward"}
	SubServiceStakeLeaderReward = SubServiceName{"StakeLeaderReward"}
	SubServiceMarketBuyToken    = SubServiceName{"MarketBuyToken"}
	SubServiceCampaignQuote     = SubServiceName{"CampaignQuote"}
	SubServiceExchangeToken     = SubServiceName{"ExchangeToken"}
)
