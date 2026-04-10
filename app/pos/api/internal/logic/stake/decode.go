package stake

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/pkg/entity"
)

// decodeSOLTx 解析 stake 奖励领取的 Solana 交易。
// 交易结构：SetComputeUnitPrice / SetComputeUnitLimit / [CreateATA] /
//
//	TransferChecked(rewardTokenAccount→userTokenAccount) /
//	System.Transfer(user→costAccount)
func (l *StakeRewardLogic) decodeSOLTx(tx *solana.Transaction, txInfo *types.ClaimStakeRewardTxInfo) (*entity.DecodedSolanaTransaction, error) {
	var decodedTx entity.DecodedSolanaTransaction
	decodedTx.FromNativeAccount = tx.Message.AccountKeys[0]
	decodedTx.Signatures = tx.Signatures

	tokenMintAccount, err := solana.PublicKeyFromBase58(txInfo.Mint)
	if err != nil {
		return nil, fmt.Errorf("invalid token mint: %v", err)
	}

	// 推导用户的 token account
	userTokenAccount, _, err := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, fmt.Errorf("can not find user token account: %v", err)
	}
	decodedTx.FromTokenAccount = userTokenAccount

	// 推导 rewardTokenAccount
	rewardNativeKey, err := solana.PublicKeyFromBase58(txInfo.RewardAccount)
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
			// 可选的 CreateATA 指令，跳过
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

				var fromNativeAccount solana.PublicKey
				if fromTokenAccount.Equals(rewardTokenAccount) {
					fromNativeAccount = rewardNativeKey
				} else {
					return nil, fmt.Errorf("unexpected from token account in TransferChecked: %v", fromTokenAccount)
				}

				var toNativeAccount solana.PublicKey
				if toTokenAccount.Equals(userTokenAccount) {
					toNativeAccount = decodedTx.FromNativeAccount
				} else {
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

		if programId.Equals(l.LightHouseAddress) {
			// lighthouse 安全校验，跳过
			continue
		}

		return nil, fmt.Errorf("unsupported program id %v", programId)
	}

	return &decodedTx, nil
}

// decodeLeaderSOLTx 解析区域经理领取奖励的 Solana 交易。
// 交易结构：SetComputeUnitPrice / SetComputeUnitLimit / [CreateATA] /
//
//	TransferChecked(leaderRewardTokenAccount → leaderTokenAccount)
func (l *StakeRewardLogic) decodeLeaderSOLTx(tx *solana.Transaction, txInfo *types.LeaderRewardTxInfo) (*entity.DecodedSolanaTransaction, error) {
	var decodedTx entity.DecodedSolanaTransaction
	decodedTx.FromNativeAccount = tx.Message.AccountKeys[0]
	decodedTx.Signatures = tx.Signatures

	tokenMintAccount, err := solana.PublicKeyFromBase58(txInfo.Mint)
	if err != nil {
		return nil, fmt.Errorf("invalid token mint: %v", err)
	}

	leaderTokenAccount, _, err := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, fmt.Errorf("can not find leader token account: %v", err)
	}
	decodedTx.FromTokenAccount = leaderTokenAccount

	rewardNativeKey, err := solana.PublicKeyFromBase58(txInfo.RewardAccount)
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

				var fromNativeAccount solana.PublicKey
				if fromTokenAccount.Equals(rewardTokenAccount) {
					fromNativeAccount = rewardNativeKey
				} else {
					return nil, fmt.Errorf("unexpected from token account in TransferChecked: %v", fromTokenAccount)
				}

				var toNativeAccount solana.PublicKey
				if toTokenAccount.Equals(leaderTokenAccount) {
					toNativeAccount = decodedTx.FromNativeAccount
				} else {
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

		if programId.Equals(l.LightHouseAddress) {
			continue
		}

		return nil, fmt.Errorf("unsupported program id %v", programId)
	}

	return &decodedTx, nil
}

// checkLeaderSOLTx 校验区域经理领取奖励交易是否符合规则
func (l *StakeRewardLogic) checkLeaderSOLTx(decodedTx *entity.DecodedSolanaTransaction, txInfo *types.LeaderRewardTxInfo) error {
	prefix := "区域经理领取奖励 checkLeaderSOLTx -"

	if len(decodedTx.TransferInstructions) != 0 {
		log.Errorf("%s 不应包含SOL转账指令，实际: %d", prefix, len(decodedTx.TransferInstructions))
		return errors.New("leader claim transaction must not contain SOL transfer")
	}
	if len(decodedTx.TransferCheckedInstructions) != 1 {
		log.Errorf("%s 必须包含1条TransferChecked指令，实际: %d", prefix, len(decodedTx.TransferCheckedInstructions))
		return errors.New("must contain exactly 1 TransferChecked instruction")
	}

	checkedInst := decodedTx.TransferCheckedInstructions[0]
	rewardNativeKey, _ := solana.PublicKeyFromBase58(txInfo.RewardAccount)
	tokenMintKey, _ := solana.PublicKeyFromBase58(txInfo.Mint)
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeKey, tokenMintKey)

	if !checkedInst.FromTokenAccount.Equals(rewardTokenAccount) {
		log.Errorf("%s TransferChecked from[%v]不是leaderRewardTokenAccount[%v]", prefix, checkedInst.FromTokenAccount, rewardTokenAccount)
		return errors.New("TransferChecked from account must be leader reward token account")
	}
	leaderTokenAccount, _, _ := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintKey)
	if !checkedInst.ToTokenAccount.Equals(leaderTokenAccount) {
		log.Errorf("%s TransferChecked to[%v]不是领导人token账户[%v]", prefix, checkedInst.ToTokenAccount, leaderTokenAccount)
		return errors.New("TransferChecked to account must be leader token account")
	}
	if checkedInst.TokenMintAccount.String() != txInfo.Mint {
		log.Errorf("%s TransferChecked mint[%v]不匹配[%v]", prefix, checkedInst.TokenMintAccount, txInfo.Mint)
		return errors.New("TransferChecked token mint not match")
	}
	if checkedInst.Amount != uint64(txInfo.TotalReward) {
		log.Errorf("%s TransferChecked金额[%d]不等于totalReward[%.0f]", prefix, checkedInst.Amount, txInfo.TotalReward)
		return errors.New("TransferChecked amount does not match total reward")
	}
	if checkedInst.Decimals != uint8(txInfo.Decimals) {
		log.Errorf("%s TransferChecked精度[%d]不匹配[%d]", prefix, checkedInst.Decimals, txInfo.Decimals)
		return errors.New("TransferChecked decimals not match token config")
	}

	return nil
}

// checkSOLTx 校验 stake 奖励领取交易是否符合规则
func (l *StakeRewardLogic) checkSOLTx(decodedTx *entity.DecodedSolanaTransaction, txInfo *types.ClaimStakeRewardTxInfo) error {
	prefix := fmt.Sprintf("%s checkSOLTx -", l.prefix)

	// 必须恰好 1 条 System.Transfer（用户→costAccount）
	if len(decodedTx.TransferInstructions) != 1 {
		log.Errorf("%s 必须包含1条SOL转账指令，实际: %d", prefix, len(decodedTx.TransferInstructions))
		return errors.New("must contain exactly 1 SOL transfer instruction")
	}
	// 必须恰好 1 条 TransferChecked（rewardTokenAccount→userTokenAccount）
	if len(decodedTx.TransferCheckedInstructions) != 1 {
		log.Errorf("%s 必须包含1条TransferChecked指令，实际: %d", prefix, len(decodedTx.TransferCheckedInstructions))
		return errors.New("must contain exactly 1 TransferChecked instruction")
	}

	// 校验 SOL Transfer
	transferInst := decodedTx.TransferInstructions[0]
	costAccount, err := solana.PublicKeyFromBase58(txInfo.CostAccount)
	if err != nil {
		return fmt.Errorf("invalid cost account: %v", err)
	}
	if !transferInst.ToNativeAccount.Equals(costAccount) {
		log.Errorf("%s SOL转账目标[%v]不是cost账户[%v]", prefix, transferInst.ToNativeAccount, costAccount)
		return errors.New("SOL transfer recipient must be cost account")
	}
	if !transferInst.FromNativeAccount.Equals(decodedTx.FromNativeAccount) {
		log.Errorf("%s SOL转账发送地址[%v]不是交易发起地址[%v]", prefix, transferInst.FromNativeAccount, decodedTx.FromNativeAccount)
		return errors.New("SOL transfer sender must be transaction fee payer")
	}
	minCostFee := txInfo.CostFee * (1 - l.srvCtx.FeeTolerance.MaxLessRate)
	log.Infof("%s 要求costFee %.0f，最小costFee %.0f，实际 %d", prefix, txInfo.CostFee, minCostFee, transferInst.Amount)
	if float64(transferInst.Amount) < minCostFee {
		log.Errorf("%s SOL转账金额[%d]小于最小要求[%.0f]", prefix, transferInst.Amount, minCostFee)
		return errors.New("SOL transfer amount is less than required cost fee")
	}

	// 校验 TransferChecked
	checkedInst := decodedTx.TransferCheckedInstructions[0]
	rewardNativeKey, _ := solana.PublicKeyFromBase58(txInfo.RewardAccount)
	tokenMintKey, _ := solana.PublicKeyFromBase58(txInfo.Mint)
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeKey, tokenMintKey)

	if !checkedInst.FromTokenAccount.Equals(rewardTokenAccount) {
		log.Errorf("%s TransferChecked from[%v]不是rewardTokenAccount[%v]", prefix, checkedInst.FromTokenAccount, rewardTokenAccount)
		return errors.New("TransferChecked from account must be reward token account")
	}
	userTokenAccount, _, _ := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintKey)
	if !checkedInst.ToTokenAccount.Equals(userTokenAccount) {
		log.Errorf("%s TransferChecked to[%v]不是用户token账户[%v]", prefix, checkedInst.ToTokenAccount, userTokenAccount)
		return errors.New("TransferChecked to account must be user token account")
	}
	if checkedInst.TokenMintAccount.String() != txInfo.Mint {
		log.Errorf("%s TransferChecked mint[%v]不匹配[%v]", prefix, checkedInst.TokenMintAccount, txInfo.Mint)
		return errors.New("TransferChecked token mint not match")
	}
	if checkedInst.Amount != uint64(txInfo.TotalReward) {
		log.Errorf("%s TransferChecked金额[%d]不等于totalReward[%.0f]", prefix, checkedInst.Amount, txInfo.TotalReward)
		return errors.New("TransferChecked amount does not match total reward")
	}
	if checkedInst.Decimals != uint8(txInfo.Decimals) {
		log.Errorf("%s TransferChecked精度[%d]不匹配[%d]", prefix, checkedInst.Decimals, txInfo.Decimals)
		return errors.New("TransferChecked decimals not match token config")
	}

	return nil
}
