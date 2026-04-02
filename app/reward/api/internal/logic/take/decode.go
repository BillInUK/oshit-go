package take

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

// decodeSOLTx 解析solana交易
func (l *TakeTokenLogic) decodeSOLTx(givenTokenInfo *types.TakeTokenTxInfo, tx *solana.Transaction) (*entity.DecodedSolanaTransaction, error) {
	var decodedTx entity.DecodedSolanaTransaction
	tokenMintAccount, _ := solana.PublicKeyFromBase58(givenTokenInfo.TokenMintAccount)
	decodedTx.FromNativeAccount = tx.Message.AccountKeys[0]
	decodedTx.Signatures = tx.Signatures
	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, tokenMintAccount)
	if err != nil {
		log.Errorf("官方转账获取奖励 - 无法找到账户[%v]的token account", decodedTx.FromNativeAccount)
		return nil, errors.New("can not find token account by transaction from native account")
	}
	decodedTx.FromTokenAccount = fromTokenAccount

	// 将邀请人信息和被奖励的人的信息放到map里面 key-TokenAccount value-RewardItem
	var rewardItemMap = map[string]types.RewardTokenItem{}
	rewardTA, _, _ := solana.FindAssociatedTokenAddress(solana.MPK(givenTokenInfo.RewardInfo.NativeAccount), tokenMintAccount)
	rewardItemMap[rewardTA.String()] = givenTokenInfo.RewardInfo
	for _, item := range givenTokenInfo.RewardInviterInfo {
		itemTA, _, _ := solana.FindAssociatedTokenAddress(solana.MPK(item.NativeAccount), tokenMintAccount)
		rewardItemMap[itemTA.String()] = item
	}

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
		} else if programId.Equals(solana.SPLAssociatedTokenAccountProgramID) {
			continue
		} else if programId.Equals(solana.TokenProgramID) {
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedTokenInst, err := token.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			if transferChecked, ok := decodedTokenInst.Impl.(*token.TransferChecked); ok {
				// TODO: 使用dubbo实现，要求base模块先查询本地数据库，如果本地数据库不存在则调用rpc查询
				fromNativeAccount, err := app_utils.QueryNativeAccountByTokenAccount(l.rpcClient, l.db, transferChecked.Accounts.Get(0).PublicKey)
				if err != nil {
					account := transferChecked.Accounts.Get(0).PublicKey
					return nil, fmt.Errorf("decode transaction failed,can not get native account of from token account[%v],error:%v", account, err)
				}
				toNativeAccount, err := app_utils.QueryNativeAccountByTokenAccount(l.rpcClient, l.db, transferChecked.Accounts.Get(2).PublicKey)
				if err != nil {
					account := transferChecked.Accounts.Get(2).PublicKey
					// 如果存在于give token info 里面的reward item
					rewardItem, exist := rewardItemMap[account.String()]
					if exist {
						toNativeAccount0, _ := solana.PublicKeyFromBase58(rewardItem.NativeAccount)
						toNativeAccount = &toNativeAccount0
					} else {
						return nil, fmt.Errorf("decode transaction failed,can not get native account of to token account[%v],error:%v", account, err)
					}
				}
				inst := entity.DecodedSolTransferCheckedInst{
					FromTokenAccount:   transferChecked.Accounts.Get(0).PublicKey,
					FromNativeAccount:  *fromNativeAccount,
					TokenMintAccount:   transferChecked.Accounts.Get(1).PublicKey,
					ToTokenAccount:     transferChecked.Accounts.Get(2).PublicKey,
					ToNativeAccount:    *toNativeAccount,
					OwnerNativeAccount: transferChecked.Accounts.Get(3).PublicKey,
					Amount:             *transferChecked.Amount,
					Decimals:           *transferChecked.Decimals,
				}
				decodedTx.TransferCheckedInstructions = append(decodedTx.TransferCheckedInstructions, inst)
			} else {
				return nil, errors.New("decode transfer checked instruction error")
			}
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
			if transfer, ok := decodedSystemInst.Impl.(*system.Transfer); ok {
				inst := entity.DecodedSolTransferInst{
					FromNativeAccount: transfer.GetFundingAccount().PublicKey,
					ToNativeAccount:   transfer.GetRecipientAccount().PublicKey,
					Amount:            *transfer.Lamports,
				}
				decodedTx.TransferInstructions = append(decodedTx.TransferInstructions, inst)
			} else {
				return nil, errors.New("decode transfer instruction error")
			}
			continue
		} else if programId.Equals(l.LightHouseAddress) {
			// light house 安全校验什么都不做
		} else {
			return nil, fmt.Errorf("unsupport program id %v", programId)
		}
	}
	return &decodedTx, nil
}

// checkDecodedSOLTx 检查solana交易当中的指令是否符合规则
func (l *TakeTokenLogic) checkDecodedSOLTx(txInfo *types.TakeTokenTxInfo, decodedTx *entity.DecodedSolanaTransaction) (*entity.DecodedServiceTransaction, error) {
	var err error
	var prefix = fmt.Sprintf("%s 检查解码后的solana交易 -", l.prefix)
	var decodedServiceTx *entity.DecodedServiceTransaction
	rewardClaims := txInfo.Claims

	// 必须包含1条转给dex地址的transfer指令
	if len(decodedTx.TransferInstructions) != 1 {
		log.Errorf("%s 交易指令数量错误: 必须包含1条转sol到dex的指令", prefix)
		return nil, errors.New("decoded solana transaction error: can not detect any transfer decodedInst to dex")
	}

	// TransferChecked的指令需要包含: 平台奖励转账人,如果转账人有上级，还需要包含平台转账给转账人上级的指令
	if len(decodedTx.TransferCheckedInstructions) != len(rewardClaims)+1 {
		log.Errorf("%s  transfer checked 转账指令数量[%d] != 需要奖励上级邀请人的数量[%d] + 1", prefix, len(decodedTx.TransferCheckedInstructions), len(rewardClaims))
		log.Errorf("%s  需要奖励的上级为 %v", prefix, txInfo.RewardInviterInfo)
		return nil, errors.New("decoded solana transaction error: transfer checked instruction number !=  reward claims+1 ")
	}

	// 检查发送SOL到DEX地址的Transfer的地址和金额
	transferInst := decodedTx.TransferInstructions[0]
	if transferInst.ToNativeAccount.String() != txInfo.DexNativeAccount {
		log.Errorf("%s 解析交易错误: solana收款地址[%v]不是规则要求的地址[%v]", prefix, transferInst.ToNativeAccount, l.serviceConfig.DexNativeAccount)
		return nil, errors.New("transfer receipt account must be dex account")
	}
	if transferInst.FromNativeAccount != decodedTx.FromNativeAccount {
		log.Errorf("%s 解析交易错误: solana发送地址[%v]不是交易发起地址[%v]", prefix, transferInst.FromNativeAccount, decodedTx.FromNativeAccount)
		return nil, errors.New("transfer funding account must be transaction fee payer")
	}
	requiredDexFee := txInfo.QuotedSOLAmount
	var minRequiredDexFee = requiredDexFee * (1 - l.srvCtx.FeeTolerance.MaxLessRate)
	log.Infof("%s 要求dex fee %d 最小 dex fee %d", prefix, uint64(requiredDexFee), uint64(minRequiredDexFee))
	if transferInst.Amount < uint64(minRequiredDexFee) {
		log.Errorf("%s 解析交易错误: 转给dex的手续费 [%d] 小于规则要求的 [%d]", prefix, transferInst.Amount, uint64(minRequiredDexFee))
		return nil, errors.New("transfer funding less than config required")
	}
	// 检查TransferChecked指令的地址和金额
	tokenMintPubKey, _ := solana.PublicKeyFromBase58(txInfo.TokenMintAccount)
	transferCheckedMap := make(map[solana.PublicKey]entity.DecodedSolTransferCheckedInst)
	inviterRewardMap := make(map[string]types.RewardTokenItem)
	inviterClaimMap := make(map[string]uint64)

	// 将所有的TransferChecked指令存放到map里面,key-ToTokenAccount value-DecodedSolTransferCheckedInst
	for _, transfer := range decodedTx.TransferCheckedInstructions {
		transferCheckedMap[transfer.ToTokenAccount] = transfer
	}
	// 所有的邀请人记录存放到map里面,key-TokenAccount value-RewardTokenItem
	for _, record := range txInfo.RewardInviterInfo {
		inviterTA, _, _ := solana.FindAssociatedTokenAddress(solana.MPK(record.NativeAccount), tokenMintPubKey)
		inviterRewardMap[inviterTA.String()] = record
	}

	// 校验奖励领取地址的指令里面的地址和金额是否正确
	var rewardTxFromAmount uint64 = 0
	var rewardTxFromTokenInstCount = 0
	for toTokenAccount, decodedInst := range transferCheckedMap {
		// 如果token的收款地址为转账发起人的TokenAccount，则认为是奖励转账人token指令
		if toTokenAccount.Equals(decodedTx.FromTokenAccount) {
			log.Infof("%s 邀请码[%s] 有效 [%t]", prefix, txInfo.InviteCode, txInfo.InviteCodeValid)
			err := l.checkRewardInst(decodedInst, decodedTx.FromNativeAccount, txInfo.InviteCodeValid)
			if err != nil {
				log.Errorf("%s 检查奖励领取地址的指令错误: %v", prefix, err)
				return nil, fmt.Errorf("check official given token reward instruction error: %v", err)
			}
			rewardTxFromTokenInstCount++
			rewardTxFromAmount = decodedInst.Amount
			break
		}
	}

	// 只能包含1条奖励转账地址的指令
	if rewardTxFromTokenInstCount != 1 {
		log.Errorf("%s 找不到任何奖励转账人的指令", prefix)
		return nil, errors.New("can not find transfer token instruction in transaction or transfer token instruction count > 1")
	}

	// 计算出来每个级别的上级应该拿到的奖励
	for index, inviteRecord := range txInfo.RewardInviterInfo {
		rewardInviterAmount := uint64(float64(rewardTxFromAmount) * rewardClaims[index].Ratio)
		inviterTA, _, _ := solana.FindAssociatedTokenAddress(solana.MPK(inviteRecord.NativeAccount), tokenMintPubKey)
		inviterClaimMap[inviterTA.String()] = rewardInviterAmount
	}

	// 校验奖励上级邀请人的指令里面的地址和金额是否正确,检查转账指令是否正确
	for toTokenAccount, decodedInst := range transferCheckedMap {
		if !toTokenAccount.Equals(decodedTx.FromTokenAccount) {
			// 如果收款地址是inviterRecordMap里面的地址，则认为是奖励转账人的上级邀请人的指令
			_, exist := inviterRewardMap[toTokenAccount.String()]
			if exist {
				// 如果查询出来应该奖励的邀请人不包含在交易里面
				rewardInviterAmount, claimExist := inviterClaimMap[toTokenAccount.String()]
				if !claimExist {
					log.Errorf("%s 交易内未包含应该奖励邀请人[%s]的指令", prefix, toTokenAccount.String())
					return nil, errors.New("the transaction does not include the inviter that should receive the reward")
				}
				if err := l.checkRewardInviterInst(decodedInst, rewardInviterAmount); err != nil {
					log.Errorf("%s 检查奖励邀请人指令错误[%v]", prefix, err)
					return nil, err
				}
				continue
			}
		}
	}

	if decodedServiceTx, err = app_utils.DecodeServiceTransaction(decodedTx); err != nil {
		return nil, fmt.Errorf("convert decoded solana transaction to service transaction error: %v", err)
	}

	log.Infof("%s 解析完成", prefix)

	return decodedServiceTx, nil
}

// checkRewardInst 检查官网领取奖励的发送token指令是否符合规则
func (l *TakeTokenLogic) checkRewardInst(decodedInst entity.DecodedSolTransferCheckedInst, txFromNativeAccount solana.PublicKey, inviteCodeValid bool) error {
	prefix := fmt.Sprintf("%s 检查奖励领取人指令 -", l.prefix)

	// 发放token的地址是否是官方指定的地址
	if decodedInst.FromTokenAccount.String() != l.serviceConfig.RewardTokenAccount {
		log.Errorf("%s 错误: token account [%v]  与规则 [%v] 不一致", prefix, decodedInst.FromTokenAccount, l.serviceConfig.RewardTokenAccount)
		return errors.New("reward instruction error: send reward token account not match service config")
	}
	// token地址是否相同
	if decodedInst.TokenMintAccount.String() != l.serviceConfig.TokenMintAccount {
		log.Errorf("%s 错误: token mint account [%v]  与规则 [%v] 不一致", prefix, decodedInst.TokenMintAccount, l.serviceConfig.TokenMintAccount)
		return errors.New("reward instruction error: send reward token mint account not match service config")
	}
	// to地址必须是支付了手续费和dex费的地址
	toTokenAccount0, _, _ := solana.FindAssociatedTokenAddress(txFromNativeAccount, decodedInst.TokenMintAccount)
	if !decodedInst.ToTokenAccount.Equals(toTokenAccount0) {
		log.Errorf("%s 错误: destination [%v]  与 fee payer [%v] 不一致", prefix, decodedInst.ToTokenAccount, toTokenAccount0)
		return errors.New("reward instruction error: the destination is not the transaction fee payer")
	}
	// TODO: 校验Owner地址

	// 金额是否跟规则规定的一样
	if inviteCodeValid && decodedInst.Amount != uint64(l.serviceConfig.InviteAmount) {
		log.Errorf("%s 错误: 奖励金额[%d]和规则规定的金额[%d]不一致", prefix, decodedInst.Amount, uint64(l.serviceConfig.InviteAmount))
		return errors.New("reward amount not equal officially rule")
	}
	if !inviteCodeValid && decodedInst.Amount != uint64(l.serviceConfig.Amount) {
		log.Errorf("%s 错误: 奖励金额[%d]和规则规定的金额[%d]不一致", prefix, decodedInst.Amount, uint64(l.serviceConfig.Amount))
		return errors.New("reward amount not equal officially rule")
	}

	// 进制是否跟规则规定的一样
	if decodedInst.Decimals != uint8(l.serviceConfig.Decimals) {
		log.Errorf("%s 错误: 奖励金额进制[%d]和规则规定的金额进制[%d]不一致", prefix, decodedInst.Decimals, uint8(l.serviceConfig.Decimals))
		return errors.New("token decimal not equal  officially rule")
	}

	return nil
}

// checkRewardInviterInst 检查官网领取奖励的奖励给上级邀请人token指令是否符合规则
func (l *TakeTokenLogic) checkRewardInviterInst(decodedInst entity.DecodedSolTransferCheckedInst, rewardAmount uint64) error {
	prefix := fmt.Sprintf("%s 检查奖励邀请人指令 -", l.prefix)

	// from地址必须是规则规定的奖励地址
	if decodedInst.FromTokenAccount.String() != l.serviceConfig.RewardTokenAccount {
		log.Errorf("%s 错误: token account [%v]  与规则 [%v] 不一致", prefix, decodedInst.FromTokenAccount, l.serviceConfig.RewardTokenAccount)
		return errors.New("reward inviter instruction error: send reward token account not match service config")
	}
	// token地址必须与规则规定的相同
	if decodedInst.TokenMintAccount.String() != l.serviceConfig.TokenMintAccount {
		log.Errorf("%s 错误: token mint account [%v]  与规则 [%v] 不一致", prefix, decodedInst.TokenMintAccount, l.serviceConfig.TokenMintAccount)
		return errors.New("reward inviter instruction error: send reward token mint account not match service config")
	}
	// owner address必须跟转账地址的native account一样
	if decodedInst.OwnerNativeAccount.String() != l.serviceConfig.RewardNativeAccount {
		log.Error("%s 解析从官方转token交易指令,错误:", prefix)
		return fmt.Errorf("token transfer owner address [%s]not equal from native address[%s]", decodedInst.OwnerNativeAccount.String(), l.serviceConfig.RewardNativeAccount)
	}
	// 金额必须一样
	if decodedInst.Amount != rewardAmount {
		log.Errorf("[sol] 解析从官方转token交易指令,错误:奖励地址[%v]的金额[%d]和规则规定的[%d]不一致",
			decodedInst.ToNativeAccount, decodedInst.Amount, rewardAmount)
		return fmt.Errorf("the reward amount [%d] is inconsistent with the rules specified [%d]", decodedInst.Amount, rewardAmount)
	}
	// 进制是否跟规则规定的一样
	if decodedInst.Decimals != uint8(l.serviceConfig.Decimals) {
		log.Error("[sol] 解析从官方转token交易指令,错误:")
		return errors.New("token decimal not equal  officially serviceConfig")
	}
	return nil
}
