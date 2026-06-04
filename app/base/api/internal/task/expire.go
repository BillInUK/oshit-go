package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"math/rand"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"time"
)

// 分布式锁 key
const txExpireScanLock = "base:sol:tx-expire:scan:lock"

// TxExpireTask 业务交易扫描任务
type TxExpireTask struct {
	db          *gorm.DB
	redis       redis.UniversalClient
	redSync     redsync.Redsync
	rpcClient   *rpc.Client
	kafkaWriter *kafka.Writer
	rpcURL      string
	prefix      string
}

// NewTxExpireTask 创建交易超时任务
func NewTxExpireTask(taskCtx *TaskContext) *TxExpireTask {
	var rpcClient *rpc.Client
	var rpcURL string
	if taskCtx.RpcPool != nil {
		rpcClient = taskCtx.RpcPool.First()
		rpcURL = taskCtx.RpcPool.FirstURL()
	} else if taskCtx.RpcClient != nil {
		rpcClient = taskCtx.RpcClient
	}

	var kafkaWriter *kafka.Writer
	if taskCtx.KafkaProducer != nil {
		if w, ok := taskCtx.KafkaProducer.(*kafka.Writer); ok {
			kafkaWriter = w
		}
	}

	return &TxExpireTask{
		db:          taskCtx.DB,
		redis:       taskCtx.Redis,
		redSync:     taskCtx.RedSync,
		rpcClient:   rpcClient,
		rpcURL:      rpcURL,
		kafkaWriter: kafkaWriter,
		prefix:      "扫描过期交易",
	}
}

// Start 启动定时扫描任务
func (t *TxExpireTask) Start() {
	go runPeriodic(&t.redSync, 5*time.Second, txExpireScanLock, 2*time.Minute, t.scanExpiredTransactions)
}

// scanExpiredTransactions 扫描过期交易
// TODO: 该用xxl-job实现分页扫描
func (t *TxExpireTask) scanExpiredTransactions() {
	//log.Infof("%s - 任务开始", t.prefix)
	ctx := context.Background()

	// 计算5分钟前的时间
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	var expiredTxs []model.ServiceTx
	if err := t.db.WithContext(ctx).
		Where(`tx_state = ? AND created_at <= ?`, constants.TxStateInit, fiveMinutesAgo).
		Find(&expiredTxs).Error; err != nil {
		log.Errorf("%s 查询过期的交易错误: %v", t.prefix, err)
		return
	}

	if len(expiredTxs) == 0 {
		return
	}

	log.Infof("%s 找到 %d 条超时记录需要处理", t.prefix, len(expiredTxs))

	for _, record := range expiredTxs {
		t.processTransaction(ctx, record)
	}
}

// processTransaction 处理单个交易
func (t *TxExpireTask) processTransaction(ctx context.Context, posTxRecord model.ServiceTx) {
	// 如果交易ID非法则直接失败
	txSig, err := solana.SignatureFromBase58(posTxRecord.TxID)
	if err != nil {
		log.Errorf("%s 交易[%s]失败 - 非法的交易ID格式: %v", t.prefix, posTxRecord.TxID, err)
		t.markTxExpired(ctx, posTxRecord)
		return
	}

	// 限制rpc的用量
	for {
		if app_utils.AcquireDistributedRateLimit(t.redis, "handleTokenTx", 200) {
			break
		}
		// 超过 300/s，sleep 一小段时间（随机 10~50ms）避免冲突
		time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)
	}

	// redis分布式锁
	handleMutex := t.redSync.NewMutex("HandlePosExpiredTx" + "-" + posTxRecord.TxID)
	if err := handleMutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 交易[%s] - 获取处理交易的分布式锁错误: %v", t.prefix, posTxRecord.TxID, err)
		}
		return
	}
	defer handleMutex.Unlock()

	// 使用 GetTransaction 获取完整的交易信息，包括 TransactionSignature
	var maxSupportVersion uint64 = 0
	txResult, err := t.rpcClient.GetTransaction(
		ctx,
		txSig,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &maxSupportVersion,
			Commitment:                     rpc.CommitmentFinalized,
			Encoding:                       solana.EncodingBase64,
		},
	)

	// 如果交易无法找到
	if errors.Is(err, rpc.ErrNotFound) {
		log.Errorf("%s 交易[%s] - 无法找到，标记为过期", t.prefix, posTxRecord.TxID)
		t.markTxExpired(ctx, posTxRecord)
		return
	}

	if err != nil {
		log.Errorf("%s 交易[%s] - 查询交易错误: %v", t.prefix, posTxRecord.TxID, err)
		// 可以根据需要决定是否重试或标记为过期
		return
	}

	// 成功找到交易，处理成功逻辑
	log.Infof("%s 交易[%s] - 成功找到交易", t.prefix, posTxRecord.TxID)

	// 从 txResult 中获取 solana.HeliusBuyTokenTx
	solTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txResult.Transaction.GetBinary()))
	if err != nil {
		log.Errorf("%s 交易[%s] - 解析交易二进制数据错误: %v", t.prefix, posTxRecord.TxID, err)
		t.markTxExpired(ctx, posTxRecord)
		return
	}

	// 解析交易
	decodedTx, err := app_utils.DecodeSolanaTransaction(t.rpcClient, t.db, solTx, txSig)
	if err != nil {
		log.Errorf("%s 交易[%s] - 解析交易错误: %v", t.prefix, posTxRecord.TxID, err)
		// 解析失败，标记为过期
		t.markTxExpired(ctx, posTxRecord)
		return
	}

	// 构建 TransactionSignature
	transactionSignature := rpc.TransactionSignature{
		Signature: txSig,
		Slot:      txResult.Slot,
		Err:       txResult.Meta.Err,
		BlockTime: txResult.BlockTime,
	}

	// 标记交易为已经发现
	t.markTxFetched(ctx, posTxRecord, transactionSignature, decodedTx)
	log.Infof("%s 交易[%s] - 处理完成，已发送成功消息", t.prefix, posTxRecord.TxID)
}

// markTxExpired 标记交易为失败并发送消息
func (t *TxExpireTask) markTxExpired(ctx context.Context, posTxRecord model.ServiceTx) {
	// 标记交易发现状态为失败（条件更新，防止与 TxScanTask 重复处理）
	changed, err := t.MarkTxFetchState(posTxRecord.TxID, constants.TxFetchFailed)
	if err != nil {
		log.Errorf("%s 交易[%s] - 标记交易发现状态为失败 - 错误: %v", t.prefix, posTxRecord.TxID, err)
		return
	}
	if !changed {
		log.Infof("%s 交易[%s] - 已被其他任务标记，跳过", t.prefix, posTxRecord.TxID)
		return
	}

	// 发送RocketMQ消息
	rmqMsg := entity.KafkaExpiredTxMsg{
		MsgType: "NewExpiredTransaction",
		MsgContent: entity.NewExpiredTx{
			Service:    posTxRecord.Service,
			SubService: posTxRecord.SubService,
			TxID:       posTxRecord.TxID,
		},
	}

	if err := t.sendMsgToKafka(rmqMsg); err != nil {
		log.Errorf("%s - 标记交易发现状态为失败 - 发送交易超时消息到RocketMQ错误: %v", t.prefix, err)
		return
	}

	log.Infof("%s 交易 %s 超时消息已经发送", t.prefix, posTxRecord.TxID)
}

// markTxFetched 标记交易为已经发现
func (t *TxExpireTask) markTxFetched(ctx context.Context, posTxRecord model.ServiceTx, txSig rpc.TransactionSignature, decodedTx *entity.DecodedSolanaTransaction) {
	// 标记交易发现状态为成功（条件更新，防止与 TxScanTask 重复处理）
	changed, err := t.MarkTxFetchState(posTxRecord.TxID, constants.TxFetchSuccess)
	if err != nil {
		log.Errorf("%s 交易[%s] - 标记交易发现状态为成功 - 错误: %v", t.prefix, posTxRecord.TxID, err)
		return
	}
	if !changed {
		log.Infof("%s 交易[%s] - 已被其他任务标记，跳过Kafka发送", t.prefix, posTxRecord.TxID)
		return
	}

	// 发送成功消息
	rmqMsg := entity.KafkaTxMsg{
		MsgType: "NewScannedTransaction",
		MsgContent: entity.NewScannedTx{
			Service:    posTxRecord.Service,
			SubService: posTxRecord.SubService,
			TxSig:      txSig,
			DecodedTx:  *decodedTx,
		},
	}

	if err := t.sendMsgToKafka(rmqMsg); err != nil {
		log.Errorf("%s 交易[%s] - 标记交易发现状态为成功 - 分发RocketMQ消息错误: %v", t.prefix, posTxRecord.TxID, err)
		return
	}

	log.Infof("%s 交易 %s 成功消息已经发送", t.prefix, posTxRecord.TxID)
}

// MarkTxFetchState 条件标记交易获取状态，仅当状态尚未被设置时才更新。
// 返回 changed=true 表示本次调用实际修改了状态，false 表示已被其他任务标记过。
func (t *TxExpireTask) MarkTxFetchState(txId string, state int) (changed bool, err error) {
	result := t.db.Table(model.TableNameServiceTx).
		Where("tx_id = ? AND tx_state <> ?", txId, state).
		Update("tx_state", state)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// sendMsgToKafka 发送消息到Kafka
func (t *TxExpireTask) sendMsgToKafka(msg entity.KafkaMsg) error {
	if t.kafkaWriter == nil {
		return fmt.Errorf("kafka producer 未初始化")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化Kafka消息失败: %v", err)
	}

	return t.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Topic: txScanKafkaTopic,
		Key:   []byte(msg.GetMsgType()),
		Value: body,
	})
}
