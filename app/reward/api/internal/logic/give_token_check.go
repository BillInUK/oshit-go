package logic

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
)

func (l *GiveTokenLogic) decodeSOLTx(rpcClient *rpc.Client, ruleTokenMintAccount, transferToNativeAccount solana.PublicKey, needCreateTokenAccount bool, tx *solana.Transaction) (*entity.DecodedSolanaTransaction, error) {
	var decodedTx entity.DecodedSolanaTransaction

	decodedTx.FromNativeAccount = tx.Message.AccountKeys[0]
	decodedTx.Signatures = tx.Signatures
	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(decodedTx.FromNativeAccount, ruleTokenMintAccount)
	if err != nil {
		log.Errorf("无法找到账户[%v]的token account", decodedTx.FromNativeAccount)
		return nil, errors.New("can not find token account by transaction from native account")
	}
	decodedTx.FromTokenAccount = fromTokenAccount

	for index, inst := range tx.Message.Instructions {
		programId, err := tx.ResolveProgramIDIndex(inst.ProgramIDIndex)
		if err != nil {
			return nil, fmt.Errorf("can not decode program id of instruction %d", index)
		}

		// 如果是 ComputeUnitPrice或者ComputeUnitLimit
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
		// 如果是AssociatedTokenAccount
		if programId.Equals(solana.SPLAssociatedTokenAccountProgramID) {
			continue
		}
		// 如果是TransferChecked
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
				// TODO: 改为从dubbo获取
				fromNativeAccount, err := app_utils.QueryNativeAccountByTokenAccount(rpcClient, l.db, transferChecked.Accounts.Get(0).PublicKey)
				if err != nil {
					account := transferChecked.Accounts.Get(0).PublicKey
					return nil, fmt.Errorf("decode transaction failed,can not get native account of from token account[%v],error:%v", account, err)
				}
				// 防止收款地址没有tokenAccount的情况发生
				// TODO: 改为从dubbo获取
				toNativeAccount, err := app_utils.QueryNativeAccountByTokenAccount(rpcClient, l.db, transferChecked.Accounts.Get(2).PublicKey)
				if err != nil && needCreateTokenAccount {
					if needCreateTokenAccount {
						toNativeAccount = &transferToNativeAccount
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
		}
		// 如果是Transfer
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
		}
	}
	return &decodedTx, nil
}

// checkRewardFromInst 检查官网转账的发送token指令是否符合规则
func (l *GiveTokenLogic) checkRewardFromInst(
	decodedInst entity.DecodedSolTransferCheckedInst,
	txFromNativeAccount solana.PublicKey,
	transferAmount uint64,
	matchRewardRule bool) error {
	// 发放token的地址是否是官方指定的地址
	if decodedInst.FromTokenAccount.String() != l.serviceConfig.RewardTokenAccount {
		return errors.New("the address for distributing rewards is not the officially designated address")
	}
	//// to地址是否在非奖励地址里面
	//rewardExcludeRecord, err := QueryRewardExcludeAccount(decodedInst.ToTokenAccount.String())
	//if err != nil {
	//	return err
	//}
	//if rewardExcludeRecord != nil {
	//	return errors.New("given token address is excluded")
	//}
	// token地址是否相同
	if decodedInst.TokenMintAccount.String() != l.serviceConfig.TokenMintAccount {
		return errors.New("the token mint account is not the officially designated address")
	}
	// to地址必须是支付了手续费和dex费的地址
	toTokenAccount0, _, _ := solana.FindAssociatedTokenAddress(txFromNativeAccount, decodedInst.TokenMintAccount)
	if !decodedInst.ToTokenAccount.Equals(toTokenAccount0) {
		log.Infof("check reward instruction failed to address %v not equal %v\n", decodedInst.ToTokenAccount, toTokenAccount0)
		return errors.New("the reward receipt address is not the transaction fee payer")
	}
	// owner address必须和奖励的native account一致
	if decodedInst.OwnerNativeAccount.String() != l.serviceConfig.RewardNativeAccount {
		return errors.New("transfer check reward instruction owner address not reward native address")
	}
	// 奖励金额不能大于费率规定的金额
	amountByRule := min(l.serviceConfig.MaxValidReward, float64(transferAmount)*l.serviceConfig.RewardRate)
	if matchRewardRule {
		amountByRule = min(l.serviceConfig.MaxValidReward, float64(transferAmount)*l.serviceConfig.ValidRate)
	}
	log.Infof("官方转账 - 转账金额 [%d] 交易指令内的奖励金额 [%d] 按照规则应该奖励金额[%d]", transferAmount, decodedInst.Amount, amountByRule)
	if decodedInst.Amount > uint64(amountByRule) {
		log.Errorf("奖励token的额度[%d]大于规则的费率规定额度[%d]", decodedInst.Amount, amountByRule)
		return fmt.Errorf("the requested amount of reward tokens in the transaction[%d], is greater than the amount received according to the rules [%d]", decodedInst.Amount, uint64(amountByRule))
	}
	// 进制是否跟规则规定的一样
	if decodedInst.Decimals != uint8(l.serviceConfig.Decimal) {
		return errors.New("token decimal not equal  officially rule")
	}
	return nil
}

// checkRewardInvitersInst 检查官网转账奖励给上级邀请人token指令是否符合规则
func (l *GiveTokenLogic) checkRewardInvitersInst(decodedInst entity.DecodedSolTransferCheckedInst, rewardAmount uint64) error {
	prefix := fmt.Sprintf("%s - %s -", l.prefix, "解析奖励邀请人token指令")
	// to地址是否在非奖励地址里面
	//rewardExcludeRecord, err := QueryRewardExcludeAccount(decodedInst.ToTokenAccount.String())
	//if err != nil {
	//	log.Errorf("%s 查询排除地址错误: %v", prefix, err)
	//	return err
	//}
	//if rewardExcludeRecord != nil {
	//	log.Errorf("%s 邀请人地址被排除:%v", prefix, err)
	//	return errors.New("reward inviter token address is excluded")
	//}
	// from地址必须是规则规定的奖励地址
	if decodedInst.FromTokenAccount.String() != l.serviceConfig.RewardTokenAccount {
		log.Errorf("%s 奖励指令的from地址[%v]与规则规定的[%v]不一致:", decodedInst.FromTokenAccount, l.serviceConfig.RewardTokenAccount)
		return errors.New("reward instruction from address not rule specified")
	}
	// token地址必须与规则规定的相同
	if decodedInst.TokenMintAccount.String() != l.serviceConfig.TokenMintAccount {
		log.Errorf("%s token地址[%v]与规则规定的[%v]不一致:", prefix, decodedInst.TokenMintAccount, l.serviceConfig.TokenMintAccount)
		return errors.New("the token mint account is not the officially designated address")
	}
	// owner address必须跟转账地址的native account一样
	if decodedInst.OwnerNativeAccount.String() != l.serviceConfig.RewardNativeAccount {
		log.Errorf("%s owner地址[%v]与规则规定的[%v]不一致:", prefix, decodedInst.OwnerNativeAccount, l.serviceConfig.RewardNativeAccount)
		return errors.New("token transfer owner address not equal from native address")
	}
	// 金额必须一样
	if decodedInst.Amount != rewardAmount {
		log.Errorf("%s 指令当中奖励邀请人 %v 的金额 %d 与按照规则应该获取的不一致 %d:", prefix, decodedInst.ToNativeAccount, decodedInst.Amount, rewardAmount)
		return fmt.Errorf("the reward amount of [%d] given to the referrer does not match the [%d] they should have received according to the rules", decodedInst.Amount, rewardAmount)
	}
	// 进制是否跟规则规定的一样
	if decodedInst.Decimals != uint8(l.serviceConfig.Decimal) {
		log.Errorf("%s token精度 %d 和规则规定的 %d 不一致:", prefix, decodedInst.Decimals, l.serviceConfig.Decimal)
		return errors.New("token decimal not equal  officially rule")
	}
	return nil
}

// checkToReceiptInst 检查官网转账转出token指令是否正确
func (l *GiveTokenLogic) checkToReceiptInst(decodedInst entity.DecodedSolTransferCheckedInst,
	txFromNativeAccount solana.PublicKey) error {
	var err error
	// to地址是否在非奖励地址里面
	//rewardExcludeRecord, err := QueryRewardExcludeAccount(decodedInst.ToTokenAccount.String())
	//if err != nil {
	//	log.Error("[sol] 解析从官方转token交易指令,错误:", err)
	//	return err
	//}
	//if rewardExcludeRecord != nil {
	//	return errors.New("given token address is excluded")
	//}
	// token地址必须与规则规定的相同
	if decodedInst.TokenMintAccount.String() != l.serviceConfig.TokenMintAccount {
		log.Error("[sol] 解析从官方转token交易指令,错误:", err)
		return errors.New("the token mint account is not the officially designated address")
	}
	// from地址必须是支付了手续费和dex费的地址
	fromTokenAccount0, _, _ := solana.FindAssociatedTokenAddress(txFromNativeAccount, decodedInst.TokenMintAccount)
	if err != nil {
		log.Error("[sol] 解析从官方转token交易指令,错误:", err)
		return errors.New("can not get spl token account for token transfer from address")
	}
	if !decodedInst.FromTokenAccount.Equals(fromTokenAccount0) {
		log.Error("[sol] 解析从官方转token交易指令,错误:", err)
		return errors.New("the reward receipt address is not the transaction fee payer")
	}
	// owner address必须跟转账地址的native account一样
	if !decodedInst.OwnerNativeAccount.Equals(txFromNativeAccount) {
		log.Errorf("[sol] 解析从官方转token交易指令错误，转token指令的owner地址[%s]和发起交易的地址[%s]不一致", decodedInst.OwnerNativeAccount.String(), txFromNativeAccount.String())
		return fmt.Errorf("token transfer owner address [%s] not equal from native address[%s]", decodedInst.OwnerNativeAccount.String(), txFromNativeAccount.String())
	}
	// 进制是否跟规则规定的一样
	if decodedInst.Decimals != uint8(l.serviceConfig.Decimal) {
		log.Error("[sol] 解析从官方转token交易指令,错误:", err)
		return errors.New("token decimal not equal  officially rule")
	}
	return nil
}

// checkSOLTx 解析交易后检查交易内的参数
func (l *GiveTokenLogic) checkSOLTx(
	decodedTx *entity.DecodedSolanaTransaction,
	upInvitersInfo []model.InviteRelation,
	receiptTokenAccount solana.PublicKey,
	valid bool) (*entity.DecodedServiceTransaction, error) {
	var err error
	var decodedServiceTx *entity.DecodedServiceTransaction
	// 必须包含1条转给dex地址的transfer指令
	if len(decodedTx.TransferInstructions) != 1 {
		log.Errorf("官方转账获取奖励 - 必须包含至1条转sol到dex的指令")
		return nil, errors.New("decoded solana transaction error: can not detect any transfer decodedInst to dex")
	}
	// transfer checked的指令需要包含: 转账指令,平台奖励转账人,如果转账人有上级，还需要包含平台转账给转账人上级的指令
	if len(decodedTx.TransferCheckedInstructions) != len(upInvitersInfo)+2 {
		log.Errorf("官方转账获取奖励 - transfer checked 转账指令数量[%d] != 需要奖励上级邀请人的数量[%d] + 2", len(decodedTx.TransferCheckedInstructions), len(upInvitersInfo)+2)
		return nil, errors.New("decoded solana transaction error: transfer checked instruction number !=  reward claims+2 ")
	}
	// 检查Transfer指令的地址和金额
	transferInst := decodedTx.TransferInstructions[0]
	if transferInst.ToNativeAccount.String() != l.serviceConfig.DexNativeAccount {
		log.Errorf("官方转账获取奖励 - 解析交易错误: solana收款地址[%v]不是规则要求的地址[%v]", transferInst.ToNativeAccount, l.serviceConfig.DexNativeAccount)
		return nil, errors.New("transfer receipt account must be dex account")
	}
	if transferInst.FromNativeAccount != decodedTx.FromNativeAccount {
		log.Errorf("官方转账获取奖励 -  解析交易错误: solana发送地址[%v]不是交易发起地址[%v]", transferInst.FromNativeAccount, decodedTx.FromNativeAccount)
		return nil, errors.New("transfer funding account must be transaction fee payer")
	}

	// 检查TransferChecked指令的地址和金额
	transferCheckedMap := make(map[solana.PublicKey]entity.DecodedSolTransferCheckedInst)
	inviterRecordMap := make(map[string]model.InviteRelation)
	inviterClaimMap := make(map[string]uint64)

	// 将所有的transfer checked指令存放到map里面,key-ToTokenAccount value-DecodedSolTransferCheckedInst
	for _, transfer := range decodedTx.TransferCheckedInstructions {
		transferCheckedMap[transfer.ToTokenAccount] = transfer
	}
	// 所有的邀请人记录存放到map里面,key-InviterTokenAccount value=SolDetermineInviteRecord
	for _, record := range upInvitersInfo {
		inviterRecordMap[record.InviterTokenAccount] = record
	}

	// 解析转账地址转出token的指令
	var transferTokenAmount uint64 = 0
	var transferTokenInstCount = 0
	for _, decodedInst := range decodedTx.TransferCheckedInstructions {
		log.Infof("官方转账获取奖励 - FromTokenAccount [%v] to TokenAccount[%v] TokenMintAccount[%v] Amount[%d] Decimals [%d]",
			decodedInst.FromTokenAccount, decodedInst.ToTokenAccount, decodedInst.TokenMintAccount, decodedInst.Amount, decodedInst.Decimals)
		// 如果From和To跟前端传入的参数一样，则认为是转账指令
		if decodedInst.FromTokenAccount.Equals(decodedTx.FromTokenAccount) && decodedInst.ToTokenAccount.Equals(receiptTokenAccount) {
			if err := l.checkToReceiptInst(decodedInst, decodedTx.FromNativeAccount); err != nil {
				return nil, err
			}
			transferTokenInstCount++
			transferTokenAmount = decodedInst.Amount
			continue
		}
	}
	if transferTokenAmount == 0 || transferTokenInstCount == 0 {
		log.Errorf("官方转账获取奖励 - 没有找到任何转出token的指令")
		return nil, errors.New("can not find any transfer token instruction")
	}
	if transferTokenInstCount != 1 {
		log.Errorf("官方转账获取奖励 - 只允许1条转出token的指令")
		return nil, errors.New("transfer token instruction count must be 1")
	}

	// 校验奖励转账人的指令里面的地址和金额是否正确
	var totalRewardAmount uint64 = 0
	var rewardTxFromAmount uint64 = 0
	var rewardTxFromTokenInstCount = 0
	for toTokenAccount, decodedInst := range transferCheckedMap {
		// 如果token的收款地址为转账发起人的TokenAccount，则认为是奖励转账人token指令
		if toTokenAccount.Equals(decodedTx.FromTokenAccount) {
			err := l.checkRewardFromInst(decodedInst, decodedTx.FromNativeAccount, transferTokenAmount, valid)
			if err != nil {
				log.Error("检查奖励转账人的指令错误:", err)
				return nil, err
			}
			rewardTxFromTokenInstCount++
			rewardTxFromAmount = decodedInst.Amount
			totalRewardAmount += rewardTxFromAmount
			break
		}
	}
	// 只能包含1条奖励转账地址的指令
	if rewardTxFromTokenInstCount != 1 {
		log.Errorf("官方转账获取奖励 - 找不到任何奖励转账人的指令")
		return nil, errors.New("can not find transfer token instruction in transaction or transfer token instruction count > 1")
	}

	// 计算出来每个级别的上级应该拿到的奖励
	for index, inviteRecord := range upInvitersInfo {
		rewardInviterAmount := uint64(float64(rewardTxFromAmount) * l.srvCtx.LevelRatio[index].Ratio)
		inviterClaimMap[inviteRecord.InviterTokenAccount] = rewardInviterAmount
	}
	// 校验奖励上级邀请人的指令里面的地址和金额是否正确,检查转账指令是否正确
	for toTokenAccount, decodedInst := range transferCheckedMap {
		// 如果收款地址是inviterRecordMap里面的地址，则认为是奖励转账人的上级邀请人的指令
		_, exist := inviterRecordMap[toTokenAccount.String()]
		if exist {
			// 如果查询出来应该奖励的邀请人不包含在交易里面
			rewardInviterAmount, claimExist := inviterClaimMap[toTokenAccount.String()]
			if !claimExist {
				log.Errorf("官方转账获取奖励 - 交易内未包含应该奖励邀请人[%s]的指令", toTokenAccount.String())
				return nil, errors.New("the transaction does not include the inviter that should receive the reward")
			}
			// 如果指令的From地址是规则规定的奖励发放地址并且收款地址是邀请人地址，则认为是奖励邀请人指令
			if decodedInst.FromTokenAccount.String() == l.serviceConfig.RewardTokenAccount &&
				decodedInst.ToTokenAccount.Equals(toTokenAccount) {
				if err := l.checkRewardInvitersInst(decodedInst, rewardInviterAmount); err != nil {
					log.Errorf("官方转账获取奖励 - 检查奖励邀请人指令错误[%v]", err)
					return nil, err
				}
				totalRewardAmount += decodedInst.Amount
			}
			continue
		}
	}

	// 检查成本费用是否符合规则
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(l.ctx)
	if err != nil {
		log.Errorf("官方转账获取奖励 - 无法获取token兑换solana的价格: %v", err)
		return nil, errors.New("can not get token quote sol price")
	}
	requiredDexFee := float64(totalRewardAmount) / l.srvCtx.TokenDecimal * quoteSOLPrice * float64(solana.LAMPORTS_PER_SOL)
	var minRequiredDexFee = requiredDexFee * (1 - l.srvCtx.FeeTolerance.MaxLessRate)
	log.Infof("官方转账获取奖励 要求dex fee %d 最小 dex fee %d", uint64(requiredDexFee), uint64(minRequiredDexFee))
	if transferInst.Amount < uint64(minRequiredDexFee) {
		log.Errorf("官方转账获取奖励 解析交易错误: 转给dex的手续费 [%d] 小于规则要求的 [%d]", transferInst.Amount, uint64(minRequiredDexFee))
		return nil, errors.New("transfer funding less than serviceConfig required")
	}

	if decodedServiceTx, err = app_utils.DecodeServiceTransaction(decodedTx); err != nil {
		log.Errorf("官方转账获取奖励 - 无法将solana交易转成业务交易,错误: %v", err)
		return nil, fmt.Errorf("convert decoded solana transaction to service transaction error: %v", err)
	}

	log.Infof("官方转账获取奖励 - 解析交易完成")

	return decodedServiceTx, nil
}
