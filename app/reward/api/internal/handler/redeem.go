package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/reward/api/internal/svc"
)

type RedeemHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewRedeemHandler(srvCtx *svc.ServiceContext) *RedeemHandler {
	return &RedeemHandler{
		prefix: "奖励码业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h *RedeemHandler) GetRedeemRewardByTxId(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *RedeemHandler) GetRedeemTxInfo(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *RedeemHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	return nil
}
