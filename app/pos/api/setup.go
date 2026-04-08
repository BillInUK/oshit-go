package main

import (
	"context"
	"oshit-go/app/pos/api/internal/logic/pos"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/common/pkg/entity"
)

// registerTasks 注册 Kafka 消息处理器并启动后台任务
func registerTasks(srvCtx *svc.ServiceContext) {
	srvCtx.TaskMgr.RegisterScannedTxHandler("PosReward", func(ctx context.Context, msg entity.NewScannedTx) error {
		return pos.NewPosLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("PosReward", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return pos.NewPosLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})
	go srvCtx.TaskMgr.StartAllTasks()
}
