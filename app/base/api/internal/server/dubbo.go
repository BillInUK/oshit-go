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
	nacosCfg := svcCtx.Config.Nacos
	port := nacosCfg.Port
	if port == 0 {
		port = 8848
	}
	nacosAddr := fmt.Sprintf("%s:%d", nacosCfg.Host, port)

	registryOpts := []registry.Option{
		registry.WithNacos(),
		registry.WithAddress(nacosAddr),
		registry.WithRegisterInterface(),
		registry.WithoutUseAsMetaReport(),
		registry.WithoutUseAsConfigCenter(),
	}
	if nacosCfg.Username != "" {
		registryOpts = append(registryOpts, registry.WithUsername(nacosCfg.Username))
		registryOpts = append(registryOpts, registry.WithPassword(nacosCfg.Password))
	}

	dubboPort := svcCtx.Config.Dubbo.Port
	if dubboPort == 0 {
		dubboPort = 20880
	}

	ins, err := dubbo.NewInstance(
		dubbo.WithName(svcCtx.Config.App.Name),
		dubbo.WithRegistry(registryOpts...),
		dubbo.WithProtocol(
			protocol.WithTriple(),
			protocol.WithPort(dubboPort),
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

	log.Infof("dubbo: BaseService listening on :%d (nacos=%s)", dubboPort, nacosAddr)
	if err := srv.Serve(); err != nil {
		if strings.Contains(err.Error(), "client not connected") {
			log.Errorf("dubbo: Nacos not reachable at %s", nacosAddr)
		} else {
			log.Errorf("dubbo: serve error: %v", err)
		}
	}
}
