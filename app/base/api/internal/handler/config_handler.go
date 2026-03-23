package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
)

type ConfigHandler struct {
	srvCtx *svc.ServiceContext
}

func NewConfigHandler(svc *svc.ServiceContext) *ConfigHandler {
	return &ConfigHandler{svc}
}

// GetTokenInfo 获取token信息
func (h *ConfigHandler) GetTokenInfo(fiberCtx *fiber.Ctx) error {
	var req types.GetTokenInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	cl := logic.NewConfigLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := cl.GetTokenInfo(&req)
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
