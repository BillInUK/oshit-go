package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
)

type ConfigHandler struct {
	srvCtx *svc.ServiceContext
}

func NewConfigHandler(svc *svc.ServiceContext) *ConfigHandler {
	return &ConfigHandler{svc}
}

// GetTokenInfo 获取token信息
func (h *ConfigHandler) GetTokenInfo(fiberCtx *fiber.Ctx) error {
	cl := logic.NewConfigLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := cl.GetTokenInfo()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return fiberCtx.JSON(resp)
}

// GetFeeTolerance 获取手续费容错
func (h *ConfigHandler) GetFeeTolerance(fiberCtx *fiber.Ctx) error {
	cl := logic.NewConfigLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := cl.GetFeeTolerance()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}
