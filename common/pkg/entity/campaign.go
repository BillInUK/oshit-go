package entity

type ScoreData struct {
	TotalAvailable uint64 `json:"totalAvailable"`
	FrozenScore    uint64 `json:"frozenScore"`
	LastModify     uint64 `json:"lastModify"`
}

// ScoreRequest 积分操作请求
type ScoreRequest struct {
	UserID       uint64 `json:"userId"`
	TxID         string `json:"txId"`
	FlowID       uint64 `json:"flowId"`
	SourceSys    string `json:"sourceSys"`
	BusinessName string `json:"businessName"`
	Reason       string `json:"reason"`
	Score        uint64 `json:"score"`
}

type FreezeRequest struct {
	UserID       uint64 `json:"userId"`
	TxID         string `json:"txId"`
	SourceSys    string `json:"sourceSys"`
	BusinessName string `json:"businessName"`
	Reason       string `json:"reason"`
	Score        uint64 `json:"score"`
}

type ConsumeFrozenRequest struct {
	UserID       uint64 `json:"userId"`
	TxID         string `json:"txId"`
	FlowID       uint64 `json:"flowId"`
	SourceSys    string `json:"sourceSys"`
	BusinessName string `json:"businessName"`
	Reason       string `json:"reason"`
}

// 操作积分日志
type ScoreLog struct {
	LogID         uint64 `json:"logId"`
	UserID        uint64 `json:"userId"`
	OpType        string `json:"opType"`
	DeltaScore    uint64 `json:"deltaScore"`
	PreTotal      uint64 `json:"preTotal"`
	PostTotal     uint64 `json:"postTotal"`
	PreFrozen     uint64 `json:"preFrozen"`
	PostFrozen    uint64 `json:"postFrozen"`
	TransactionID string `json:"transactionId"`
	CreatedAt     string `json:"createdAt"`
	BusinessName  string `json:"businessName"`
	Reason        string `json:"reason"`
	SourceSys     string `json:"sourceSys"`
}

// 操作结果通用结构
type ScoreOperationResult struct {
	OperationResult struct {
		TxID string   `json:"txId"`
		Log  ScoreLog `json:"log"`
	} `json:"operationResult"`
}
