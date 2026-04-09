package main

import (
	"context"
	"oshit-go/app/pos/api/internal/logic/pos"
	"oshit-go/app/pos/api/internal/logic/stake"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/common/pkg/entity"
)

// registerTasks 注册 Kafka 消息处理器并启动后台任务
func registerTasks(srvCtx *svc.ServiceContext) {
	// 注册pos相关消息处理
	srvCtx.TaskMgr.RegisterSnapShotHandler("PosSnapShot", func(ctx context.Context, msg entity.KafkaNewSnapShotMsg) error {
		return pos.NewPosSnapShotLogic(ctx, srvCtx).HandeSnapShot(msg)
	})
	srvCtx.TaskMgr.RegisterScannedTxHandler("PosReward", func(ctx context.Context, msg entity.NewScannedTx) error {
		return pos.NewPosRewardLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("PosReward", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return pos.NewPosRewardLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})

	// 注册stake相关消息处理
	srvCtx.TaskMgr.RegisterScannedTxHandler("StakeToken", func(ctx context.Context, msg entity.NewScannedTx) error {
		return stake.NewStakeLogic(ctx, srvCtx).HandleStakeTx(msg)
	})
	srvCtx.TaskMgr.RegisterSnapShotHandler("StakeSnapShot", func(ctx context.Context, msg entity.KafkaNewSnapShotMsg) error {
		return stake.NewStakeSnapShotLogic(ctx, srvCtx).HandeSnapShot(msg)
	})
	srvCtx.TaskMgr.RegisterScannedTxHandler("StakeReward", func(ctx context.Context, msg entity.NewScannedTx) error {
		return stake.NewStakeRewardLogic(ctx, srvCtx).HandleScannedTx(msg)
	})
	srvCtx.TaskMgr.RegisterExpiredTxHandler("StakeReward", func(ctx context.Context, msg entity.NewExpiredTx) error {
		return stake.NewStakeRewardLogic(ctx, srvCtx).HandleExpiredTx(msg)
	})

	go srvCtx.TaskMgr.StartAllTasks()
}
