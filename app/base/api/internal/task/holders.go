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
	"net/http"
	"time"
)

// 分布式锁 key
const holdersFetchLock = "base:sol:holders:fetch:lock"

// Redis 数据 key
const holdersCount = "base:sol:holders:count"

// HoldersTask 手续费统计任务
type HoldersTask struct {
	db        *gorm.DB
	redis     redis.UniversalClient
	redSync   redsync.Redsync
	rpcClient *rpc.Client
	rpcURL    string
}

// NewHoldersTask 创建手续费任务
func NewHoldersTask(taskCtx *TaskContext) *HoldersTask {
	rpcURL := taskCtx.ChainConfig.RPCURL
	rpcClient := rpc.New(rpcURL)
	return &HoldersTask{
		db:        taskCtx.DB,
		redis:     taskCtx.Redis,
		redSync:   taskCtx.RedSync,
		rpcClient: rpcClient,
		rpcURL:    rpcURL,
	}
}

// Start 启动任务
func (t *HoldersTask) Start() {
	go runPeriodic(&t.redSync, 15*time.Second, holdersFetchLock, 5*time.Minute, t.fetchHolderTask)
}

func (t *HoldersTask) fetchHolderTask() {
	holders, err := t.fetchHoldersNumber()
	if err != nil {
		log.Errorf("持有者查询失败: %v", err)
		return
	}
	if err := t.redis.Set(
		context.Background(),
		holdersCount,
		holders,
		-1,
	).Err(); err != nil {
		log.Errorf("Redis写入失败: %v", err)
	}
}

func (t *HoldersTask) fetchHoldersNumber() (int, error) {
	url := "https://dev-omega-one.vercel.app/api/v1/getShitHolders?tokenAddress=ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("非200状态码: %d", resp.StatusCode)
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			HoldersNumber int `json:"holdersNumber"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("JSON解析失败: %w", err)
	}

	if result.Code != 200 {
		return 0, fmt.Errorf("业务状态码异常: %d", result.Code)
	}

	return result.Data.HoldersNumber, nil
}
