package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"io"
	"math"
	"net/http"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"strconv"
	"strings"
	"time"
)

// 分布式锁 key
const priceFetchRaydiumLock = "base:sol:price:fetch-raydium:lock"

// PriceTask 手续费统计任务
type PriceTask struct {
	db          *gorm.DB
	redis       redis.UniversalClient
	redSync     redsync.Redsync
	rpcClient   *rpc.Client
	rpcURL      string
	tokenDec    float64
	tokenConfig *model.TokenConfig
}

// NewPriceTask 创建手续费任务
func NewPriceTask(taskCtx *TaskContext) *PriceTask {
	rpcURL := taskCtx.ChainConfig.RPCURL
	rpcClient := rpc.New(rpcURL)
	return &PriceTask{
		db:          taskCtx.DB,
		redis:       taskCtx.Redis,
		redSync:     taskCtx.RedSync,
		rpcClient:   rpcClient,
		rpcURL:      rpcURL,
		tokenDec:    taskCtx.TokenDecimal,
		tokenConfig: taskCtx.TokenConfig,
	}
}

func (t *PriceTask) Start() {
	go runPeriodic(&t.redSync, 15*time.Second, priceFetchRaydiumLock, 5*time.Minute, t.fetchRaydiumPrice)
}

func (t *PriceTask) fetchRaydiumPrice() {
	if err := t.fetchRaydiumQuoteTokenPrice("SOL", "So11111111111111111111111111111111111111112", 9); err != nil {
		log.Errorf("获取token兑换solana价格获取失败: %v", err)
		return
	}
	if err := t.fetchRaydiumQuoteTokenPrice("USDT", "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", 6); err != nil {
		log.Errorf("获取token兑换usdt价格获取失败: %v", err)
		return
	}
	if err := t.fetchRaydiumUSDTQuoteSOLPrice("USDT", "SOL", "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", "So11111111111111111111111111111111111111112", 1000000, 6, 9); err != nil {
		log.Errorf("获取usdt兑换sol价格获取失败: %v", err)
		return
	}
}

func (t *PriceTask) fetchRaydiumQuoteTokenPrice(symbol, account string, decimal int) error {
	priceTTL := 2 * time.Minute
	redisKey := fmt.Sprintf("base:sol:price:raydium-quote-%s", strings.ToLower(symbol))
	url := fmt.Sprintf("https://transaction-v1.raydium.io/compute/swap-base-in?inputMint=ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH&outputMint=%s&amount=1000&slippageBps=50&txVersion=V0", account)
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// Send the GET request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check for non-200 HTTP status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse the JSON response
	var result entity.RaydiumPriceRes
	if err = json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 安全数值转换
	inputAmount, err := strconv.ParseUint(result.Data.InputAmount, 10, 64)
	if err != nil {
		log.Errorf("输入金额解析失败: %v", err)
		return fmt.Errorf("parse input amount error: %v", err)
	}

	outputAmount, err := strconv.ParseUint(result.Data.OutputAmount, 10, 64)
	if err != nil {
		log.Errorf("输出金额解析失败: %v", err)
		return fmt.Errorf("parse output amount error: %v", err)
	}

	// 防除零保护
	tokenAmount := float64(inputAmount) / t.tokenDec
	if tokenAmount <= 0 {
		log.Error("无效的Token数量")
		return fmt.Errorf("invalid token amount")
	}

	// 价格计算
	solAmount := float64(outputAmount) / math.Pow10(decimal)
	price := solAmount / tokenAmount
	if math.IsInf(price, 0) || math.IsNaN(price) {
		log.Error("价格计算异常（无穷或非数字）")
		return fmt.Errorf("calculate price exception")
	}

	scaled := int64(price * math.Pow10(decimal))

	// 原子化写入Redis
	if err := t.redis.Set(
		context.Background(),
		redisKey,
		strconv.FormatInt(scaled, 10),
		priceTTL,
	).Err(); err != nil {
		log.Errorf("Redis写入失败: %v", err)
	}

	return nil
}

func (t *PriceTask) fetchRaydiumUSDTQuoteSOLPrice(inputSymbol, outputSymbol, inputMint, outputMint string, inputAmount uint64, inputDecimal, outputDecimal int) error {
	priceTTL := 2 * time.Minute
	redisKey := fmt.Sprintf("base:sol:price:raydium-%s-quote-%s", strings.ToLower(inputSymbol), strings.ToLower(outputSymbol))
	url := fmt.Sprintf("https://transaction-v1.raydium.io/compute/swap-base-in?inputMint=%s&outputMint=%s&amount=%d&slippageBps=50&txVersion=V0", inputMint, outputMint, inputAmount)
	prefix := fmt.Sprintf("获取 raydium %s 兑换 %s 价格 -", inputSymbol, outputSymbol)
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// Send the GET request
	resp, err := client.Get(url)
	if err != nil {
		log.Errorf("%s 发送请求到raydium错误: %v", prefix, err)
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("%s 发送请求到raydium错误，错误码: %s", prefix, resp.Status)
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("%s 发送请求到raydium错误，无法读取返回body: %s", prefix, err)
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse the JSON response
	var result entity.RaydiumPriceRes
	if err = json.Unmarshal(body, &result); err != nil {
		log.Errorf("%s 发送请求到raydium错误，解析body为json错误: %s", prefix, err)
		return fmt.Errorf("failed to parse JSON: %v", err)
	}
	log.Tracef("%s raydium 返回价格 inputAmount %s outputAmount %s", prefix, result.Data.InputAmount, result.Data.OutputAmount)

	// 安全数值转换
	inputAmountBaseUnit, err := strconv.ParseUint(result.Data.InputAmount, 10, 64)
	if err != nil {
		log.Errorf("%s 发送请求到raydium错误，输入金额解析失败: %v", prefix, err)
		return fmt.Errorf("parse input amount error: %v", err)
	}

	outputAmountBaseUnit, err := strconv.ParseUint(result.Data.OutputAmount, 10, 64)
	if err != nil {
		log.Errorf("%s 发送请求到raydium错误，输出金额解析失败: %v", prefix, err)
		return fmt.Errorf("parse output amount error: %v", err)
	}
	log.Tracef("%s raydium 返回价格 inputAmountBaseUnit %d outputAmountBaseUnit %d", prefix, inputAmountBaseUnit, outputAmountBaseUnit)

	// 防除零保护
	inputAmountDisplay := float64(inputAmountBaseUnit) / math.Pow10(inputDecimal)
	if inputAmountDisplay <= 0 {
		log.Errorf("%s 发送请求到raydium错误，raydium返回无效的输入金额", prefix)
		return fmt.Errorf("invalid token amount")
	}

	// 价格计算
	outputAmountDisplay := float64(outputAmountBaseUnit) / math.Pow10(outputDecimal)

	log.Tracef("%s raydium 返回价格 inputAmountDisplay %v outputAmountDisplay %v", prefix, inputAmountDisplay, outputAmountDisplay)

	price := outputAmountDisplay / inputAmountDisplay
	if math.IsInf(price, 0) || math.IsNaN(price) {
		log.Errorf("%s 发送请求到raydium错误，价格计算异常（无穷或非数字)", prefix)
		return fmt.Errorf("calculate price exception")
	}
	log.Tracef("%s raydium 返回价格 %v", prefix, price)

	priceStr := strconv.FormatFloat(price, 'f', outputDecimal, 64)
	if err := t.redis.Set(
		context.Background(),
		redisKey,
		priceStr,
		priceTTL,
	).Err(); err != nil {
		log.Errorf("%s 发送请求到raydium错误，Redis写入失败: %v", prefix, err)
		return fmt.Errorf("faild to store price in redis")
	}

	return nil
}
