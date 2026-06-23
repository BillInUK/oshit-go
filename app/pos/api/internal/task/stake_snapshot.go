package task

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
)

const (
	MaxStakeRecordNum   = 10
	RateTierCutoffSlot  = uint64(425359465) // 2026-06-10 00:00:00 HKT
)

// StakeSnapshotTask 质押快照服务
type StakeSnapshotTask struct {
	db          *gorm.DB
	rpcClient   *rpc.Client
	rpcURL      string
	kafkaWriter *kafka.Writer
	programID   solana.PublicKey
	mint        solana.PublicKey
	dec         uint64
}

// StakeRecordLocal 本地质押记录结构
type StakeRecordLocal struct {
	StakeType      uint8
	StakedAmount   uint64
	StakeStartSlot uint64
	StakeEndSlot   uint64
}

// StakeInfoLocal 本地质押信息结构
type StakeInfoLocal struct {
	UserWallet solana.PublicKey
	Stakes     [MaxStakeRecordNum]StakeRecordLocal
}

// NewStakeSnapShotTask 创建质押快照服务
func NewStakeSnapShotTask(taskCtx *TaskContext) *StakeSnapshotTask {
	var kafkaWriter *kafka.Writer
	if taskCtx.KafkaProducer != nil {
		if w, ok := taskCtx.KafkaProducer.(*kafka.Writer); ok {
			kafkaWriter = w
		}
	}
	dec := math.Pow10(int(taskCtx.TokenConfig.Decimals))
	return &StakeSnapshotTask{
		db:          taskCtx.DB,
		rpcClient:   taskCtx.RpcClient,
		rpcURL:      taskCtx.RpcURL,
		kafkaWriter: kafkaWriter,
		programID:   solana.MustPublicKeyFromBase58(taskCtx.RewardConfig.ProgramID),
		mint:        solana.MustPublicKeyFromBase58(taskCtx.TokenConfig.Mint),
		dec:         uint64(dec),
	}
}

// Start 启动服务的函数
func (t *StakeSnapshotTask) Start() {
	go t.startTask()
}

func (t *StakeSnapshotTask) startTask() {
	//计算第一次执行的时间
	location, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		log.Fatalf("Stake业务 - 无法加载时区: %v", err)
	}
	now := time.Now().In(location)
	nextExecution := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, location)
	if now.After(nextExecution) {
		nextExecution = nextExecution.Add(24 * time.Hour)
	}
	timeUntilNextExecution := time.Until(nextExecution)

	//等待到第一次执行时间
	log.Infof("Stake - 距离快照任务的首次执行时间还有: %v", timeUntilNextExecution)
	time.Sleep(timeUntilNextExecution)

	for {
		log.Infof("Taking stake snapshot...")

		// 获取所有质押信息账户
		stakeInfos, err := t.findAllStakeInfoAccounts()
		if err != nil {
			log.Infof("Failed to find stake info accounts: %v", err)
			return
		}

		// 获取当前slot用于检查是否过期
		currentSlot, err := utils.GetCurrentSlot(t.rpcClient, rpc.CommitmentConfirmed)
		if err != nil {
			log.Errorf("Warning: Could not get current slot, expiration check may be inaccurate")
			return
		}

		// 处理每个用户的质押信息
		var snapshots []model.StakeSnapShot
		today := time.Now().UTC().Truncate(24 * time.Hour) // 获取今天的日期

		for _, stakeInfo := range stakeInfos {
			type groupKey struct {
				stakeType uint8
				rateTier  int32
			}
			groupMap := make(map[groupKey]uint64)

			for _, record := range stakeInfo.Stakes {
				if record.StakedAmount == 0 {
					continue
				}
				if currentSlot != 0 && currentSlot >= record.StakeEndSlot {
					continue
				}
				rateTier := int32(1)
				if record.StakeStartSlot >= RateTierCutoffSlot {
					rateTier = 2
				}
				key := groupKey{stakeType: record.StakeType, rateTier: rateTier}
				groupMap[key] += record.StakedAmount
			}

			for key, amount := range groupMap {
				snapshots = append(snapshots, model.StakeSnapShot{
					NativeAccount: stakeInfo.UserWallet.String(),
					Amount:        float64(amount),
					StakeType:     int32(key.stakeType),
					RateTier:      key.rateTier,
					SnapDay:       today,
				})
			}
		}

		// 批量插入或更新到数据库
		if len(snapshots) > 0 {
			if err := t.saveSnapshots(snapshots); err != nil {
				log.Infof("Failed to save snapshots to database: %v", err)
			} else {
				log.Infof("Successfully saved %d stake snapshots to database", len(snapshots))
			}
		} else {
			log.Infof("No active stakes found to snapshot")
		}

		day := time.Now()
		rmqMsg := entity.KafkaNewSnapShotMsg{
			MsgType:    "NewStakeSnapShot",
			MsgContent: day,
		}
		if err := t.sendMsgToKafka(rmqMsg); err != nil {
			log.Errorf("Stake业务 - 分发Kafka消息错误: %v", err)
			break
		}

		// 等待到第二天的新加坡时间12点
		log.Infof("Stake业务 - 快照结束")
		nextExecution = nextExecution.Add(24 * time.Hour)
		timeUntilNextExecution = time.Until(nextExecution)
		log.Infof("Stake业务 - 距离下次快照任务执行时间还有: %v", timeUntilNextExecution)
		time.Sleep(timeUntilNextExecution)

	}
}

// StartTaskManually 手动执行执行质押快照
func (t *StakeSnapshotTask) StartTaskManually() {
	log.Infof("Taking stake snapshot...")

	// 获取所有质押信息账户
	stakeInfos, err := t.findAllStakeInfoAccounts()
	if err != nil {
		log.Infof("Failed to find stake info accounts: %v", err)
		return
	}

	// 获取当前slot用于检查是否过期
	currentSlot, err := utils.GetCurrentSlot(t.rpcClient, rpc.CommitmentConfirmed)
	if err != nil {
		log.Errorf("Warning: Could not get current slot, expiration check may be inaccurate")
		return
	}

	// 处理每个用户的质押信息
	var snapshots []model.StakeSnapShot
	today := time.Now().UTC().Truncate(24 * time.Hour) // 获取今天的日期

	for _, stakeInfo := range stakeInfos {
		type groupKey struct {
			stakeType uint8
			rateTier  int32
		}
		groupMap := make(map[groupKey]uint64)

		for _, record := range stakeInfo.Stakes {
			if record.StakedAmount == 0 {
				continue
			}
			if currentSlot != 0 && currentSlot >= record.StakeEndSlot {
				continue
			}
			rateTier := int32(1)
			if record.StakeStartSlot >= RateTierCutoffSlot {
				rateTier = 2
			}
			key := groupKey{stakeType: record.StakeType, rateTier: rateTier}
			groupMap[key] += record.StakedAmount
		}

		for key, amount := range groupMap {
			snapshots = append(snapshots, model.StakeSnapShot{
				NativeAccount: stakeInfo.UserWallet.String(),
				Amount:        float64(amount),
				StakeType:     int32(key.stakeType),
				RateTier:      key.rateTier,
				SnapDay:       today,
			})
		}
	}

	// 批量插入或更新到数据库
	if len(snapshots) > 0 {
		if err := t.saveSnapshots(snapshots); err != nil {
			log.Infof("Failed to save snapshots to database: %v", err)
		} else {
			log.Infof("Successfully saved %d stake snapshots to database", len(snapshots))
		}
	} else {
		log.Infof("No active stakes found to snapshot")
	}

	rmqMsg := entity.KafkaNewSnapShotMsg{
		MsgType:    "NewStakeSnapShot",
		MsgContent: today,
	}
	if err := t.sendMsgToKafka(rmqMsg); err != nil {
		log.Errorf("Stake业务 - 分发Kafka消息错误: %v", err)
		return
	}
}

// sendMsgToKafka 发送消息到Kafka
func (t *StakeSnapshotTask) sendMsgToKafka(msg entity.KafkaMsg) error {
	if t.kafkaWriter == nil {
		return fmt.Errorf("kafka producer 未初始化")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化Kafka消息失败: %v", err)
	}

	return t.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Topic: "StakeTopic",
		Key:   []byte(msg.GetMsgType()),
		Value: body,
	})
}

// saveSnapshots 保存快照到数据库
func (t *StakeSnapshotTask) saveSnapshots(snapshots []model.StakeSnapShot) error {
	if len(snapshots) == 0 {
		return nil
	}

	return t.db.Table(model.TableNameStakeSnapShot).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "native_account"},
				{Name: "snap_day"},
				{Name: "stake_type"},
				{Name: "rate_tier"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"amount", "updated_at"}),
		},
	).CreateInBatches(snapshots, 100).Error
}

// stakeInfoAccountSize = 8(discriminator) + 32(wallet) + 10*32(StakeRecord with alignment) = 360
// Rust aligns StakeRecord{u8,u64,u64,u64} to 32 bytes (8-byte alignment adds 7 padding after u8)
const stakeInfoAccountSize = 360

// Helius getProgramAccountsV2 请求/响应结构
type heliusProgramAccountsReq struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type heliusProgramAccountsRsp struct {
	Result struct {
		Accounts []struct {
			Pubkey  string `json:"pubkey"`
			Account struct {
				Data []string `json:"data"` // [base64_data, "base64"]
			} `json:"account"`
		} `json:"accounts"`
		PaginationKey *string `json:"paginationKey"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// findAllStakeInfoAccounts 获取所有质押信息账户
// 如果 RPC 是 Helius，使用 getProgramAccountsV2 分页；否则用标准 RPC
func (t *StakeSnapshotTask) findAllStakeInfoAccounts() ([]*StakeInfoLocal, error) {
	if t.rpcURL == "" {
		return t.findAllStakeInfoAccountsStandard()
	}

	var stakeInfos []*StakeInfoLocal
	var paginationKey *string
	page := 0
	httpClient := &http.Client{Timeout: 30 * time.Second}

	for {
		page++
		opts := map[string]interface{}{
			"encoding": "base64",
			"filters":  []map[string]interface{}{{"dataSize": stakeInfoAccountSize}},
			"limit":    5000,
		}
		if paginationKey != nil {
			opts["paginationKey"] = *paginationKey
		}

		reqBody := heliusProgramAccountsReq{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "getProgramAccountsV2",
			Params:  []interface{}{t.programID.String(), opts},
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}

		resp, err := httpClient.Post(t.rpcURL, "application/json", bytes.NewReader(jsonData))
		if err != nil {
			return nil, fmt.Errorf("page %d request failed: %w", page, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("page %d read response: %w", page, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("page %d HTTP %d: %s", page, resp.StatusCode, string(body))
		}

		var result heliusProgramAccountsRsp
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("page %d parse response: %w", page, err)
		}
		if result.Error != nil {
			return nil, fmt.Errorf("page %d RPC error %d: %s", page, result.Error.Code, result.Error.Message)
		}

		for _, acct := range result.Result.Accounts {
			if len(acct.Account.Data) < 1 {
				continue
			}
			data, err := base64.StdEncoding.DecodeString(acct.Account.Data[0])
			if err != nil {
				log.Infof("Failed to decode account %s: %v", acct.Pubkey, err)
				continue
			}
			if len(data) < stakeInfoAccountSize {
				continue
			}
			stakeInfo, err := t.parseStakeInfoAccount(data)
			if err != nil {
				log.Infof("Failed to parse account %s: %v", acct.Pubkey, err)
				continue
			}
			stakeInfos = append(stakeInfos, stakeInfo)
		}

		log.Infof("Page %d: fetched %d accounts, total parsed: %d", page, len(result.Result.Accounts), len(stakeInfos))

		paginationKey = result.Result.PaginationKey
		if paginationKey == nil {
			break
		}
	}

	log.Infof("Found %d stake info accounts in %d pages", len(stakeInfos), page)
	return stakeInfos, nil
}

// findAllStakeInfoAccountsStandard 标准 RPC fallback（无分页，适用于小规模数据）
func (t *StakeSnapshotTask) findAllStakeInfoAccountsStandard() ([]*StakeInfoLocal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dataSize := uint64(stakeInfoAccountSize)
	accounts, err := t.rpcClient.GetProgramAccountsWithOpts(ctx, t.programID, &rpc.GetProgramAccountsOpts{
		Filters: []rpc.RPCFilter{{DataSize: dataSize}},
	})
	if err != nil {
		return nil, fmt.Errorf("get program accounts: %w", err)
	}

	log.Infof("Found %d program accounts (standard RPC)", len(accounts))
	var stakeInfos []*StakeInfoLocal
	for _, account := range accounts {
		data := account.Account.Data.GetBinary()
		if len(data) < stakeInfoAccountSize {
			continue
		}
		stakeInfo, err := t.parseStakeInfoAccount(data)
		if err != nil {
			log.Infof("Failed to parse account %s: %v", account.Pubkey, err)
			continue
		}
		stakeInfos = append(stakeInfos, stakeInfo)
	}
	return stakeInfos, nil
}

// parseStakeInfoAccount 解析质押信息账户
func (t *StakeSnapshotTask) parseStakeInfoAccount(data []byte) (*StakeInfoLocal, error) {
	if len(data) < 8+32 {
		return nil, fmt.Errorf("data too short for stake info account")
	}

	stakeInfo := &StakeInfoLocal{}

	// 跳过 8 字节的 discriminator
	offset := 8

	// 解析 user_wallet (32 bytes)
	copy(stakeInfo.UserWallet[:], data[offset:offset+32])
	offset += 32

	// 解析 StakeRecord 数组
	// Anchor 使用 Borsh 序列化（无 padding）: u8(1) + u64(8) + u64(8) + u64(8) = 25 bytes per record
	const stakeRecordSize = 25
	for i := 0; i < MaxStakeRecordNum && offset+stakeRecordSize <= len(data); i++ {
		record := StakeRecordLocal{}

		// stake_type (1 byte, no padding in Borsh)
		record.StakeType = data[offset]
		offset += 1

		// staked_amount (8 bytes, little endian)
		// 合约存的是不含精度的基础数量，乘以 10^decimals 还原为原始值（最小单位）
		record.StakedAmount = binary.LittleEndian.Uint64(data[offset:offset+8]) * t.dec
		offset += 8

		// stake_start_slot (8 bytes, little endian)
		record.StakeStartSlot = binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8

		// stake_end_slot (8 bytes, little endian)
		record.StakeEndSlot = binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8

		stakeInfo.Stakes[i] = record
	}

	return stakeInfo, nil
}
