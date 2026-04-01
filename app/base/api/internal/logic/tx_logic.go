package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"

	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/pkg/dal/model"
)

type TxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	db     *gorm.DB
}

func NewTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TxLogic {
	return &TxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		db:     svcCtx.DB,
	}
}

// SendTransaction 签名并异步广播 encodedTx，将结果写入 t_service_tx，同步返回 record_id 和 tx_id
func (l *TxLogic) SendTransaction(req *types.SendTransactionReq) (*types.SendTransactionRsp, error) {
	// 1. base64 解码
	txBytes, err := base64.StdEncoding.DecodeString(req.EncodedTx)
	if err != nil {
		return nil, fmt.Errorf("decode encoded_tx base64 error: %v", err)
	}

	// 2. 反序列化 Solana 交易
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return nil, fmt.Errorf("deserialize transaction error: %v", err)
	}

	// 3. 找私钥
	subSvcMap, ok := l.svcCtx.ServiceKeyMap[req.Service]
	if !ok {
		return nil, fmt.Errorf("no key configured for service [%s]", req.Service)
	}
	privateKey, ok := subSvcMap[req.SubService]
	if !ok {
		return nil, fmt.Errorf("no key configured for service [%s] subService [%s]", req.Service, req.SubService)
	}

	// 4. 签名交易消息
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal transaction message error: %v", err)
	}
	sig, err := privateKey.Sign(messageContent)
	if err != nil {
		return nil, fmt.Errorf("sign transaction error: %v", err)
	}
	if len(tx.Signatures) >= 2 {
		tx.Signatures[1] = sig
	} else {
		tx.Signatures = append(tx.Signatures, sig)
	}

	// 5. tx_id = 用户签名（索引0），即链上的交易ID
	if len(tx.Signatures) == 0 {
		return nil, fmt.Errorf("transaction has no signatures")
	}
	txID := tx.Signatures[0].String()

	// 6. 写 t_service_tx（state=0 待确认）
	record := model.ServiceTx{
		Service:       req.Service,
		SubService:    req.SubService,
		TxID:          txID,
		State:         0,
		RetryCount:    0,
		NextRetryTime: time.Now(),
		MaxRetries:    5,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := l.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create service_tx record error: %v", err)
	}

	// 7. 异步广播
	go l.broadcastTx(record.RecordID, tx)

	return &types.SendTransactionRsp{
		RecordID: record.RecordID,
		TxID:     txID,
	}, nil
}

// broadcastTx 异步广播交易并更新 t_service_tx 状态
func (l *TxLogic) broadcastTx(recordID string, tx *solana.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rpcClient := l.svcCtx.RpcClient
	if rpcClient == nil {
		log.Errorf("broadcast tx [%s] error: rpc client not initialized", recordID)
		l.updateServiceTxState(recordID, -1)
		return
	}

	_, err := rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		log.Errorf("broadcast tx [%s] error: %v", recordID, err)
		l.retryOrFail(recordID, tx)
		return
	}

	log.Infof("broadcast tx [%s] success", recordID)
	l.updateServiceTxState(recordID, 1)
}

// retryOrFail 若还有重试次数则递增重试计数，否则标记失败
func (l *TxLogic) retryOrFail(recordID string, tx *solana.Transaction) {
	var record model.ServiceTx
	if err := l.db.Where("record_id = ?", recordID).First(&record).Error; err != nil {
		return
	}
	if record.RetryCount < record.MaxRetries {
		l.db.Model(&model.ServiceTx{}).Where("record_id = ?", recordID).Updates(map[string]interface{}{
			"retry_count":     record.RetryCount + 1,
			"next_retry_time": time.Now().Add(time.Duration(record.RetryCount+1) * 30 * time.Second),
			"state":           0,
			"updated_at":      time.Now(),
		})
		// 等待后重试
		time.Sleep(time.Duration(record.RetryCount+1) * 30 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := l.svcCtx.RpcClient.SendTransaction(ctx, tx); err != nil {
			log.Errorf("retry broadcast tx [%s] error: %v", recordID, err)
			l.updateServiceTxState(recordID, -1)
			return
		}
		l.updateServiceTxState(recordID, 1)
	} else {
		l.updateServiceTxState(recordID, -1)
	}
}

func (l *TxLogic) updateServiceTxState(recordID string, state int32) {
	l.db.Model(&model.ServiceTx{}).Where("record_id = ?", recordID).Updates(map[string]interface{}{
		"state":      state,
		"updated_at": time.Now(),
	})
}

// commitWithSkipPreflight 使用 skip-preflight 选项广播，适合某些场景
func (l *TxLogic) sendWithOptions(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	skipPreflight := true
	return l.svcCtx.RpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight: skipPreflight,
	})
}
