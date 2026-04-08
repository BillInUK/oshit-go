package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/pos/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/snap")
	posHandler := NewPosHandler(srvCtx)
	stakeHandler := NewStakeHandler(srvCtx)

	// pos 相关路由
	pos := api.Group("/pos")
	{
		pos.Post("/reward/config", posHandler.GetConfig)
		pos.Post("/reward/record", posHandler.GetRecord)
		pos.Post("/reward/tx-info", posHandler.GetTxInfo)
		pos.Post("/reward/commit-tx", posHandler.CommitTx)
	}

	// stake 相关路由
	stake := api.Group("/stake")
	{
		// 快照相关接口
		stake.Post("/shot/take", stakeHandler.TakeSnapShot)
		stake.Post("/shot/reset", stakeHandler.ResetSnapShot)

		// 用户奖励相关接口
		stake.Post("/reward/config", stakeHandler.GetConfig)
		stake.Post("/reward/record", stakeHandler.GetRewardRecord)
		stake.Post("/reward/claim-record", stakeHandler.GetClaimRecord)
		stake.Post("/reward/tx-info", stakeHandler.GetTxInfo)
		stake.Post("/reward/commit-tx", stakeHandler.CommitTx)
	}
}
