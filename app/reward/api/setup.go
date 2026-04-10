package main

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic/give"
	"oshit-go/app/reward/api/internal/logic/lottery"
	"oshit-go/app/reward/api/internal/logic/take"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/common/pkg/entity"
)

// registerTasks 注册 Kafka 消息处理器并启动后台任务
func registerTasks(srvCtx *svc.ServiceContext) {
	// 已确认交易：统一注册为 "Reward"，内部按 SubService 路由
	srvCtx.TaskMgr.RegisterScannedTxHandler("Reward", func(ctx context.Context, msg entity.NewScannedTx) error {
		switch msg.SubService {
		case "TakeToken":
			return take.NewTakeLogic(ctx, srvCtx).HandleScannedTx(msg)
		case "GiveToken":
			return give.NewGiveTokenLogic(ctx, srvCtx).HandleScannedTx(msg)
		case "Lottery":
			return lottery.NewLotteryLogic(ctx, srvCtx).HandleScannedTx(msg)
		default:
			log.Warnf("registerTasks: 未知 SubService=%s 的已扫描交易，跳过", msg.SubService)
			return nil
		}
	})

	// 超时交易：统一注册为 "Reward"，内部按 SubService 路由
	srvCtx.TaskMgr.RegisterExpiredTxHandler("Reward", func(ctx context.Context, msg entity.NewExpiredTx) error {
		switch msg.SubService {
		case "TakeToken":
			return take.NewTakeLogic(ctx, srvCtx).HandleExpiredTx(msg)
		case "GiveToken":
			return give.NewGiveTokenLogic(ctx, srvCtx).HandleExpiredTx(msg)
		case "Lottery":
			return lottery.NewLotteryLogic(ctx, srvCtx).HandleExpiredTx(msg)
		default:
			log.Warnf("registerTasks: 未知 SubService=%s 的超时交易，跳过", msg.SubService)
			return nil
		}
	})

	go srvCtx.TaskMgr.StartAllTasks()
}
