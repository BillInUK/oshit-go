package main

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/pos/api/internal/logic/pos"
	"oshit-go/app/pos/api/internal/logic/stake"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/common/pkg/entity"
)

// registerTasks 注册 Kafka 消息处理器并启动后台任务
func registerTasks(srvCtx *svc.ServiceContext) {
	// 快照：PosSnapShot 和 StakeSnapShot 统一注册为 "Pos"，内部按 MsgType 路由
	srvCtx.TaskMgr.RegisterSnapShotHandler("Pos", func(ctx context.Context, msg entity.KafkaNewSnapShotMsg) error {
		switch msg.MsgType {
		case "NewPosSnapShot":
			return pos.NewPosSnapShotLogic(ctx, srvCtx).HandeSnapShot(msg)
		case "NewStakeSnapShot":
			return stake.NewStakeSnapShotLogic(ctx, srvCtx).HandeSnapShot(msg)
		default:
			log.Warnf("registerTasks: 未知快照类型 MsgType=%s", msg.MsgType)
			return nil
		}
	})

	// 已确认交易：统一注册为 "Pos"，内部按 SubService 路由
	srvCtx.TaskMgr.RegisterScannedTxHandler("Pos", func(ctx context.Context, msg entity.NewScannedTx) error {
		switch msg.SubService {
		case "PosReward":
			return pos.NewPosRewardLogic(ctx, srvCtx).HandleScannedTx(msg)
		case "StakeToken":
			return stake.NewStakeLogic(ctx, srvCtx).HandleStakeTx(msg)
		case "StakeReward":
			return stake.NewStakeRewardLogic(ctx, srvCtx).HandleScannedTx(msg)
		case "StakeLeaderReward":
			return stake.NewStakeRewardLogic(ctx, srvCtx).HandleLeaderScannedTx(msg)
		default:
			log.Warnf("registerTasks: 未知 SubService=%s 的已扫描交易，跳过", msg.SubService)
			return nil
		}
	})

	// 超时交易：统一注册为 "Pos"，内部按 SubService 路由
	srvCtx.TaskMgr.RegisterExpiredTxHandler("Pos", func(ctx context.Context, msg entity.NewExpiredTx) error {
		switch msg.SubService {
		case "PosReward":
			return pos.NewPosRewardLogic(ctx, srvCtx).HandleExpiredTx(msg)
		case "StakeReward":
			return stake.NewStakeRewardLogic(ctx, srvCtx).HandleExpiredTx(msg)
		case "StakeLeaderReward":
			return stake.NewStakeRewardLogic(ctx, srvCtx).HandleLeaderExpiredTx(msg)
		default:
			log.Warnf("registerTasks: 未知 SubService=%s 的超时交易，跳过", msg.SubService)
			return nil
		}
	})

	go srvCtx.TaskMgr.StartAllTasks()
}
