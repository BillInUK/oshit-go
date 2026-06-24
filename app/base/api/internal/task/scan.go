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
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"sort"
	"sync"
	"time"
)

const txScanKafkaTopic = "ServiceTransaction"

// 分布式锁 key 格式
const (
	txScanLockFmt   = "base:sol:tx-scan:%s-%s-%s:lock" // args: service, subService, pdaAccount
	txHandleLockFmt = "base:sol:tx-handle:%s:lock"     // args: txSig
)

// marketBuyTokenProgramID Raydium 相关市场购买程序地址，用于识别 MarketBuyToken 交易
const marketBuyTokenProgramID = "HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i"

// raydiumPoolSourceAccount Raydium (SHIT-USDT) Pool 1 的 token account，作为 TransferChecked source 的过滤条件
const raydiumPoolSourceAccount = "GjkvqFpZ5gqbzYEUAGsn5ozmFgM52JJDgso426DiLXbQ"

// TxScanTask 业务交易扫描任务
type TxScanTask struct {
	db               *gorm.DB
	redis            redis.UniversalClient
	redSync          redsync.Redsync
	rpcClient        *rpc.Client
	mainnetRpcClient *rpc.Client
	rpcURL           string
	heliusAPIKey     string
	kafkaWriter      *kafka.Writer
	mu               sync.Mutex
	runningScans     map[string]context.CancelFunc
}

type ScanConfig struct {
	Service          string
	SubService       string
	Enabled          bool
	NativeAccount    string
	PdaAccount       string
	InitialUntilTxID string
	InitialSlot      uint64
}

// NewTxScanTask 创建交易扫描任务
func NewTxScanTask(taskCtx *TaskContext) *TxScanTask {
	var mainnetRpcClient *rpc.Client
	if taskCtx.MainnetRpcClient != nil {
		mainnetRpcClient = taskCtx.MainnetRpcClient
	}

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

	return &TxScanTask{
		db:               taskCtx.DB,
		redis:            taskCtx.Redis,
		redSync:          taskCtx.RedSync,
		rpcClient:        rpcClient,
		mainnetRpcClient: mainnetRpcClient,
		rpcURL:           rpcURL,
		heliusAPIKey:     taskCtx.HeliusAPIKey,
		kafkaWriter:      kafkaWriter,
		runningScans:     make(map[string]context.CancelFunc),
	}
}

// getRpcClient 根据 subService 选择合适的 RPC 客户端
// MarketBuyToken 强制使用主网 RPC，其余业务使用默认 RPC
func (t *TxScanTask) getRpcClient(subService string) *rpc.Client {
	if subService == constants.SubServiceMarketBuyToken.String() && t.mainnetRpcClient != nil {
		return t.mainnetRpcClient
	}
	return t.rpcClient
}

// Start 启动任务
func (t *TxScanTask) Start() {
	scanInfo, err := t.QueryAllScanInfo()
	if err != nil {
		panic(err)
	}
	configs := make([]ScanConfig, 0, len(scanInfo))
	for _, info := range scanInfo {
		configs = append(configs, ScanConfig{
			Service:          info.Service,
			SubService:       info.SubService,
			Enabled:          true,
			NativeAccount:    info.NativeAccount,
			PdaAccount:       info.PdaAccount,
			InitialUntilTxID: info.UntilTxID,
			InitialSlot:      uint64(info.Slot),
		})
	}
	if err := t.Reconcile(configs); err != nil {
		panic(err)
	}
}

func scanTaskKey(service, subService, pdaAccount string) string {
	return fmt.Sprintf("%s:%s:%s", service, subService, pdaAccount)
}

func (t *TxScanTask) Reconcile(configs []ScanConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	desired := make(map[string]ScanConfig, len(configs))
	for _, cfg := range configs {
		if cfg.Service == "" || cfg.SubService == "" || cfg.PdaAccount == "" {
			return fmt.Errorf("invalid scan config: service/subService/pdaAccount required")
		}
		key := scanTaskKey(cfg.Service, cfg.SubService, cfg.PdaAccount)
		desired[key] = cfg

		if !cfg.Enabled {
			if cancel, ok := t.runningScans[key]; ok {
				cancel()
				delete(t.runningScans, key)
				log.Infof("停止扫描任务 [%s]", key)
			}
			continue
		}

		if err := t.ensureScanCheckpoint(cfg); err != nil {
			return err
		}
		if _, ok := t.runningScans[key]; ok {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		t.runningScans[key] = cancel
		go t.startTxScanTasks(ctx, cfg.Service, cfg.SubService, cfg.PdaAccount)
		log.Infof("启动扫描任务 [%s]", key)
	}

	for key, cancel := range t.runningScans {
		cfg, ok := desired[key]
		if !ok || !cfg.Enabled {
			cancel()
			delete(t.runningScans, key)
			log.Infof("停止扫描任务 [%s]", key)
		}
	}

	return nil
}

func (t *TxScanTask) ensureScanCheckpoint(cfg ScanConfig) error {
	var count int64
	err := t.db.Model(&model.TxScanInfo{}).
		Where("service = ? and sub_service = ? and pda_account = ?", cfg.Service, cfg.SubService, cfg.PdaAccount).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("query scan checkpoint [%s/%s/%s] error: %v", cfg.Service, cfg.SubService, cfg.PdaAccount, err)
	}
	if count > 0 {
		return nil
	}

	record := model.TxScanInfo{
		Service:       cfg.Service,
		SubService:    cfg.SubService,
		NativeAccount: cfg.NativeAccount,
		PdaAccount:    cfg.PdaAccount,
		UntilTxID:     cfg.InitialUntilTxID,
		Slot:          float64(cfg.InitialSlot),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := t.db.Create(&record).Error; err != nil {
		return fmt.Errorf("create scan checkpoint [%s/%s/%s] error: %v", cfg.Service, cfg.SubService, cfg.PdaAccount, err)
	}
	log.Infof("创建扫描检查点 [%s/%s/%s] untilTxId=%s slot=%d", cfg.Service, cfg.SubService, cfg.PdaAccount, cfg.InitialUntilTxID, cfg.InitialSlot)
	return nil
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

func (t *TxScanTask) QueryScanInfo(service, subService, pdaAccount string) (*model.TxScanInfo, error) {
	var err error
	var record model.TxScanInfo
	db := t.db.Table(model.TableNameTxScanInfo)
	err = db.Where("service = ? and sub_service = ? and pda_account = ?", service, subService, pdaAccount).First(&record).Error
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
		Where("service = ? and sub_service = ? and pda_account = ? and slot < ?", service, subService, pdaAccount, slot).
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

// MarkTxFetchState 条件标记交易获取状态，仅当状态尚未被设置时才更新。
// 返回 changed=true 表示本次调用实际修改了状态（应继续发送 Kafka），
// changed=false 表示已被其他 goroutine 标记过（应跳过 Kafka 发送）。
func (t *TxScanTask) MarkTxFetchState(txId string, state int) (changed bool, err error) {
	result := t.db.Table(model.TableNameServiceTx).
		Where("tx_id = ? AND tx_state <> ?", txId, state).
		Update("tx_state", state)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 获取最新交易
func (t *TxScanTask) getLatestTransaction(ctx context.Context, rpcClient *rpc.Client, pdaAccountStr, untilTxId string) ([]*rpc.TransactionSignature, error) {
	pdaAccount, _ := solana.PublicKeyFromBase58(pdaAccountStr)
	untilTx, _ := solana.SignatureFromBase58(untilTxId)

	var limit int = 300
	outs, err := rpcClient.GetSignaturesForAddressWithOpts(
		ctx,
		pdaAccount,
		&rpc.GetSignaturesForAddressOpts{
			Commitment: rpc.CommitmentConfirmed,
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
func (t *TxScanTask) startTxScanTasks(ctx context.Context, service, subService, pdaAccountStr string) {
	prefix := fmt.Sprintf("%s服务 - %s业务", service, subService)
	lockKey := fmt.Sprintf(txScanLockFmt, service, subService, pdaAccountStr)

	// 初始化配置（无需加锁）
	initScanInfo, err := t.QueryScanInfo(service, subService, pdaAccountStr)
	if err != nil || initScanInfo == nil {
		log.Errorf("%s 初始化配置错误: %v", prefix, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Infof("%s 扫描任务退出", prefix)
			return
		default:
		}

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

			// 4. 获取最新交易（MarketBuyToken 强制使用主网 RPC）
			outs, err := t.getLatestTransaction(ctx, t.getRpcClient(subService), pdaAccountStr, scanInfo.UntilTxID)
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
					go t.handleServiceTx(service, subService, *tx)
				}
			}
		}()

		timer := time.NewTimer(1 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Infof("%s 扫描任务退出", prefix)
			return
		case <-timer.C:
		}
	}
}

func (t *TxScanTask) handleServiceTx(service, subService string, txSig rpc.TransactionSignature) {
	prefix := fmt.Sprintf("%s业务 - %s - solana 处理扫描到的交易 -", service, subService)
	log.Infof("%s 交易Id[%s]", prefix, txSig.Signature.String())

	// 使用 Redis 限流，每秒最多 200 次
	for {
		if app_utils.AcquireDistributedRateLimit(t.redis, "handleServiceTx", 200) {
			break
		}
		// 超过 300/s，sleep 一小段时间（随机 10~50ms）避免冲突
		time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)
	}

	handleMutex := t.redSync.NewMutex(
		fmt.Sprintf(txHandleLockFmt, txSig.Signature.String()),
		redsync.WithTries(1), // 同一笔交易可能被多个扫描任务发现（如 take token 和 lottery 共用地址），只允许一个处理
	)
	if err := handleMutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 交易Id[%s] 获取分布式锁出错: %v", prefix, txSig.Signature.String(), err)
		}
		return
	}
	defer handleMutex.Unlock()

	// 根据 subService 选择 RPC 客户端
	rpcClient := t.getRpcClient(subService)

	// 获取交易执行结果
	var err error
	var tr *rpc.GetTransactionResult
	maxRetries := 5
	for i := 0; i <= maxRetries; i++ {
		tr, err = utils.GetTransactionResultByTxId(context.Background(), rpcClient, txSig.Signature)
		if err == nil {
			break
		}
		if errors.Is(err, rpc.ErrNotFound) {
			// tx 刚被 getSignaturesForAddress 发现，ErrNotFound 是 RPC 内部短暂传播延迟，用短退避
			delay := time.Duration(1<<uint(i)) * time.Second // 1s, 2s, 4s, 8s, 16s, 32s
			randomDelay := time.Duration(rand.Intn(500)) * time.Millisecond
			time.Sleep(delay + randomDelay)
			continue
		}
		if app_utils.IsRpcRateLimitedError(err) {
			delay := 5 * time.Duration(1<<uint(i)) * time.Second // 保持原有长退避
			randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(delay + randomDelay)
			continue
		}
		log.Errorf("%s 交易Id[%s] 查询交易错误[%v]", prefix, txSig.Signature.String(), err)
		return
	}

	if tr == nil {
		log.Errorf("%s 交易Id[%s] 获取交易数据为空", prefix, txSig.Signature.String())
		return
	}

	// MarketBuyToken 走独立的解析流程（DEX inner instruction）
	if subService == constants.SubServiceMarketBuyToken.String() {
		t.handleMarketBuyTokenTx(prefix, service, subService, txSig, tr)
		return
	}

	// 原有的处理逻辑，适用于 Stake, StakeToken, Pos, Reward 等
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(tr.Transaction.GetBinary()))
	if err != nil {
		log.Errorf("%s 交易Id[%s] 从交易执行结果获取交错误: %v", prefix, txSig.Signature.String(), err)
		return
	}

	decodedTx, err := app_utils.DecodeSolanaTransaction(rpcClient, t.db, tx, txSig.Signature)
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

	// 将交易设置为已经发现（条件更新，防止多个扫描任务重复发送 Kafka）
	changed, err := t.MarkTxFetchState(rewardTxRecord.TxID, constants.TxFetchSuccess)
	if err != nil {
		log.Errorf("%s 交易Id[%s],设置交易Id为已经发现错误: %v", prefix, txSig.Signature.String(), err)
		return
	}
	if !changed {
		log.Infof("%s 交易Id[%s] 已被其他任务标记，跳过Kafka发送", prefix, txSig.Signature.String())
		return
	}

	rmqMsg := entity.KafkaTxMsg{
		MsgType: "NewScannedTransaction",
		MsgContent: entity.NewScannedTx{
			Service:    rewardTxRecord.Service,
			SubService: rewardTxRecord.SubService,
			TxSig:      txSig,
			DecodedTx:  *decodedTx,
		},
	}
	if err = t.sendMsgToKafka(rmqMsg); err != nil {
		log.Errorf("%s 分发Kafka消息错误: %v", prefix, err)
		return
	}
}

// handleMarketBuyTokenTx 处理 MarketBuyToken DEX 购买交易：
// 解析 inner instructions 中从 Raydium Pool 发出的 TransferChecked 指令，
// 不依赖 t_service_tx，直接将解析结果发送到 Kafka。
//func (t *TxScanTask) handleMarketBuyTokenTx(prefix, service, subService string, txSig rpc.TransactionSignature, tr *rpc.GetTransactionResult) {
//	log.Infof("%s 处理交易所购买token交易，交易Id[%s] ", prefix, txSig.Signature.String())
//	// 只处理链上成功的交易
//	if txSig.Err != nil {
//		log.Infof("%s 交易Id[%s] 交易失败，跳过", prefix, txSig.Signature.String())
//		return
//	}
//
//	// 解码外层交易以获取 account keys
//	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(tr.Transaction.GetBinary()))
//	if err != nil {
//		log.Errorf("%s 交易Id[%s] 解码交易失败: %v", prefix, txSig.Signature.String(), err)
//		return
//	}
//
//	if tr.Meta == nil {
//		log.Warnf("%s 交易Id[%s] meta 为空，跳过", prefix, txSig.Signature.String())
//		return
//	}
//
//	// 构建完整账户列表：静态账户 + ALT 动态加载的账户（v0 交易）
//	// 顺序：static keys → loaded writable → loaded readonly
//	accountKeys := make([]solana.PublicKey, len(tx.Message.AccountKeys))
//	copy(accountKeys, tx.Message.AccountKeys)
//	accountKeys = append(accountKeys, tr.Meta.LoadedAddresses.Writable...)
//	accountKeys = append(accountKeys, tr.Meta.LoadedAddresses.ReadOnly...)
//
//	// 1. 检查完整账户列表是否包含目标程序地址
//	marketProgramPubKey := solana.MPK(marketBuyTokenProgramID)
//	found := false
//	for _, acc := range accountKeys {
//		if acc.Equals(marketProgramPubKey) {
//			found = true
//			break
//		}
//	}
//	if !found {
//		log.Infof("%s 交易Id[%s] 不包含目标程序 %s，跳过", prefix, txSig.Signature.String(), marketBuyTokenProgramID)
//		return
//	}
//
//	// 2. 遍历 inner instructions，找到符合条件的 TransferChecked
//	raydiumSrcPubKey := solana.MPK(raydiumPoolSourceAccount)
//	var matchedInsts []entity.DecodedSolTransferCheckedInst
//
//	for _, innerGroup := range tr.Meta.InnerInstructions {
//		for _, innerInst := range innerGroup.Instructions {
//			// 检查 program 是否为 Token Program
//			if int(innerInst.ProgramIDIndex) >= len(accountKeys) {
//				continue
//			}
//			programID := accountKeys[innerInst.ProgramIDIndex]
//			if !programID.Equals(solana.TokenProgramID) {
//				continue
//			}
//
//			// 检查指令数据：data[0]=12 表示 TransferChecked，data 至少 10 字节
//			data := []byte(innerInst.Data)
//			if len(data) < 10 || data[0] != 12 {
//				continue
//			}
//
//			// 需要至少 4 个 account indices
//			if len(innerInst.Accounts) < 4 {
//				continue
//			}
//
//			// 解析 account 索引
//			srcIdx := int(innerInst.Accounts[0])
//			mintIdx := int(innerInst.Accounts[1])
//			dstIdx := int(innerInst.Accounts[2])
//			authIdx := int(innerInst.Accounts[3])
//
//			nAccounts := len(accountKeys)
//			if srcIdx >= nAccounts || mintIdx >= nAccounts || dstIdx >= nAccounts || authIdx >= nAccounts {
//				continue
//			}
//
//			sourceAccount := accountKeys[srcIdx]
//
//			// 3. source 必须是 Raydium Pool source account
//			if !sourceAccount.Equals(raydiumSrcPubKey) {
//				continue
//			}
//
//			mintAccount := accountKeys[mintIdx]
//			dstAccount := accountKeys[dstIdx]
//			authAccount := accountKeys[authIdx]
//
//			// 解析 amount（uint64 little-endian）和 decimals
//			amount := binary.LittleEndian.Uint64(data[1:9])
//			decimals := data[9]
//
//			matchedInsts = append(matchedInsts, entity.DecodedSolTransferCheckedInst{
//				FromTokenAccount:   sourceAccount,
//				FromNativeAccount:  authAccount,
//				TokenMintAccount:   mintAccount,
//				ToTokenAccount:     dstAccount,
//				OwnerNativeAccount: authAccount,
//				Amount:             amount,
//				Decimals:           decimals,
//			})
//		}
//	}
//
//	if len(matchedInsts) == 0 {
//		log.Infof("%s 交易Id[%s] 未找到符合条件的 TransferChecked 指令，跳过", prefix, txSig.Signature.String())
//		return
//	}
//
//	// 4. 填充 DecodedSolanaTransaction
//	decodedTx := entity.DecodedSolanaTransaction{
//		TxID:                        txSig.Signature,
//		FromNativeAccount:           accountKeys[0],
//		TransferCheckedInstructions: matchedInsts,
//	}
//
//	// 5. 发送 Kafka 消息
//	rmqMsg := entity.KafkaTxMsg{
//		MsgType: "NewScannedTransaction",
//		MsgContent: entity.NewScannedTx{
//			Service:    service,
//			SubService: subService,
//			TxSig:      txSig,
//			DecodedTx:  decodedTx,
//		},
//	}
//	if err := t.sendMsgToKafka(rmqMsg); err != nil {
//		log.Errorf("%s 交易Id[%s] 发送 Kafka 消息失败: %v", prefix, txSig.Signature.String(), err)
//	}
//}

func (t *TxScanTask) handleMarketBuyTokenTx(prefix, service, subService string, txSig rpc.TransactionSignature, tr *rpc.GetTransactionResult) {
	log.Infof("%s 处理交易所购买token交易，交易Id[%s] ", prefix, txSig.Signature.String())
	// 只处理链上成功的交易
	if txSig.Err != nil {
		log.Infof("%s 交易Id[%s] 交易失败，跳过", prefix, txSig.Signature.String())
		return
	}
	decodedTx, err := utils.HeliusParseMarketBuyTx(t.heliusAPIKey, txSig.Signature)
	if err != nil {
		log.Errorf("%s 交易Id[%s] 交易解析错误: %v", prefix, txSig.Signature.String(), err)
		return
	}
	// 5. 发送 Kafka 消息
	rmqMsg := entity.KafkaTxMsg{
		MsgType: "NewScannedTransaction",
		MsgContent: entity.NewScannedTx{
			Service:    service,
			SubService: subService,
			TxSig:      txSig,
			DecodedTx:  *decodedTx,
		},
	}
	if err := t.sendMsgToKafka(rmqMsg); err != nil {
		log.Errorf("%s 交易Id[%s] 发送 Kafka 消息失败: %v", prefix, txSig.Signature.String(), err)
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
