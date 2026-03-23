package handler

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	"oshit-go/common/pkg/response"
)

type TradeHandler struct {
	srvCtx *svc.ServiceContext
}

func NewTradeHandler(srvCtx *svc.ServiceContext) *TradeHandler {
	return &TradeHandler{srvCtx}
}

func (auth *TradeHandler) FetchUserOrders(fiberCtx *fiber.Ctx) error {
	prefix := "交易模块 - 查询用户订单信息 -"
	var req types.UserOrdersReq
	if err := json.Unmarshal(fiberCtx.Body(), &req); err != nil {
		log.Errorf("%s failed to parse message: %v", prefix, err)
		return response.FailWithMsg(fiberCtx, "failed to parse message")
	}

	ol := logic.NewOrderLogic(context.Background(), auth.srvCtx)
	resp, _ := ol.FetchUserOrders(&req)
	return fiberCtx.JSON(resp)
}
