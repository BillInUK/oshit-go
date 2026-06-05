package utils

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"

	"oshit-go/common/pkg/entity"
)

// TransferSPLTokenWithPriorityFee 发送SPLToken
func TransferSPLTokenWithPriorityFee(rpcClient *rpc.Client,
	tokenMintAccount solana.PublicKey,
	fromPrivateKey solana.PrivateKey,
	toNativeAccount solana.PublicKey,
	amount uint64,
	decimals uint8,
	memo string,
	priorityFee entity.PriorityFee,
	computeUnitDetail entity.ComputeUnitDetail) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativeAccount := fromPrivateKey.PublicKey()
	recent, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(fromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}

	toTokenAccountExist := true
	toTokenAccount, _, err := solana.FindAssociatedTokenAddress(toNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}
	if _, err := GetNativeAccountByTokenAccount(rpcClient, toTokenAccount); err != nil {
		toTokenAccountExist = false
	}

	computePrice := priorityFee.PerComputeUnit.High
	if computePrice > priorityFee.PerComputeUnit.Medium*2 {
		computePrice = priorityFee.PerComputeUnit.Medium * 2
	}
	computeUnitLimit := uint32(0)
	var instructions []solana.Instruction
	if !toTokenAccountExist {
		associated := associatedtokenaccount.NewCreateInstruction(fromNativeAccount, toNativeAccount, tokenMintAccount).Build()
		instructions = append(instructions, associated)
		computeUnitLimit += uint32(computeUnitDetail.AssociatedAccount)
	}

	transferInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(fromTokenAccount).
		SetDestinationAccount(toTokenAccount).
		SetDecimals(decimals).
		SetMintAccount(tokenMintAccount).
		SetOwnerAccount(fromNativeAccount).
		Build()

	instructions = append(instructions, transferInst)
	computeUnitLimit += uint32(computeUnitDetail.TransferChecked)
	if memo != "" {
		programID := solana.MemoProgramID
		accounts := []*solana.AccountMeta{solana.Meta(fromNativeAccount).SIGNER().WRITE()}
		data := []byte(memo)
		memoInst := solana.NewInstruction(programID, accounts, data)
		instructions = append(instructions, memoInst)
		computeUnitLimit += uint32(computeUnitDetail.Memo)
	}
	computeUnitLimit = uint32(float64(computeUnitLimit) * 1.2)
	computeBudgetLimitInst := computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(computeUnitLimit).Build()
	computeBudgetPriceInst := computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build()
	instructions = append([]solana.Instruction{computeBudgetLimitInst}, instructions...)
	instructions = append([]solana.Instruction{computeBudgetPriceInst}, instructions...)

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(fromNativeAccount))
	if err != nil {
		return nil, err
	}

	if _, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativeAccount.Equals(key) {
				return &fromPrivateKey
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	simulateResult, err := rpcClient.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
		ReplaceRecentBlockhash: true,
	})
	if err != nil {
		return nil, fmt.Errorf("simulate transaction error %v", err)
	}
	if simulateResult.Value.Err != nil {
		return nil, fmt.Errorf("simulate transaction failed %v", simulateResult.Value.Err)
	}

	txId, err := rpcClient.SendTransaction(ctx, tx)
	return &txId, err
}

// TransferSPLToken 发送SPLToken
func TransferSPLToken(rpcClient *rpc.Client,
	tokenMintAccount solana.PublicKey,
	fromPrivateKey solana.PrivateKey,
	toNativeAccount solana.PublicKey,
	amount uint64,
	decimals uint8,
	memo string) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativeAccount := fromPrivateKey.PublicKey()
	recent, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(fromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}

	toTokenAccountExist := true
	toTokenAccount, _, err := solana.FindAssociatedTokenAddress(toNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}
	if _, err := GetNativeAccountByTokenAccount(rpcClient, toTokenAccount); err != nil {
		toTokenAccountExist = false
	}

	var instructions []solana.Instruction
	if !toTokenAccountExist {
		associated := associatedtokenaccount.NewCreateInstruction(fromNativeAccount, toNativeAccount, tokenMintAccount).Build()
		instructions = append(instructions, associated)
	}

	transferInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(fromTokenAccount).
		SetDestinationAccount(toTokenAccount).
		SetDecimals(decimals).
		SetMintAccount(tokenMintAccount).
		SetOwnerAccount(fromNativeAccount).
		Build()

	instructions = append(instructions, transferInst)
	if memo != "" {
		programID := solana.MemoProgramID
		accounts := []*solana.AccountMeta{solana.Meta(fromNativeAccount).SIGNER().WRITE()}
		data := []byte(memo)
		memoInst := solana.NewInstruction(programID, accounts, data)
		instructions = append(instructions, memoInst)
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(fromNativeAccount))
	if err != nil {
		return nil, err
	}

	if _, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativeAccount.Equals(key) {
				return &fromPrivateKey
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	simulateResult, err := rpcClient.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
		ReplaceRecentBlockhash: true,
	})
	if err != nil {
		return nil, fmt.Errorf("simulate transaction error %v", err)
	}
	if simulateResult.Value.Err != nil {
		return nil, fmt.Errorf("simulate transaction failed %v", simulateResult.Value.Err)
	}

	txId, err := rpcClient.SendTransaction(ctx, tx)
	return &txId, err
}

func TransferSPLTokenWithFeePayer(rpcClient *rpc.Client, wssClient *ws.Client, tokenMintAccount solana.PublicKey,
	fromPrivKey solana.PrivateKey, feePayerPrivKey solana.PrivateKey, fromTokenPubKey solana.PublicKey, toTokenPubKey solana.PublicKey, amount uint64, decimals uint8) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativePubKey := fromPrivKey.PublicKey()
	feePayerNativePubKey := feePayerPrivKey.PublicKey()

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	transferInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(amount).
		SetDecimals(decimals).
		SetSourceAccount(fromTokenPubKey).
		SetMintAccount(tokenMintAccount).
		SetDestinationAccount(toTokenPubKey).
		SetOwnerAccount(fromNativePubKey).
		Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			transferInst,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayerNativePubKey),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			if feePayerNativePubKey.Equals(key) {
				return &feePayerPrivKey
			}
			return nil
		},
	)

	txSig, err := confirm.SendAndConfirmTransaction(ctx, rpcClient, wssClient, tx)
	return &txSig, err
}

// MultiTransferSPLToken 发送SPLToken
func MultiTransferSPLToken(rpcClient *rpc.Client, wssClient *ws.Client, tokenMintAccount solana.PublicKey,
	fromPrivKey solana.PrivateKey, fromTokenPubKey solana.PublicKey, toTokenPubKey []solana.PublicKey, amount uint64, decimals uint8) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativePubKey := fromPrivKey.PublicKey()

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	var instList []solana.Instruction
	for _, to := range toTokenPubKey {
		transferInst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(amount).
			SetDecimals(decimals).
			SetSourceAccount(fromTokenPubKey).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(to).
			SetOwnerAccount(fromNativePubKey).
			Build()
		instList = append(instList, transferInst)
	}

	tx, err := solana.NewTransaction(
		instList,
		recent.Value.Blockhash,
		solana.TransactionPayer(fromNativePubKey),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			return nil
		},
	)

	txSig, err := confirm.SendAndConfirmTransaction(ctx, rpcClient, wssClient, tx)
	return &txSig, err
}

// MultiTransferSPLTokenWithFeePayer 发送SPLToken
func MultiTransferSPLTokenWithFeePayer(rpcClient *rpc.Client, wssClient *ws.Client, tokenMintAccount solana.PublicKey,
	fromPrivKey solana.PrivateKey, feePayerPrivKey solana.PrivateKey, fromTokenPubKey solana.PublicKey,
	toTokenPubKey []solana.PublicKey, amount uint64, decimals uint8) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativePubKey := fromPrivKey.PublicKey()
	feePayerNativePubKey := feePayerPrivKey.PublicKey()

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	var instList []solana.Instruction
	for _, to := range toTokenPubKey {
		transferInst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(amount).
			SetDecimals(decimals).
			SetSourceAccount(fromTokenPubKey).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(to).
			SetOwnerAccount(fromNativePubKey).
			Build()
		instList = append(instList, transferInst)
	}

	tx, err := solana.NewTransaction(
		instList,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayerNativePubKey),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			if feePayerNativePubKey.Equals(key) {
				return &feePayerPrivKey
			}
			return nil
		},
	)

	txSig, err := confirm.SendAndConfirmTransaction(ctx, rpcClient, wssClient, tx)
	return &txSig, err
}

// MultiTransferSPLRewardToken TODO: 手续费优化
func MultiTransferSPLRewardToken(ctx context.Context,
	rpcClient *rpc.Client,
	fromPrivKey solana.PrivateKey,
	toTokenAccounts []solana.PublicKey,
	tokenMintAccounts []solana.PublicKey,
	rewardAmounts []uint64,
	decimals []uint8) (*solana.Signature, uint64, error) {

	fromNativePubKey := fromPrivKey.PublicKey()

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, 0, err
	}

	var instList []solana.Instruction
	computeBudgetPriceInst := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(1000).
		Build()
	computeBudgetLimitInst := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(600000).
		Build()
	instList = append(instList, computeBudgetPriceInst)
	instList = append(instList, computeBudgetLimitInst)

	for index, to := range toTokenAccounts {
		fromTokenAccount, _, _ := solana.FindAssociatedTokenAddress(fromNativePubKey, tokenMintAccounts[index])
		transferInst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(rewardAmounts[index]).
			SetDecimals(decimals[index]).
			SetSourceAccount(fromTokenAccount).
			SetMintAccount(tokenMintAccounts[index]).
			SetDestinationAccount(to).
			SetOwnerAccount(fromNativePubKey).
			Build()
		instList = append(instList, transferInst)
	}

	tx, err := solana.NewTransaction(
		instList,
		recent.Value.Blockhash,
		solana.TransactionPayer(fromNativePubKey),
	)
	if err != nil {
		return nil, 0, err
	}
	messageBin, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, 0, err
	}
	base64EncodeMsg := base64.StdEncoding.EncodeToString(messageBin)
	fee, err := rpcClient.GetFeeForMessage(ctx, base64EncodeMsg, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, 0, err
	}
	if fee == nil || fee.Value == nil {
		return nil, 0, err
	}
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			return nil
		},
	)
	txId, err := rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight: true,
	})
	return &txId, *fee.Value, err
}
