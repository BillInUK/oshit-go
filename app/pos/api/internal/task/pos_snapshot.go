package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"time"
)

// PosSnapShotTask 消费 base 模块发送的 Kafka 消息
type PosSnapShotTask struct {
	reader           *kafka.Reader
	db               *gorm.DB
	redis            redis.UniversalClient
	rpcClient        *rpc.Client
	kafkaWriter      *kafka.Writer
	decimals         uint8
	tokenMintAccount string
}

// NewPosSnapShotTask 创建 pos 快照业务
func NewPosSnapShotTask(taskCtx *TaskContext) *PosSnapShotTask {
	reader, _ := taskCtx.KafkaConsumer.(*kafka.Reader)
	var kafkaWriter *kafka.Writer
	if taskCtx.KafkaProducer != nil {
		if w, ok := taskCtx.KafkaProducer.(*kafka.Writer); ok {
			kafkaWriter = w
		}
	}
	return &PosSnapShotTask{
		reader:           reader,
		db:               taskCtx.DB,
		redis:            taskCtx.Redis,
		rpcClient:        taskCtx.RpcClient,
		tokenMintAccount: taskCtx.TokenConfig.Mint,
		decimals:         uint8(taskCtx.TokenConfig.Decimals),
		kafkaWriter:      kafkaWriter,
	}
}

// Start 快照数据，并记录星级用户的数据
func (t *PosSnapShotTask) Start() {
	ctx := context.Background()
	tokenMintAccount := solana.MPK(t.tokenMintAccount)

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
	log.Infof("Pos业务 - 距离快照任务的首次执行时间还有: %v", timeUntilNextExecution)
	time.Sleep(timeUntilNextExecution)

	for {
		day := time.Now()
		log.Infof("Pos业务 - 快照开始")
		var holders uint64 = 0
		out, err := t.rpcClient.GetProgramAccountsWithOpts(ctx, solana.TokenProgramID, &rpc.GetProgramAccountsOpts{
			Encoding: solana.EncodingJSONParsed,
			Filters: []rpc.RPCFilter{
				{
					DataSize: 165,
				},
				{
					Memcmp: &rpc.RPCFilterMemcmp{Offset: 0, Bytes: tokenMintAccount.Bytes()},
				},
			},
		})
		if err != nil {
			log.Errorf("Pos业务 - 获取token余额错误: %v", err)
			continue
		}
		if len(out) > 0 {
			var snapShotRecords []model.PosSnapShot
			for _, account := range out {
				var parsedData utils.ParsedData
				dataBytes, err := account.Account.Data.MarshalJSON()
				if err != nil {
					log.Errorf("Pos业务 - 获取token余额错误: %v", err)
					break
				}
				if err = json.Unmarshal(dataBytes, &parsedData); err != nil {
					log.Errorf("Pos业务 - 获取token余额错误: %v", err)
					break
				}
				rawAmount := parsedData.Parsed.TokenAccountInfo.TokenAmount.UIAmount * math.Pow10(int(t.decimals))
				if rawAmount < 500000 {
					continue
				}
				owner := parsedData.Parsed.TokenAccountInfo.Owner.String()
				snapShotRecord := model.PosSnapShot{
					NativeAccount: owner,
					Amount:        rawAmount,
					StarLevel:     0,
					Rate:          0,
					SnapDay:       day,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}

				// 将每条记录加入到切片中
				snapShotRecords = append(snapShotRecords, snapShotRecord)
			}

			// 批量插入，避免重复记录
			if len(snapShotRecords) > 0 {
				table := t.db.Table(model.TableNamePosSnapShot)
				err := table.Clauses(
					clause.OnConflict{
						Columns: []clause.Column{
							{Name: "native_account"},
							{Name: "snap_day"},
						},
						DoNothing: true,
					}).
					CreateInBatches(snapShotRecords, 100).Error
				if err != nil {
					log.Errorf("Pos业务 - 批量插入快照数据失败: %v", err)
				}
			}
		}
		log.Infof("Pos业务 - 快照完成，需要奖励的持币人数量为 %d", holders)

		rmqMsg := entity.KafkaNewSnapShotMsg{
			MsgType:    "NewPosSnapShot",
			MsgContent: day,
		}
		if err := t.sendMsgToKafka(rmqMsg); err != nil {
			log.Errorf("Pos业务 - 分发RocketMQ消息错误: %v", err)
			break
		}

		// 等待到第二天的新加坡时间12点
		log.Infof("Pos业务 - 快照结束")
		nextExecution = nextExecution.Add(24 * time.Hour)
		timeUntilNextExecution = time.Until(nextExecution)
		log.Infof("Pos业务 - 距离下次快照任务执行时间还有: %v", timeUntilNextExecution)
		time.Sleep(timeUntilNextExecution)
		//return
	}
}

// sendMsgToKafka 发送消息到Kafka
func (t *PosSnapShotTask) sendMsgToKafka(msg entity.KafkaMsg) error {
	if t.kafkaWriter == nil {
		return fmt.Errorf("kafka producer 未初始化")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化Kafka消息失败: %v", err)
	}

	return t.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Topic: "PosTopic",
		Key:   []byte(msg.GetMsgType()),
		Value: body,
	})
}
