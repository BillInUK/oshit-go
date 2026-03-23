package task

import (
	"context"
	"fmt"
	"oshit-go/app/base/dal/model"
	"oshit-go/app/base/dal/query"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"
)

// ServiceTask 服务交易扫描任务
type ServiceTask struct {
	BaseTask
	db     *gorm.DB
	rpc    *rpc.Client
	ticker *time.Ticker
}

// NewServiceTask 创建服务交易扫描任务
func NewServiceTask(db *gorm.DB, rpcClient *rpc.Client) *ServiceTask {
	return &ServiceTask{
		BaseTask: BaseTask{
			name: "service-transaction-scan",
		},
		db:  db,
		rpc: rpcClient,
	}
}

// Start 启动任务
func (t *ServiceTask) Start(ctx context.Context) error {
	t.ticker = time.NewTicker(30 * time.Second) // 每30秒扫描一次
	go t.run(ctx)
	return nil
}

// Stop 停止任务
func (t *ServiceTask) Stop() error {
	if t.ticker != nil {
		t.ticker.Stop()
	}
	return nil
}

// run 任务主循环
func (t *ServiceTask) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.ticker.C:
			t.scanPendingTransactions(ctx)
			t.scanTimeoutTransactions(ctx)
		}
	}
}

// scanPendingTransactions 扫描待处理交易
func (t *ServiceTask) scanPendingTransactions(ctx context.Context) {
	q := query.Use(t.db)
	serviceTx := q.ServiceTx

	// 查找待处理交易（state=0）
	transactions, err := serviceTx.WithContext(ctx).
		Where(serviceTx.State.Eq(0)).
		Find()

	if err != nil {
		fmt.Printf("查询待处理交易失败: %v\n", err)
		return
	}

	for _, tx := range transactions {
		t.processTransaction(ctx, tx)
	}
}

// scanTimeoutTransactions 扫描超时交易
func (t *ServiceTask) scanTimeoutTransactions(ctx context.Context) {
	q := query.Use(t.db)
	serviceTx := q.ServiceTx

	// 查找超时交易（state=0且created_at距离现在超过5分钟）
	timeoutTime := time.Now().Add(-5 * time.Minute)
	transactions, err := serviceTx.WithContext(ctx).
		Where(
			serviceTx.State.Eq(0),
			serviceTx.CreatedAt.Lte(timeoutTime),
		).
		Find()

	if err != nil {
		fmt.Printf("查询超时交易失败: %v\n", err)
		return
	}

	for _, tx := range transactions {
		t.processTransaction(ctx, tx)
	}
}

// processTransaction 处理单个交易
func (t *ServiceTask) processTransaction(ctx context.Context, tx *model.ServiceTx) {
	// 查询交易信息
	txSignature, err := solana.SignatureFromBase58(tx.TxID)
	if err != nil {
		fmt.Printf("解析交易签名失败: %v\n", err)
		t.markTransactionAsFailed(ctx, tx, "invalid transaction signature")
		return
	}

	// 查询链上交易状态
	txInfo, err := t.rpc.GetTransaction(ctx, txSignature, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		// 交易未找到，需要重试
		t.handleTransactionNotFound(ctx, tx)
		return
	}

	// 交易找到，更新状态
	if txInfo != nil && txInfo.Meta != nil {
		if txInfo.Meta.Err != nil {
			// 交易失败
			t.markTransactionAsFailed(ctx, tx, fmt.Sprintf("transaction failed: %v", txInfo.Meta.Err))
		} else {
			// 交易成功
			t.markTransactionAsSuccess(ctx, tx)
		}
	}
}

// handleTransactionNotFound 处理未找到的交易（简化版）
func (t *ServiceTask) handleTransactionNotFound(ctx context.Context, tx *model.ServiceTx) {
	// 简化版：直接标记为失败
	t.markTransactionAsFailed(ctx, tx, "transaction not found on chain")
}

// markTransactionAsSuccess 标记交易为成功
func (t *ServiceTask) markTransactionAsSuccess(ctx context.Context, tx *model.ServiceTx) {
	q := query.Use(t.db)
	serviceTx := q.ServiceTx

	_, err := serviceTx.WithContext(ctx).
		Where(serviceTx.RecordID.Eq(tx.RecordID)).
		UpdateColumns(map[string]interface{}{
			"state":      1, // 成功状态
			"updated_at": time.Now(),
		})

	if err != nil {
		fmt.Printf("标记交易成功失败: %v\n", err)
		return
	}

	// 通知业务模块
	t.notifyService(ctx, tx, 1)
}

// markTransactionAsFailed 标记交易为失败
func (t *ServiceTask) markTransactionAsFailed(ctx context.Context, tx *model.ServiceTx, reason string) {
	q := query.Use(t.db)
	serviceTx := q.ServiceTx

	_, err := serviceTx.WithContext(ctx).
		Where(serviceTx.RecordID.Eq(tx.RecordID)).
		UpdateColumns(map[string]interface{}{
			"state":      -1, // 失败状态
			"updated_at": time.Now(),
		})

	if err != nil {
		fmt.Printf("标记交易失败失败: %v\n", err)
		return
	}

	// 通知业务模块
	t.notifyService(ctx, tx, -1)
}

// notifyService 通知业务模块
func (t *ServiceTask) notifyService(ctx context.Context, tx *model.ServiceTx, state int32) {
	q := query.Use(t.db)
	serviceInfo := q.ServiceInfo

	// 查询服务配置
	service, err := serviceInfo.WithContext(ctx).
		Where(serviceInfo.Service.Eq(tx.Service)).
		First()

	if err != nil {
		fmt.Printf("查询服务配置失败: %v\n", err)
		return
	}

	// 根据配置类型通知
	switch service.HookType {
	case 0: // Kafka
		t.notifyViaKafka(ctx, service, tx, state)
	case 1: // Webhook
		t.notifyViaWebhook(ctx, service, tx, state)
	default:
		fmt.Printf("未知的通知类型: %d\n", service.HookType)
	}
}

// notifyViaKafka 通过Kafka通知
func (t *ServiceTask) notifyViaKafka(ctx context.Context, service *model.ServiceInfo, tx *model.ServiceTx, state int32) {
	// TODO: 实现Kafka通知逻辑
	fmt.Printf("通过Kafka通知服务 %s: 交易 %s 状态变为 %d\n",
		service.Service, tx.TxID, state)
}

// notifyViaWebhook 通过Webhook通知
func (t *ServiceTask) notifyViaWebhook(ctx context.Context, service *model.ServiceInfo, tx *model.ServiceTx, state int32) {
	// TODO: 实现Webhook通知逻辑
	fmt.Printf("通过Webhook通知服务 %s: 交易 %s 状态变为 %d\n",
		service.Service, tx.TxID, state)
}
