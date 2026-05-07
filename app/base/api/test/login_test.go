package test

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
	//baseURL = "http://localhost:1100"
	baseURL = "https://beta.testnet.oshit.io/meme/base/api/v1"
	brand   = "OShit" // 按实际配置修改
	symbol  = "OShit" // 按实际配置修改
)

type loginReq struct {
	Brand      string `json:"brand"`
	Symbol     string `json:"symbol"`
	Account    string `json:"account"`
	Sign       string `json:"sign"`
	Nonce      uint64 `json:"nonce"`
	InviteCode string `json:"invite_code"`
}

type loginRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token struct {
			Access  string `json:"Access"`
			Refresh string `json:"Refresh"`
		} `json:"token"`
	} `json:"data"`
}

func TestLogin(t *testing.T) {
	// 生成临时 Solana keypair（每次运行都是新账户，会自动注册）
	keypair := solana.NewWallet()
	account := keypair.PublicKey().String()
	nonce := uint64(time.Now().UnixMilli())

	// 拼接签名消息（与服务端保持一致）
	msg := fmt.Sprintf("I am login %s for token %s with my address %s with nonce %d", brand, symbol, account, nonce)
	t.Logf("sign msg: %s", msg)

	// 用私钥签名
	sig, err := keypair.PrivateKey.Sign([]byte(msg))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	req := loginReq{
		Brand:   brand,
		Symbol:  symbol,
		Account: account,
		Sign:    sig.String(),
		Nonce:   nonce,
	}

	body, _ := json.Marshal(req)
	t.Logf("request: %s", body)

	resp, err := http.Post(baseURL+"/base/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http post failed: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	t.Logf("status: %d", resp.StatusCode)
	t.Logf("response: %s", raw)

	var rsp loginRsp
	if err := json.Unmarshal(raw, &rsp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	fmt.Println("=== JWT Token ===")
	fmt.Println("Access :", rsp.Data.Token.Access)
	fmt.Println("Refresh:", rsp.Data.Token.Refresh)
}
