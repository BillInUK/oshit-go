package model

import "time"

const TableNameRpcEndpoint = "t_rpc_endpoint"

type RpcEndpoint struct {
	RecordID    string    `gorm:"column:record_id;not null;default:gen_ulid()" json:"recordId"`
	Scope       string    `gorm:"column:scope;not null" json:"scope"`             // 'env' or 'mainnet'
	Provider    string    `gorm:"column:provider;not null" json:"provider"`       // 'quicknode', 'helius', etc.
	Endpoint    string    `gorm:"column:endpoint;not null" json:"endpoint"`       // base URL without api key
	APIKey      string    `gorm:"column:api_key;default:''" json:"apiKey"`
	WssEndpoint string    `gorm:"column:wss_endpoint;default:''" json:"wssEndpoint"`
	WssAPIKey   string    `gorm:"column:wss_api_key;default:''" json:"wssApiKey"`
	Weight      int       `gorm:"column:weight;not null;default:1" json:"weight"` // 0 = disabled
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (*RpcEndpoint) TableName() string {
	return TableNameRpcEndpoint
}
