package test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
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

// ---- 常量 ----

const (
	Brand           = "OShit"
	Symbol          = "OShit"
	SOLLoginSignMsg = "I am login %s for token %s with my address %s with nonce %d"

	// devnet
	RpcUrl           = "https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/"
	TokenMintAddress = "wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM"

	// 测试账户（devnet）
	DavidPrivate      = "4LRdCeZ4EYHsGtRr99zbh6WLZCa7jACVuhunzrbgyKvP4NxKuHPnaJ52t1AKzREbaD5n2NMsKYkdPMcnYRMacjgU"
	DavidNativePubKey = "JAZtFeZfLeeVtWS4vrruCpTa5LdASRDJuMe7yLKbkJk"

	BaseURL = "http://localhost:1100/base"
	SnapURL = "http://localhost:1300/snap"

	//BaseURL = "https://beta.testnet.oshit.io/meme/base/api/v1"
	//SnapURL = "https://beta.testnet.oshit.io/meme/snap/api/v1"

	systemTransferCU uint64 = 500 // System.Transfer 固定预留 CU
)

var rpcClient = rpc.New(RpcUrl)

// ---- 公共数据类型 ----

type JwtToken = string

type ApiResponse[T any] struct {
	Code  int    `json:"code"`
	Count int    `json:"count"`
	Data  T      `json:"data"`
	Msg   string `json:"msg"`
}

type InstUnitsRsp struct {
	MiniRent          uint64 `json:"mini_rent"`
	AssociatedAccount uint64 `json:"associated_account"`
	TransferChecked   uint64 `json:"transfer_checked"`
	Memo              uint64 `json:"memo"`
}

type loginReqBody struct {
	Brand      string `json:"brand"`
	Symbol     string `json:"symbol"`
	Account    string `json:"account"`
	Sign       string `json:"sign"`
	Nonce      uint64 `json:"nonce"`
	InviteCode string `json:"invite_code"`
}

type loginRspData struct {
	Token string `json:"token"`
}

// ---- HTTP 工具函数 ----

func postJsonRequest[T any](reqURL string, body any, headers map[string]string) (*ApiResponse[T], error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body error: %w", err)
	}
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var apiResp ApiResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}
	return &apiResp, nil
}

func getJsonRequest[T any](reqURL string, headers map[string]string) (*ApiResponse[T], error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var apiResp ApiResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}
	return &apiResp, nil
}

// ---- 通用业务函数 ----

func loginForToken(brand, symbol string, nativeAccount solana.PublicKey, privKey solana.PrivateKey) (string, error) {
	nonce := uint64(time.Now().UnixMilli())
	msg := fmt.Sprintf(SOLLoginSignMsg, brand, symbol, nativeAccount.String(), nonce)
	sign, err := privKey.Sign([]byte(msg))
	if err != nil {
		return "", fmt.Errorf("sign message error: %w", err)
	}
	body := loginReqBody{
		Brand:   brand,
		Symbol:  symbol,
		Account: nativeAccount.String(),
		Sign:    sign.String(),
		Nonce:   nonce,
	}
	rsp, err := postJsonRequest[loginRspData](BaseURL+"/auth/login", body, nil)
	if err != nil {
		return "", err
	}
	return rsp.Data.Token, nil
}

func getPriorityFee() (uint64, error) {
	type FeeLevel struct {
		Low     uint64 `json:"Low"`
		Medium  uint64 `json:"Medium"`
		High    uint64 `json:"High"`
		Extreme uint64 `json:"Extreme"`
	}
	type PriorityFeeRsp struct {
		PerComputeUnit FeeLevel `json:"PerComputeUnit"`
		PerTransaction FeeLevel `json:"PerTransaction"`
	}
	rsp, err := getJsonRequest[PriorityFeeRsp](BaseURL+"/fee/priority", nil)
	if err != nil {
		return 0, err
	}
	return rsp.Data.PerComputeUnit.Medium, nil
}

func getInstUnits() (*InstUnitsRsp, error) {
	rsp, err := getJsonRequest[InstUnitsRsp](BaseURL+"/fee/inst-units", nil)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

// ---- Stake 奖励 API 调用 ----

// getStakeRewardRecords 查询未领取的 stake 奖励列表（需要 JWT）
func getStakeRewardRecords(jwtToken string) ([]model.StakeReward, error) {
	rsp, err := postJsonRequest[[]model.StakeReward](
		SnapURL+"/stake/reward/record",
		struct{}{},
		map[string]string{"Authorization": "Bearer " + jwtToken},
	)
	if err != nil {
		return nil, err
	}
	return rsp.Data, nil
}

// getStakeRewardTxInfo 获取领取交易所需参数（需要 JWT）
func getStakeRewardTxInfo(jwtToken string) (*types.ClaimStakeRewardTxInfo, error) {
	rsp, err := postJsonRequest[types.ClaimStakeRewardTxInfo](
		SnapURL+"/stake/reward/tx-info",
		struct{}{},
		map[string]string{"Authorization": "Bearer " + jwtToken},
	)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

// commitStakeRewardTx 提交已签名的领取交易
func commitStakeRewardTx(encodedTx string) (string, error) {
	rsp, err := postJsonRequest[string](
		SnapURL+"/stake/reward/commit-tx",
		types.CommitStakeRewardTxReq{EncodedTx: encodedTx},
		nil,
	)
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

// getStakeClaimRecord 根据 txId 查询领取记录
func getStakeClaimRecord(txId string) (*model.StakeRewardClaim, error) {
	rsp, err := postJsonRequest[model.StakeRewardClaim](
		SnapURL+"/stake/reward/claim-record",
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

// createStakeRewardHexEncodedTx 构造 stake 奖励领取交易并返回 hex 编码字节。
//
// 指令顺序：
//  1. SetComputeUnitPrice（Medium 档）
//  2. SetComputeUnitLimit（instUnits.TransferChecked + systemTransferCU）× 1.2
//  3. [可选] CreateAssociatedTokenAccount（用户 token account 不存在时）
//  4. TransferChecked：rewardTokenAccount → userTokenAccount（amount = TotalReward）
//  5. System.Transfer：userNativeAccount → costAccount（amount = CostFee）
func createStakeRewardHexEncodedTx(ctx context.Context, privKey solana.PrivateKey, txInfo *types.ClaimStakeRewardTxInfo) (string, error) {
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
		// 平台发放 stake 奖励
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

// TestGetStakeRewardRecords 查询未领取的 stake 奖励列表
func TestGetStakeRewardRecords(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	rewards, err := getStakeRewardRecords(jwtToken)
	if err != nil {
		t.Fatalf("getStakeRewardRecords failed: %v", err)
	}

	fmt.Printf("=== Stake Reward Records (count=%d) ===\n", len(rewards))
	for i, r := range rewards {
		fmt.Printf("[%d] RecordID=%s Amount=%.0f StakeType=%d RewardType=%d SnapDay=%s\n",
			i, r.RecordID, r.RewardAmount, r.StakeType, r.RewardType, r.SnapDay.Format("2006-01-02"))
	}
}

// TestGetStakeRewardTxInfo 获取领取 stake 奖励的交易参数
func TestGetStakeRewardTxInfo(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	jwtToken, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}

	txInfo, err := getStakeRewardTxInfo(jwtToken)
	if err != nil {
		t.Fatalf("getStakeRewardTxInfo failed: %v", err)
	}

	fmt.Printf("=== Stake Reward TxInfo ===\n")
	fmt.Printf("RewardAccount : %s\n", txInfo.RewardAccount)
	fmt.Printf("Mint          : %s\n", txInfo.Mint)
	fmt.Printf("CostAccount   : %s\n", txInfo.CostAccount)
	fmt.Printf("Decimals      : %d\n", txInfo.Decimals)
	fmt.Printf("TotalReward   : %.0f\n", txInfo.TotalReward)
	fmt.Printf("QuoteAmount   : %.0f lamports\n", txInfo.QuoteAmount)
	fmt.Printf("CostFeeRate   : %d\n", txInfo.CostFeeRate)
	fmt.Printf("CostFee       : %.0f lamports\n", txInfo.CostFee)
}

// TestClaimStakeReward 完整领取流程：tx-info → 构建交易 → commit-tx → 轮询确认
func TestClaimStakeReward(t *testing.T) {
	ctx := context.Background()
	privKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	pubKey := privKey.PublicKey()

	// 1. 登录获取 JWT
	jwtToken, err := loginForToken(Brand, Symbol, pubKey, privKey)
	if err != nil {
		t.Fatalf("loginForToken failed: %v", err)
	}
	fmt.Printf("登录账户: %s\n", pubKey.String())

	// 2. 查询待领取奖励（仅展示，不阻断流程）
	rewards, err := getStakeRewardRecords(jwtToken)
	if err != nil {
		t.Logf("getStakeRewardRecords failed (non-fatal): %v", err)
	} else {
		var total float64
		for _, r := range rewards {
			total += r.RewardAmount
		}
		fmt.Printf("待领取奖励: %d 条，合计 raw=%.0f\n", len(rewards), total)
	}

	// 3. 获取交易参数（服务端汇总可领取总额，不信任第 2 步的结果）
	txInfo, err := getStakeRewardTxInfo(jwtToken)
	if err != nil {
		t.Fatalf("getStakeRewardTxInfo failed: %v", err)
	}
	fmt.Printf("交易参数: TotalReward=%.0f CostFee=%.0f\n", txInfo.TotalReward, txInfo.CostFee)

	// 4. 构建并签名交易
	encodedTx, err := createStakeRewardHexEncodedTx(ctx, privKey, txInfo)
	if err != nil {
		t.Fatalf("createStakeRewardHexEncodedTx failed: %v", err)
	}

	// 5. 提交交易
	txId, err := commitStakeRewardTx(encodedTx)
	if err != nil {
		t.Fatalf("commitStakeRewardTx failed: %v", err)
	}
	fmt.Printf("交易提交成功: https://explorer.solana.com/tx/%s?cluster=devnet\n", txId)

	// 6. 轮询 claim record，等待链上确认（最多 3 分钟）
	fmt.Println("轮询链上确认状态...")
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		record, err := getStakeClaimRecord(txId)
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

// TestGetStakeClaimRecord 根据 txId 查询领取记录（用于调试已知交易）
func TestGetStakeClaimRecord(t *testing.T) {
	// 替换为实际的 txId
	txId := "replace-with-actual-tx-id"

	record, err := getStakeClaimRecord(txId)
	if err != nil {
		t.Fatalf("getStakeClaimRecord failed: %v", err)
	}

	fmt.Printf("=== Stake Claim Record ===\n")
	fmt.Printf("RecordID  : %s\n", record.RecordID)
	fmt.Printf("RewardIds : %s\n", record.RewardIds)
	fmt.Printf("TxID      : %s\n", record.TxID)
	fmt.Printf("TxState   : %d (0=pending, 1=success, -1=failed)\n", record.TxState)
	fmt.Printf("CreatedAt : %s\n", record.CreatedAt.Format("2006-01-02 15:04:05"))
}
