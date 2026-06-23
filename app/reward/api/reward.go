package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jsoniter "github.com/json-iterator/go"

	"oshit-go/app/reward/api/internal/handler"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/common/pkg/logging"
)

func main() {
	// 初始化日志：JSON 结构化输出 + lumberjack 轮转
	logging.Setup(logging.Config{
		ServiceName: "reward-api",
		LogDir:      "./log",
		MaxSizeMB:   100,
		MaxAgeDays:  30,
		MaxBackups:  10,
		Compress:    true,
		Level:       slog.LevelInfo,
	})

	// 创建服务上下文
	srvCtx, err := svc.NewServiceContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL] %v\n", err)
		log.Fatal(err)
	}
	defer srvCtx.Close()

	// 注册 Kafka 消息处理器并启动任务
	registerTasks(srvCtx)

	// 读取启动参数
	appName := srvCtx.Config.App.Name
	appPort := strconv.Itoa(srvCtx.Config.App.Port)

	// 配置 json-iterator
	// 重要!! 需要禁用 6位小数截断
	json := jsoniter.Config{
		EscapeHTML:              true,
		MarshalFloatWith6Digits: false,
	}.Froze()

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		AppName:     appName,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// 中间件
	app.Use(logger.New())
	app.Use(recover.New())

	// 注册路由
	handler.RegisterRoutes(app, srvCtx)

	// 启动服务
	log.Printf("Reward API starting on :%s", appPort)
	if err := app.Listen(":" + appPort); err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL] %v\n", err)
		log.Fatal(err)
	}
}
