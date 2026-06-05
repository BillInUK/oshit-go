package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"oshit-go/common/utils"
)

// 分布式锁 key
const holdersFetchLock = "base:sol:holders:fetch:lock"

// Redis 数据 key
const holdersCount = "base:sol:holders:count"

// 主网 SHIT token mint
const holdersTokenMint = "ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH"

// HoldersTask token 持有者统计任务
type HoldersTask struct {
	db         *gorm.DB
	redis      redis.UniversalClient
	redSync    redsync.Redsync
	mainnetURL string
}

// NewHoldersTask 创建持有者统计任务
func NewHoldersTask(taskCtx *TaskContext) *HoldersTask {
	mainnetURL := taskCtx.MainnetRpcURL
	if mainnetURL == "" && taskCtx.HeliusAPIKey != "" {
		mainnetURL = utils.BuildRPCURL("helius", "https://mainnet.helius-rpc.com", taskCtx.HeliusAPIKey)
	}
	return &HoldersTask{
		db:         taskCtx.DB,
		redis:      taskCtx.Redis,
		redSync:    taskCtx.RedSync,
		mainnetURL: mainnetURL,
	}
}

// Start 启动任务（每 5 分钟统计一次）
func (t *HoldersTask) Start() {
	go runPeriodic(&t.redSync, 5*time.Minute, holdersFetchLock, 10*time.Minute, t.fetchHolderTask)
}

func (t *HoldersTask) fetchHolderTask() {
	if t.mainnetURL == "" {
		log.Warnf("持有者统计跳过: mainnet RPC URL 未配置")
		return
	}
	holders, err := t.countTokenHolders()
	if err != nil {
		log.Warnf("持有者查询失败: %v", err)
		return
	}
	if err := t.redis.Set(context.Background(), holdersCount, holders, 0).Err(); err != nil {
		log.Errorf("持有者数量 Redis 写入失败: %v", err)
	}
	log.Infof("token 持有者数量更新: %d", holders)
}

// Helius getTokenAccounts 请求/响应
type heliusGetTokenAccountsReq struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type heliusGetTokenAccountsRsp struct {
	Result struct {
		TokenAccounts []json.RawMessage `json:"token_accounts"`
		Cursor        string            `json:"cursor"`
	} `json:"result"`
}

// countTokenHolders 通过 Helius getTokenAccounts 分页统计持有者数量
func (t *HoldersTask) countTokenHolders() (int, error) {
	total := 0
	cursor := ""
	client := &http.Client{Timeout: 15 * time.Second}

	for {
		params := map[string]interface{}{
			"mint":    holdersTokenMint,
			"limit":   1000,
			"options": map[string]bool{"showZeroBalance": false},
		}
		if cursor != "" {
			params["cursor"] = cursor
		}

		reqBody := heliusGetTokenAccountsReq{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "getTokenAccounts",
			Params:  params,
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return 0, fmt.Errorf("marshal request: %w", err)
		}

		resp, err := client.Post(t.mainnetURL, "application/json", bytes.NewReader(jsonData))
		if err != nil {
			return 0, fmt.Errorf("send request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return 0, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		var result heliusGetTokenAccountsRsp
		if err := json.Unmarshal(body, &result); err != nil {
			return 0, fmt.Errorf("parse response: %w", err)
		}

		count := len(result.Result.TokenAccounts)
		total += count
		cursor = result.Result.Cursor

		if count < 1000 || cursor == "" {
			break
		}
	}

	return total, nil
}
