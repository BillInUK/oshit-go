package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/reward/api/internal/svc"
)

type LotteryHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewLotteryHandler(srvCtx *svc.ServiceContext) *LotteryHandler {
	return &LotteryHandler{
		prefix: "Lottery业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h LotteryHandler) GetStatus(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h LotteryHandler) GetUnClaimedRecord(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h LotteryHandler) ExecuteLottery(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h LotteryHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h LotteryHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h LotteryHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	return nil
}
