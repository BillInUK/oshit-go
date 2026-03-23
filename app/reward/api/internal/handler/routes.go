package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/reward/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/api")
	handler := NewTradeHandler(srvCtx)

	// 用户相关路由
	user := api.Group("/trade")
	{
		user.Post("/fetchUserOrders", handler.FetchUserOrders)
	}
}
