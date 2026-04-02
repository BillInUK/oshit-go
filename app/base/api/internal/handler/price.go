package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/pkg/response"
)

type PriceHandler struct {
	srvCtx *svc.ServiceContext
}

func NewPriceHandler(svc *svc.ServiceContext) *PriceHandler {
	return &PriceHandler{svc}
}

// GetTokenQuoteSOLPrice 获取token兑换SOL价格
func (h *PriceHandler) GetTokenQuoteSOLPrice(fiberCtx *fiber.Ctx) error {
	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetTokenQuoteSOLPrice()
	if err != nil {
		return response.ServerError(fiberCtx, "GetTokenQuoteSOLPrice failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}

// GetTokenQuoteUSDTPrice 获取token兑换USDT价格
func (h *PriceHandler) GetTokenQuoteUSDTPrice(fiberCtx *fiber.Ctx) error {
	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetTokenQuoteUSDTPrice()
	if err != nil {
		return response.ServerError(fiberCtx, "GetTokenQuoteUSDTPrice failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}

// GetUSDTQuoteSOLPrice 获取USDT兑换SOL价格
func (h *PriceHandler) GetUSDTQuoteSOLPrice(fiberCtx *fiber.Ctx) error {
	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetUSDTQuoteSOLPrice()
	if err != nil {
		return response.ServerError(fiberCtx, "GetUSDTQuoteSOLPrice failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}

// GetBirdEyePrice 获取BirdEye价格数据
func (h *PriceHandler) GetBirdEyePrice(fiberCtx *fiber.Ctx) error {
	var req types.GetBirdEyePriceReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "Invalid request body")
	}

	// 验证时间间隔参数
	validIntervals := map[string]bool{"1D": true, "1W": true, "1M": true}
	if req.Interval == "" {
		return response.BadRequest(fiberCtx, "interval must be 1D,1W,1M")
	}

	if !validIntervals[req.Interval] {
		return response.BadRequest(fiberCtx, "interval must be 1D,1W,1M")
	}

	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetBirdEyePrice(&req)
	if err != nil {
		return response.ServerError(fiberCtx, "GetBirdEyePrice failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}
