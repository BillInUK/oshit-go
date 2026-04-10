package rewardcode

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/entity"
)

// decodedRewardCodeTx 解析后的奖励码交易中间结构
type decodedRewardCodeTx struct {
	FromNativeAccount solana.PublicKey
	FromTokenAccount  solana.PublicKey
	// SOL 成本转账指令
	TransferInst *entity.DecodedSolTransferInst
	// Token 奖励转账指令
	TransferCheckedInst *entity.DecodedSolTransferCheckedInst
}

// decodeSOLTx 解析奖励码领取交易
// 期望包含: 1条 SOL Transfer (用户→cost_account) + 1条 TransferChecked (reward_account→用户)
func (l *RewardCodeLogic) decodeSOLTx(tx *solana.Transaction, txInfo *types.RewardCodeTxInfo) (*decodedRewardCodeTx, error) {
	tokenMintAccount, _ := solana.PublicKeyFromBase58(txInfo.Mint)

	result := &decodedRewardCodeTx{}
	result.FromNativeAccount = tx.Message.AccountKeys[0]

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(result.FromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, errors.New("can not find token account by transaction from native account")
	}
	result.FromTokenAccount = fromTokenAccount

	for index, inst := range tx.Message.Instructions {
		programId, err := tx.ResolveProgramIDIndex(inst.ProgramIDIndex)
		if err != nil {
			return nil, fmt.Errorf("can not decode program id of instruction %d", index)
		}

		if programId.Equals(solana.ComputeBudget) {
			// compute budget 指令忽略
			continue
		} else if programId.Equals(solana.SPLAssociatedTokenAccountProgramID) {
			continue
		} else if programId.Equals(solana.SystemProgramID) {
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedSystemInst, err := system.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			transfer, ok := decodedSystemInst.Impl.(*system.Transfer)
			if !ok {
				return nil, errors.New("decode system instruction error: expected Transfer")
			}
			decoded := entity.DecodedSolTransferInst{
				FromNativeAccount: transfer.GetFundingAccount().PublicKey,
				ToNativeAccount:   transfer.GetRecipientAccount().PublicKey,
				Amount:            *transfer.Lamports,
			}
			result.TransferInst = &decoded
		} else if programId.Equals(solana.TokenProgramID) {
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedTokenInst, err := token.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			transferChecked, ok := decodedTokenInst.Impl.(*token.TransferChecked)
			if !ok {
				return nil, errors.New("decode token instruction error: expected TransferChecked")
			}
			fromNativeAccount, err := app_utils.QueryNativeAccountByTokenAccount(l.rpcClient, l.db, transferChecked.Accounts.Get(0).PublicKey)
			if err != nil {
				account := transferChecked.Accounts.Get(0).PublicKey
				return nil, fmt.Errorf("can not get native account of from token account[%v]: %v", account, err)
			}
			toNativeAccount, err := app_utils.QueryNativeAccountByTokenAccount(l.rpcClient, l.db, transferChecked.Accounts.Get(2).PublicKey)
			if err != nil {
				// to address may be reward recipient (not yet on-chain), derive from tx sender
				toNativeAccount = &result.FromNativeAccount
			}
			decoded := entity.DecodedSolTransferCheckedInst{
				FromTokenAccount:   transferChecked.Accounts.Get(0).PublicKey,
				FromNativeAccount:  *fromNativeAccount,
				TokenMintAccount:   transferChecked.Accounts.Get(1).PublicKey,
				ToTokenAccount:     transferChecked.Accounts.Get(2).PublicKey,
				ToNativeAccount:    *toNativeAccount,
				OwnerNativeAccount: transferChecked.Accounts.Get(3).PublicKey,
				Amount:             *transferChecked.Amount,
				Decimals:           *transferChecked.Decimals,
			}
			result.TransferCheckedInst = &decoded
		} else if programId.Equals(l.srvCtx.LightHouseAddress) {
			// light house 安全校验，忽略
		} else {
			// 检查是否 compute budget（有些版本 program id 不同）
			accounts, _ := inst.ResolveInstructionAccounts(&tx.Message)
			cbInst, cbErr := computebudget.DecodeInstruction(accounts, inst.Data)
			if cbErr == nil && cbInst != nil {
				continue
			}
			return nil, fmt.Errorf("unsupported program id %v", programId)
		}
	}

	if result.TransferInst == nil {
		return nil, errors.New("missing SOL transfer instruction to cost account")
	}
	if result.TransferCheckedInst == nil {
		return nil, errors.New("missing TransferChecked instruction for token reward")
	}

	return result, nil
}

// checkSOLTx 校验奖励码交易内容是否符合规则
func (l *RewardCodeLogic) checkSOLTx(decodedTx *decodedRewardCodeTx, txInfo *types.RewardCodeTxInfo) error {
	prefix := fmt.Sprintf("%s 校验交易 -", l.prefix)

	// 校验 SOL Transfer: from 必须是交易发起人，to 必须是 cost_account
	transfer := decodedTx.TransferInst
	if transfer.FromNativeAccount != decodedTx.FromNativeAccount {
		log.Errorf("%s SOL转账发起地址[%v]不是交易发起地址[%v]", prefix, transfer.FromNativeAccount, decodedTx.FromNativeAccount)
		return errors.New("transfer funding account must be transaction fee payer")
	}
	if transfer.ToNativeAccount.String() != txInfo.CostAccount {
		log.Errorf("%s SOL收款地址[%v]不是规则要求的cost_account[%v]", prefix, transfer.ToNativeAccount, txInfo.CostAccount)
		return errors.New("transfer recipient must be cost account")
	}
	// 允许一定容差
	minCostFee := txInfo.CostFee * (1 - l.srvCtx.FeeTolerance.MaxLessRate)
	if float64(transfer.Amount) < minCostFee {
		log.Errorf("%s SOL成本费[%d] 小于规则要求的最小值[%d]", prefix, transfer.Amount, uint64(minCostFee))
		return errors.New("cost fee is less than required minimum")
	}

	// 校验 TransferChecked: from 必须是 reward_account 的 token account，to 必须是用户的 token account
	tc := decodedTx.TransferCheckedInst
	tokenMintPubKey, _ := solana.PublicKeyFromBase58(txInfo.Mint)

	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(
		solana.MPK(txInfo.RewardAccount),
		tokenMintPubKey,
	)
	if tc.FromTokenAccount.String() != rewardTokenAccount.String() {
		log.Errorf("%s token发送地址[%v]不是规则要求的reward_account token account[%v]", prefix, tc.FromTokenAccount, rewardTokenAccount)
		return errors.New("token transfer source must be reward account token account")
	}

	expectedToTA, _, _ := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintPubKey)
	if tc.ToTokenAccount.String() != expectedToTA.String() {
		log.Errorf("%s token接收地址[%v]不是交易发起人的token account[%v]", prefix, tc.ToTokenAccount, expectedToTA)
		return errors.New("token transfer destination must be transaction sender token account")
	}

	if tc.TokenMintAccount.String() != txInfo.Mint {
		log.Errorf("%s token mint[%v]与规则要求的[%v]不一致", prefix, tc.TokenMintAccount, txInfo.Mint)
		return errors.New("token mint account mismatch")
	}

	if tc.Amount != uint64(txInfo.RewardAmount) {
		log.Errorf("%s token奖励金额[%d]与规则要求的[%d]不一致", prefix, tc.Amount, uint64(txInfo.RewardAmount))
		return errors.New("reward token amount mismatch")
	}

	if tc.Decimals != uint8(txInfo.Decimals) {
		log.Errorf("%s token精度[%d]与规则要求的[%d]不一致", prefix, tc.Decimals, txInfo.Decimals)
		return errors.New("token decimals mismatch")
	}

	return nil
}
