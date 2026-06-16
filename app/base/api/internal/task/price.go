package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 分布式锁 key
const priceFetchLock = "base:sol:price:fetch:lock"

// Redis keys（保持与消费方一致）
const (
	redisKeyTokenQuoteSOL  = "base:sol:price:raydium-quote-sol"
	redisKeyTokenQuoteUSDT = "base:sol:price:raydium-quote-usdt"
	redisKeyUSDTQuoteSOL   = "base:sol:price:raydium-usdt-quote-sol"
)

// Raydium mint/price API 响应结构
type raydiumPriceRsp struct {
	ID      string            `json:"id"`
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"` // mint -> price string
}

// 主网 token mint 地址
const (
	shitTokenMint  = "ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"
	wrappedSOLMint = "So11111111111111111111111111111111111111112"
)

// Raydium API 地址
const raydiumPriceAPI = "https://api-v3.raydium.io/mint/price"

// PriceTask 价格获取任务
type PriceTask struct {
	db       *gorm.DB
	redis    redis.UniversalClient
	redSync  redsync.Redsync
	tokenDec float64 // 10^decimals
}

// NewPriceTask 创建价格任务
func NewPriceTask(taskCtx *TaskContext) *PriceTask {
	return &PriceTask{
		db:       taskCtx.DB,
		redis:    taskCtx.Redis,
		redSync:  taskCtx.RedSync,
		tokenDec: taskCtx.TokenDecimal,
	}
}

func (t *PriceTask) Start() {
	go runPeriodic(&t.redSync, 15*time.Second, priceFetchLock, 5*time.Minute, t.fetchPrice)
}

func (t *PriceTask) fetchPrice() {
	// 一次请求同时获取 SHIT 和 SOL 的 USDC 价格
	prices, err := t.getRaydiumPrices(shitTokenMint, wrappedSOLMint)
	if err != nil {
		log.Errorf("Raydium 获取价格失败: %v", err)
		return
	}

	shitPrice, ok := prices[shitTokenMint]
	if !ok || shitPrice <= 0 {
		log.Errorf("获取 SHIT token 价格无效: %v", shitPrice)
		return
	}

	solPrice, ok := prices[wrappedSOLMint]
	if !ok || solPrice <= 0 {
		log.Errorf("获取 SOL 价格无效: %v", solPrice)
		return
	}

	// 计算各种价格并写入 Redis
	ctx := context.Background()

	// token/SOL 价格: 1 个 token 值多少 lamports
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

	log.Infof("价格更新成功: SHIT=$%.8f, SOL=$%.2f, token/SOL=%d lamports, token/USDT=%d, USDT/SOL=%s",
		shitPrice, solPrice, tokenPerSOLLamports, tokenPerUSDT, usdtPerSOLStr)
}

// getRaydiumPrices 通过 Raydium API 批量获取 token 的 USDC 价格
func (t *PriceTask) getRaydiumPrices(mints ...string) (map[string]float64, error) {
	url := raydiumPriceAPI + "?mints="
	for i, mint := range mints {
		if i > 0 {
			url += ","
		}
		url += mint
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result raydiumPriceRsp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("raydium API returned success=false")
	}

	prices := make(map[string]float64, len(result.Data))
	for mint, priceStr := range result.Data {
		p, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parse price for %s: %w", mint, err)
		}
		prices[mint] = p
	}
	return prices, nil
}
