package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/order/api/internal/logic"
	"oshit-go/app/order/api/internal/svc"
)

type OrderHandler struct {
	srvCtx *svc.ServiceContext
}

func NewOrderHandler(srvCtx *svc.ServiceContext) *OrderHandler {
	return &OrderHandler{srvCtx}
}

func (auth *OrderHandler) FetchUserOrders(fiberCtx *fiber.Ctx) error {
	ol := logic.NewOrderLogic(context.Background(), auth.srvCtx)
	resp, _ := ol.FetchUserOrders(nil)
	return fiberCtx.JSON(resp)
}
