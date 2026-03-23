package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io"
	"net/http"
	"time"
)

// PriceTask 价格数据获取任务
type PriceTask struct {
	*BaseTask
	redis *redis.Client
}

// NewPriceTask 创建价格任务
func NewPriceTask(redis *redis.Client) *PriceTask {
	base := NewBaseTask("price_task", 5*time.Minute)
	return &PriceTask{
		BaseTask: base,
		redis:    redis,
	}
}

// Start 启动任务
func (t *PriceTask) Start(ctx context.Context) error {
	// 启动不同时间间隔的价格获取任务
	go t.RunWithInterval(ctx, func(ctx context.Context) error {
		return t.fetchBirdEyePrice(ctx, "1D")
	})

	go t.RunWithInterval(ctx, func(ctx context.Context) error {
		return t.fetchBirdEyePrice(ctx, "1W")
	})

	go t.RunWithInterval(ctx, func(ctx context.Context) error {
		return t.fetchBirdEyePrice(ctx, "1M")
	})

	return nil
}

// fetchBirdEyePrice 获取BirdEye价格数据
func (t *PriceTask) fetchBirdEyePrice(ctx context.Context, interval string) error {
	// 构建Redis键
	redisKey := fmt.Sprintf("BIRD-EYE-APP-PRICE-%s", interval)

	// 模拟从BirdEye API获取数据
	// 实际实现应该调用BirdEye API
	priceData := map[string]interface{}{
		"interval": interval,
		"prices": []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-24 * time.Hour).Unix(),
				"open":      0.001,
				"high":      0.0015,
				"low":       0.0008,
				"close":     0.0012,
				"volume":    1000000,
			},
			{
				"timestamp": time.Now().Unix(),
				"open":      0.0012,
				"high":      0.0018,
				"low":       0.0010,
				"close":     0.0015,
				"volume":    1500000,
			},
		},
		"current_price":               0.0015,
		"price_change_24h":            0.0003,
		"price_change_percentage_24h": 25.0,
	}

	// 转换为JSON
	priceJSON, err := json.Marshal(priceData)
	if err != nil {
		return fmt.Errorf("marshal price data error: %v", err)
	}

	// 更新Redis缓存
	if err := t.redis.Set(ctx, redisKey, priceJSON, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("set redis cache error: %v", err)
	}

	fmt.Printf("PriceTask updated %s cache at %s\n", redisKey, time.Now().Format(time.RFC3339))
	return nil
}

// fetchRealBirdEyePrice 实际从BirdEye API获取价格数据
func (t *PriceTask) fetchRealBirdEyePrice(ctx context.Context, interval string) error {
	// BirdEye API配置
	tokenAddress := "ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH" // 示例token地址
	apiKey := "your_birdeye_api_key"                              // 需要配置API Key

	// 构建API URL
	url := fmt.Sprintf("https://public-api.birdeye.so/defi/ohlcv?address=%s&type=%s&time_from=%d&time_to=%d",
		tokenAddress,
		interval,
		time.Now().Add(-30*24*time.Hour).Unix(), // 30天前
		time.Now().Unix(),
	)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bird eye API error: %s", string(body))
	}

	// 解析响应
	var priceData interface{}
	if err := json.Unmarshal(body, &priceData); err != nil {
		return fmt.Errorf("unmarshal response error: %v", err)
	}

	// 更新Redis缓存
	redisKey := fmt.Sprintf("BIRD-EYE-APP-PRICE-%s", interval)
	if err := t.redis.Set(ctx, redisKey, body, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("set redis cache error: %v", err)
	}

	return nil
}

// GetBirdEyePrice 获取BirdEye价格数据（供handler调用）
func (t *PriceTask) GetBirdEyePrice(ctx context.Context, interval string) (interface{}, error) {
	// 验证时间间隔
	validIntervals := map[string]bool{"1D": true, "1W": true, "1M": true}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("invalid interval: %s", interval)
	}

	// 构建Redis键
	redisKey := fmt.Sprintf("BIRD-EYE-APP-PRICE-%s", interval)

	// 从Redis获取数据
	val, err := t.redis.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get redis cache error: %v", err)
	}

	// 解析JSON
	var priceData interface{}
	if err := json.Unmarshal([]byte(val), &priceData); err != nil {
		return nil, fmt.Errorf("unmarshal price data error: %v", err)
	}

	return priceData, nil
}
