package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"oshit-go/app/order/api/internal/config"
	"oshit-go/app/order/api/internal/handler"
	"oshit-go/app/order/api/internal/rpc"
	"oshit-go/app/order/api/internal/svc"
	"strconv"
)

func main() {
	if err := config.InitNacosConfig(); err != nil {
		fmt.Printf("Nacos配置中心初始化失败，程序退出：%v\n", err)
		return
	}

	// 2. 第二步：初始化dubbo-go消费端（调用上面的InitDubboClient）
	if err := rpc.InitDubboClient(); err != nil {
		fmt.Printf("❌ dubbo-go消费端初始化失败，程序退出: %v\n", err)
		return
	}

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
	log.Printf("Order API starting on :%s", appPort)
	log.Fatal(app.Listen(":" + appPort))
}
