package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
)

const (
	SOLLoginSignMsg = "I am login %s for token %s with my address %s with nonce %d"
	// 使用旧工程中的Alice私钥（测试网）
	AlicePrivate      = "46dPKNS2nHDuaGrdJ9tyr7mJ1H6Gu8EKcvPNDmC8zZ2xf1SpuhzR9kfK7uA3mf1RsfwoYvJDYdjA4WPkjtSv1r4E"
	AliceNativePubKey = "5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN"
	// Token配置
	Brand   = "OShit"
	Symbol  = "OShit"
	BaseURL = "http://localhost:1100/api"
)

type LoginRequest struct {
	Brand   string `json:"brand"`
	Symbol  string `json:"symbol"`
	Account string `json:"account"`
	Sign    string `json:"sign"`
	Nonce   uint64 `json:"nonce"`
}

type TokenResponse struct {
	Token struct {
		Access  string `json:"Access"`
		Refresh string `json:"Refresh"`
	} `json:"token"`
}

// HTTPClient 封装HTTP请求
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Post 发送POST请求
func (c *HTTPClient) Post(path string, body interface{}, headers map[string]string) ([]byte, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get 发送GET请求
func (c *HTTPClient) Get(path string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending GET request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// loginForToken 参考旧工程的实现
func loginForToken(client *HTTPClient, brand, symbol string, nativeAccount solana.PublicKey, privKey solana.PrivateKey) (*TokenResponse, error) {
	nonce := uint64(time.Now().UnixMilli())
	msg := fmt.Sprintf(SOLLoginSignMsg, brand, symbol, nativeAccount.String(), nonce)
	sign, _ := privKey.Sign([]byte(msg))

	loginReq := LoginRequest{
		Brand:   brand,
		Symbol:  symbol,
		Account: nativeAccount.String(),
		Nonce:   nonce,
		Sign:    sign.String(),
	}

	respBody, err := client.Post("/auth/login", loginReq, nil)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %v", err)
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return &tokenResp, nil
}

// TestLogin 测试登录功能
func TestLogin(t *testing.T) {
	// 1. 创建HTTP客户端
	client := NewHTTPClient(BaseURL)

	// 2. 检查API服务是否运行
	t.Log("Checking API service...")

	// 尝试直接访问根路径，忽略404错误（可能没有定义根路由）
	resp, err := http.Get("http://localhost:1100")
	if err != nil {
		t.Skipf("API service may not be running: %v. Please start the API service with: go run app/base/api/base_api.go", err)
	}
	defer resp.Body.Close()

	// 只要服务响应（即使是404），就认为服务在运行
	t.Logf("✅ API service is responding (status: %d)", resp.StatusCode)

	// 3. 解析私钥
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey := privateKey.PublicKey()
	if publicKey.String() != AliceNativePubKey {
		t.Fatalf("Public key mismatch: got %s, expected %s", publicKey.String(), AliceNativePubKey)
	}
	t.Logf("✅ Using Alice account: %s", publicKey.String())

	// 4. 测试登录
	t.Log("Testing login...")
	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	t.Log("✅ Login successful!")
	if tokenResp.Token.Access == "" {
		t.Error("Access token should not be empty")
	}
	if tokenResp.Token.Refresh == "" {
		t.Error("Refresh token should not be empty")
	}

	t.Logf("Access Token: %s", tokenResp.Token.Access)
	t.Logf("Refresh Token: %s", tokenResp.Token.Refresh)

	// 5. 验证token可以用于其他API调用
	t.Log("Testing token usage...")
	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	// 测试检查邀请记录
	checkReq := map[string]string{
		"native_account": publicKey.String(),
	}
	_, err = client.Post("/invite/check-record", checkReq, headers)
	if err != nil {
		t.Logf("Note: Check invite record failed (may be expected): %v", err)
	} else {
		t.Log("✅ Token works for authenticated API calls")
	}
}

// TestLoginWithWrongSignature 测试使用错误签名的登录
func TestLoginWithWrongSignature(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 解析正确的私钥获取公钥
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	// 使用错误的私钥生成签名
	wrongPrivateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate random private key: %v", err)
	}

	nonce := uint64(time.Now().UnixMilli())
	msg := fmt.Sprintf(SOLLoginSignMsg, Brand, Symbol, publicKey.String(), nonce)
	wrongSign, _ := wrongPrivateKey.Sign([]byte(msg))

	loginReq := LoginRequest{
		Brand:   Brand,
		Symbol:  Symbol,
		Account: publicKey.String(),
		Nonce:   nonce,
		Sign:    wrongSign.String(),
	}

	_, err = client.Post("/auth/login", loginReq, nil)
	if err == nil {
		t.Error("Login with wrong signature should fail")
	} else {
		t.Logf("✅ Login with wrong signature correctly failed: %v", err)
	}
}

// TestLoginWithInvalidAccount 测试使用无效账户登录
func TestLoginWithInvalidAccount(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 生成随机账户
	privateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate random private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	nonce := uint64(time.Now().UnixMilli())
	msg := fmt.Sprintf(SOLLoginSignMsg, Brand, Symbol, publicKey.String(), nonce)
	sign, _ := privateKey.Sign([]byte(msg))

	loginReq := LoginRequest{
		Brand:   Brand,
		Symbol:  Symbol,
		Account: publicKey.String(),
		Nonce:   nonce,
		Sign:    sign.String(),
	}

	// 新账户应该被注册并登录成功
	respBody, err := client.Post("/auth/login", loginReq, nil)
	if err != nil {
		t.Logf("New account registration/login may have failed: %v", err)
		return
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		t.Errorf("Error decoding response: %v", err)
		return
	}

	if tokenResp.Token.Access == "" {
		t.Error("New account should receive access token")
	} else {
		t.Log("✅ New account registered and logged in successfully")
	}
}

// ==================== Config API Tests ====================

type GetTokenInfoReq struct {
	Brand  string `json:"brand"`
	Symbol string `json:"symbol"`
}

type GetTokenInfoRsp struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Decimal   int32  `json:"decimal"`
	Mint      string `json:"mint"`
	CreatedAt string `json:"created_at"`
}

type GetFeeToleranceRsp struct {
	MinFee uint64 `json:"min_fee"`
	MaxFee uint64 `json:"max_fee"`
}

// TestGetTokenInfo 测试获取token信息
func TestGetTokenInfo(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	req := GetTokenInfoReq{
		Brand:  Brand,
		Symbol: Symbol,
	}

	respBody, err := client.Post("/config/token", req, nil)
	if err != nil {
		t.Fatalf("Get token info failed: %v", err)
	}

	var resp GetTokenInfoRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.Name == "" {
		t.Error("Token name should not be empty")
	}
	if resp.Symbol == "" {
		t.Error("Token symbol should not be empty")
	}
	if resp.Mint == "" {
		t.Error("Token mint address should not be empty")
	}

	t.Logf("✅ Get token info successful")
	t.Logf("   Name: %s", resp.Name)
	t.Logf("   Symbol: %s", resp.Symbol)
	t.Logf("   Decimal: %d", resp.Decimal)
	t.Logf("   Mint: %s", resp.Mint)
}

// TestGetFeeTolerance 测试获取手续费容错
func TestGetFeeTolerance(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/config/fee-tolerance", nil)
	if err != nil {
		t.Fatalf("Get fee tolerance failed: %v", err)
	}

	var resp GetFeeToleranceRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.MinFee == 0 {
		t.Error("Min fee should not be zero")
	}
	if resp.MaxFee == 0 {
		t.Error("Max fee should not be zero")
	}
	if resp.MinFee > resp.MaxFee {
		t.Error("Min fee should not be greater than max fee")
	}

	t.Logf("✅ Get fee tolerance successful")
	t.Logf("   Min Fee: %d", resp.MinFee)
	t.Logf("   Max Fee: %d", resp.MaxFee)
}

// ==================== Fee API Tests ====================

type FeeDetail struct {
	P50  uint64 `json:"p50"`
	P90  uint64 `json:"p90"`
	P95  uint64 `json:"p95"`
	P99  uint64 `json:"p99"`
	Mean uint64 `json:"mean"`
}

type PriorityFeeRsp struct {
	PerComputeUnit FeeDetail `json:"per_compute_unit"`
	PerTransaction FeeDetail `json:"per_transaction"`
}

type ComputeUnitConsumedRsp struct {
	MiniRent          uint64 `json:"mini_rent"`
	AssociatedAccount uint64 `json:"associated_account"`
	TransferChecked   uint64 `json:"transfer_checked"`
	Memo              uint64 `json:"memo"`
}

// TestGetPriorityFee 测试获取优先手续费
func TestGetPriorityFee(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/fee/priority", nil)
	if err != nil {
		t.Fatalf("Get priority fee failed: %v", err)
	}

	var resp PriorityFeeRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// 检查基本字段
	if resp.PerComputeUnit.P50 == 0 {
		t.Error("Per compute unit P50 should not be zero")
	}
	if resp.PerTransaction.P50 == 0 {
		t.Error("Per transaction P50 should not be zero")
	}

	t.Logf("✅ Get priority fee successful")
	t.Logf("   Per Compute Unit - P50: %d, P90: %d", resp.PerComputeUnit.P50, resp.PerComputeUnit.P90)
	t.Logf("   Per Transaction - P50: %d, P90: %d", resp.PerTransaction.P50, resp.PerTransaction.P90)
}

// TestGetPriorityFeeOnBlockchain 测试获取链上优先手续费
func TestGetPriorityFeeOnBlockchain(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/fee/priority/on-chain", nil)
	if err != nil {
		t.Fatalf("Get priority fee on blockchain failed: %v", err)
	}

	var resp PriorityFeeRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	t.Logf("✅ Get priority fee on blockchain successful")
}

// TestGetComputeUnitConsumed 测试获取计算单元消耗
func TestGetComputeUnitConsumed(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/fee/compute-units", nil)
	if err != nil {
		t.Fatalf("Get compute unit consumed failed: %v", err)
	}

	var resp ComputeUnitConsumedRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.TransferChecked == 0 {
		t.Error("Transfer checked compute units should not be zero")
	}

	t.Logf("✅ Get compute unit consumed successful")
	t.Logf("   Transfer Checked: %d", resp.TransferChecked)
	t.Logf("   Associated Account: %d", resp.AssociatedAccount)
}

// ==================== Price API Tests ====================

type PriceQuoteRsp struct {
	Price float64 `json:"price"`
}

type GetBirdEyePriceReq struct {
	Interval string `json:"interval"` // 1D, 1W, 1M
}

type BirdEyePriceRsp struct {
	Data interface{} `json:"data"`
}

// TestGetTokenQuoteSOLPrice 测试获取token兑换SOL价格
func TestGetTokenQuoteSOLPrice(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/price/token/sol", nil)
	if err != nil {
		t.Fatalf("Get token quote SOL price failed: %v", err)
	}

	var resp PriceQuoteRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.Price <= 0 {
		t.Error("Token price should be positive")
	}

	t.Logf("✅ Get token quote SOL price successful")
	t.Logf("   Price: %.9f", resp.Price)
}

// TestGetTokenQuoteUSDTPrice 测试获取token兑换USDT价格
func TestGetTokenQuoteUSDTPrice(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/price/token/usdt", nil)
	if err != nil {
		t.Fatalf("Get token quote USDT price failed: %v", err)
	}

	var resp PriceQuoteRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.Price <= 0 {
		t.Error("Token price should be positive")
	}

	t.Logf("✅ Get token quote USDT price successful")
	t.Logf("   Price: %.9f", resp.Price)
}

// TestGetUSDTQuoteSOLPrice 测试获取USDT兑换SOL价格
func TestGetUSDTQuoteSOLPrice(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	respBody, err := client.Get("/price/usdt/sol", nil)
	if err != nil {
		t.Fatalf("Get USDT quote SOL price failed: %v", err)
	}

	var resp PriceQuoteRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.Price <= 0 {
		t.Error("USDT/SOL price should be positive")
	}

	t.Logf("✅ Get USDT quote SOL price successful")
	t.Logf("   Price: %.9f", resp.Price)
}

// TestGetBirdEyePrice 测试获取BirdEye价格数据
func TestGetBirdEyePrice(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	testCases := []struct {
		name     string
		interval string
	}{
		{"1D", "1D"},
		{"1W", "1W"},
		{"1M", "1M"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := GetBirdEyePriceReq{
				Interval: tc.interval,
			}

			respBody, err := client.Post("/price/birdeye", req, nil)
			if err != nil {
				t.Logf("Get BirdEye price for interval %s failed: %v", tc.interval, err)
				return
			}

			var resp BirdEyePriceRsp
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				t.Errorf("Error decoding response: %v", err)
				return
			}

			if resp.Data == nil {
				t.Error("BirdEye price data should not be nil")
			}

			t.Logf("✅ Get BirdEye price for interval %s successful", tc.interval)
		})
	}
}

// ==================== Invite API Tests ====================

type GetAccountByInviteCodeReq struct {
	InviteCode string `json:"invite_code"`
}

type GetAccountByInviteCodeRsp struct {
	RecordID      string `json:"record_id"`
	NativeAccount string `json:"native_account"`
	TokenAccount  string `json:"token_account"`
	InviteCode    string `json:"invite_code"`
	CreatedAt     string `json:"created_at"`
}

type CheckInviteRecordReq struct {
	NativeAccount string `json:"native_account"`
}

type CheckInviteRecordRsp struct {
	Exists bool `json:"exists"`
}

type RecursiveQueryReq struct {
	Depth         int    `json:"depth"`
	NativeAccount string `json:"native_account"`
}

type InviteRelation struct {
	RecordID  string `json:"record_id"`
	Inviter   string `json:"inviter"`
	Invitee   string `json:"invitee"`
	Channel   string `json:"channel"`
	Level     int32  `json:"level"`
	TxID      string `json:"tx_id"`
	CreatedAt string `json:"created_at"`
}

type RecursiveQueryRsp struct {
	Records []InviteRelation `json:"records"`
}

type RewardDistributionRsp struct {
	Level int32 `json:"level"`
}

type RewardClaimRsp struct {
	Level int32   `json:"level"`
	Rate  float64 `json:"rate"`
}

type TokenHoldersRsp struct {
	HoldersNumber int `json:"holders_number"`
}

// TestGetAccountByInviteCode 测试根据邀请码查询账户
func TestGetAccountByInviteCode(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	// 测试不存在的邀请码
	req := GetAccountByInviteCodeReq{
		InviteCode: "INVALID123",
	}

	_, err = client.Post("/invite/account-by-code", req, headers)
	if err == nil {
		t.Error("Query with invalid invite code should fail")
	} else {
		t.Logf("✅ Query with invalid invite code correctly failed: %v", err)
	}
}

// TestCheckInviteRecord 测试检查邀请记录
func TestCheckInviteRecord(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	req := CheckInviteRecordReq{
		NativeAccount: publicKey.String(),
	}

	respBody, err := client.Post("/invite/check-record", req, headers)
	if err != nil {
		t.Fatalf("Check invite record failed: %v", err)
	}

	var resp CheckInviteRecordRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	t.Logf("✅ Check invite record successful")
	t.Logf("   Exists: %v", resp.Exists)
}

// TestGetUpInviterRecords 测试递归查询上级邀请人
func TestGetUpInviterRecords(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	req := RecursiveQueryReq{
		Depth:         3,
		NativeAccount: publicKey.String(),
	}

	respBody, err := client.Post("/invite/up-records", req, headers)
	if err != nil {
		t.Logf("Get up inviter records failed (may be expected if no invite records): %v", err)
		return
	}

	var resp RecursiveQueryRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	t.Logf("✅ Get up inviter records successful")
	t.Logf("   Records count: %d", len(resp.Records))
}

// TestGetDownInviteeRecords 测试递归查询下级被邀请人
func TestGetDownInviteeRecords(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	req := RecursiveQueryReq{
		Depth:         3,
		NativeAccount: publicKey.String(),
	}

	respBody, err := client.Post("/invite/down-records", req, headers)
	if err != nil {
		t.Logf("Get down invitee records failed (may be expected if no invite records): %v", err)
		return
	}

	var resp RecursiveQueryRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	t.Logf("✅ Get down invitee records successful")
	t.Logf("   Records count: %d", len(resp.Records))
}

// TestGetRewardDistribution 测试获取奖励分布
func TestGetRewardDistribution(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	respBody, err := client.Get("/invite/reward-distribution", headers)
	if err != nil {
		t.Fatalf("Get reward distribution failed: %v", err)
	}

	var resp RewardDistributionRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.Level <= 0 {
		t.Error("Reward distribution level should be positive")
	}

	t.Logf("✅ Get reward distribution successful")
	t.Logf("   Level: %d", resp.Level)
}

// TestGetRewardClaims 测试获取奖励声明
func TestGetRewardClaims(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	// 测试不同level
	testLevels := []string{"1", "2", "3"}
	for _, level := range testLevels {
		respBody, err := client.Get("/invite/reward-claims?level="+level, headers)
		if err != nil {
			t.Logf("Get reward claims for level %s failed: %v", level, err)
			continue
		}

		var resp RewardClaimRsp
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			t.Errorf("Error decoding response for level %s: %v", level, err)
			continue
		}

		if resp.Level <= 0 {
			t.Errorf("Reward claim level should be positive for level %s", level)
		}

		t.Logf("✅ Get reward claims for level %s successful", level)
		t.Logf("   Level: %d, Rate: %.2f", resp.Level, resp.Rate)
	}
}

// TestGetTokenHolders 测试获取token持有者数量
func TestGetTokenHolders(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先登录获取token
	privateKey, err := solana.PrivateKeyFromBase58(AlicePrivate)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	publicKey := privateKey.PublicKey()

	tokenResp, err := loginForToken(client, Brand, Symbol, publicKey, privateKey)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + tokenResp.Token.Access,
	}

	respBody, err := client.Get("/invite/token-holders", headers)
	if err != nil {
		t.Fatalf("Get token holders failed: %v", err)
	}

	var resp TokenHoldersRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.HoldersNumber < 0 {
		t.Error("Token holders number should not be negative")
	}

	t.Logf("✅ Get token holders successful")
	t.Logf("   Holders Number: %d", resp.HoldersNumber)
}

// ==================== Service API Tests ====================

type ServiceRegisterReq struct {
	Service       string `json:"service"`
	NativeAccount string `json:"native_account"`
	PdaAccount    string `json:"pda_account"`
	UntilTxID     string `json:"until_tx_id"`
	Slot          int64  `json:"slot"`
	Webhook       string `json:"webhook,omitempty"`
	MqGroup       string `json:"mq_group,omitempty"`
	MqTopic       string `json:"mq_topic,omitempty"`
	HookType      int32  `json:"hook_type"` // 0=Kafka, 1=Webhook
}

type ServiceRegisterRsp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ServiceUpdateReq struct {
	Service       string `json:"service"`
	NativeAccount string `json:"native_account,omitempty"`
	PdaAccount    string `json:"pda_account,omitempty"`
	UntilTxID     string `json:"until_tx_id,omitempty"`
	Slot          int64  `json:"slot,omitempty"`
	Webhook       string `json:"webhook,omitempty"`
	MqGroup       string `json:"mq_group,omitempty"`
	MqTopic       string `json:"mq_topic,omitempty"`
	HookType      int32  `json:"hook_type,omitempty"`
}

type ServiceUpdateRsp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ServiceQueryReq struct {
	Service string `json:"service"`
}

type ServiceQueryRsp struct {
	Service       string `json:"service"`
	NativeAccount string `json:"native_account"`
	PdaAccount    string `json:"pda_account"`
	UntilTxID     string `json:"until_tx_id"`
	Slot          int64  `json:"slot"`
	Webhook       string `json:"webhook,omitempty"`
	MqGroup       string `json:"mq_group,omitempty"`
	MqTopic       string `json:"mq_topic,omitempty"`
	HookType      int32  `json:"hook_type"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// TestServiceRegister 测试服务注册
func TestServiceRegister(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	serviceName := fmt.Sprintf("test_service_%d", time.Now().Unix())

	req := ServiceRegisterReq{
		Service:       serviceName,
		NativeAccount: "native1234567890abcdef",
		PdaAccount:    "pda1234567890abcdef",
		UntilTxID:     "",
		Slot:          0,
		Webhook:       "https://example.com/webhook",
		MqTopic:       "test_topic",
		HookType:      0, // Kafka
	}

	respBody, err := client.Post("/service/register", req, nil)
	if err != nil {
		t.Fatalf("Service register failed: %v", err)
	}

	var resp ServiceRegisterRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Service register should be successful: %s", resp.Message)
	}

	t.Logf("✅ Service register successful")
	t.Logf("   Message: %s", resp.Message)

	// 清理：测试查询
	queryReq := ServiceQueryReq{
		Service: serviceName,
	}

	queryRespBody, err := client.Post("/service/query", queryReq, nil)
	if err != nil {
		t.Logf("Service query after register failed: %v", err)
		return
	}

	var queryResp ServiceQueryRsp
	err = json.Unmarshal(queryRespBody, &queryResp)
	if err != nil {
		t.Errorf("Error decoding query response: %v", err)
		return
	}

	if queryResp.Service != serviceName {
		t.Errorf("Service name mismatch: got %s, expected %s", queryResp.Service, serviceName)
	}

	t.Logf("✅ Service query after register successful")
	t.Logf("   Service: %s", queryResp.Service)
	t.Logf("   Native Account: %s", queryResp.NativeAccount)
}

// TestServiceUpdate 测试服务更新
func TestServiceUpdate(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 先注册一个服务
	serviceName := fmt.Sprintf("test_service_update_%d", time.Now().Unix())

	registerReq := ServiceRegisterReq{
		Service:       serviceName,
		NativeAccount: "native1234567890abcdef",
		PdaAccount:    "pda1234567890abcdef",
		UntilTxID:     "",
		Slot:          0,
		Webhook:       "https://example.com/webhook",
		MqTopic:       "test_topic",
		HookType:      0,
	}

	_, err := client.Post("/service/register", registerReq, nil)
	if err != nil {
		t.Fatalf("Service register for update test failed: %v", err)
	}

	// 更新服务
	updateReq := ServiceUpdateReq{
		Service:  serviceName,
		Webhook:  "https://example.com/webhook_updated",
		MqTopic:  "test_topic_updated",
		HookType: 1, // Webhook
	}

	respBody, err := client.Post("/service/update", updateReq, nil)
	if err != nil {
		t.Fatalf("Service update failed: %v", err)
	}

	var resp ServiceUpdateRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Service update should be successful: %s", resp.Message)
	}

	t.Logf("✅ Service update successful")
	t.Logf("   Message: %s", resp.Message)
}

// TestServiceQuery 测试服务查询
func TestServiceQuery(t *testing.T) {
	client := NewHTTPClient(BaseURL)

	// 测试查询不存在的服务
	req := ServiceQueryReq{
		Service: "non_existent_service",
	}

	_, err := client.Post("/service/query", req, nil)
	if err == nil {
		t.Error("Query non-existent service should fail")
	} else {
		t.Logf("✅ Query non-existent service correctly failed: %v", err)
	}

	// 测试查询存在的服务（需要先注册）
	serviceName := fmt.Sprintf("test_service_query_%d", time.Now().Unix())

	registerReq := ServiceRegisterReq{
		Service:       serviceName,
		NativeAccount: "native1234567890abcdef",
		PdaAccount:    "pda1234567890abcdef",
		UntilTxID:     "",
		Slot:          0,
		Webhook:       "https://example.com/webhook",
		MqTopic:       "test_topic",
		HookType:      0,
	}

	_, err = client.Post("/service/register", registerReq, nil)
	if err != nil {
		t.Fatalf("Service register for query test failed: %v", err)
	}

	// 查询刚注册的服务
	queryReq := ServiceQueryReq{
		Service: serviceName,
	}

	respBody, err := client.Post("/service/query", queryReq, nil)
	if err != nil {
		t.Fatalf("Service query failed: %v", err)
	}

	var resp ServiceQueryRsp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if resp.Service != serviceName {
		t.Errorf("Service name mismatch: got %s, expected %s", resp.Service, serviceName)
	}
	if resp.NativeAccount != "native1234567890abcdef" {
		t.Errorf("Native account mismatch: got %s, expected %s", resp.NativeAccount, "native1234567890abcdef")
	}

	t.Logf("✅ Service query successful")
	t.Logf("   Service: %s", resp.Service)
	t.Logf("   Native Account: %s", resp.NativeAccount)
	t.Logf("   PDA Account: %s", resp.PdaAccount)
	t.Logf("   Hook Type: %d", resp.HookType)
}
