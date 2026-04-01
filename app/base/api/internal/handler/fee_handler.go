package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/common/pkg/response"
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
		return response.ServerError(fiberCtx, "GetPriorityFee failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}

// GetInstUnits 获取计算单元消耗
func (h *FeeHandler) GetInstUnits(fiberCtx *fiber.Ctx) error {
	fl := logic.NewFeeLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := fl.GetInstUnits()
	if err != nil {
		return response.ServerError(fiberCtx, "GetInstUnits failed", err)
	}
	return response.OkWithData(fiberCtx, resp)
}

// GetFeeTolerance 获取手续费容错
func (h *FeeHandler) GetFeeTolerance(fiberCtx *fiber.Ctx) error {
	cl := logic.NewFeeLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := cl.GetFeeTolerance()
	if err != nil {
		return response.ServerError(fiberCtx, "GetFeeTolerance failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}
