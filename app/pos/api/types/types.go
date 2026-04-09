package types

import "github.com/gagliardetto/solana-go"

const (
	PosFixedIncome  = 0 // Pos 持币固定收益
	PosTwitterLike  = 1 // Pos 关注twitter奖励
	PosRetweet      = 2 // Pos 推特转贴奖励
	PosTwitterReply = 3 // Pos 推特点赞或回复奖励：每次点赞或回复奖励 0.25

	StakeStateInit    = 0
	StakeStateSuccess = 1
	StakeStateFailed  = -1

	StakeFixed          = 0 // 质押每日固定利息
	StakeInvite         = 1 // 质押邀请奖励
	StakeStarIndividual = 2 // 质押激励奖励 -  个人奖励
	StakeStar           = 3 // 质押激励奖励 - 星级奖励
	StakeStarGroup      = 4 // 质押激励奖励 - 团队奖励
	StakeAreaLeader     = 5 // 质押激励奖励 - 区域领导奖励

	AreaLeaderRewardTypeDirect      = 0 // 区域经理奖励 - 直接区域经理（无Leader），10%
	AreaLeaderRewardTypeLevel1      = 1 // 区域经理奖励 - level=1 区域经理，7%
	AreaLeaderRewardTypeLeader      = 2 // 区域经理奖励 - level=1 的上级 Leader，3%
	AreaLeaderRewardTypeTotalLeader = 3 // 区域经理奖励 - 总区域经理
)

// StakeInstructionData 质押指令数据结构
type StakeInstructionData struct {
	Discriminator []byte // 8字节
	Amount        uint64 // 质押数量
	StakeType     uint8  // 质押类型
}

// StakeAccounts 质押指令相关的账户
type StakeAccounts struct {
	Staker           solana.PublicKey // 质押者
	Deployer         solana.PublicKey // 部署者
	ConfigAccount    solana.PublicKey // 配置账户
	StakeInfoAccount solana.PublicKey // 质押信息账户
	StakeAccount     solana.PublicKey // 质押账户
	UserTokenAccount solana.PublicKey // 用户代币账户
	MintAccount      solana.PublicKey // Mint账户
	ProgramID        solana.PublicKey // 程序ID
}

// ParsedStakeTx 解析后的质押交易
type ParsedStakeTx struct {
	TransactionSignature solana.Signature
	InstructionData      StakeInstructionData
	Accounts             StakeAccounts
	Signatures           []solana.Signature
	Success              bool
	RawData              []byte // 原始指令数据
}

type GetByTxIdReq struct {
	TxId string `json:"txId"`
}

type PosGroupInfo struct {
	Inviter         string
	StarLevel       int32
	GroupHoldAmount float64
	GroupFixReward  float64
}

type StakeSnapShotDetail struct {
	NativeAccount string
	SnapBase      float64
	SnapTotal     float64
}

type InviteNode struct {
	Inviter   string  `gorm:"column:inviter"`
	Invitee   string  `gorm:"column:invitee"`
	Level     int     `gorm:"column:level"`
	Amount    float64 `gorm:"column:amount"`
	StarLevel int     `gorm:"column:star_level"`
	Rate      float64 `gorm:"column:rate"`
	GroupId   string  `gorm:"column:group_id"`
	Base      float64 `gorm:"column:base"`
}

type StakeRewardStat struct {
	StarLevel           int32   `json:"starLevel"`           // 星级
	StakeAmount         float64 `json:"stakeAmount"`         // 个人质押量
	AccumulatedInterest float64 `json:"accumulatedInterest"` // 质押累计固定利息
	InviteReward        float64 `json:"inviteReward"`        // 质押邀请奖励
	StarReward          float64 `json:"starReward"`          // 质押激励奖励星级部分
	GroupReward         float64 `json:"groupReward"`         // 质押激励奖励团队部分
	TotalReward         float64 `json:"totalReward"`         // 质押当日总奖励
	GroupTotalStake     float64 `json:"groupTotalStake"`     // 团队总质押量
}

type EncodedTxReq struct {
	EncodedTx string `json:"encodedTx"`
}

type CommitStakeRewardTxReq struct {
	EncodedTx string `json:"encodedTx"`
}

type ClaimStakeRewardTxInfo struct {
	RewardAccount string  `json:"rewardAccount"`
	Mint          string  `json:"mint"`
	Decimals      int32   `json:"decimals"`
	TotalReward   float64 `json:"totalReward"`
	QuoteAmount   float64 `json:"quoteAmount"`
	CostAccount   string  `json:"costAccount"`
	CostFeeRate   int32   `json:"costFeeRate"`
	CostFee       float64 `json:"costFee"`
}
