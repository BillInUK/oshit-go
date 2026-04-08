package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/pos/api/internal/logic/pos"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/common/pkg/response"
)

type PosHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewPosHandler(srvCtx *svc.ServiceContext) *PosHandler {
	return &PosHandler{
		prefix: "Pos业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h *PosHandler) GetConfig(fiberCtx *fiber.Ctx) error {
	l := pos.NewPosLogic(fiberCtx.Context(), h.srvCtx)
	config, err := l.GetConfig()
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get pos reward config error")
	}
	return response.OkWithData(fiberCtx, config)
}

func (h *PosHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *PosHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *PosHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	return nil
}
