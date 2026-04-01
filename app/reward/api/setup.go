package main

import (
	"context"
	"oshit-go/app/reward/api/internal/logic"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/common/pkg/entity"
)

// registerTasks 注册 Kafka 消息处理器并启动后台任务
func registerTasks(srvCtx *svc.ServiceContext) {
	srvCtx.TaskMgr.RegisterScannedTxHandler("TakeToken", func(ctx context.Context, msg entity.NewScannedTx) error {
		return logic.NewTakeLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterScannedTxHandler("GiveToken", func(ctx context.Context, msg entity.NewScannedTx) error {
		return logic.NewGiveTokenLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("TakeToken", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return logic.NewTakeLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("GiveToken", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return logic.NewGiveTokenLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})
	go srvCtx.TaskMgr.StartAllTasks()
}
