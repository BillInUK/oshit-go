package test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"oshit-go/app/reward/api/types"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
)

// ---- RewardCode API 调用 ----

func getRewardCodeInfo(rewardCode string) (*model.RewardCode, error) {
	rsp, err := postJsonRequest[model.RewardCode](RewardURL+"/reward-code/info", types.GetRewardCodeTxInfoReq{
		RewardCode: rewardCode,
	}, map[string]string{})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func getRewardCodeTxInfo(rewardCode string) (*types.RewardCodeTxInfo, error) {
	rsp, err := postJsonRequest[types.RewardCodeTxInfo](RewardURL+"/reward-code/tx-info", types.GetRewardCodeTxInfoReq{
		RewardCode: rewardCode,
	}, map[string]string{})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func commitRewardCodeTx(req types.CommitRewardCodeTxReq) (string, error) {
	rsp, err := postJsonRequest[string](RewardURL+"/reward-code/commit-tx", req, map[string]string{})
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

// createRewardCodeHexEncodedTx 根据 privKey 构造并签名奖励码领取交易，返回 hex 编码的交易字节。
//
// 指令顺序：
//  1. SetComputeUnitPrice（Medium 档）
//  2. SetComputeUnitLimit
//  3. [可选] CreateAssociatedTokenAccount（用户 PDA 不存在时）
//  4. TransferChecked：rewardTokenAccount → userTokenAccount（平台发放 token 奖励）
//  5. System.Transfer：userNativeAccount → costAccount（SOL 成本费）
func createRewardCodeHexEncodedTx(ctx context.Context, privKey solana.PrivateKey, txInfo *types.RewardCodeTxInfo) (string, error) {
	userPubKey := privKey.PublicKey()

	priorityFee, err := getPriorityFee()
	if err != nil {
		return "", fmt.Errorf("getPriorityFee failed: %w", err)
	}
	instUnits, err := getInstUnits()
	if err != nil {
		return "", fmt.Errorf("getInstUnits failed: %w", err)
	}

	fmt.Printf("user native account : %v\n", userPubKey)
	fmt.Printf("reward amount       : %v\n", txInfo.RewardAmount)
	fmt.Printf("cost fee (lamports) : %v\n", txInfo.CostFee)

	rewardNativeAccount := solana.MPK(txInfo.RewardAccount)
	tokenMintAccount := solana.MPK(txInfo.Mint)
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeAccount, tokenMintAccount)
	costAccount := solana.MPK(txInfo.CostAccount)
	decimals := uint8(txInfo.Decimals)

	// 检查用户是否已有 token account
	userTokenAccount, err := utils.GetSPLTokenAccountByNative(rpcClient, userPubKey, tokenMintAccount)
	if err != nil {
		return "", fmt.Errorf("GetSPLTokenAccountByNative failed: %w", err)
	}
	userPDAExist := userTokenAccount != nil
	if !userPDAExist {
		derived, _, _ := solana.FindAssociatedTokenAddress(userPubKey, tokenMintAccount)
		userTokenAccount = &derived
	}

	computePrice := priorityFee.PerComputeUnit.Medium
	// CU: 1 TransferChecked + 1 System.Transfer
	computeUnitLimit := uint32(float64(instUnits.TransferChecked+systemTransferCU) * 1.2)
	fmt.Printf("compute price: %d\n", computePrice)
	fmt.Printf("computeUnitLimit: %d\n", computeUnitLimit)

	var instructions []solana.Instruction
	instructions = append(instructions,
		computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build(),
		computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(computeUnitLimit).Build(),
	)
	if !userPDAExist {
		instructions = append(instructions,
			associatedtokenaccount.NewCreateInstruction(userPubKey, userPubKey, tokenMintAccount).Build(),
		)
	}
	instructions = append(instructions,
		// 平台从 reward_account 发放 token 给用户
		token.NewTransferCheckedInstructionBuilder().
			SetAmount(uint64(txInfo.RewardAmount)).
			SetDecimals(decimals).
			SetSourceAccount(rewardTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(*userTokenAccount).
			SetOwnerAccount(rewardNativeAccount).
			Build(),
		// 用户支付 SOL 成本费到 cost_account
		system.NewTransferInstructionBuilder().
			SetLamports(uint64(txInfo.CostFee)).
			SetFundingAccount(userPubKey).
			SetRecipientAccount(costAccount).
			Build(),
	)

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("GetLatestBlockhash failed: %w", err)
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(userPubKey))
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

// ---- RewardCode 测试用例 ----

// TestGetRewardCodeInfo 根据奖励码查询基本信息
func TestGetRewardCodeInfo(t *testing.T) {
	// 从 1000000~1000010 中随机选一个
	rewardCode := fmt.Sprintf("%d", 1000000+rand.Intn(11))
	fmt.Printf("查询奖励码: %s\n", rewardCode)

	info, err := getRewardCodeInfo(rewardCode)
	if err != nil {
		t.Fatalf("getRewardCodeInfo failed: %v", err)
	}

	fmt.Printf("=== RewardCode Info ===\n")
	fmt.Printf("RecordID     : %s\n", info.RecordID)
	fmt.Printf("RewardCode   : %s\n", info.RewardCode)
	fmt.Printf("RewardAmount : %v\n", info.RewardAmount)
	fmt.Printf("TxID         : %s\n", info.TxID)
	fmt.Printf("TxState      : %d\n", info.TxState)
	fmt.Printf("ExpireTime   : %s\n", info.ExpiredAt.Format("2006-01-02 15:04:05"))
}

// TestGetRewardCodeTxInfo 根据奖励码查询交易构建所需信息
func TestGetRewardCodeTxInfo(t *testing.T) {
	rewardCode := fmt.Sprintf("%d", 1000000+rand.Intn(11))
	fmt.Printf("查询奖励码交易信息: %s\n", rewardCode)

	txInfo, err := getRewardCodeTxInfo(rewardCode)
	if err != nil {
		t.Fatalf("getRewardCodeTxInfo failed: %v", err)
	}

	fmt.Printf("=== RewardCode TxInfo ===\n")
	fmt.Printf("RewardAccount : %s\n", txInfo.RewardAccount)
	fmt.Printf("Mint          : %s\n", txInfo.Mint)
	fmt.Printf("CostAccount   : %s\n", txInfo.CostAccount)
	fmt.Printf("Decimals      : %d\n", txInfo.Decimals)
	fmt.Printf("RewardAmount  : %v\n", txInfo.RewardAmount)
	fmt.Printf("CostFee       : %v lamports\n", txInfo.CostFee)
}

// TestRewardCode 执行完整奖励码领取流程：随机选码 → tx-info → 构建交易 → commit-tx → 轮询确认
func TestRewardCode(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}

	// 1. 随机选择奖励码（1000000~1000010）
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rewardCode := fmt.Sprintf("%d", 100000+rand.Intn(11))
	fmt.Printf("=== 使用奖励码: %s ===\n", rewardCode)

	// 2. 查询奖励码基本信息
	info, err := getRewardCodeInfo(rewardCode)
	if err != nil {
		t.Fatalf("getRewardCodeInfo failed: %v", err)
	}
	fmt.Printf("奖励码信息: RewardAmount=%v TxState=%d ExpireTime=%s\n",
		info.RewardAmount, info.TxState, info.ExpiredAt.Format("2006-01-02 15:04:05"))

	// 3. 获取交易构建参数
	txInfo, err := getRewardCodeTxInfo(rewardCode)
	if err != nil {
		t.Fatalf("getRewardCodeTxInfo failed: %v", err)
	}
	fmt.Printf("交易信息: RewardAmount=%v CostFee=%v lamports\n", txInfo.RewardAmount, txInfo.CostFee)

	// 4. 构建并签名交易
	encodedTx, err := createRewardCodeHexEncodedTx(context.Background(), privKey, txInfo)
	if err != nil {
		t.Fatalf("createRewardCodeHexEncodedTx failed: %v", err)
	}

	// 5. 提交交易
	txId, err := commitRewardCodeTx(types.CommitRewardCodeTxReq{
		EncodedTx:  encodedTx,
		RewardCode: rewardCode,
	})
	if err != nil {
		t.Fatalf("commitRewardCodeTx failed: %v", err)
	}
	fmt.Printf("交易已提交: https://explorer.solana.com/tx/%s?cluster=devnet\n", txId)

	// 6. 轮询等待链上确认（通过查询奖励码 tx_state 判断）
	fmt.Println("轮询等待链上确认...")
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		result, err := getRewardCodeInfo(rewardCode)
		if err != nil {
			fmt.Printf("  [轮询] 查询奖励码状态失败: %v，继续等待...\n", err)
			time.Sleep(3 * time.Second)
			continue
		}
		switch result.TxState {
		case 1:
			fmt.Printf("  [轮询] 交易已上链确认，tx_state=1\n")
			fmt.Printf("=== TestRewardCode 完成 ===\n")
			return
		case -1:
			t.Fatalf("交易失败，tx_state=-1，txId=%s", txId)
		default:
			fmt.Printf("  [轮询] tx_state=%d，等待确认中...\n", result.TxState)
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("等待链上确认超时，txId=%s", txId)
}
