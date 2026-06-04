package lottery

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"oshit-go/app/reward/api/types"
	"oshit-go/common/pkg/entity"
)

// decodeSOLTx 解析抽奖领取的 solana 交易
// 不需要从链上查询 native account，因为我们已知所有账户信息（来自 rewardInfo）
func (l *LotteryLogic) decodeSOLTx(tx *solana.Transaction, rewardInfo *types.ClaimLotteryTxInfo) (*entity.DecodedSolanaTransaction, error) {
	var decodedTx entity.DecodedSolanaTransaction

	decodedTx.FromNativeAccount = tx.Message.AccountKeys[0]
	decodedTx.Signatures = tx.Signatures

	tokenMintAccount, err := solana.PublicKeyFromBase58(rewardInfo.Mint)
	if err != nil {
		return nil, fmt.Errorf("invalid token mint: %v", err)
	}

	// 推导用户的 token account
	userTokenAccount, _, err := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, fmt.Errorf("can not find user token account: %v", err)
	}
	decodedTx.FromTokenAccount = userTokenAccount

	// 解析 rewardTokenAccount（from RewardAccount）
	rewardNativeKey, err := solana.PublicKeyFromBase58(l.serviceConfig.RewardAccount)
	if err != nil {
		return nil, fmt.Errorf("invalid reward account: %v", err)
	}
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeKey, tokenMintAccount)

	for index, inst := range tx.Message.Instructions {
		programId, err := tx.ResolveProgramIDIndex(inst.ProgramIDIndex)
		if err != nil {
			return nil, fmt.Errorf("can not decode program id of instruction %d", index)
		}

		if programId.Equals(solana.ComputeBudget) {
			accounts, _ := inst.ResolveInstructionAccounts(&tx.Message)
			computeBudgetInst, _ := computebudget.DecodeInstruction(accounts, inst.Data)
			if computeBudgetInst.TypeID.Uint8() == computebudget.Instruction_SetComputeUnitPrice {
				computeUnitPriceInst, _ := computeBudgetInst.Impl.(*computebudget.SetComputeUnitPrice)
				decodedTx.ComputeUnitPrice = computeUnitPriceInst.MicroLamports
			}
			if computeBudgetInst.TypeID.Uint8() == computebudget.Instruction_SetComputeUnitLimit {
				computeUnitLimitInst, _ := computeBudgetInst.Impl.(*computebudget.SetComputeUnitLimit)
				decodedTx.ComputeUnitLimit = uint64(computeUnitLimitInst.Units)
			}
			continue
		}

		if programId.Equals(solana.SPLAssociatedTokenAccountProgramID) {
			// 可选的 CreateATA 指令，直接跳过
			continue
		}

		if programId.Equals(solana.TokenProgramID) {
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedTokenInst, err := token.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			if transferChecked, ok := decodedTokenInst.Impl.(*token.TransferChecked); ok {
				fromTokenAccount := transferChecked.Accounts.Get(0).PublicKey
				toTokenAccount := transferChecked.Accounts.Get(2).PublicKey

				// 推导 from 的 native account：fromTokenAccount 应该是 rewardTokenAccount
				var fromNativeAccount solana.PublicKey
				if fromTokenAccount.Equals(rewardTokenAccount) {
					fromNativeAccount = rewardNativeKey
				} else {
					return nil, fmt.Errorf("unexpected from token account in TransferChecked: %v", fromTokenAccount)
				}

				// 推导 to 的 native account：toTokenAccount 应该是 userTokenAccount
				var toNativeAccount solana.PublicKey
				if toTokenAccount.Equals(userTokenAccount) {
					toNativeAccount = decodedTx.FromNativeAccount
				} else {
					// 尝试通过 ATA 推导：也许用户的 ATA 还不存在，用 fromNativeAccount + mint 推算
					return nil, fmt.Errorf("unexpected to token account in TransferChecked: %v", toTokenAccount)
				}

				checkedInst := entity.DecodedSolTransferCheckedInst{
					FromTokenAccount:   fromTokenAccount,
					FromNativeAccount:  fromNativeAccount,
					TokenMintAccount:   transferChecked.Accounts.Get(1).PublicKey,
					ToTokenAccount:     toTokenAccount,
					ToNativeAccount:    toNativeAccount,
					OwnerNativeAccount: transferChecked.Accounts.Get(3).PublicKey,
					Amount:             *transferChecked.Amount,
					Decimals:           *transferChecked.Decimals,
				}
				decodedTx.TransferCheckedInstructions = append(decodedTx.TransferCheckedInstructions, checkedInst)
			} else {
				return nil, errors.New("decode transfer checked instruction error")
			}
			continue
		}

		if programId.Equals(solana.SystemProgramID) {
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedSystemInst, err := system.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			if transfer, ok := decodedSystemInst.Impl.(*system.Transfer); ok {
				transferInst := entity.DecodedSolTransferInst{
					FromNativeAccount: transfer.GetFundingAccount().PublicKey,
					ToNativeAccount:   transfer.GetRecipientAccount().PublicKey,
					Amount:            *transfer.Lamports,
				}
				decodedTx.TransferInstructions = append(decodedTx.TransferInstructions, transferInst)
			} else {
				return nil, errors.New("decode transfer instruction error")
			}
			continue
		}

		if programId.Equals(l.srvCtx.LightHouseAddress) {
			// lighthouse 安全校验，跳过
			continue
		}

		return nil, fmt.Errorf("unsupported program id %v", programId)
	}

	return &decodedTx, nil
}

// checkSOLTx 校验抽奖领取交易中的指令是否符合规则
func (l *LotteryLogic) checkSOLTx(decodedTx *entity.DecodedSolanaTransaction, rewardInfo *types.ClaimLotteryTxInfo) error {
	prefix := fmt.Sprintf("%s 检查解码后的solana交易 -", l.prefix)

	// 必须包含 1 条 SOL Transfer 指令（用户→costAccount）
	if len(decodedTx.TransferInstructions) != 1 {
		log.Errorf("%s 必须包含1条转SOL到cost账户的指令，实际: %d", prefix, len(decodedTx.TransferInstructions))
		return errors.New("must contain exactly 1 SOL transfer instruction to cost account")
	}

	// 必须包含 1 条 TransferChecked 指令（rewardTokenAccount→userTokenAccount）
	if len(decodedTx.TransferCheckedInstructions) != 1 {
		log.Errorf("%s 必须包含1条TransferChecked指令，实际: %d", prefix, len(decodedTx.TransferCheckedInstructions))
		return errors.New("must contain exactly 1 TransferChecked instruction")
	}

	// 校验 SOL Transfer
	transferInst := decodedTx.TransferInstructions[0]
	costAccount, err := solana.PublicKeyFromBase58(rewardInfo.CostAccount)
	if err != nil {
		return fmt.Errorf("invalid cost account: %v", err)
	}
	if !transferInst.ToNativeAccount.Equals(costAccount) {
		log.Errorf("%s SOL转账目标账户[%v]不是规则要求的cost账户[%v]", prefix, transferInst.ToNativeAccount, costAccount)
		return errors.New("SOL transfer recipient must be cost account")
	}
	if !transferInst.FromNativeAccount.Equals(decodedTx.FromNativeAccount) {
		log.Errorf("%s SOL转账发送地址[%v]不是交易发起地址[%v]", prefix, transferInst.FromNativeAccount, decodedTx.FromNativeAccount)
		return errors.New("SOL transfer sender must be transaction fee payer")
	}

	// 校验 SOL 金额（允许费率容错）
	requiredCostFee := rewardInfo.CostFee
	minRequiredCostFee := requiredCostFee * (1 - l.srvCtx.FeeTolerance.MaxLessRate)
	log.Infof("%s 要求cost fee %d, 最小cost fee %d, 实际 %d",
		prefix, uint64(requiredCostFee), uint64(minRequiredCostFee), transferInst.Amount)
	if transferInst.Amount < uint64(minRequiredCostFee) {
		log.Errorf("%s SOL转账金额[%d]小于最小要求金额[%d]", prefix, transferInst.Amount, uint64(minRequiredCostFee))
		return errors.New("SOL transfer amount is less than required cost fee")
	}

	// 校验 TransferChecked
	transferCheckedInst := decodedTx.TransferCheckedInstructions[0]

	// from 必须是 rewardTokenAccount
	rewardNativeKey, _ := solana.PublicKeyFromBase58(l.serviceConfig.RewardAccount)
	tokenMintKey, _ := solana.PublicKeyFromBase58(rewardInfo.Mint)
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeKey, tokenMintKey)

	if !transferCheckedInst.FromTokenAccount.Equals(rewardTokenAccount) {
		log.Errorf("%s TransferChecked的from账户[%v]不是规则要求的rewardTokenAccount[%v]",
			prefix, transferCheckedInst.FromTokenAccount, rewardTokenAccount)
		return errors.New("TransferChecked from account must be reward token account")
	}

	// to 必须是 userTokenAccount
	userTokenAccount, _, _ := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintKey)
	if !transferCheckedInst.ToTokenAccount.Equals(userTokenAccount) {
		log.Errorf("%s TransferChecked的to账户[%v]不是用户的token账户[%v]",
			prefix, transferCheckedInst.ToTokenAccount, userTokenAccount)
		return errors.New("TransferChecked to account must be user token account")
	}

	// mint 必须正确
	if transferCheckedInst.TokenMintAccount.String() != rewardInfo.Mint {
		log.Errorf("%s TransferChecked的mint账户[%v]不是规则要求的[%v]",
			prefix, transferCheckedInst.TokenMintAccount, rewardInfo.Mint)
		return errors.New("TransferChecked token mint not match")
	}

	// 金额必须等于 LotteryAmount
	if transferCheckedInst.Amount != uint64(rewardInfo.LotteryAmount) {
		log.Errorf("%s TransferChecked的金额[%d]不是规则要求的[%d]",
			prefix, transferCheckedInst.Amount, uint64(rewardInfo.LotteryAmount))
		return errors.New("TransferChecked amount does not match lottery amount")
	}

	// decimals 必须正确
	expectedDecimals := uint8(l.srvCtx.TokenConfig.Decimals)
	if transferCheckedInst.Decimals != expectedDecimals {
		log.Errorf("%s TransferChecked的精度[%d]不是规则要求的[%d]",
			prefix, transferCheckedInst.Decimals, expectedDecimals)
		return errors.New("TransferChecked decimals not match token config")
	}

	return nil
}
