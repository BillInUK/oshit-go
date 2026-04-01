package test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"net/http"
	"net/url"
	"oshit-go/common/utils"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"oshit-go/app/reward/api/types"
	"oshit-go/common/pkg/entity"
)

const (
	SOLLoginSignMsg               = "I am login %s for token %s with my address %s with nonce %d"
	SOLLoginWithInviteCodeSignMsg = "I am login %s for token %s with my address %s with nonce %d inviteCode %s"
	// mainnet:
	//RpcUrl           = "https://radial-purple-sailboat.solana-mainnet.quiknode.pro/0afcb192bb26b0dbcba3d49df6ad2ee2829c529b/"
	//WssUrl           = "wss://radial-purple-sailboat.solana-mainnet.quiknode.pro/0afcb192bb26b0dbcba3d49df6ad2ee2829c529b/"
	//TokenMintAddress = "ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"
	//TokenMintAddress = "3zyKg7L471V9FGb4LGMLxpygY7M8srCAzQtM1HWgCixy"

	// testnet:
	RpcUrl           = "https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/"
	WssUrl           = "wss://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/"
	TokenMintAddress = "wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM"
	//TokenMintAddress = "9TXFPq4UnFismeQyQmnJ6waJguhkevSLvxVyvTA29QMD"

	MainPrivate      = "3coeLXLqdWi9kwSGDrNZZWf5bv6mBW2Zmd8gBSYJpqiGjy73cpmFYHM33Nn5Sxr2i2jJcE4xRXqDuDPTRtn3iHsz"
	MainNativePubKey = "GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E"

	AlicePrivate      = "46dPKNS2nHDuaGrdJ9tyr7mJ1H6Gu8EKcvPNDmC8zZ2xf1SpuhzR9kfK7uA3mf1RsfwoYvJDYdjA4WPkjtSv1r4E"
	AliceNativePubKey = "5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN"

	BobPrivate      = "43HXEVBSJzsfA9PQpYbE9GFuAg1YwcKuLiecqvqUVMg8p8EcDSXgQShBMBsqURWogs7k173BehCWF2y17VEXnZkn"
	BobNativePubKey = "27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx"

	DavidPrivate      = "4LRdCeZ4EYHsGtRr99zbh6WLZCa7jACVuhunzrbgyKvP4NxKuHPnaJ52t1AKzREbaD5n2NMsKYkdPMcnYRMacjgU"
	DavidNativePubKey = "JAZtFeZfLeeVtWS4vrruCpTa5LdASRDJuMe7yLKbkJk"

	RobertPrivate      = "MzJGrbzW1yqSzAAbGLHbSkHKmGFS6kuhACV8wFxePqSwe6rTsv7N3eRVozdJcJSBQAPT6rjtnmoCFxv6YuA5hGq"
	RobertNativePubKey = "6HLScqNL4EQWLk8DTcB4hXUrHjDkVbeP2a3Sc5VtHozM"

	DexNativePubKey = "6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K"

	//BaseURL = "https://testnet.oshit.io/meme/api/v1"
	//BaseURL = "https://oshit.io/meme/api/v1"

	BaseURL   = "http://localhost:1100/base"
	RewardURL = "http://localhost:1200/reward"

	Brand  = "OShit"
	Symbol = "OShit"
)

var rpcClient = rpc.New(RpcUrl)

type JwtToken struct {
	Access  string `json:"Access"`
	Refresh string `json:"Refresh"`
}

// ApiResponse represents a generic API response structure.
type ApiResponse[T any] struct {
	Code  int    `json:"code"`
	Count int    `json:"count"`
	Data  T      `json:"data"`
	Msg   string `json:"msg"`
}

type TransactionParams struct {
	TokenQuoteSOLPrice float64                  `json:"tokenQuoteSOLPrice"`
	PriorityFee        entity.PriorityFee       `json:"priorityFee"`
	ConsumedUnits      entity.ComputeUnitDetail `json:"consumedUnits"`
	DiscountRate       float64                  `json:"discountRate"`
}

// postJsonRequest sends a POST request with a JSON body and parses the response.
func postJsonRequest[T any](reqURL string, body any, headers map[string]string) (*ApiResponse[T], error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse ApiResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	return &apiResponse, nil
}

// postFormRequest sends a POST request with form data and parses the response.
func postFormRequest[T any](url string, formData url.Values, headers map[string]string) (*ApiResponse[T], error) {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse ApiResponse[T]
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	return &apiResponse, nil
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
	Token JwtToken `json:"token"`
}

func loginForToken(brand, symbol string, nativeAccount solana.PublicKey, privKey solana.PrivateKey) (*JwtToken, error) {
	nonce := uint64(time.Now().UnixMilli())
	msg := fmt.Sprintf(SOLLoginSignMsg, brand, symbol, nativeAccount.String(), nonce)
	sign, err := privKey.Sign([]byte(msg))
	if err != nil {
		return nil, fmt.Errorf("sign message error: %w", err)
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
		return nil, err
	}
	return &rsp.Data.Token, nil
}

// getJsonRequest sends a GET request and parses the response.
func getJsonRequest[T any](reqURL string, headers map[string]string) (*ApiResponse[T], error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse ApiResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}
	return &apiResponse, nil
}

// InstUnitsRsp 与 base 模块 /fee/inst-units 响应字段保持一致
type InstUnitsRsp struct {
	MiniRent          uint64 `json:"mini_rent"`
	AssociatedAccount uint64 `json:"associated_account"`
	TransferChecked   uint64 `json:"transfer_checked"`
	Memo              uint64 `json:"memo"`
}

func getPriorityFee() (*entity.PriorityFee, error) {
	rsp, err := getJsonRequest[entity.PriorityFee](BaseURL+"/fee/priority", nil)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func getInstUnits() (*InstUnitsRsp, error) {
	rsp, err := getJsonRequest[InstUnitsRsp](BaseURL+"/fee/inst-units", nil)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func commitTakeTokenTx(req types.CommitTakeTokenTxInfoReq) (string, error) {
	//headers := map[string]string{
	//	"Authorization": fmt.Sprintf("Bearer %s", jwtToken.Access),
	//}
	headers := map[string]string{}
	rsp, err := postJsonRequest[string](RewardURL+"/take/commit-tx", req, headers)
	if err != nil {
		return "", err
	}
	return rsp.Data, nil
}

func getTakeTokenTxInfo(req types.GetTakeTokenTxInfoReq) (*types.TakeTokenTxInfo, error) {
	//headers := map[string]string{
	//	"Authorization": fmt.Sprintf("Bearer %s", jwtToken.Access),
	//}
	headers := map[string]string{}
	rsp, err := postJsonRequest[types.TakeTokenTxInfo](RewardURL+"/take/tx-info", req, headers)
	if err != nil {
		return nil, err
	}
	return &rsp.Data, nil
}

func TestGetTakeTokenTxInfo(t *testing.T) {
	//privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	//if err != nil {
	//	t.Fatalf("parse private key failed: %v", err)
	//}

	//token, err := loginForToken(Brand, Symbol, privKey.PublicKey(), privKey)
	//if err != nil {
	//	t.Fatalf("login failed: %v", err)
	//}

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

func TestLogin(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("parse private key failed: %v", err)
	}
	nativeAccount := privKey.PublicKey()
	t.Logf("账户地址: %s", nativeAccount.String())

	token, err := loginForToken(Brand, Symbol, nativeAccount, privKey)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	fmt.Println("=== JWT Token ===")
	fmt.Println("Access :", token.Access)
	fmt.Println("Refresh:", token.Refresh)
}

func createHexEncodedTx(ctx context.Context, fromPrivateKey solana.PrivateKey, inviteCode string) (string, error) {
	fromNativeAccount := fromPrivateKey.PublicKey()

	req := types.GetTakeTokenTxInfoReq{
		ReceiptAccount: AliceNativePubKey,
	}
	fmt.Printf("from native account %v: \n", fromNativeAccount)
	txInfo, err := getTakeTokenTxInfo(req)
	if err != nil {
		panic(err)
	}
	fmt.Printf("give token info %v \n", txInfo)
	fmt.Printf("total reward amount %v \n", txInfo.TotalRewardAmount)
	fmt.Printf("quote sol price %v \n", txInfo.QuoteSOLPrice)
	fmt.Printf("quoted sol amount %v \n", txInfo.QuotedSOLAmount)

	var fromTokenAccountExist = true
	rewardNativeAccount := solana.MPK(txInfo.RewardNativeAccount)
	rewardTokenAccount := solana.MPK(txInfo.RewardTokenAccount)
	tokenMintAccount := solana.MPK(txInfo.TokenMintAccount)
	dexNativeAccount := solana.MPK(txInfo.DexNativeAccount)
	decimals := uint8(txInfo.Decimals)
	fromTokenAccount, err := utils.GetSPLTokenAccountByNative(rpcClient, fromNativeAccount, solana.MPK(txInfo.TokenMintAccount))

	priorityFee, err := getPriorityFee()
	if err != nil {
		panic(err)
	}
	computePrice := priorityFee.PerComputeUnit.Medium
	computeBudgetPriceInst := computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build()

	fmt.Printf("compute price is %d\n", computePrice)

	// 奖励领取人token指令
	rewardInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(txInfo.RewardInfo.Amount).
		SetDecimals(decimals).
		SetSourceAccount(rewardTokenAccount).
		SetMintAccount(tokenMintAccount).
		SetDestinationAccount(*fromTokenAccount).
		SetOwnerAccount(rewardNativeAccount).
		Build()

	// 奖励邀请人指令
	var rewardInviterInstructions []solana.Instruction
	if txInfo.InviteDetermine {
		for _, rewardInfo := range txInfo.RewardInviterInfo {
			inst := token.NewTransferCheckedInstructionBuilder().
				SetAmount(rewardInfo.Amount).
				SetDecimals(decimals).
				SetSourceAccount(rewardTokenAccount).
				SetMintAccount(tokenMintAccount).
				SetDestinationAccount(solana.MPK(rewardInfo.TokenAccount)).
				SetOwnerAccount(rewardNativeAccount).
				Build()
			rewardInviterInstructions = append(rewardInviterInstructions, inst)
		}
	}

	// 发送给dex的sol
	toDexInst := system.NewTransferInstructionBuilder().
		SetLamports(uint64(txInfo.QuotedSOLAmount)).
		SetFundingAccount(fromNativeAccount).
		SetRecipientAccount(dexNativeAccount).
		Build()

	// 用 instUnits API 计算 computeUnitLimit：
	//   每条 TransferChecked 按 instUnits.TransferChecked 计，
	//   加上 System Transfer 的固定预留（500 CU），再留 20% buffer。
	// 注意：不使用模拟交易方式，原因是模拟时需要放入假签名，
	//       Solana 节点对假签名的模拟结果不可信，会严重低估实际消耗。
	instUnits, err := getInstUnits()
	if err != nil {
		panic(err)
	}
	transferCheckedCount := 1 + len(rewardInviterInstructions)
	const systemTransferCU uint64 = 500 // System Transfer 固定预留
	computeUnitLimit := uint32(float64(uint64(transferCheckedCount)*instUnits.TransferChecked+systemTransferCU) * 1.2)
	computeBudgetLimitInst := computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(computeUnitLimit).Build()
	fmt.Printf("computeUnitLimit: %d (transferCheckedCount=%d, perTC=%d)\n", computeUnitLimit, transferCheckedCount, instUnits.TransferChecked)

	var instructions1 []solana.Instruction
	instructions1 = append(instructions1, computeBudgetPriceInst)
	instructions1 = append(instructions1, computeBudgetLimitInst)
	if !fromTokenAccountExist {
		associated := associatedtokenaccount.NewCreateInstruction(fromNativeAccount, fromNativeAccount, tokenMintAccount).Build()
		instructions1 = append(instructions1, associated)
	}
	instructions1 = append(instructions1, rewardInst)
	instructions1 = append(instructions1, toDexInst)
	if len(rewardInviterInstructions) > 0 {
		instructions1 = append(instructions1, rewardInviterInstructions...)
	}

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	// 构造最终要发送的交易
	tx, err := solana.NewTransaction(
		instructions1,
		recent.Value.Blockhash,
		solana.TransactionPayer(fromNativeAccount),
	)
	if err != nil {
		panic(err)
	}

	messageContent1, err := tx.Message.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fromSign, err := fromPrivateKey.Sign(messageContent1)
	if err != nil {
		panic(err)
	}
	tx.Signatures = append(tx.Signatures, fromSign)

	//actualFee := utils.CalcGasFee(2, 5000, computePrice, uint64(unitConsumed), true)
	//fmt.Printf("actual fee %v\n", actualFee)

	txBytes, _ := tx.MarshalBinary()
	hexEncodedTx := hex.EncodeToString(txBytes)
	return hexEncodedTx, nil
}

func TestTakeToken(t *testing.T) {
	fromPrivateKey, _ := solana.PrivateKeyFromBase58(AlicePrivate)
	hexEncodedTx0, err := createHexEncodedTx(context.Background(), fromPrivateKey, "")
	if err != nil {
		panic(err)
	}
	// 调用 officialGiveToken
	txId, err := commitTakeTokenTx(
		types.CommitTakeTokenTxInfoReq{
			EncodedTx: hexEncodedTx0,
		},
	)
	fmt.Printf("请求成功: https://solscan.io/tx/%s?cluster=devnet\n", txId)
}
