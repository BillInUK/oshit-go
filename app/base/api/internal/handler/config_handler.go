package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/common/pkg/response"
)

type ConfigHandler struct {
	srvCtx *svc.ServiceContext
}

func NewConfigHandler(svc *svc.ServiceContext) *ConfigHandler {
	return &ConfigHandler{svc}
}

// GetFeeTolerance 获取手续费容错
func (h *ConfigHandler) GetFeeTolerance(fiberCtx *fiber.Ctx) error {
	cl := logic.NewConfigLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := cl.GetFeeTolerance()
	if err != nil {
		return response.ServerError(fiberCtx, "GetFeeTolerance failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}
