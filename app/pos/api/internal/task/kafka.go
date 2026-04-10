package task

import (
	"context"
	"encoding/json"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"oshit-go/common/pkg/entity"
)

// KafkaConsumerTask 消费 base 模块发送的 Kafka 消息
type KafkaConsumerTask struct {
	reader          *kafka.Reader
	db              *gorm.DB
	redis           redis.UniversalClient
	rpcClient       *rpc.Client
	scannedHandlers map[string]ScannedTxHandler
	expiredHandlers map[string]ExpiredTxHandler
}

// NewKafkaConsumerTask 创建 Kafka 消费任务
func NewKafkaConsumerTask(taskCtx *TaskContext) *KafkaConsumerTask {
	reader, _ := taskCtx.KafkaConsumer.(*kafka.Reader)
	return &KafkaConsumerTask{
		reader:          reader,
		db:              taskCtx.DB,
		redis:           taskCtx.Redis,
		rpcClient:       taskCtx.RpcClient,
		scannedHandlers: taskCtx.ScannedHandlers,
		expiredHandlers: taskCtx.ExpiredHandlers,
	}
}

// Start 启动消费任务
func (t *KafkaConsumerTask) Start() {
	if t.reader == nil {
		log.Warn("KafkaConsumerTask: reader 未初始化，跳过启动")
		return
	}
	go t.consume()
}

// consume 持续读取并分发消息
func (t *KafkaConsumerTask) consume() {
	for {
		msg, err := t.reader.ReadMessage(context.Background())
		if err != nil {
			log.Errorf("KafkaConsumerTask: 读取消息失败: %v", err)
			continue
		}
		t.dispatch(msg)
	}
}

// dispatch 根据消息外层结构中的 MsgType 分发到对应处理器
func (t *KafkaConsumerTask) dispatch(msg kafka.Message) {
	// 先解析 MsgType
	var envelope struct {
		MsgType string `json:"MsgType"`
	}
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		log.Errorf("KafkaConsumerTask: 解析消息 MsgType 失败: %v", err)
		return
	}

	switch envelope.MsgType {
	case "NewScannedTransaction":
		var m entity.KafkaTxMsg
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Errorf("KafkaConsumerTask: 解析 NewScannedTransaction 失败: %v", err)
			return
		}
		t.handleScannedTx(m.MsgContent)

	case "NewExpiredTransaction":
		var m entity.KafkaExpiredTxMsg
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Errorf("KafkaConsumerTask: 解析 NewExpiredTransaction 失败: %v", err)
			return
		}
		t.handleExpiredTx(m.MsgContent)

	default:
		log.Warnf("KafkaConsumerTask: 未知消息类型: %s", envelope.MsgType)
	}
}

// handleScannedTx 根据 Service 将已确认交易路由到对应的业务处理器，内部再按 SubService 细分
func (t *KafkaConsumerTask) handleScannedTx(tx entity.NewScannedTx) {
	log.Infof("KafkaConsumerTask: 收到已扫描交易 service=%s subService=%s txID=%s", tx.Service, tx.SubService, tx.TxSig.Signature)
	if t.scannedHandlers == nil {
		log.Warnf("KafkaConsumerTask: scannedHandlers 未注册，跳过处理 service=%s", tx.Service)
		return
	}
	handler, ok := t.scannedHandlers[tx.Service]
	if !ok {
		log.Warnf("KafkaConsumerTask: 未找到 service=%s 的已扫描交易处理器", tx.Service)
		return
	}
	if err := handler(context.Background(), tx); err != nil {
		log.Errorf("KafkaConsumerTask: 处理已扫描交易失败 service=%s subService=%s txID=%s: %v", tx.Service, tx.SubService, tx.TxSig.Signature, err)
	}
}

// handleExpiredTx 根据 Service 将超时交易路由到对应的业务处理器，内部再按 SubService 细分
func (t *KafkaConsumerTask) handleExpiredTx(tx entity.NewExpiredTx) {
	log.Infof("KafkaConsumerTask: 收到超时交易 service=%s subService=%s txID=%s", tx.Service, tx.SubService, tx.TxID)
	if t.expiredHandlers == nil {
		log.Warnf("KafkaConsumerTask: expiredHandlers 未注册，跳过处理 service=%s", tx.Service)
		return
	}
	handler, ok := t.expiredHandlers[tx.Service]
	if !ok {
		log.Warnf("KafkaConsumerTask: 未找到 service=%s 的超时交易处理器", tx.Service)
		return
	}
	if err := handler(context.Background(), tx); err != nil {
		log.Errorf("KafkaConsumerTask: 处理超时交易失败 service=%s subService=%s txID=%s: %v", tx.Service, tx.SubService, tx.TxID, err)
	}
}

// SnapShotConsumerTask 消费 PosTopic 和 StakeTopic 的快照消息
type SnapShotConsumerTask struct {
	reader           *kafka.Reader
	snapShotHandlers map[string]SnapShotHandler
}

// NewSnapShotConsumerTask 创建快照消费任务
func NewSnapShotConsumerTask(taskCtx *TaskContext) *SnapShotConsumerTask {
	reader, _ := taskCtx.SnapShotKafkaConsumer.(*kafka.Reader)
	return &SnapShotConsumerTask{
		reader:           reader,
		snapShotHandlers: taskCtx.SnapShotHandlers,
	}
}

// Start 启动快照消费任务
func (t *SnapShotConsumerTask) Start() {
	if t.reader == nil {
		log.Warn("SnapShotConsumerTask: reader 未初始化，跳过启动")
		return
	}
	go t.consume()
}

// consume 持续读取并分发快照消息
func (t *SnapShotConsumerTask) consume() {
	for {
		msg, err := t.reader.ReadMessage(context.Background())
		if err != nil {
			log.Errorf("SnapShotConsumerTask: 读取消息失败: %v", err)
			continue
		}
		t.dispatch(msg)
	}
}

// dispatch 根据 MsgType 分发到对应的快照处理器
func (t *SnapShotConsumerTask) dispatch(msg kafka.Message) {
	var envelope struct {
		MsgType string `json:"MsgType"`
	}
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		log.Errorf("SnapShotConsumerTask: 解析消息 MsgType 失败: %v", err)
		return
	}

	if envelope.MsgType != "NewPosSnapShot" && envelope.MsgType != "NewStakeSnapShot" {
		log.Warnf("SnapShotConsumerTask: 未知消息类型: %s", envelope.MsgType)
		return
	}

	var m entity.KafkaNewSnapShotMsg
	if err := json.Unmarshal(msg.Value, &m); err != nil {
		log.Errorf("SnapShotConsumerTask: 解析 %s 失败: %v", envelope.MsgType, err)
		return
	}

	if t.snapShotHandlers == nil {
		log.Warnf("SnapShotConsumerTask: snapShotHandlers 未注册，跳过 %s", envelope.MsgType)
		return
	}
	handler, ok := t.snapShotHandlers["Pos"]
	if !ok {
		log.Warnf("SnapShotConsumerTask: 未找到 Pos 快照处理器")
		return
	}
	if err := handler(context.Background(), m); err != nil {
		log.Errorf("SnapShotConsumerTask: 处理 %s 失败: %v", envelope.MsgType, err)
	}
}
