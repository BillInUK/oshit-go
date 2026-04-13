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

// ---- GiveToken API 调用 ----

func getGiveTokenTxInfo(jwtToken string, req types.GetGiveTokenTxInfoReq) (*types.GiveTokenTxInfo, error) {
	rsp, err := postJsonRequest[types.GiveTokenTxInfo](RewardURL+"/give/tx-info", req, map[string]string{
		"Authorization": "Bearer " + jwtToken,
	})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func commitGiveTokenTx(req types.CommitGiveTokenTxInfoReq) (string, error) {
	rsp, err := postJsonRequest[string](RewardURL+"/give/commit-tx", req, map[string]string{})
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

// createGiveTokenHexEncodedTx 根据 privKey 和目标地址构造并签名 GiveToken 交易，返回 hex 编码的交易字节。
//
// 指令顺序：
//  1. SetComputeUnitPrice（Medium 档）
//  2. SetComputeUnitLimit（根据 instUnits 计算）
//  3. [可选] CreateAssociatedTokenAccount（接收人 PDA 不存在时）
//  4. TransferChecked：fromTokenAccount → toTokenAccount（转账人转出 token）
//  5. TransferChecked：rewardTokenAccount → fromTokenAccount（平台奖励转账人）
//  6. System.Transfer：fromNativeAccount → costAccount（SOL 成本费）
//  7. TransferChecked×N：rewardTokenAccount → inviterTokenAccount[i]（邀请人奖励）
func createGiveTokenHexEncodedTx(ctx context.Context, privKey solana.PrivateKey, toAccount string, amount float64) (string, error) {
	fromPubKey := privKey.PublicKey()

	// 获取 JWT token
	jwtToken, err := loginForToken(Brand, Symbol, fromPubKey, privKey)
	if err != nil {
		return "", fmt.Errorf("loginForToken failed: %w", err)
	}

	txInfo, err := getGiveTokenTxInfo(jwtToken, types.GetGiveTokenTxInfoReq{
		To:     toAccount,
		Amount: amount,
	})
	if err != nil {
		return "", fmt.Errorf("getGiveTokenTxInfo failed: %w", err)
	}

	priorityFee, err := getPriorityFee()
	if err != nil {
		return "", fmt.Errorf("getPriorityFee failed: %w", err)
	}

	instUnits, err := getInstUnits()
	if err != nil {
		return "", fmt.Errorf("getInstUnits failed: %w", err)
	}

	fmt.Printf("from native account: %v\n", fromPubKey)
	fmt.Printf("to native account: %v\n", toAccount)
	fmt.Printf("total reward amount: %v\n", txInfo.TotalReward)
	fmt.Printf("quote sol price: %v\n", txInfo.QuoteSOLPrice)
	fmt.Printf("quoted sol amount: %v\n", txInfo.QuotedSOLAmount)

	rewardNativeAccount := solana.MPK(txInfo.RewardAccount)
	tokenMintAccount := solana.MPK(txInfo.Mint)
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeAccount, tokenMintAccount)
	costAccount := solana.MPK(txInfo.CostAccount)
	decimals := uint8(txInfo.Decimals)

	// from 地址的 token account（转账人必须已持有 token）
	fromTokenAccount, err := utils.GetSPLTokenAccountByNative(rpcClient, fromPubKey, tokenMintAccount)
	if err != nil || fromTokenAccount == nil {
		return "", fmt.Errorf("GetSPLTokenAccountByNative for from account failed: %w", err)
	}

	// to 地址的 token account（可能不存在，需要创建）
	toNativeKey := solana.MPK(toAccount)
	toTokenAccount, _, _ := solana.FindAssociatedTokenAddress(toNativeKey, tokenMintAccount)
	toTokenAccountInfo, _ := rpcClient.GetAccountInfo(ctx, toTokenAccount)
	toPDAExist := toTokenAccountInfo != nil && toTokenAccountInfo.Value != nil

	computePrice := priorityFee.PerComputeUnit.Medium
	fmt.Printf("compute price: %d\n", computePrice)

	// 奖励邀请人指令
	var inviterInsts []solana.Instruction
	for _, rewardInfo := range txInfo.RewardInviterInfo {
		inviterTA, _, _ := solana.FindAssociatedTokenAddress(solana.MPK(rewardInfo.ReceiptAccount), tokenMintAccount)
		inst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(rewardInfo.Amount).
			SetDecimals(decimals).
			SetSourceAccount(rewardTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(inviterTA).
			SetOwnerAccount(rewardNativeAccount).
			Build()
		inviterInsts = append(inviterInsts, inst)
	}

	// CU 计算：1 条用户转账 + 1 条奖励转账人 + N 条奖励邀请人 + 1 条 System.Transfer
	transferCheckedCount := uint64(2 + len(inviterInsts))
	computeUnitLimit := uint32(float64(transferCheckedCount*instUnits.TransferChecked+systemTransferCU) * 1.2)
	fmt.Printf("computeUnitLimit: %d (transferCheckedCount=%d, perTC=%d)\n", computeUnitLimit, transferCheckedCount, instUnits.TransferChecked)

	var instructions []solana.Instruction
	instructions = append(instructions,
		computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build(),
		computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(computeUnitLimit).Build(),
	)
	// 如果接收人没有 token account，创建 ATA
	if !toPDAExist {
		instructions = append(instructions,
			associatedtokenaccount.NewCreateInstruction(fromPubKey, toNativeKey, tokenMintAccount).Build(),
		)
	}
	instructions = append(instructions,
		// 转账人转出 token 给接收人
		token.NewTransferCheckedInstructionBuilder().
			SetAmount(txInfo.GiveInfo.Amount).
			SetDecimals(decimals).
			SetSourceAccount(*fromTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(toTokenAccount).
			SetOwnerAccount(fromPubKey).
			Build(),
		// 平台奖励转账人
		token.NewTransferCheckedInstructionBuilder().
			SetAmount(txInfo.RewardInfo.Amount).
			SetDecimals(decimals).
			SetSourceAccount(rewardTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(*fromTokenAccount).
			SetOwnerAccount(rewardNativeAccount).
			Build(),
		// 转账人支付 SOL 成本费
		system.NewTransferInstructionBuilder().
			SetLamports(uint64(txInfo.QuotedSOLAmount)).
			SetFundingAccount(fromPubKey).
			SetRecipientAccount(costAccount).
			Build(),
	)
	instructions = append(instructions, inviterInsts...)

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("GetLatestBlockhash failed: %w", err)
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(fromPubKey))
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

// ---- GiveToken 测试用例 ----

func TestGetGiveTokenTxInfo(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	fromPubKey := privKey.PublicKey()

	jwtToken, err := loginForToken(Brand, Symbol, fromPubKey, privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	txInfo, err := getGiveTokenTxInfo(jwtToken, types.GetGiveTokenTxInfoReq{
		To:     BobNativePubKey,
		Amount: 1000,
	})
	if err != nil {
		t.Fatalf("getGiveTokenTxInfo failed: %v", err)
	}

	fmt.Printf("=== GiveToken TxInfo ===\n")
	fmt.Printf("RewardAccount       : %s\n", txInfo.RewardAccount)
	fmt.Printf("Mint                : %s\n", txInfo.Mint)
	fmt.Printf("CostAccount         : %s\n", txInfo.CostAccount)
	fmt.Printf("Decimals            : %d\n", txInfo.Decimals)
	fmt.Printf("QuoteSOLPrice       : %v\n", txInfo.QuoteSOLPrice)
	fmt.Printf("TotalReward         : %v\n", txInfo.TotalReward)
	fmt.Printf("QuotedSOLAmount     : %v\n", txInfo.QuotedSOLAmount)
	fmt.Printf("GiveInfo            : %+v\n", txInfo.GiveInfo)
	fmt.Printf("RewardInfo          : %+v\n", txInfo.RewardInfo)
	fmt.Printf("RewardInviterInfo   : %+v\n", txInfo.RewardInviterInfo)
}

func TestGiveToken(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	to := RobertNativePubKey
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}

	encodedTx, err := createGiveTokenHexEncodedTx(context.Background(), privKey, to, 1000)
	if err != nil {
		t.Fatalf("createGiveTokenHexEncodedTx failed: %v", err)
	}

	txId, err := commitGiveTokenTx(types.CommitGiveTokenTxInfoReq{
		EncodedTx: encodedTx,
		To:        to,
	})
	if err != nil {
		t.Fatalf("commitGiveTokenTx failed: %v", err)
	}
	fmt.Printf("请求成功: https://solscan.io/tx/%s?cluster=devnet\n", txId)
}
