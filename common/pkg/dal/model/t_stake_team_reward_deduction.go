package model

import "time"

const TableNameStakeTeamRewardDeduction = "t_stake_team_reward_deduction"

// StakeTeamRewardDeduction mapped from table <t_stake_team_reward_deduction>
type StakeTeamRewardDeduction struct {
	NativeAccount string    `gorm:"column:native_account;primaryKey" json:"native_account"`
	Total         float64   `gorm:"column:total;not null" json:"total"`
	Deducted      float64   `gorm:"column:deducted;not null;default:0" json:"deducted"`
	Remaining     float64   `gorm:"column:remaining;not null;default:0" json:"remaining"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (*StakeTeamRewardDeduction) TableName() string {
	return TableNameStakeTeamRewardDeduction
}
