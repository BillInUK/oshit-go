package test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
)

// ---- Pos 奖励 API 调用 ----

// getPosRewardConfig 查询 pos 奖励配置（无需 JWT）
func getPosRewardConfig() (*model.PosRewardConfig, error) {
	rsp, err := postJsonRequest[model.PosRewardConfig](
		SnapURL+"/pos/reward/config",
		struct{}{},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

// getPosRewardRecords 查询当日未领取的 pos 奖励列表（需要 JWT）
func getPosRewardRecords(jwtToken string) ([]model.PosReward, error) {
	rsp, err := postJsonRequest[[]model.PosReward](
		SnapURL+"/pos/reward/record",
		struct{}{},
		map[string]string{"Authorization": "Bearer " + jwtToken},
	)
	if err != nil {
		return nil, err
	}
	return rsp.Data, nil
}

// getPosRewardTxInfo 获取领取交易所需参数（需要 JWT）
func getPosRewardTxInfo(jwtToken string) (*types.ClaimPosRewardTxInfo, error) {
	rsp, err := postJsonRequest[types.ClaimPosRewardTxInfo](
		SnapURL+"/pos/reward/tx-info",
		struct{}{},
		map[string]string{"Authorization": "Bearer " + jwtToken},
	)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

// commitPosRewardTx 提交已签名的领取交易
func commitPosRewardTx(encodedTx string) (string, error) {
	rsp, err := postJsonRequest[string](
		SnapURL+"/pos/reward/commit-tx",
		types.CommitPosRewardTxReq{EncodedTx: encodedTx},
		nil,
	)
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

// getPosClaimRecord 根据 txId 查询 pos 奖励领取记录
func getPosClaimRecord(txId string) (*model.PosRewardClaim, error) {
	rsp, err := postJsonRequest[model.PosRewardClaim](
		SnapURL+"/pos/reward/claim-record",
		struct {
			TxId string `json:"txId"`
		}{TxId: txId},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

// ---- 交易构建 ----

// createPosRewardHexEncodedTx 构造 pos 奖励领取交易并返回 hex 编码字节。
//
// 指令顺序：
//  1. SetComputeUnitPrice（Medium 档）
//  2. SetComputeUnitLimit（instUnits.TransferChecked + systemTransferCU）× 1.2
//  3. [可选] CreateAssociatedTokenAccount（用户 token account 不存在时）
//  4. TransferChecked：rewardTokenAccount → userTokenAccount（amount = TotalReward）
//  5. System.Transfer：userNativeAccount → costAccount（amount = CostFee）
func createPosRewardHexEncodedTx(ctx context.Context, privKey solana.PrivateKey, txInfo *types.ClaimPosRewardTxInfo) (string, error) {
	userPubKey := privKey.PublicKey()
	tokenMintAccount := solana.MPK(TokenMintAddress)
	rewardNativeAccount := solana.MPK(txInfo.RewardAccount)
	rewardTokenAccount, _, _ := solana.FindAssociatedTokenAddress(rewardNativeAccount, tokenMintAccount)
	costAccount := solana.MPK(txInfo.CostAccount)
	decimals := uint8(txInfo.Decimals)

	fmt.Printf("  user native account : %s\n", userPubKey.String())
	fmt.Printf("  reward account      : %s\n", txInfo.RewardAccount)
	fmt.Printf("  total reward (raw)  : %.0f\n", txInfo.TotalReward)
	fmt.Printf("  cost fee (lamports) : %.0f\n", txInfo.CostFee)

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

	// 优先费
	computePrice, err := getPriorityFee()
	if err != nil {
		return "", fmt.Errorf("getPriorityFee failed: %w", err)
	}
	// CU 预算
	instUnits, err := getInstUnits()
	if err != nil {
		return "", fmt.Errorf("getInstUnits failed: %w", err)
	}
	computeUnitLimit := uint32(float64(instUnits.TransferChecked+systemTransferCU) * 1.2)
	fmt.Printf("  computePrice=%d computeUnitLimit=%d\n", computePrice, computeUnitLimit)

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
		// 平台发放 pos 奖励
		token.NewTransferCheckedInstructionBuilder().
			SetAmount(uint64(txInfo.TotalReward)).
			SetDecimals(decimals).
			SetSourceAccount(rewardTokenAccount).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(*userTokenAccount).
			SetOwnerAccount(rewardNativeAccount).
			Build(),
		// 用户支付 SOL 成本费
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

	msgBin, err := tx.Message.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("MarshalBinary failed: %w", err)
	}
	sig, err := privKey.Sign(msgBin)
	if err != nil {
		return "", fmt.Errorf("Sign failed: %w", err)
	}
	tx.Signatures = append(tx.Signatures, sig)

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("tx.MarshalBinary failed: %w", err)
	}
	return hex.EncodeToString(txBytes), nil
}

// ---- 测试用例 ----

// TestGetPosRewardConfig 查询 pos 奖励配置（无需 JWT）
func TestGetPosRewardConfig(t *testing.T) {
	config, err := getPosRewardConfig()
	if err != nil {
		t.Fatalf("getPosRewardConfig failed: %v", err)
	}

	fmt.Printf("=== Pos Reward Config ===\n")
	fmt.Printf("RewardAccount    : %s\n", config.RewardAccount)
	fmt.Printf("CostAccount      : %s\n", config.CostAccount)
	fmt.Printf("CostFeeRate      : %v\n", config.CostFeeRate)
	fmt.Printf("MaxCostFee       : %v\n", config.MaxCostFee)
	fmt.Printf("QuoteTokenAmount : %v\n", config.QuoteTokenAmount)
}

// TestGetPosRewardRecords 查询未领取的 pos 奖励列表
func TestGetPosRewardRecords(t *testing.T) {
	privKey := loadTestPrivateKey(t, "david")
	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	rewards, err := getPosRewardRecords(jwtToken)
	if err != nil {
		t.Fatalf("getPosRewardRecords failed: %v", err)
	}

	fmt.Printf("=== Pos Reward Records (count=%d) ===\n", len(rewards))
	for i, r := range rewards {
		fmt.Printf("[%d] RecordID=%s Amount=%.0f RewardType=%d SnapDay=%s Pending=%v\n",
			i, r.RecordID, r.RewardAmount, r.RewardType, r.SnapDay.Format("2006-01-02"), r.Pending)
	}
}

// TestGetPosRewardTxInfo 获取领取 pos 奖励的交易参数
func TestGetPosRewardTxInfo(t *testing.T) {
	privKey := loadTestPrivateKey(t, "david")
	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	txInfo, err := getPosRewardTxInfo(jwtToken)
	if err != nil {
		t.Fatalf("getPosRewardTxInfo failed: %v", err)
	}

	fmt.Printf("=== Pos Reward TxInfo ===\n")
	fmt.Printf("RewardAccount : %s\n", txInfo.RewardAccount)
	fmt.Printf("Mint          : %s\n", txInfo.Mint)
	fmt.Printf("CostAccount   : %s\n", txInfo.CostAccount)
	fmt.Printf("Decimals      : %d\n", txInfo.Decimals)
	fmt.Printf("TotalReward   : %.0f\n", txInfo.TotalReward)
	fmt.Printf("CostFee       : %.0f lamports\n", txInfo.CostFee)
}

// TestClaimPosReward 完整领取流程：tx-info → 构建交易 → commit-tx → 轮询确认
func TestClaimPosReward(t *testing.T) {
	ctx := context.Background()
	privKey := loadTestPrivateKey(t, "david")
	pubKey := privKey.PublicKey()

	// 1. 登录获取 JWT
	jwtToken, err := loginForToken(Brand, Symbol, pubKey, privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}
	fmt.Printf("登录账户: %s\n", pubKey.String())

	// 2. 查询待领取奖励（仅展示，不阻断流程）
	rewards, err := getPosRewardRecords(jwtToken)
	if err != nil {
		t.Logf("getPosRewardRecords failed (non-fatal): %v", err)
	} else {
		var total float64
		for _, r := range rewards {
			total += r.RewardAmount
		}
		fmt.Printf("待领取奖励: %d 条，合计 raw=%.0f\n", len(rewards), total)
	}

	// 3. 获取交易参数（服务端汇总可领取总额，不信任第 2 步的结果）
	txInfo, err := getPosRewardTxInfo(jwtToken)
	if err != nil {
		t.Fatalf("getPosRewardTxInfo failed: %v", err)
	}
	if txInfo.TotalReward == 0 {
		t.Skip("当前账户无可领取 pos 奖励，跳过测试")
	}
	fmt.Printf("交易参数: TotalReward=%.0f CostFee=%.0f\n", txInfo.TotalReward, txInfo.CostFee)

	// 4. 构建并签名交易
	encodedTx, err := createPosRewardHexEncodedTx(ctx, privKey, txInfo)
	if err != nil {
		t.Fatalf("createPosRewardHexEncodedTx failed: %v", err)
	}

	// 5. 提交交易
	txId, err := commitPosRewardTx(encodedTx)
	if err != nil {
		t.Fatalf("commitPosRewardTx failed: %v", err)
	}
	fmt.Printf("交易提交成功: https://explorer.solana.com/tx/%s?cluster=devnet\n", txId)

	// 6. 轮询 claim record，等待链上确认（最多 3 分钟）
	fmt.Println("轮询链上确认状态...")
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		record, err := getPosClaimRecord(txId)
		if err != nil {
			fmt.Printf("  [轮询] 查询 claim record 失败 (可能尚未写入): %v\n", err)
			time.Sleep(3 * time.Second)
			continue
		}
		switch record.TxState {
		case 1:
			fmt.Printf("  [轮询] 链上确认成功！TxState=%d\n", record.TxState)
			fmt.Printf("=== 领取完成 ===\n")
			fmt.Printf("ClaimRecord RecordID : %s\n", record.RecordID)
			fmt.Printf("RewardIds            : %s\n", record.RewardIds)
			fmt.Printf("TxID                 : %s\n", record.TxID)
			return
		case -1:
			t.Fatalf("交易链上执行失败，TxState=%d", record.TxState)
		default:
			fmt.Printf("  [轮询] 等待链上确认，TxState=%d ...\n", record.TxState)
			time.Sleep(3 * time.Second)
		}
	}
	t.Fatalf("等待链上确认超时（3 分钟）")
}

// TestGetPosClaimRecord 根据 txId 查询 pos 奖励领取记录（用于调试已知交易）
func TestGetPosClaimRecord(t *testing.T) {
	// 替换为实际的 txId
	txId := "replace-with-actual-tx-id"

	record, err := getPosClaimRecord(txId)
	if err != nil {
		t.Fatalf("getPosClaimRecord failed: %v", err)
	}

	fmt.Printf("=== Pos Claim Record ===\n")
	fmt.Printf("RecordID  : %s\n", record.RecordID)
	fmt.Printf("RewardIds : %s\n", record.RewardIds)
	fmt.Printf("TxID      : %s\n", record.TxID)
	fmt.Printf("TxState   : %d (0=pending, 1=success, -1=failed)\n", record.TxState)
	fmt.Printf("CreatedAt : %s\n", record.CreatedAt.Format("2006-01-02 15:04:05"))
}
