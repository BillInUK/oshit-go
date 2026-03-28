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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"math/rand"
	"oshit-go/app/base/dal/model"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"sort"
	"time"
)

const txScanKafkaTopic = "OShitPos"

// 分布式锁 key 格式
const (
	txScanLockFmt   = "base:sol:tx-scan:%s-%s:lock"  // args: service, subService
	txHandleLockFmt = "base:sol:tx-handle:%s:lock"    // args: txSig
)

// TxScanTask 业务交易扫描任务
type TxScanTask struct {
	db          *gorm.DB
	redis       redis.UniversalClient
	redSync     redsync.Redsync
	rpcClient   *rpc.Client
	rpcURL      string
	kafkaWriter *kafka.Writer
}

// NewTxScanTask 创建交易扫描任务
func NewTxScanTask(taskCtx *TaskContext) *TxScanTask {
	rpcURL := taskCtx.ChainConfig.RPCURL
	rpcClient := rpc.New(rpcURL)

	var kafkaWriter *kafka.Writer
	if taskCtx.KafkaProducer != nil {
		if w, ok := taskCtx.KafkaProducer.(*kafka.Writer); ok {
			kafkaWriter = w
		}
	}

	return &TxScanTask{
		db:          taskCtx.DB,
		redis:       taskCtx.Redis,
		redSync:     taskCtx.RedSync,
		rpcClient:   rpcClient,
		rpcURL:      rpcURL,
		kafkaWriter: kafkaWriter,
	}
}

// Start 启动任务
func (t *TxScanTask) Start() {
	scanInfo, err := t.QueryAllScanInfo()
	if err != nil {
		panic(err)
	}
	for _, info := range scanInfo {
		go t.startTxScanTasks(info.Service, info.SubService)
	}
}

func (t *TxScanTask) QueryAllScanInfo() ([]model.TxScanInfo, error) {
	var err error
	var records []model.TxScanInfo
	db := t.db.Table(model.TableNameTxScanInfo)
	err = db.Find(&records).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return records, nil
}

func (t *TxScanTask) QueryScanInfo(service, subService string) (*model.TxScanInfo, error) {
	var err error
	var record model.TxScanInfo
	db := t.db.Table(model.TableNameTxScanInfo)
	err = db.Where("service = ? and sub_service = ?", service, subService).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, nil
}

// QueryScanInfoWithLock 带锁查询（事务内）
func (t *TxScanTask) QueryScanInfoWithLock(tx *gorm.DB, service, subService, pdaAccount string) (*model.TxScanInfo, error) {
	var scanInfo model.TxScanInfo
	err := tx.Set("gorm:query_option", "FOR UPDATE NOWAIT").
		Where("service = ? and sub_service = ? and pda_account = ?", service, subService, pdaAccount).
		First(&scanInfo).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("记录不存在")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
		return nil, nil
	}
	return &scanInfo, err
}

// UpdateScanUntilTx 条件更新（事务内）
func (t *TxScanTask) UpdateScanUntilTx(tx *gorm.DB, service, subService, pdaAccount, untilTxId string, slot uint64) error {
	return tx.Model(&model.TxScanInfo{}).
		Where("service = ? and sub_service = ? pda_account = ? AND slot < ?", service, subService, pdaAccount, slot).
		Updates(map[string]interface{}{
			"until_tx_id": untilTxId,
			"slot":        slot,
		}).Error
}

// GetTxFetchState 查询交易获取状态
func (t *TxScanTask) GetTxFetchState(txId string) (*model.ServiceTx, error) {
	var record model.ServiceTx
	table := t.db.Table(model.TableNameServiceTx)
	if err := table.Where("tx_id = ?", txId).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// MarkTxFetchState 标记交易获取状态
func (t *TxScanTask) MarkTxFetchState(txId string, state int) error {
	table := t.db.Table(model.TableNameServiceTx)
	if err := table.Where("tx_id = ?", txId).Update("state", state).Error; err != nil {
		return err
	}
	return nil
}

// 获取最新交易
func (t *TxScanTask) getLatestTransaction(ctx context.Context, pdaAccountStr, untilTxId string) ([]*rpc.TransactionSignature, error) {
	pdaAccount, _ := solana.PublicKeyFromBase58(pdaAccountStr)
	untilTx, _ := solana.SignatureFromBase58(untilTxId)

	var limit int = 300
	outs, err := t.rpcClient.GetSignaturesForAddressWithOpts(
		ctx,
		pdaAccount,
		&rpc.GetSignaturesForAddressOpts{
			Commitment: rpc.CommitmentFinalized,
			Until:      untilTx,
			Limit:      &limit,
		},
	)
	if err != nil || len(outs) == 0 {
		return nil, err
	}
	return outs, nil
}

// startTxScanTasks 开启单独的协程扫描业务相关交易
func (t *TxScanTask) startTxScanTasks(service, subService string) {
	prefix := fmt.Sprintf("%s服务 - %s业务", service, subService)
	lockKey := fmt.Sprintf(txScanLockFmt, service, subService)
	ctx := context.Background()

	// 初始化配置（无需加锁）
	initScanInfo, err := t.QueryScanInfo(service, subService)
	if err != nil || initScanInfo == nil {
		panic(fmt.Sprintf("%s 初始化配置错误: %v", prefix, err))
	}
	pdaAccountStr := initScanInfo.PdaAccount

	for {
		func() {
			// 每个循环独立作用域
			// 1. 获取分布式锁
			distMutex := t.redSync.NewMutex(
				lockKey,
				redsync.WithTries(1), // 只尝试一次
			)
			if err := distMutex.Lock(); err != nil {
				return // 其他进程持有锁时静默退出
			}
			defer distMutex.Unlock()

			// 2. 开启数据库事务
			tx := t.db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			// 3. 行级锁定
			scanInfo, err := t.QueryScanInfoWithLock(tx, service, subService, pdaAccountStr)
			if err != nil || scanInfo == nil {
				tx.Rollback()
				return
			}

			// 4. 获取最新交易
			outs, err := t.getLatestTransaction(ctx, pdaAccountStr, scanInfo.UntilTxID)
			if err != nil || len(outs) == 0 {
				tx.Rollback()
				return
			}
			// 5.交易按照slot降序排序
			if len(outs) > 0 {
				sort.Slice(outs, func(i, j int) bool {
					return outs[i].Slot > outs[j].Slot
				})
			}
			latestTx := outs[0]
			// 6. 条件更新
			if latestTx.Slot > uint64(scanInfo.Slot) {
				if err := t.UpdateScanUntilTx(tx, service, subService, pdaAccountStr, latestTx.Signature.String(), latestTx.Slot); err != nil {
					tx.Rollback()
					return
				}
			}

			// 7. 提交事务（自动释放行锁）
			if err := tx.Commit().Error; err != nil {
				log.Errorf("%s 提交事务失败: %v", prefix, err)
				return
			}

			// 8. 处理交易（在事务外异步执行）
			for _, tx := range outs {
				if tx.Slot > uint64(scanInfo.Slot) {
					go t.handleServiceTx(service, *tx)
				}
			}
		}()

		time.Sleep(3 * time.Second)
	}
}

func (t *TxScanTask) handleServiceTx(service string, txSig rpc.TransactionSignature) {
	prefix := fmt.Sprintf("%s业务 - solana 处理扫描到的交易 -", service)
	log.Infof("%s 交易Id[%s]", prefix, txSig.Signature.String())

	// 使用 Redis 限流，每秒最多 200 次
	for {
		if app_utils.AcquireDistributedRateLimit(t.redis, "handleServiceTx", 200) {
			break
		}
		// 超过 300/s，sleep 一小段时间（随机 10~50ms）避免冲突
		time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)
	}

	handleMutex := t.redSync.NewMutex(fmt.Sprintf(txHandleLockFmt, txSig.Signature.String()))
	if err := handleMutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 交易Id[%s] 获取分布式锁出错: %v", prefix, txSig.Signature.String(), err)
		}
		return
	}
	defer handleMutex.Unlock()

	// 获取交易执行结果
	var err error
	var tr *rpc.GetTransactionResult
	maxRetries := 5
	for i := 0; i <= maxRetries; i++ {
		tr, err = utils.GetTransactionResultByTxId(context.Background(), t.rpcClient, txSig.Signature)
		if err == nil {
			break
		}
		if errors.Is(err, rpc.ErrNotFound) || app_utils.IsRpcRateLimitedError(err) {
			delay := 5 * time.Duration(1<<i) * time.Second
			randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond
			totalDelay := delay + randomDelay
			time.Sleep(totalDelay)
			continue
		}
		log.Errorf("%s 交易Id[%s],查询交易错误[%v]", prefix, txSig.Signature.String(), err)
		return
	}

	if tr == nil {
		log.Errorf("%s 交易Id[%s] 获取交易数据为空", prefix, txSig.Signature.String())
		return
	}

	// 原有的处理逻辑，适用于Stake, StakeToken, Pos
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(tr.Transaction.GetBinary()))
	if err != nil {
		log.Errorf("%s 交易Id[%s] 从交易执行结果获取交错误: %v", prefix, txSig.Signature.String(), err)
		return
	}

	decodedTx, err := app_utils.DecodeSolanaTransaction(t.rpcClient, t.db, tx, txSig.Signature)
	if err != nil {
		log.Errorf("%s 交易Id[%s],解析交易错误[%v]", prefix, txSig.Signature.String(), err)
		return
	}
	// 查询数据库里面是否存在该交易Id
	rewardTxRecord, err := t.GetTxFetchState(txSig.Signature.String())
	if err != nil || rewardTxRecord == nil {
		log.Errorf("%s 交易Id[%s],查询数据库错误: %v", prefix, txSig.Signature.String(), err)
		return
	}

	// 将交易设置为已经发现
	if err := t.MarkTxFetchState(rewardTxRecord.TxID, constants.TxFetchStateSuccess); err != nil {
		log.Errorf("%s 交易Id[%s],设置交易Id为已经发现错误: %v", prefix, txSig.Signature.String(), err)
		return
	}

	rmqMsg := entity.KafkaTxMsg{
		MsgType: "NewScannedTransaction",
		MsgContent: entity.NewScannedTx{
			Service:    rewardTxRecord.Service,
			SubService: rewardTxRecord.Service,
			TxSig:      txSig,
			DecodedTx:  *decodedTx,
		},
	}
	if err = t.sendMsgToKafka(rmqMsg); err != nil {
		log.Errorf("%s 分发Kafka消息错误: %v", prefix, err)
		return
	}
}

// sendMsgToKafka 发送消息到Kafka
func (t *TxScanTask) sendMsgToKafka(msg entity.KafkaMsg) error {
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
