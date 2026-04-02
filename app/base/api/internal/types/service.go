package types

// ServiceRegisterReq 服务注册请求
type ServiceRegisterReq struct {
	Service       string `json:"service"`
	NativeAccount string `json:"native_account"`
	PdaAccount    string `json:"pda_account"`
	UntilTxID     string `json:"until_tx_id"`
	Slot          int64  `json:"slot"`
	Webhook       string `json:"webhook,omitempty"`
	MqGroup       string `json:"mq_group,omitempty"`
	MqTopic       string `json:"mq_topic,omitempty"`
	HookType      int32  `json:"hook_type"` // 0=Kafka, 1=Webhook
}

// ServiceRegisterRsp 服务注册响应
type ServiceRegisterRsp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ServiceUpdateReq 服务更新请求
type ServiceUpdateReq struct {
	Service       string `json:"service"`
	NativeAccount string `json:"native_account,omitempty"`
	PdaAccount    string `json:"pda_account,omitempty"`
	UntilTxID     string `json:"until_tx_id,omitempty"`
	Slot          int64  `json:"slot,omitempty"`
	Webhook       string `json:"webhook,omitempty"`
	MqGroup       string `json:"mq_group,omitempty"`
	MqTopic       string `json:"mq_topic,omitempty"`
	HookType      int32  `json:"hook_type,omitempty"`
}

// ServiceUpdateRsp 服务更新响应
type ServiceUpdateRsp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ServiceQueryReq 服务查询请求
type ServiceQueryReq struct {
	Service string `json:"service"`
}

// ServiceQueryRsp 服务查询响应
type ServiceQueryRsp struct {
	Service       string `json:"service"`
	NativeAccount string `json:"native_account"`
	PdaAccount    string `json:"pda_account"`
	UntilTxID     string `json:"until_tx_id"`
	Slot          int64  `json:"slot"`
	Webhook       string `json:"webhook,omitempty"`
	MqGroup       string `json:"mq_group,omitempty"`
	MqTopic       string `json:"mq_topic,omitempty"`
	HookType      int32  `json:"hook_type"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// KafkaMessage Kafka消息结构
type KafkaMessage struct {
	MsgType   string      `json:"msg_type"`
	Service   string      `json:"service"`
	TxID      string      `json:"tx_id"`
	TxSlot    uint64      `json:"tx_slot"`
	State     int32       `json:"state"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// WebhookMessage Webhook消息结构
type WebhookMessage struct {
	Service   string      `json:"service"`
	TxID      string      `json:"tx_id"`
	TxSlot    uint64      `json:"tx_slot"`
	State     int32       `json:"state"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}
