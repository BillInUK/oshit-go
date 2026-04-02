package test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"oshit-go/app/reward/api/types"
	"oshit-go/common/utils"
)

const (
	systemTransferCU uint64 = 500 // System Transfer 固定预留 CU
)

// ---- TakeToken API 调用 ----

func getTakeTokenTxInfo(req types.GetTakeTokenTxInfoReq) (*types.TakeTokenTxInfo, error) {
	rsp, err := postJsonRequest[types.TakeTokenTxInfo](RewardURL+"/take/tx-info", req, map[string]string{})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func commitTakeTokenTx(req types.CommitTakeTokenTxInfoReq) (string, error) {
	rsp, err := postJsonRequest[string](RewardURL+"/take/commit-tx", req, map[string]string{})
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

// createHexEncodedTx 根据 privKey 和邀请码构造并签名 TakeToken 交易，返回 hex 编码的交易字节。
//
// 指令顺序：
//  1. SetComputeUnitPrice（Medium 档）
//  2. SetComputeUnitLimit（根据 instUnits 计算）
//  3. [可选] CreateAssociatedTokenAccount（领取人 PDA 不存在时）
//  4. TransferChecked：rewardTokenAccount → receiptTokenAccount（领取人奖励）
//  5. System.Transfer：receiptNativeAccount → dexNativeAccount（SOL 成本费）
//  6. TransferChecked×N：rewardTokenAccount → inviterTokenAccount[i]（邀请人奖励）
func createHexEncodedTx(ctx context.Context, privKey solana.PrivateKey, inviteCode string) (string, error) {
	receiptPubKey := privKey.PublicKey()

	txInfo, err := getTakeTokenTxInfo(types.GetTakeTokenTxInfoReq{
		ReceiptAccount: receiptPubKey.String(),
		InviteCode:     inviteCode,
	})
	if err != nil {
		return "", fmt.Errorf("getTakeTokenTxInfo failed: %w", err)
	}

	priorityFee, err := getPriorityFee()
	if err != nil {
		return "", fmt.Errorf("getPriorityFee failed: %w", err)
	}

	// 用 instUnits API 计算 computeUnitLimit：
	//   每条 TransferChecked 按 instUnits.TransferChecked 计，
	//   加上 System Transfer 的固定预留（500 CU），再留 20% buffer。
	// 注意：不使用模拟交易方式，原因是模拟时需要放入假签名，
	//       Solana 节点对假签名的模拟结果不可信，会严重低估实际消耗。
	instUnits, err := getInstUnits()
	if err != nil {
		return "", fmt.Errorf("getInstUnits failed: %w", err)
	}

	fmt.Printf("receipt native account: %v\n", receiptPubKey)
	fmt.Printf("total reward amount: %v\n", txInfo.TotalRewardAmount)
	fmt.Printf("quote sol price: %v\n", txInfo.QuoteSOLPrice)
	fmt.Printf("quoted sol amount: %v\n", txInfo.QuotedSOLAmount)

	rewardNativeAccount := solana.MPK(txInfo.RewardNativeAccount)
	rewardTokenAccount := solana.MPK(txInfo.RewardTokenAccount)
	tokenMintAccount := solana.MPK(txInfo.TokenMintAccount)
	dexNativeAccount := solana.MPK(txInfo.DexNativeAccount)
	decimals := uint8(txInfo.Decimals)

	receiptTokenAccount, err := utils.GetSPLTokenAccountByNative(rpcClient, receiptPubKey, solana.MPK(txInfo.TokenMintAccount))
	if err != nil {
		return "", fmt.Errorf("GetSPLTokenAccountByNative failed: %w", err)
	}
	receiptPDAExist := receiptTokenAccount != nil

	computePrice := priorityFee.PerComputeUnit.Medium
	fmt.Printf("compute price: %d\n", computePrice)

	// 奖励邀请人指令
	var inviterInsts []solana.Instruction
	for _, rewardInfo := range txInfo.RewardInviterInfo {
		inst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(rewardInfo.Amount).
			SetDecimals(decimals).
			SetSourceAccount(rewardTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(solana.MPK(rewardInfo.TokenAccount)).
			SetOwnerAccount(rewardNativeAccount).
			Build()
		inviterInsts = append(inviterInsts, inst)
	}

	transferCheckedCount := uint64(1 + len(inviterInsts))
	computeUnitLimit := uint32(float64(transferCheckedCount*instUnits.TransferChecked+systemTransferCU) * 1.2)
	fmt.Printf("computeUnitLimit: %d (transferCheckedCount=%d, perTC=%d)\n", computeUnitLimit, transferCheckedCount, instUnits.TransferChecked)

	var instructions []solana.Instruction
	instructions = append(instructions,
		computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build(),
		computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(computeUnitLimit).Build(),
	)
	if !receiptPDAExist {
		instructions = append(instructions,
			associatedtokenaccount.NewCreateInstruction(receiptPubKey, receiptPubKey, tokenMintAccount).Build(),
		)
	}
	instructions = append(instructions,
		token.NewTransferCheckedInstructionBuilder().
			SetAmount(txInfo.RewardInfo.Amount).
			SetDecimals(decimals).
			SetSourceAccount(rewardTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(*receiptTokenAccount).
			SetOwnerAccount(rewardNativeAccount).
			Build(),
		system.NewTransferInstructionBuilder().
			SetLamports(uint64(txInfo.QuotedSOLAmount)).
			SetFundingAccount(receiptPubKey).
			SetRecipientAccount(dexNativeAccount).
			Build(),
	)
	instructions = append(instructions, inviterInsts...)

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("GetLatestBlockhash failed: %w", err)
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(receiptPubKey))
	if err != nil {
		return "", fmt.Errorf("NewTransaction failed: %w", err)
	}

	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("MarshalBinary failed: %w", err)
	}

	fromSign, err := privKey.Sign(messageContent)
	if err != nil {
		return "", fmt.Errorf("Sign failed: %w", err)
	}
	tx.Signatures = append(tx.Signatures, fromSign)

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("tx.MarshalBinary failed: %w", err)
	}
	return hex.EncodeToString(txBytes), nil
}

// ---- TakeToken 测试用例 ----

func TestGetTakeTokenTxInfo(t *testing.T) {
	txInfo, err := getTakeTokenTxInfo(types.GetTakeTokenTxInfoReq{
		ReceiptAccount: AliceNativePubKey,
	})
	if err != nil {
		t.Fatalf("getTakeTokenTxInfo failed: %v", err)
	}

	fmt.Printf("=== TakeToken TxInfo ===\n")
	fmt.Printf("RewardNativeAccount : %s\n", txInfo.RewardNativeAccount)
	fmt.Printf("RewardTokenAccount  : %s\n", txInfo.RewardTokenAccount)
	fmt.Printf("TokenMintAccount    : %s\n", txInfo.TokenMintAccount)
	fmt.Printf("DexAccount          : %s\n", txInfo.DexNativeAccount)
	fmt.Printf("Decimals            : %d\n", txInfo.Decimals)
	fmt.Printf("QuoteSOLPrice       : %v\n", txInfo.QuoteSOLPrice)
	fmt.Printf("TotalRewardAmount   : %v\n", txInfo.TotalRewardAmount)
	fmt.Printf("QuotedSOLAmount     : %v\n", txInfo.QuotedSOLAmount)
	fmt.Printf("InviteCode          : %s\n", txInfo.InviteCode)
	fmt.Printf("InviteCodeValid     : %v\n", txInfo.InviteCodeValid)
	fmt.Printf("InviteDetermine     : %v\n", txInfo.InviteDetermine)
	fmt.Printf("RewardInfo          : %+v\n", txInfo.RewardInfo)
	fmt.Printf("RewardInviterInfo   : %+v\n", txInfo.RewardInviterInfo)
}

func TestTakeToken(t *testing.T) {
	inviteCode := "ogG0W1OK"
	privKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}

	encodedTx, err := createHexEncodedTx(context.Background(), privKey, inviteCode)
	if err != nil {
		t.Fatalf("createHexEncodedTx failed: %v", err)
	}

	txId, err := commitTakeTokenTx(types.CommitTakeTokenTxInfoReq{
		EncodedTx:  encodedTx,
		InviteCode: inviteCode,
	})
	if err != nil {
		t.Fatalf("commitTakeTokenTx failed: %v", err)
	}
	fmt.Printf("请求成功: https://solscan.io/tx/%s?cluster=devnet\n", txId)
}
