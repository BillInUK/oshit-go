package main

import (
	"context"
	"oshit-go/app/reward/api/internal/logic/give"
	"oshit-go/app/reward/api/internal/logic/lottery"
	"oshit-go/app/reward/api/internal/logic/take"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/common/pkg/entity"
)

// registerTasks 注册 Kafka 消息处理器并启动后台任务
func registerTasks(srvCtx *svc.ServiceContext) {
	srvCtx.TaskMgr.RegisterScannedTxHandler("TakeToken", func(ctx context.Context, msg entity.NewScannedTx) error {
		return take.NewTakeLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterScannedTxHandler("GiveToken", func(ctx context.Context, msg entity.NewScannedTx) error {
		return give.NewGiveTokenLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterScannedTxHandler("Lottery", func(ctx context.Context, msg entity.NewScannedTx) error {
		return lottery.NewLotteryLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("TakeToken", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return take.NewTakeLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("GiveToken", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return give.NewGiveTokenLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("Lottery", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return lottery.NewLotteryLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})
	go srvCtx.TaskMgr.StartAllTasks()
}
