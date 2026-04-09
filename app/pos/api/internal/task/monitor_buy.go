package task

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"math"
)

// MonitorBuyTask 质押快照服务
type MonitorBuyTask struct {
	db          *gorm.DB
	rpcClient   *rpc.Client
	kafkaWriter *kafka.Writer
	programID   solana.PublicKey
	mint        solana.PublicKey
	dec         uint64
}

// NewMonitorBuyTask 创建交易所购买token检测任务
func NewMonitorBuyTask(taskCtx *TaskContext) *MonitorBuyTask {
	var kafkaWriter *kafka.Writer
	if taskCtx.KafkaProducer != nil {
		if w, ok := taskCtx.KafkaProducer.(*kafka.Writer); ok {
			kafkaWriter = w
		}
	}
	dec := math.Pow10(int(taskCtx.TokenConfig.Decimals))
	return &MonitorBuyTask{
		db:          taskCtx.DB,
		rpcClient:   taskCtx.RpcClient,
		kafkaWriter: kafkaWriter,
		programID:   solana.MustPublicKeyFromBase58(taskCtx.RewardConfig.ProgramID),
		mint:        solana.MustPublicKeyFromBase58(taskCtx.TokenConfig.Mint),
		dec:         uint64(dec),
	}
}

func (t *MonitorBuyTask) Start() {

}
