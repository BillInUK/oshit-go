package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/pos/api/internal/svc"
)

type StakeHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewStakeHandler(srvCtx *svc.ServiceContext) *StakeHandler {
	return &StakeHandler{
		prefix: "Stake业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h *StakeHandler) GetConfig(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *StakeHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *StakeHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *StakeHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	return nil
}
