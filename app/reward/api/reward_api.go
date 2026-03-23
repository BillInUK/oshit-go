package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"oshit-go/app/reward/api/internal/handler"
	"oshit-go/app/reward/api/internal/svc"
	"strconv"
)

func main() {
	// 创建服务上下文
	srvCtx, err := svc.NewServiceContext()
	if err != nil {
		log.Fatal(err)
	}
	defer srvCtx.Close()

	// 读取启动参数
	appName := srvCtx.Config.App.Name
	appPort := strconv.Itoa(srvCtx.Config.App.Port)

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		AppName: appName,
	})

	// 中间件
	app.Use(logger.New())
	app.Use(recover.New())

	// 注册路由
	handler.RegisterRoutes(app, srvCtx)

	// 启动服务
	log.Printf("Reward API starting on :%s", appPort)
	log.Fatal(app.Listen(":" + appPort))
}
