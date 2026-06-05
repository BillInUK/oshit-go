package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"oshit-go/common/utils"
)

// 分布式锁 key
const priceFetchLock = "base:sol:price:fetch:lock"

// Redis keys（保持与消费方一致）
const (
	redisKeyTokenQuoteSOL  = "base:sol:price:raydium-quote-sol"
	redisKeyTokenQuoteUSDT = "base:sol:price:raydium-quote-usdt"
	redisKeyUSDTQuoteSOL   = "base:sol:price:raydium-usdt-quote-sol"
)

// Helius getAsset 请求/响应结构
type heliusGetAssetReq struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type heliusGetAssetRsp struct {
	Result struct {
		TokenInfo struct {
			Symbol    string `json:"symbol"`
			Supply    uint64 `json:"supply"`
			Decimals  int    `json:"decimals"`
			PriceInfo struct {
				PricePerToken float64 `json:"price_per_token"`
				Currency      string  `json:"currency"`
			} `json:"price_info"`
		} `json:"token_info"`
	} `json:"result"`
}

// 主网 token mint 地址
const (
	shitTokenMint  = "ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"
	wrappedSOLMint = "So11111111111111111111111111111111111111112"
)

// PriceTask 价格获取任务
type PriceTask struct {
	db           *gorm.DB
	redis        redis.UniversalClient
	redSync      redsync.Redsync
	heliusAPIKey string
	mainnetURL   string
	tokenDec     float64 // 10^decimals
}

// NewPriceTask 创建价格任务
func NewPriceTask(taskCtx *TaskContext) *PriceTask {
	mainnetURL := taskCtx.MainnetRpcURL
	heliusKey := taskCtx.HeliusAPIKey
	if mainnetURL == "" && heliusKey != "" {
		mainnetURL = utils.BuildRPCURL("helius", "https://mainnet.helius-rpc.com", heliusKey)
	}
	return &PriceTask{
		db:           taskCtx.DB,
		redis:        taskCtx.Redis,
		redSync:      taskCtx.RedSync,
		heliusAPIKey: heliusKey,
		mainnetURL:   mainnetURL,
		tokenDec:     taskCtx.TokenDecimal,
	}
}

func (t *PriceTask) Start() {
	go runPeriodic(&t.redSync, 15*time.Second, priceFetchLock, 5*time.Minute, t.fetchPrice)
}

func (t *PriceTask) fetchPrice() {
	if t.mainnetURL == "" {
		log.Warnf("价格获取跳过: mainnet RPC URL 未配置")
		return
	}

	// 1. 获取 SHIT token 的 USDC 价格
	shitPrice, err := t.getAssetPrice(shitTokenMint)
	if err != nil {
		log.Errorf("获取 SHIT token 价格失败: %v", err)
		return
	}
	if shitPrice <= 0 {
		log.Errorf("获取 SHIT token 价格无效: %v", shitPrice)
		return
	}

	// 2. 获取 SOL 的 USDC 价格
	solPrice, err := t.getAssetPrice(wrappedSOLMint)
	if err != nil {
		log.Errorf("获取 SOL 价格失败: %v", err)
		return
	}
	if solPrice <= 0 {
		log.Errorf("获取 SOL 价格无效: %v", solPrice)
		return
	}

	// 3. 计算各种价格并写入 Redis
	ctx := context.Background()

	// token/SOL 价格: 1 个 token 值多少 lamports
	// shitPrice(USDC) / solPrice(USDC) = token/SOL (display)
	// × LAMPORTS_PER_SOL = lamports
	tokenPerSOL := shitPrice / solPrice
	tokenPerSOLLamports := int64(tokenPerSOL * 1e9)
	if err := t.redis.Set(ctx, redisKeyTokenQuoteSOL, strconv.FormatInt(tokenPerSOLLamports, 10), 0).Err(); err != nil {
		log.Errorf("Redis 写入 token/SOL 价格失败: %v", err)
	}

	// token/USDT 价格: 1 个 token 值多少 USDT 最小单位（1e6）
	tokenPerUSDT := int64(shitPrice * 1e6)
	if err := t.redis.Set(ctx, redisKeyTokenQuoteUSDT, strconv.FormatInt(tokenPerUSDT, 10), 0).Err(); err != nil {
		log.Errorf("Redis 写入 token/USDT 价格失败: %v", err)
	}

	// USDT/SOL 价格: 1 USDT 值多少 SOL (display float)
	usdtPerSOL := 1.0 / solPrice
	usdtPerSOLStr := strconv.FormatFloat(usdtPerSOL, 'f', 9, 64)
	if err := t.redis.Set(ctx, redisKeyUSDTQuoteSOL, usdtPerSOLStr, 0).Err(); err != nil {
		log.Errorf("Redis 写入 USDT/SOL 价格失败: %v", err)
	}

	log.Infof("价格更新成功: SHIT=$%.6f, SOL=$%.2f, token/SOL=%d lamports, token/USDT=%d, USDT/SOL=%s",
		shitPrice, solPrice, tokenPerSOLLamports, tokenPerUSDT, usdtPerSOLStr)
}

// getAssetPrice 通过 Helius getAsset API 获取 token 的 USDC 价格
func (t *PriceTask) getAssetPrice(mintAddress string) (float64, error) {
	reqBody := heliusGetAssetReq{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "getAsset",
		Params: map[string]interface{}{
			"id": mintAddress,
			"displayOptions": map[string]bool{
				"showFungible": true,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(t.mainnetURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	var result heliusGetAssetRsp
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("parse response: %w", err)
	}

	price := result.Result.TokenInfo.PriceInfo.PricePerToken
	if math.IsInf(price, 0) || math.IsNaN(price) {
		return 0, fmt.Errorf("invalid price for %s: %v", mintAddress, price)
	}

	return price, nil
}
