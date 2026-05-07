package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"oshit-go/common/pkg/entity"
)

const (
	Brand           = "OShit"
	Symbol          = "OShit"
	SOLLoginSignMsg = "I am login %s for token %s with my address %s with nonce %d"

	// mainnet:
	//RpcUrl           = "https://radial-purple-sailboat.solana-mainnet.quiknode.pro/0afcb192bb26b0dbcba3d49df6ad2ee2829c529b/"
	//WssUrl           = "wss://radial-purple-sailboat.solana-mainnet.quiknode.pro/0afcb192bb26b0dbcba3d49df6ad2ee2829c529b/"
	//TokenMintAddress = "ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"

	// testnet:
	RpcUrl           = "https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/"
	WssUrl           = "wss://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/"
	TokenMintAddress = "wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM"

	AlicePrivate      = "46dPKNS2nHDuaGrdJ9tyr7mJ1H6Gu8EKcvPNDmC8zZ2xf1SpuhzR9kfK7uA3mf1RsfwoYvJDYdjA4WPkjtSv1r4E"
	AliceNativePubKey = "5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN"

	BobPrivate      = "43HXEVBSJzsfA9PQpYbE9GFuAg1YwcKuLiecqvqUVMg8p8EcDSXgQShBMBsqURWogs7k173BehCWF2y17VEXnZkn"
	BobNativePubKey = "27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx"

	DavidPrivate      = "4LRdCeZ4EYHsGtRr99zbh6WLZCa7jACVuhunzrbgyKvP4NxKuHPnaJ52t1AKzREbaD5n2NMsKYkdPMcnYRMacjgU"
	DavidNativePubKey = "JAZtFeZfLeeVtWS4vrruCpTa5LdASRDJuMe7yLKbkJk"

	RobertPrivate      = "MzJGrbzW1yqSzAAbGLHbSkHKmGFS6kuhACV8wFxePqSwe6rTsv7N3eRVozdJcJSBQAPT6rjtnmoCFxv6YuA5hGq"
	RobertNativePubKey = "6HLScqNL4EQWLk8DTcB4hXUrHjDkVbeP2a3Sc5VtHozM"

	BaseURL   = "http://localhost:1100/base"
	RewardURL = "http://localhost:1200/reward"
	//BaseURL   = "https://beta.testnet.oshit.io/meme/base/api/v1"
	//RewardURL = "https://beta.testnet.oshit.io/meme/reward/api/v1"
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

type TransactionParams struct {
	TokenQuoteSOLPrice float64                  `json:"tokenQuoteSOLPrice"`
	PriorityFee        entity.PriorityFee       `json:"priorityFee"`
	ConsumedUnits      entity.ComputeUnitDetail `json:"consumedUnits"`
	DiscountRate       float64                  `json:"discountRate"`
}

// InstUnitsRsp 与 base 模块 /fee/inst-units 响应字段保持一致
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

func postFormRequest[T any](url string, formData url.Values, headers map[string]string) (*ApiResponse[T], error) {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

// ---- 通用测试用例 ----

func TestLogin(t *testing.T) {
	privKey, err := solana.PrivateKeyFromBase58(BobPrivate)
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
	fmt.Println("Token:", token)
}
