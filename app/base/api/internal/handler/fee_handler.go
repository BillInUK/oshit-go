package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
)

type FeeHandler struct {
	srvCtx *svc.ServiceContext
}

func NewFeeHandler(svc *svc.ServiceContext) *FeeHandler {
	return &FeeHandler{svc}
}

// GetPriorityFee 获取优先手续费
func (h *FeeHandler) GetPriorityFee(fiberCtx *fiber.Ctx) error {
	fl := logic.NewFeeLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := fl.GetPriorityFee()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// GetInstUnits 获取计算单元消耗
func (h *FeeHandler) GetInstUnits(fiberCtx *fiber.Ctx) error {
	fl := logic.NewFeeLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := fl.GetInstUnits()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return fiberCtx.JSON(resp)
}
