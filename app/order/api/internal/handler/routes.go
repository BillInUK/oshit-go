package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/order/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/api")
	handler := NewOrderHandler(srvCtx)

	// 用户相关路由
	user := api.Group("/order")
	{
		user.Post("/fetchUserOrders", handler.FetchUserOrders)
	}
}
