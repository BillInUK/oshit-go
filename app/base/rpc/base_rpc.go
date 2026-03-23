package main

import (
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"log"
	"oshit-go/app/base/rpc/internal/server"
	"oshit-go/app/base/rpc/svc"
)

func main() {
	// 创建服务上下文
	srvCtx, err := svc.NewServiceContext()
	if err != nil {
		log.Fatal(err)
	}
	defer srvCtx.Close()

	// 启动Dubbo-go RPC服务
	log.Println("Starting Account RPC Service...")
	server.StartServer(srvCtx)

	// 保持运行
	select {}
}
