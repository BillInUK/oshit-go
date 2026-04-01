package constants

const (
	TxFetchStateInit    = 0  // 交易获取状态
	TxFetchStateFailed  = -1 //交易获取状态 - 失败
	TxFetchStateSuccess = 1  // 交易获取状态 - 成功

	ServiceAirDrop            = 0 // 业务类型 - 空投
	ServiceOfficialGive       = 1 // 业务类型 - 官方领取奖励
	ServiceUnOfficialTransfer = 2 // 业务类型 - 非官方转账
	ServiceOfficialTransfer   = 3 // 业务类型 - 官方转账
	ServiceSwapNewToken       = 4 // 业务类型 - 兑换旧Token到新Token
	ServicePosDaily           = 5 // 业务类型 - POS每日奖励
	ServiceGame               = 6 // 业务类型 - Game
	ServiceStake              = 7 // 业务类型 - Stake

	FlowInput  = 0 // 流水方向 - 入账
	FlowOutput = 1 // 流水方向 - 出账

	FlowAirDropGasFee = 0 // 流水类型 - 出账 - 空投token时的gas费
	FlowAirDropToken  = 1 // 流水类型 - 出账 - 空投token出账

	FlowOfficialGiveDexFee             = 0 // 流水类型 - 入账 - 官网领取奖励时转给dex的sol
	FlowOfficialGiveTokenReward        = 1 // 流水类型 - 出账 - 官网领取奖励token出账给领取人
	FlowOfficialGiveTokenRewardInviter = 2 // 流水类型 - 出账 - 官网领取奖励token出账给领取人上级邀请人

	FlowUnOfficialTransferGasFee     = 0 // 流水类型 - 出账 - 发放非官方奖励时的gas费
	FlowUnOfficialTokenReward        = 1 // 流水类型 - 出账 - 发放非官方奖励给转账人的token
	FlowUnOfficialTokenRewardInviter = 2 // 流水类型 - 出装 - 发放非官方奖励给转账人上级邀请人的token

	FlowOfficialTransferDexFee     = 0 // 流水类型 - 入账 - 官方转账时的转给dex的sol
	FlowOfficialTokenReward        = 1 // 流水类型 - 出账 - 官方转账时奖励转账人的token
	FlowOfficialTokenRewardInviter = 2 // 流水类型 - 出账 - 官方转账时奖励转账人上级邀请人的的token

	FlowSwapNewToken = 0 // 流水类型 - 出账 - swap旧token到新token的转账

	FlowPosDexFee        = 0 // 流水类型 - 入账 - POS领取奖励时转给dex的sol
	FlowPosRewardToken   = 1 // 流水类型 - 出账 - POS领取奖励token出账给领取人
	FlowStakeRewardToken = 2 // 流水类型 - 出账 - POS领取奖励token出账给领取人
)
