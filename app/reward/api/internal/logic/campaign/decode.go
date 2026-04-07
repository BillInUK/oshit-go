package campaign

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/entity"
)

func (l *CampaignLogic) checkSOLTx(txInfo *types.CampaignExchangeTxInfo, decodedTx *entity.DecodedSolanaTransaction, receiptAccount string) (*entity.DecodedServiceTransaction, error) {
	var err error
	var prefix = fmt.Sprintf("%s 检查解码后的solana交易 -", l.prefix)
	var decodedServiceTx *entity.DecodedServiceTransaction

	// 检查成本费指令
	if len(decodedTx.TransferInstructions) != 1 {
		log.Errorf("%s 解析交易错误，只能包含1条transfer指令", prefix)
		return nil, errors.New("transfer instructions count must be 1")
	}
	transferInst := decodedTx.TransferInstructions[0]
	if transferInst.ToNativeAccount.String() != txInfo.CostAccount {
		log.Errorf("%s 解析交易错误，成本费的收款地址必须是平台配置的地址", prefix)
		return nil, errors.New("cost receipt account error")
	}
	if transferInst.FromNativeAccount.String() != receiptAccount {
		log.Errorf("%s 解析交易错误，成本费的支付地址必须是交易发起地址", prefix)
		return nil, errors.New("cost from account error")
	}

	// 只能包含1条transfer checked指令
	if len(decodedTx.TransferCheckedInstructions) != 1 {
		log.Errorf("%s 解析交易错误，只能包含1条transfer checked指令", prefix)
		return nil, errors.New("transfer checked instructions count must be 1")
	}
	// 校验transfer check指令
	inst := decodedTx.TransferCheckedInstructions[0]
	if inst.FromNativeAccount.String() != txInfo.RewardAccount {
		log.Errorf("%s 解析交易错误，发放奖励地址 %v 不是campaign发放奖励的地址 %s", prefix, inst.FromNativeAccount, txInfo.RewardAccount)
		return nil, errors.New("reward from address error")
	}
	if inst.ToNativeAccount.String() != receiptAccount {
		log.Errorf("%s 解析交易错误，接收奖励的地址: %v 不是签名交易的地址: %s", prefix, inst.ToNativeAccount, receiptAccount)
		return nil, errors.New("receipt to address error")
	}
	if int32(inst.Decimals) != txInfo.Decimals {
		log.Errorf("%s 解析交易错误，交易当中的进制 %d 根系统配置的不一致 %d", prefix, inst.Decimals, l.srvCtx.TokenConfig.Decimals)
		return nil, errors.New("decimals in transfer checked error")
	}
	if inst.Amount != uint64(txInfo.TokenAmount) {
		log.Errorf("%s 解析交易错误，交易当中兑换的金额: %d 与实际能兑换的金额 %d 不一致", prefix, inst.Amount, uint64(txInfo.TokenAmount))
		return nil, errors.New("transfer instructions count must be 0")
	}

	dexMinInstFee := uint64(txInfo.CostFee * (1 - l.srvCtx.FeeTolerance.MaxLessRate) * float64(solana.LAMPORTS_PER_SOL))
	if transferInst.Amount < dexMinInstFee {
		log.Errorf("%s 解析交易错误: 转给dex的手续费 %d 小于规则要求的 %d ", prefix, transferInst.Amount, dexMinInstFee)
		return nil, errors.New("transfer funding less than rule required")
	}

	if decodedServiceTx, err = app_utils.DecodeServiceTransaction(decodedTx); err != nil {
		return nil, fmt.Errorf("%s convert decoded solana transaction to service transaction error: %v", prefix, err)
	}

	log.Infof("%s 解析完成", prefix)

	return decodedServiceTx, nil
}
