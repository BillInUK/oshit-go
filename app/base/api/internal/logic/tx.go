package logic

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"

	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/constants"
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

// SendTransaction 根据 t_service_info 中的 multi_sign/confirm 配置，决定是否签名、是否写 t_service_tx，
// 最后异步广播（不重试），同步返回 record_id 和 tx_id。
func (l *TxLogic) SendTransaction(req *types.SendTransactionReq) (*types.SendTransactionRsp, error) {
	// 1. 查询服务配置
	l.svcCtx.ConfigMu.RLock()
	subSvcInfoMap, ok := l.svcCtx.ServiceInfoMap[req.Service]
	if !ok {
		l.svcCtx.ConfigMu.RUnlock()
		return nil, fmt.Errorf("no service info configured for service [%s]", req.Service)
	}
	serviceInfo, ok := subSvcInfoMap[req.SubService]
	if !ok {
		l.svcCtx.ConfigMu.RUnlock()
		return nil, fmt.Errorf("no service info configured for service [%s] subService [%s]", req.Service, req.SubService)
	}
	multiSign := serviceInfo.MultiSign
	confirm := serviceInfo.Confirm
	serviceKeyMap := l.svcCtx.ServiceKeyMap
	l.svcCtx.ConfigMu.RUnlock()

	// 2. base64 解码
	txBytes, err := base64.StdEncoding.DecodeString(req.EncodedTx)
	if err != nil {
		return nil, fmt.Errorf("decode encoded_tx base64 error: %v", err)
	}

	// 3. 反序列化 Solana 交易
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return nil, fmt.Errorf("deserialize transaction error: %v", err)
	}

	// 4. tx_id = 用户签名（索引0），即链上的交易ID
	if len(tx.Signatures) == 0 {
		return nil, fmt.Errorf("transaction has no signatures")
	}
	txID := tx.Signatures[0].String()

	// 5. 若 confirm=true，写 t_service_tx（state=0 待确认）
	var recordID string
	if confirm {
		record := model.ServiceTx{
			Service:    req.Service,
			SubService: req.SubService,
			TxID:       txID,
			TxState:    int32(constants.TxStateInit),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := l.db.Create(&record).Error; err != nil {
			return nil, fmt.Errorf("create service_tx record error: %v", err)
		}
		recordID = record.RecordID
	}

	// 6. 若 multi_sign=true，用服务私钥在索引1处补签
	if multiSign {
		subSvcMap, ok := serviceKeyMap[req.Service]
		if !ok {
			return nil, fmt.Errorf("no key configured for service [%s]", req.Service)
		}
		privateKey, ok := subSvcMap[req.SubService]
		if !ok {
			return nil, fmt.Errorf("no key configured for service [%s] subService [%s]", req.Service, req.SubService)
		}
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
	}

	// 7. 模拟执行交易，失败则直接返回错误
	l.svcCtx.ConfigMu.RLock()
	rpcClient := l.svcCtx.RpcClient
	l.svcCtx.ConfigMu.RUnlock()
	if rpcClient == nil {
		if confirm && recordID != "" {
			l.updateServiceTxState(recordID, constants.TxStateFailed)
		}
		return nil, fmt.Errorf("rpc client is nil")
	}

	simCtx, simCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer simCancel()

	simResult, simErr := rpcClient.SimulateTransactionWithOpts(simCtx, tx, &rpc.SimulateTransactionOpts{
		SigVerify:              false,
		ReplaceRecentBlockhash: true,
	})
	if simErr != nil {
		// 模拟 RPC 调用本身出错：区分网络层错误和 HTTP 层错误
		if isNetworkError(simErr) {
			log.Errorf("simulate tx [%s] network error: %v", txID, simErr)
			if confirm && recordID != "" {
				l.updateServiceTxState(recordID, constants.TxStateFailed)
			}
			return nil, fmt.Errorf("simulate transaction network error: %v", simErr)
		}
		// HTTP 层错误（400/500/timeout 等），跳过模拟，继续广播（保守策略）
		log.Warnf("simulate tx [%s] non-fatal error, skip simulation: %v", txID, simErr)
	} else if simResult != nil && simResult.Value != nil && simResult.Value.Err != nil {
		// 模拟执行链上失败
		log.Errorf("simulate tx [%s] execution failed: %v, logs: %v", txID, simResult.Value.Err, simResult.Value.Logs)
		if confirm && recordID != "" {
			l.updateServiceTxState(recordID, constants.TxStateFailed)
		}
		return nil, fmt.Errorf("simulate transaction failed: %v", simResult.Value.Err)
	}

	// 8. 异步广播
	go l.broadcastTx(recordID, confirm, tx)

	return &types.SendTransactionRsp{
		RecordID: recordID,
		TxID:     txID,
	}, nil
}

// broadcastTx 异步广播交易：
// - 网络传输层错误（connection reset/refused 等）：标记失败，交易根本没发出去
// - HTTP 层/RPC 层错误（400/500/timeout 等）：仅打日志，不标记失败，留给 TxScanTask/TxExpireTask 兜底
func (l *TxLogic) broadcastTx(recordID string, confirm bool, tx *solana.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	l.svcCtx.ConfigMu.RLock()
	rpcClient := l.svcCtx.RpcClient
	l.svcCtx.ConfigMu.RUnlock()
	if rpcClient == nil {
		log.Errorf("broadcast tx [%v] error: rpc client is nil", tx.Signatures[0])
		if confirm && recordID != "" {
			l.updateServiceTxState(recordID, constants.TxStateFailed)
		}
		return
	}

	if _, err := rpcClient.SendTransaction(ctx, tx); err != nil {
		if isNetworkError(err) {
			log.Errorf("broadcast tx [%v] network error: %v", tx.Signatures[0], err)
			if confirm && recordID != "" {
				l.updateServiceTxState(recordID, constants.TxStateFailed)
			}
		} else {
			// HTTP 层错误，交易可能已到达节点，不标记失败，等兜底任务处理
			log.Warnf("broadcast tx [%v] non-fatal error (will be handled by scan/expire task): %v", tx.Signatures[0], err)
		}
		return
	}

	log.Infof("broadcast tx [%v] success", tx.Signatures[0])
}

// isNetworkError 判断是否为网络传输层错误（交易根本没发出去）
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// 标准库 net 包的错误（connection refused, connection reset, no such host 等）
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	// 部分场景下错误被包装为字符串，兜底用关键字匹配
	msg := strings.ToLower(err.Error())
	networkKeywords := []string{
		"connection refused",
		"connection reset by peer",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"dial tcp",
		"eof",
	}
	for _, kw := range networkKeywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

func (l *TxLogic) updateServiceTxState(recordID string, state constants.TxState) {
	l.db.Model(&model.ServiceTx{}).Where("record_id = ?", recordID).Updates(map[string]interface{}{
		"tx_state":   state,
		"updated_at": time.Now(),
	})
}
