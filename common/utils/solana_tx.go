package utils

import (
	"context"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetTransactionByTxId(ctx context.Context, rpcClient *rpc.Client, txId solana.Signature) (*solana.Transaction, error) {
	var maxSupportVersion uint64 = 0
	out, err := rpcClient.GetTransaction(
		ctx,
		txId,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &maxSupportVersion,
			Commitment:                     rpc.CommitmentFinalized,
			Encoding:                       solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, err
	}
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func GetTransactionResultByTxId(ctx context.Context, rpcClient *rpc.Client, txId solana.Signature) (*rpc.GetTransactionResult, error) {
	var maxSupportVersion uint64 = 0
	tr, err := rpcClient.GetTransaction(
		ctx,
		txId,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &maxSupportVersion,
			Commitment:                     rpc.CommitmentConfirmed,
			Encoding:                       solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func GetParsedTransactionById(ctx context.Context, rpcClient *rpc.Client, txId solana.Signature) (*rpc.GetParsedTransactionResult, error) {
	out, err := rpcClient.GetParsedTransaction(ctx, txId, nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func DecodeSPLTransferBySig(rpcClient *rpc.Client, txSig solana.Signature) (*token.TransferChecked, error) {
	out, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return nil, err
	}
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	if err != nil {
		return nil, err
	}
	return DecodeSPLTokenTransfer(tx)
}

func DecodeSPLTokenTransfer(tx *solana.Transaction) (*token.TransferChecked, error) {
	for _, inst := range tx.Message.Instructions {
		accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
		if err != nil {
			continue
		}
		decodedInst, err := token.DecodeInstruction(accounts, inst.Data)
		if err != nil {
			continue
		}
		if transferInfo, ok := decodedInst.Impl.(*token.TransferChecked); ok {
			fmt.Printf("amount %d\n", *transferInfo.Amount)
			fmt.Printf("decimals %d\n", *transferInfo.Decimals)
			return transferInfo, nil
		}
	}
	return nil, errors.New("can not find any transfer checked instruction in transaction")
}
