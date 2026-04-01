package utils

import (
	"encoding/hex"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
)

type PreCheckedTx struct {
	From  solana.PublicKey
	TxId  solana.Signature
	SOLTx solana.Transaction
}

func CalDecodedTxFee(sigNum uint64, decodedTx entity.DecodedSolanaTransaction) uint64 {
	return utils.CalcGasFee(sigNum, 5000, decodedTx.ComputeUnitPrice, decodedTx.ComputeUnitLimit, decodedTx.ComputeUnitLimit != 0)
}

// QueryNativeAccountInfoByTokenAccount 查询原生账户信息
func QueryNativeAccountInfoByTokenAccount(db *gorm.DB, tokenAccount string) (*model.NativeAccountInfo, error) {
	var record model.NativeAccountInfo
	err := db.Where("token_account = ?", tokenAccount).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func QueryNativeAccountByTokenAccount(rpcClient *rpc.Client, db *gorm.DB, tokenAccount solana.PublicKey) (*solana.PublicKey, error) {
	// 先查询NativeAccountInfo，如果查询不到再查询rpc
	accountInfo, err := QueryNativeAccountInfoByTokenAccount(db, tokenAccount.String())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if accountInfo != nil {
		nativeAccount, _ := solana.PublicKeyFromBase58(accountInfo.NativeAccount)
		return &nativeAccount, nil
	}
	return utils.GetSPLTokenAccountOwner(rpcClient, tokenAccount)
}

func PreCheckEncodedTx(encodedTx string) (*PreCheckedTx, error) {
	// 解析交易
	txBytes, err := hex.DecodeString(encodedTx)
	if err != nil {
		return nil, fmt.Errorf("decode hex encoded transaction error: %v", err)
	}
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return nil, fmt.Errorf("decode hex encoded transaction to solana transaction error: %v", err)
	}
	// Account0是交易的发起地址，同时也是转账token的地址，也是转sol到dex的地址，同时也是手续费的支付地址
	txFromNativeAccount, err := tx.Message.Account(0)
	if err != nil {
		return nil, fmt.Errorf("can not get any signer from transaction error: %v", err)
	}
	messageBin, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal transaction to binary error: %v", err)
	}
	// 验证交易签名
	if !txFromNativeAccount.Verify(messageBin, tx.Signatures[0]) {
		return nil, fmt.Errorf("verify transaction signer failed")
	}
	return &PreCheckedTx{
		From:  txFromNativeAccount,
		TxId:  tx.Signatures[0],
		SOLTx: *tx,
	}, nil
}

// DecodeSolanaTransaction 解析solana交易
func DecodeSolanaTransaction(rpcClient *rpc.Client, db *gorm.DB, tx *solana.Transaction, txID solana.Signature) (*entity.DecodedSolanaTransaction, error) {
	var decodedTx entity.DecodedSolanaTransaction
	decodedTx.RefBlockHash = tx.Message.RecentBlockhash
	decodedTx.FromNativeAccount = tx.Message.AccountKeys[0]
	decodedTx.Signatures = tx.Signatures
	decodedTx.TxID = txID
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
		associatedMap := make(map[string]entity.PubKeyPair)
		// 如果是AssociatedTokenAccount
		if programId.Equals(solana.SPLAssociatedTokenAccountProgramID) {
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedAssociatedInst, err := associatedtokenaccount.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			if create, ok := decodedAssociatedInst.Impl.(*associatedtokenaccount.Create); ok {
				tokenAccount := create.Get(1).PublicKey
				nativeAccount := create.Wallet
				tokenMintAccount := create.Mint
				associatedMap[tokenAccount.String()] = entity.PubKeyPair{Key: tokenMintAccount, Value: nativeAccount}
			}
			continue
		}
		// 如果是TransferChecked
		if programId.Equals(solana.TokenProgramID) {
			//log.Infof("find transfer checked instruction")
			accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
			if err != nil {
				return nil, err
			}
			decodedTokenInst, err := token.DecodeInstruction(accounts, inst.Data)
			if err != nil {
				return nil, err
			}
			if transferChecked, ok := decodedTokenInst.Impl.(*token.TransferChecked); ok {
				// 因为transfer checked指令里面的from(source)和to(destination)用的是token account，而不是native account，所以需要根据token account找到native account
				// 如果From的TokenAccount不存在，则报错，因为这种情况不存在
				fromNativeAccount, err := QueryNativeAccountByTokenAccount(rpcClient, db, transferChecked.Accounts.Get(0).PublicKey)
				if err != nil {
					account := transferChecked.Accounts.Get(0).PublicKey
					return nil, fmt.Errorf("decode transaction failed,can not get native account of from token account[%v],error:%v", account, err)
				}
				// 如果To的NativeAccount没有，可能是需要新建地址，这个时候需要查询AssociatedTokenAccount里面的地址
				toNativeAccount, err := QueryNativeAccountByTokenAccount(rpcClient, db, transferChecked.Accounts.Get(2).PublicKey)
				if err != nil {
					pair, exist := associatedMap[transferChecked.Accounts.Get(2).PublicKey.String()]
					if !exist || !pair.Key.Equals(transferChecked.Accounts.Get(1).PublicKey) {
						return nil, fmt.Errorf("decode transaction failed")
					}
					toNativeAccount = &pair.Value
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

// DecodeServiceTransaction 根据业务类型转成相应的交易，方便业务更好的处理
func DecodeServiceTransaction(decodedTx *entity.DecodedSolanaTransaction) (*entity.DecodedServiceTransaction, error) {
	var decodedServiceTx entity.DecodedServiceTransaction

	decodedServiceTx.TxID = decodedTx.TxID.String()
	decodedServiceTx.RefBlockHash = decodedTx.RefBlockHash.String()
	decodedServiceTx.FromNativeAccount = decodedTx.FromNativeAccount.String()
	decodedServiceTx.FromTokenAccount = decodedTx.FromTokenAccount.String()
	decodedServiceTx.FeePayer = decodedTx.FeePayer.String()

	for index, account := range decodedTx.Accounts {
		decodedServiceTx.Accounts[index] = account.String()
	}

	// 解析出来转dex交易
	if len(decodedTx.TransferInstructions) > 0 {
		decodedServiceTx.ToDexInst.FromNativeAccount = decodedTx.TransferInstructions[0].FromNativeAccount.String()
		decodedServiceTx.ToDexInst.ToNativeAccount = decodedTx.TransferInstructions[0].ToNativeAccount.String()
		decodedServiceTx.ToDexInst.Amount = float64(decodedTx.TransferInstructions[0].Amount)
	}

	// 解析出来转账指令
	for _, inst := range decodedTx.TransferCheckedInstructions {
		if inst.FromNativeAccount.Equals(decodedTx.FromNativeAccount) {
			// 如果发送token的transfer checked指令里面的 from native account跟发起交易的native account地址一致，则认为是交易发起人发送token到其他地址的指令
			decodedServiceTx.TransferTokenInst.FromNativeAccount = inst.FromNativeAccount.String()
			decodedServiceTx.TransferTokenInst.FromTokenAccount = inst.FromTokenAccount.String()
			decodedServiceTx.TransferTokenInst.ToNativeAccount = inst.ToNativeAccount.String()
			decodedServiceTx.TransferTokenInst.ToTokenAccount = inst.ToTokenAccount.String()
			decodedServiceTx.TransferTokenInst.OwnerNativeAccount = inst.OwnerNativeAccount.String()
			decodedServiceTx.TransferTokenInst.TokenMintAccount = inst.TokenMintAccount.String()
			decodedServiceTx.TransferTokenInst.Amount = float64(inst.Amount)
			decodedServiceTx.TransferTokenInst.Decimals = int32(inst.Decimals)
		} else if inst.ToNativeAccount.Equals(decodedTx.FromNativeAccount) {
			// 如果发送token的transfer checked指令里面的 to native account跟发起交易的native account地址一致，则认为是奖励交易发起人的token指令
			decodedServiceTx.RewardInst.FromNativeAccount = inst.FromNativeAccount.String()
			decodedServiceTx.RewardInst.FromTokenAccount = inst.FromTokenAccount.String()
			decodedServiceTx.RewardInst.ToNativeAccount = inst.ToNativeAccount.String()
			decodedServiceTx.RewardInst.ToTokenAccount = inst.ToTokenAccount.String()
			decodedServiceTx.RewardInst.OwnerNativeAccount = inst.OwnerNativeAccount.String()
			decodedServiceTx.RewardInst.TokenMintAccount = inst.TokenMintAccount.String()
			decodedServiceTx.RewardInst.Amount = float64(inst.Amount)
			decodedServiceTx.RewardInst.Decimals = int32(inst.Decimals)
		} else {
			// 否则认为是奖励代理人的奖励
			decodedServiceInst := entity.DecodedServiceTransferCheckedInst{
				FromNativeAccount:  inst.FromNativeAccount.String(),
				FromTokenAccount:   inst.FromTokenAccount.String(),
				ToNativeAccount:    inst.ToNativeAccount.String(),
				ToTokenAccount:     inst.ToTokenAccount.String(),
				OwnerNativeAccount: inst.OwnerNativeAccount.String(),
				TokenMintAccount:   inst.TokenMintAccount.String(),
				Amount:             float64(inst.Amount),
				Decimals:           int32(inst.Decimals),
			}
			decodedServiceTx.RewardInviterInst = append(decodedServiceTx.RewardInviterInst, decodedServiceInst)
		}
	}
	return &decodedServiceTx, nil
}
