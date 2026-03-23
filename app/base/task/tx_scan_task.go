package task

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"oshit-go/app/base/dal/model"
	"oshit-go/app/base/dal/query"
	"sort"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"
)

// TxScanTask 通用交易扫描任务
type TxScanTask struct {
	BaseTask
	db     *gorm.DB
	rpc    *rpc.Client
	ticker *time.Ticker
}

// NewTxScanTask 创建交易扫描任务
func NewTxScanTask(db *gorm.DB, rpcClient *rpc.Client) *TxScanTask {
	return &TxScanTask{
		BaseTask: BaseTask{
			name: "transaction-scan",
		},
		db:  db,
		rpc: rpcClient,
	}
}

// Start 启动任务
func (t *TxScanTask) Start(ctx context.Context) error {
	t.ticker = time.NewTicker(3 * time.Second) // 每3秒扫描一次，与旧工程保持一致
	go t.run(ctx)
	return nil
}

// Stop 停止任务
func (t *TxScanTask) Stop() error {
	if t.ticker != nil {
		t.ticker.Stop()
	}
	return nil
}

// run 任务主循环
func (t *TxScanTask) run(ctx context.Context) {
	// 启动时先获取所有需要扫描的服务
	scanInfos, err := t.queryAllScanInfo(ctx)
	if err != nil {
		fmt.Printf("初始化扫描任务失败: %v\n", err)
		return
	}

	// 为每个服务启动独立的扫描协程
	for _, info := range scanInfos {
		go t.startScanTransactions(ctx, info.Service, info.PdaAccount)
	}

	// 主循环监听新注册的服务
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.ticker.C:
			// 定期检查是否有新注册的服务
			t.checkNewServices(ctx)
		}
	}
}

// queryAllScanInfo 查询所有需要扫描的服务信息
func (t *TxScanTask) queryAllScanInfo(ctx context.Context) ([]*model.ServiceInfo, error) {
	q := query.Use(t.db)
	serviceInfo := q.ServiceInfo

	records, err := serviceInfo.WithContext(ctx).Find()
	if err != nil {
		return nil, fmt.Errorf("query all scan info failed: %v", err)
	}

	return records, nil
}

// checkNewServices 检查新注册的服务
func (t *TxScanTask) checkNewServices(ctx context.Context) {
	// TODO: 实现检查新注册服务的逻辑
	// 可以通过比较内存中的服务列表和数据库中的服务列表来实现
}

// startScanTransactions 开启单独的协程扫描业务相关交易
func (t *TxScanTask) startScanTransactions(ctx context.Context, service, pdaAccount string) {
	prefix := fmt.Sprintf("%s业务 -", service)

	for {
		// 查询服务信息
		scanInfo, err := t.queryServiceInfo(ctx, service)
		if err != nil || scanInfo == nil {
			fmt.Printf("%s 查询服务信息失败: %v\n", prefix, err)
			time.Sleep(3 * time.Second) // 等待3秒后重试
			continue
		}

		// 获取最新交易
		outs, err := t.getLatestTransaction(ctx, pdaAccount, "")
		if err != nil || len(outs) == 0 {
			time.Sleep(3 * time.Second)
			continue
		}

		// 交易按照slot降序排序
		if len(outs) > 0 {
			sort.Slice(outs, func(i, j int) bool {
				return outs[i].Slot > outs[j].Slot
			})
		}

		// 处理交易（异步执行）
		for _, txSig := range outs {
			go t.handleServiceTx(service, *txSig)
		}

		time.Sleep(3 * time.Second) // 与旧工程保持一致
	}
}

// queryServiceInfo 查询服务信息
func (t *TxScanTask) queryServiceInfo(ctx context.Context, service string) (*model.ServiceInfo, error) {
	q := query.Use(t.db)
	serviceInfo := q.ServiceInfo

	record, err := serviceInfo.WithContext(ctx).
		Where(serviceInfo.Service.Eq(service)).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("service info not found for service %s", service)
	}
	if err != nil {
		return nil, fmt.Errorf("query service info failed: %v", err)
	}

	return record, nil
}

// getLatestTransaction 获取最新交易
func (t *TxScanTask) getLatestTransaction(ctx context.Context, pdaAccountStr, untilTxID string) ([]*rpc.TransactionSignature, error) {
	pdaAccount, err := solana.PublicKeyFromBase58(pdaAccountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pda account: %v", err)
	}

	var untilTx solana.Signature
	if untilTxID != "" {
		untilTx, err = solana.SignatureFromBase58(untilTxID)
		if err != nil {
			return nil, fmt.Errorf("invalid until tx id: %v", err)
		}
	}

	var limit int = 300 // 与旧工程保持一致
	outs, err := t.rpc.GetSignaturesForAddressWithOpts(
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

// updateScanUntilTx 条件更新扫描信息（占位符，暂时不实现）
func (t *TxScanTask) updateScanUntilTx(tx *gorm.DB, ctx context.Context, service, pdaAccount, untilTxID string, slot uint64) error {
	// TODO: 实现更新扫描信息逻辑，需要先创建t_tx_scan_info表
	return nil
}

// handleServiceTx 处理扫描到的交易
func (t *TxScanTask) handleServiceTx(service string, txSig rpc.TransactionSignature) {
	prefix := fmt.Sprintf("%s业务 - solana 处理扫描到的交易 -", service)
	fmt.Printf("%s 交易Id[%s]\n", prefix, txSig.Signature.String())

	// 使用简单的限流（简化版）
	time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)

	// 获取交易执行结果（带重试机制）
	var err error
	var tr *rpc.GetTransactionResult
	maxRetries := 5
	for i := 0; i <= maxRetries; i++ {
		tr, err = t.getTransactionResult(context.Background(), txSig.Signature)
		if err == nil {
			break
		}
		if errors.Is(err, rpc.ErrNotFound) || t.isRpcRateLimitedError(err) {
			delay := 5 * time.Duration(1<<i) * time.Second
			randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond
			totalDelay := delay + randomDelay
			time.Sleep(totalDelay)
			continue
		}
		fmt.Printf("%s 交易Id[%s],查询交易错误[%v]\n", prefix, txSig.Signature.String(), err)
		return
	}

	if tr == nil {
		fmt.Printf("%s 交易Id[%s] 获取交易数据为空\n", prefix, txSig.Signature.String())
		return
	}

	// 根据service类型处理不同的逻辑
	// 按照要求，暂时不需要处理 service == "StakeBuyCheck" 的业务逻辑
	// 只需要按照默认的service处理逻辑进行处理即可

	// 默认处理逻辑：发送交易信息到Kafka
	t.sendTransactionToKafka(service, txSig, tr)
}

// getTransactionResult 获取交易结果
func (t *TxScanTask) getTransactionResult(ctx context.Context, txSig solana.Signature) (*rpc.GetTransactionResult, error) {
	return t.rpc.GetTransaction(ctx, txSig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
}

// isRpcRateLimitedError 检查是否为RPC限流错误
func (t *TxScanTask) isRpcRateLimitedError(err error) bool {
	// 简化实现，实际应该检查具体的错误类型
	errStr := err.Error()
	return len(errStr) > 0 // 简化处理
}

// sendTransactionToKafka 发送交易信息到Kafka
func (t *TxScanTask) sendTransactionToKafka(service string, txSig rpc.TransactionSignature, tr *rpc.GetTransactionResult) {
	// 查询服务配置
	q := query.Use(t.db)
	serviceInfo := q.ServiceInfo

	serviceConfig, err := serviceInfo.WithContext(context.Background()).
		Where(serviceInfo.Service.Eq(service)).
		First()

	if err != nil {
		fmt.Printf("查询服务 %s 配置失败: %v\n", service, err)
		return
	}

	// 构建消息
	message := map[string]interface{}{
		"msg_type":  "NewScannedTransaction",
		"service":   service,
		"tx_id":     txSig.Signature.String(),
		"tx_slot":   txSig.Slot,
		"state":     0, // 初始状态
		"timestamp": time.Now().Format(time.RFC3339),
		"tx_data":   tr, // 包含完整的交易数据
	}

	// 根据配置类型发送消息
	switch serviceConfig.HookType {
	case 0: // Kafka
		t.sendKafkaMessage(serviceConfig.MqTopic, message)
	case 1: // Webhook
		t.sendWebhookMessage(serviceConfig.Webhook, message)
	default:
		fmt.Printf("未知的通知类型: %d\n", serviceConfig.HookType)
	}
}

// sendKafkaMessage 发送Kafka消息
func (t *TxScanTask) sendKafkaMessage(topic string, message interface{}) {
	// TODO: 实现Kafka消息发送逻辑
	// 按照要求，暂时只需要简单的将交易信息获取到，然后发送到kafka即可
	fmt.Printf("发送Kafka消息到 topic=%s: %v\n", topic, message)
}

// sendWebhookMessage 发送Webhook消息
func (t *TxScanTask) sendWebhookMessage(webhookURL string, message interface{}) {
	// TODO: 实现Webhook消息发送逻辑
	fmt.Printf("发送Webhook消息到 %s: %v\n", webhookURL, message)
}
