package server

import (
	"fmt"
	"strings"

	"dubbo.apache.org/dubbo-go/v3"
	_ "dubbo.apache.org/dubbo-go/v3/imports" // 注册所有 SPI（nacos registry、triple protocol 等）
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/registry"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/base/api/internal/svc"
	basepb "oshit-go/common/pkg/pb/base"
)

// StartDubboServer 启动 Dubbo Triple 服务，作为独立 goroutine 运行
func StartDubboServer(svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config.Dubbo
	nacosAddr := fmt.Sprintf("%s:%d", cfg.Nacos.Host, cfg.Nacos.Port)

	ins, err := dubbo.NewInstance(
		dubbo.WithName(cfg.Protocol.Name),
		dubbo.WithRegistry(
			registry.WithNacos(),
			registry.WithAddress(nacosAddr),
		),
		dubbo.WithProtocol(
			protocol.WithTriple(),
			protocol.WithPort(cfg.Protocol.Port),
		),
	)
	if err != nil {
		log.Errorf("dubbo: create instance failed: %v", err)
		return
	}

	srv, err := ins.NewServer()
	if err != nil {
		log.Errorf("dubbo: create server failed: %v", err)
		return
	}

	if err := basepb.RegisterBaseServiceHandler(srv, NewBaseRpcService(svcCtx)); err != nil {
		log.Errorf("dubbo: register BaseService failed: %v", err)
		return
	}

	log.Infof("dubbo: BaseService listening on :%d (nacos=%s)", cfg.Protocol.Port, nacosAddr)
	if err := srv.Serve(); err != nil {
		if strings.Contains(err.Error(), "client not connected") {
			log.Errorf("dubbo: Nacos not reachable at %s (gRPC port %d)", nacosAddr, cfg.Nacos.GrpcPort)
		} else {
			log.Errorf("dubbo: serve error: %v", err)
		}
	}
}
