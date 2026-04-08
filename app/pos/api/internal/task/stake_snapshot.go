package task

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"time"
)

const (
	MaxStakeRecordNum = 10
)

// StakeSnapshotTask 质押快照服务
type StakeSnapshotTask struct {
	db          *gorm.DB
	rpcClient   *rpc.Client
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
		log.Fatalf("Pos业务 - 无法加载时区: %v", err)
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
			// 按类型统计未过期的质押金额
			type0Amount := uint64(0) // 类型0总金额
			type1Amount := uint64(0) // 类型1总金额

			for _, record := range stakeInfo.Stakes {
				if record.StakedAmount > 0 {
					// 检查是否过期：如果没有当前slot信息或者当前slot小于结束slot，则认为未过期
					if currentSlot == 0 || currentSlot < record.StakeEndSlot {
						switch record.StakeType {
						case 0:
							type0Amount += record.StakedAmount
						case 1:
							type1Amount += record.StakedAmount
						}
					}
				}
			}

			// 创建快照记录
			if type0Amount > 0 {
				snapshots = append(snapshots, model.StakeSnapShot{
					NativeAccount: stakeInfo.UserWallet.String(),
					Amount:        float64(type0Amount), // 注意：这里存储的是基础单位
					StakeType:     0,
					SnapDay:       today,
				})
			}

			if type1Amount > 0 {
				snapshots = append(snapshots, model.StakeSnapShot{
					NativeAccount: stakeInfo.UserWallet.String(),
					Amount:        float64(type1Amount), // 注意：这里存储的是基础单位
					StakeType:     1,
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
			log.Errorf("Stake业务 - 分发RocketMQ消息错误: %v", err)
			break
		}

		// 等待到第二天的新加坡时间12点
		log.Infof("Pos业务 - 快照结束")
		nextExecution = nextExecution.Add(24 * time.Hour)
		timeUntilNextExecution = time.Until(nextExecution)
		log.Infof("Pos业务 - 距离下次快照任务执行时间还有: %v", timeUntilNextExecution)
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
		// 按类型统计未过期的质押金额
		type0Amount := uint64(0) // 类型0总金额
		type1Amount := uint64(0) // 类型1总金额

		for _, record := range stakeInfo.Stakes {
			if record.StakedAmount > 0 {
				// 检查是否过期：如果没有当前slot信息或者当前slot小于结束slot，则认为未过期
				if currentSlot == 0 || currentSlot < record.StakeEndSlot {
					switch record.StakeType {
					case 0:
						type0Amount += record.StakedAmount
					case 1:
						type1Amount += record.StakedAmount
					}
				}
			}
		}

		// 创建快照记录
		if type0Amount > 0 {
			snapshots = append(snapshots, model.StakeSnapShot{
				NativeAccount: stakeInfo.UserWallet.String(),
				Amount:        float64(type0Amount), // 注意：这里存储的是基础单位
				StakeType:     0,
				SnapDay:       today,
			})
		}

		if type1Amount > 0 {
			snapshots = append(snapshots, model.StakeSnapShot{
				NativeAccount: stakeInfo.UserWallet.String(),
				Amount:        float64(type1Amount), // 注意：这里存储的是基础单位
				StakeType:     1,
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
		log.Errorf("Pos业务 - 分发RocketMQ消息错误: %v", err)
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
			},
			DoUpdates: clause.AssignmentColumns([]string{"amount", "updated_at"}),
		},
	).CreateInBatches(snapshots, 100).Error
}

// findAllStakeInfoAccounts 查找所有质押信息账户
func (t *StakeSnapshotTask) findAllStakeInfoAccounts() ([]*StakeInfoLocal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用GetProgramAccounts获取所有程序账户
	accounts, err := t.rpcClient.GetProgramAccountsWithOpts(
		ctx,
		t.programID,
		&rpc.GetProgramAccountsOpts{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get program accounts: %w", err)
	}

	log.Infof("Found %d program accounts", len(accounts))

	var stakeInfos []*StakeInfoLocal

	for _, account := range accounts {
		data := account.Account.Data.GetBinary()

		// 只处理大小合适的账户（质押信息账户）
		if len(data) > 100 {
			stakeInfo, err := t.parseStakeInfoAccount(data)
			if err != nil {
				log.Infof("Failed to parse account %s as stake info: %v", account.Pubkey, err)
				continue
			}

			stakeInfos = append(stakeInfos, stakeInfo)
		}
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
	for i := 0; i < MaxStakeRecordNum && offset+25 <= len(data); i++ {
		record := StakeRecordLocal{}

		// stake_type (1 byte)
		record.StakeType = data[offset]
		offset += 1

		// staked_amount (8 bytes, little endian)
		if offset+8 <= len(data) {
			record.StakedAmount = binary.LittleEndian.Uint64(data[offset : offset+8])
			record.StakedAmount = record.StakedAmount * t.dec
			offset += 8
		}

		// stake_start_slot (8 bytes, little endian)
		if offset+8 <= len(data) {
			record.StakeStartSlot = binary.LittleEndian.Uint64(data[offset : offset+8])
			offset += 8
		}

		// stake_end_slot (8 bytes, little endian)
		if offset+8 <= len(data) {
			record.StakeEndSlot = binary.LittleEndian.Uint64(data[offset : offset+8])
			offset += 8
		}

		stakeInfo.Stakes[i] = record
	}

	return stakeInfo, nil
}
