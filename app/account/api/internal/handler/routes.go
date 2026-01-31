package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/account/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/api")
	handler := NewAuthHandler(srvCtx)

	// 用户相关路由
	user := api.Group("/users")
	{
		user.Post("/login", handler.Login)
	}
}
