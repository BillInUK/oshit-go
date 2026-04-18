package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
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
	subSvcInfoMap, ok := l.svcCtx.ServiceInfoMap[req.Service]
	if !ok {
		return nil, fmt.Errorf("no service info configured for service [%s]", req.Service)
	}
	serviceInfo, ok := subSvcInfoMap[req.SubService]
	if !ok {
		return nil, fmt.Errorf("no service info configured for service [%s] subService [%s]", req.Service, req.SubService)
	}
	multiSign := serviceInfo.MultiSign
	confirm := serviceInfo.Confirm

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
		subSvcMap, ok := l.svcCtx.ServiceKeyMap[req.Service]
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

	// 7. 异步广播，不重试
	go l.broadcastTx(recordID, confirm, tx)

	return &types.SendTransactionRsp{
		RecordID: recordID,
		TxID:     txID,
	}, nil
}

// broadcastTx 异步广播交易，失败直接标记 state=-1（仅 confirm=true 时写库），不重试
func (l *TxLogic) broadcastTx(recordID string, confirm bool, tx *solana.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if _, err := l.svcCtx.RpcClient.SendTransaction(ctx, tx); err != nil {
		log.Errorf("broadcast tx record[%v] error: %v", tx.Signatures[0], err)
		if confirm && recordID != "" {
			l.updateServiceTxState(recordID, constants.TxStateFailed)
		}
		return
	}

	log.Infof("broadcast tx record[%v] success", tx.Signatures[0])
}

func (l *TxLogic) updateServiceTxState(recordID string, state constants.TxState) {
	l.db.Model(&model.ServiceTx{}).Where("record_id = ?", recordID).Updates(map[string]interface{}{
		"tx_state":   state,
		"updated_at": time.Now(),
	})
}
