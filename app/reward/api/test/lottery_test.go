package test

import (
	"context"
	"encoding/hex"
	"fmt"
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

// ---- Lottery API 调用 ----

func getLotteryStatus(jwtToken string) (*model.DailyClaimStats, error) {
	rsp, err := postJsonRequest[*model.DailyClaimStats](RewardURL+"/lottery/status", struct{}{}, map[string]string{
		"Authorization": "Bearer " + jwtToken,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Data, nil
}

func executeLottery(jwtToken string) (*model.LotteryReward, error) {
	rsp, err := postJsonRequest[model.LotteryReward](RewardURL+"/lottery/execute", struct{}{}, map[string]string{
		"Authorization": "Bearer " + jwtToken,
	})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func getUnclaimedLotteryRewards(nativeAccount string) ([]*model.LotteryReward, error) {
	rsp, err := postJsonRequest[[]*model.LotteryReward](RewardURL+"/lottery/unclaimed", types.GetUnclaimedLotteryReq{
		NativeAccount: nativeAccount,
	}, map[string]string{})
	if err != nil {
		return nil, err
	}
	return rsp.Data, nil
}

func getLotteryTxInfo(recordId string) (*types.ClaimLotteryTxInfo, error) {
	rsp, err := postJsonRequest[types.ClaimLotteryTxInfo](RewardURL+"/lottery/tx-info", types.GetLotteryTxInfoReq{
		RecordId: recordId,
	}, map[string]string{})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func commitLotteryTx(req types.CommitLotteryTxReq) (string, error) {
	rsp, err := postJsonRequest[string](RewardURL+"/lottery/commit-tx", req, map[string]string{})
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

func getLotteryRecord(txId string) (*model.LotteryClaimRecord, error) {
	rsp, err := postJsonRequest[model.LotteryClaimRecord](RewardURL+"/lottery/record", types.GetByTxIdReq{
		TxId: txId,
	}, map[string]string{})
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

// createLotteryHexEncodedTx 根据 privKey 构造并签名抽奖领取交易，返回 hex 编码的交易字节。
//
// 指令顺序：
//  1. SetComputeUnitPrice（Medium 档）
//  2. SetComputeUnitLimit（根据 instUnits 计算）
//  3. [可选] CreateAssociatedTokenAccount（用户 PDA 不存在时）
//  4. TransferChecked：rewardTokenAccount → userTokenAccount（平台发放抽奖奖励）
//  5. System.Transfer：userNativeAccount → costAccount（SOL 成本费）
func createLotteryHexEncodedTx(ctx context.Context, privKey solana.PrivateKey, txInfo *types.ClaimLotteryTxInfo) (string, error) {
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
	fmt.Printf("lottery amount      : %v\n", txInfo.LotteryAmount)
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

	// 若不存在，使用 PDA 推导地址
	if !userPDAExist {
		derived, _, _ := solana.FindAssociatedTokenAddress(userPubKey, tokenMintAccount)
		userTokenAccount = &derived
	}

	computePrice := priorityFee.PerComputeUnit.Medium
	fmt.Printf("compute price: %d\n", computePrice)

	// CU 计算：1 条 TransferChecked + 1 条 System.Transfer
	computeUnitLimit := uint32(float64(instUnits.TransferChecked+systemTransferCU) * 1.2)
	fmt.Printf("computeUnitLimit: %d (perTC=%d)\n", computeUnitLimit, instUnits.TransferChecked)

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
		// 平台发放抽奖奖励给用户
		token.NewTransferCheckedInstructionBuilder().
			SetAmount(uint64(txInfo.LotteryAmount)).
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

// ---- Lottery 测试用例 ----

func TestGetLotteryStatus(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}
	stats, err := getLotteryStatus(jwtToken)
	if err != nil {
		t.Fatalf("getLotteryStatus failed: %v", err)
	}

	if stats == nil {
		fmt.Println("=== Lottery Status: 今日暂无领取记录 ===")
		return
	}

	fmt.Printf("=== Lottery Status ===\n")
	fmt.Printf("NativeAccount : %s\n", stats.NativeAccount)
	fmt.Printf("TakeDate      : %s\n", stats.TakeDate)
	fmt.Printf("TakeCount     : %d\n", stats.TakeCount)
	fmt.Printf("NeedLottery   : %v\n", stats.NeedLottery)
	fmt.Printf("LotteryCount  : %d\n", stats.LotteryCount)
}

func TestGetUnclaimedLotteryRewards(t *testing.T) {
	rewards, err := getUnclaimedLotteryRewards(AliceNativePubKey)
	if err != nil {
		t.Fatalf("getUnclaimedLotteryRewards failed: %v", err)
	}

	fmt.Printf("=== Unclaimed Lottery Rewards (count=%d) ===\n", len(rewards))
	for i, r := range rewards {
		fmt.Printf("[%d] RecordID=%s Amount=%v RewardState=%d Pending=%v RewardDay=%s\n",
			i, r.RecordID, r.RewardAmount, r.State, r.Pending, r.RewardDay.Format("2006-01-02"))
	}
}

func TestExecuteLottery(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}

	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	reward, err := executeLottery(jwtToken)
	if err != nil {
		t.Fatalf("executeLottery failed: %v", err)
	}

	fmt.Printf("=== Execute Lottery ===\n")
	fmt.Printf("RecordID     : %s\n", reward.RecordID)
	fmt.Printf("RewardAmount : %v\n", reward.RewardAmount)
	fmt.Printf("RewardState        : %d\n", reward.State)
	fmt.Printf("Pending      : %v\n", reward.Pending)
	fmt.Printf("RewardDay    : %s\n", reward.RewardDay.Format("2006-01-02"))
}

func TestGetLotteryTxInfo(t *testing.T) {
	// 替换为实际执行抽奖后返回的 RecordID
	recordId := "replace-with-actual-record-id"
	txInfo, err := getLotteryTxInfo(recordId)
	if err != nil {
		t.Fatalf("getLotteryTxInfo failed: %v", err)
	}

	fmt.Printf("=== Lottery TxInfo ===\n")
	fmt.Printf("RecordId      : %s\n", txInfo.RecordId)
	fmt.Printf("RewardAccount : %s\n", txInfo.RewardAccount)
	fmt.Printf("Mint          : %s\n", txInfo.Mint)
	fmt.Printf("CostAccount   : %s\n", txInfo.CostAccount)
	fmt.Printf("Decimals      : %d\n", txInfo.Decimals)
	fmt.Printf("LotteryAmount : %v\n", txInfo.LotteryAmount)
	fmt.Printf("CostFee       : %v\n", txInfo.CostFee)
}

// TestLottery 执行完整抽奖领取流程：execute → tx-info → 构建交易 → commit-tx
func TestLottery(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	pubKey := privKey.PublicKey()

	// 1. 登录获取 JWT
	jwtToken, err := loginForToken(Brand, Symbol, pubKey, privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	// 2. 执行抽奖
	reward, err := executeLottery(jwtToken)
	if err != nil {
		t.Fatalf("executeLottery failed: %v", err)
	}
	fmt.Printf("抽奖记录: RecordID=%s Amount=%v\n", reward.RecordID, reward.RewardAmount)

	// 3. 获取交易信息
	txInfo, err := getLotteryTxInfo(reward.RecordID)
	if err != nil {
		t.Fatalf("getLotteryTxInfo failed: %v", err)
	}
	fmt.Printf("交易信息: LotteryAmount=%v CostFee=%v\n", txInfo.LotteryAmount, txInfo.CostFee)

	// 4. 构建并签名交易
	encodedTx, err := createLotteryHexEncodedTx(context.Background(), privKey, txInfo)
	if err != nil {
		t.Fatalf("createLotteryHexEncodedTx failed: %v", err)
	}

	// 5. 提交交易
	txId, err := commitLotteryTx(types.CommitLotteryTxReq{
		EncodedTx: encodedTx,
		RewardId:  txInfo.RecordId,
	})
	if err != nil {
		t.Fatalf("commitLotteryTx failed: %v", err)
	}
	fmt.Printf("请求成功: https://solscan.io/tx/%s?cluster=devnet\n", txId)
}

// ---- Take & Lottery 联合测试辅助函数 ----

// doOneTakeToken 构建并提交一次 take token，返回 txId
func doOneTakeToken(t *testing.T, n int, privKey solana.PrivateKey, inviteCode string) string {
	t.Helper()
	encodedTx, err := createHexEncodedTx(context.Background(), privKey, inviteCode)
	if err != nil {
		t.Fatalf("第 %d 次 createHexEncodedTx failed: %v", n, err)
	}
	txId, err := commitTakeTokenTx(types.CommitTakeTokenTxInfoReq{
		EncodedTx:  encodedTx,
		InviteCode: inviteCode,
	})
	if err != nil {
		t.Fatalf("第 %d 次 commitTakeTokenTx failed: %v", n, err)
	}
	fmt.Printf("  第 %d 次 take token: https://solscan.io/tx/%s?cluster=devnet\n", n, txId)
	return txId
}

// waitNeedLottery 轮询直到 need_lottery=true，超时则 Fatal
func waitNeedLottery(t *testing.T, jwtToken string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		stats, err := getLotteryStatus(jwtToken)
		if err == nil && stats != nil && stats.NeedLottery {
			fmt.Printf("  [轮询] need_lottery=true (take_count=%d, lottery_count=%d)\n",
				stats.TakeCount, stats.LotteryCount)
			return
		}
		takeCount := int32(0)
		if err == nil && stats != nil {
			takeCount = stats.TakeCount
		}
		fmt.Printf("  [轮询] 等待 need_lottery=true，当前 take_count=%d ...\n", takeCount)
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("等待 need_lottery=true 超时 (%v)", timeout)
}

// waitLotteryDone 轮询直到 need_lottery=false，超时则 Fatal
func waitLotteryDone(t *testing.T, jwtToken string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		stats, err := getLotteryStatus(jwtToken)
		if err == nil && stats != nil && !stats.NeedLottery {
			fmt.Printf("  [轮询] need_lottery=false，抽奖已结算\n")
			return
		}
		fmt.Printf("  [轮询] 等待 need_lottery=false（lottery tx 上链中）...\n")
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("等待 need_lottery=false 超时 (%v)", timeout)
}

// doFullLotteryFlow 执行完整抽奖流程：execute → tx-info → commit-tx
func doFullLotteryFlow(t *testing.T, privKey solana.PrivateKey) {
	t.Helper()
	pubKey := privKey.PublicKey()

	jwtToken, err := loginForToken(Brand, Symbol, pubKey, privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	reward, err := executeLottery(jwtToken)
	if err != nil {
		t.Fatalf("executeLottery failed: %v", err)
	}
	fmt.Printf("  抽奖记录: RecordID=%s Amount=%v\n", reward.RecordID, reward.RewardAmount)

	txInfo, err := getLotteryTxInfo(reward.RecordID)
	if err != nil {
		t.Fatalf("getLotteryTxInfo failed: %v", err)
	}
	fmt.Printf("  交易信息: LotteryAmount=%v CostFee=%v\n", txInfo.LotteryAmount, txInfo.CostFee)

	encodedTx, err := createLotteryHexEncodedTx(context.Background(), privKey, txInfo)
	if err != nil {
		t.Fatalf("createLotteryHexEncodedTx failed: %v", err)
	}

	lotteryTxId, err := commitLotteryTx(types.CommitLotteryTxReq{
		EncodedTx: encodedTx,
		RewardId:  txInfo.RecordId,
	})
	if err != nil {
		t.Fatalf("commitLotteryTx failed: %v", err)
	}
	fmt.Printf("  lottery tx: https://solscan.io/tx/%s?cluster=devnet\n", lotteryTxId)
}

// TestTakeAndLottery 模拟连续 take token，分别在第 5、10、20 次触发抽奖
func TestTakeAndLottery(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	inviteCode := ""
	pollTimeout := 3 * time.Minute

	// 登录一次，JWT 用于状态轮询（有效期内无需重新登录）
	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	// ── 第 1～4 次：不触发抽奖 ──
	fmt.Println("=== Take 1~4 ===")
	for i := 1; i <= 4; i++ {
		doOneTakeToken(t, i, privKey, inviteCode)
	}

	// ── 第 5 次：触发抽奖阈值 ──
	fmt.Println("=== Take 5（触发抽奖）===")
	doOneTakeToken(t, 5, privKey, inviteCode)
	waitNeedLottery(t, jwtToken, pollTimeout)

	fmt.Println("=== Lottery #1（第 5 次后）===")
	doFullLotteryFlow(t, privKey)
	waitLotteryDone(t, jwtToken, pollTimeout)

	// ── 第 6～9 次：不触发抽奖 ──
	fmt.Println("=== Take 6~9 ===")
	for i := 6; i <= 9; i++ {
		doOneTakeToken(t, i, privKey, inviteCode)
	}

	// ── 第 10 次：触发抽奖阈值 ──
	fmt.Println("=== Take 10（触发抽奖）===")
	doOneTakeToken(t, 10, privKey, inviteCode)
	waitNeedLottery(t, jwtToken, pollTimeout)

	fmt.Println("=== Lottery #2（第 10 次后）===")
	doFullLotteryFlow(t, privKey)
	waitLotteryDone(t, jwtToken, pollTimeout)

	// ── 第 11～19 次：不触发抽奖 ──
	fmt.Println("=== Take 11~19 ===")
	for i := 11; i <= 19; i++ {
		doOneTakeToken(t, i, privKey, inviteCode)
	}

	// ── 第 20 次：触发抽奖阈值 ──
	fmt.Println("=== Take 20（触发抽奖）===")
	doOneTakeToken(t, 20, privKey, inviteCode)
	waitNeedLottery(t, jwtToken, pollTimeout)

	fmt.Println("=== Lottery #3（第 20 次后）===")
	doFullLotteryFlow(t, privKey)

	fmt.Println("=== TestTakeAndLottery 完成 ===")
}

// TestGetLotteryRecord 根据 txId 查询抽奖领取记录
func TestGetLotteryRecord(t *testing.T) {
	// 替换为实际的 txId
	txId := "replace-with-actual-tx-id"

	record, err := getLotteryRecord(txId)
	if err != nil {
		t.Fatalf("getLotteryRecord failed: %v", err)
	}

	fmt.Printf("=== Lottery Claim Record ===\n")
	fmt.Printf("RecordID  : %s\n", record.RecordID)
	fmt.Printf("RewardIds : %s\n", record.RewardIds)
	fmt.Printf("TxID      : %s\n", record.TxID)
	fmt.Printf("RewardState     : %d\n", record.State)
	fmt.Printf("CreatedAt : %s\n", record.CreatedAt.Format("2006-01-02 15:04:05"))
}
