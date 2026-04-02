package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/reward/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/reward")
	takeTokenHandler := NewTakeTokenHandler(srvCtx)
	giveTokenHandler := NewGiveTokenHandler(srvCtx)

	// take token 相关路由
	take := api.Group("/take")
	{
		take.Post("/config", takeTokenHandler.GetConfig)
		take.Post("/record", takeTokenHandler.GetRecord)
		take.Post("/tx-info", takeTokenHandler.GetTxInfo)
		take.Post("/commit-tx", takeTokenHandler.CommitTx)
	}

	// give token 相关路由
	give := api.Group("/give")
	{
		give.Post("/config", giveTokenHandler.GetConfig)
		give.Post("/record", giveTokenHandler.GetRecord)
		give.Post("/tx-info", giveTokenHandler.GetTxInfo)
		give.Post("/commit-tx", giveTokenHandler.CommitTx)
	}
}
