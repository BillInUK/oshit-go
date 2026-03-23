package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
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
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// GetTokenQuoteUSDTPrice 获取token兑换USDT价格
func (h *PriceHandler) GetTokenQuoteUSDTPrice(fiberCtx *fiber.Ctx) error {
	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetTokenQuoteUSDTPrice()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// GetUSDTQuoteSOLPrice 获取USDT兑换SOL价格
func (h *PriceHandler) GetUSDTQuoteSOLPrice(fiberCtx *fiber.Ctx) error {
	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetUSDTQuoteSOLPrice()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// GetBirdEyePrice 获取BirdEye价格数据
func (h *PriceHandler) GetBirdEyePrice(fiberCtx *fiber.Ctx) error {
	var req types.GetBirdEyePriceReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	pl := logic.NewPriceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := pl.GetBirdEyePrice(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}
