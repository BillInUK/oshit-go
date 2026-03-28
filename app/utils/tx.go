package utils

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/app/base/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
)

// QueryNativeAccountInfoByTokenAccount 查询原生账户信息
func QueryNativeAccountInfoByTokenAccount(db *gorm.DB, tokenAccount string) (*model.NativeAccountInfo, error) {
	var record model.NativeAccountInfo
	err := db.Where("\"TokenAccount\"= ?", tokenAccount).Take(&record).Error
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
