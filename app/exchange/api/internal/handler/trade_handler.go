package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/exchange/api/internal/logic"
	"oshit-go/app/exchange/api/internal/svc"
)

type TradeHandler struct {
	srvCtx *svc.ServiceContext
}

func NewTradeHandler(srvCtx *svc.ServiceContext) *TradeHandler {
	return &TradeHandler{srvCtx}
}

func (auth *TradeHandler) FetchUserOrders(fiberCtx *fiber.Ctx) error {
	ol := logic.NewOrderLogic(context.Background(), auth.srvCtx)
	resp, _ := ol.FetchUserOrders(nil)
	return fiberCtx.JSON(resp)
}
