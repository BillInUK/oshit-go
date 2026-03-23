package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
)

type ServiceHandler struct {
	srvCtx *svc.ServiceContext
}

func NewServiceHandler(svc *svc.ServiceContext) *ServiceHandler {
	return &ServiceHandler{svc}
}

// RegisterService 注册服务
func (h *ServiceHandler) RegisterService(fiberCtx *fiber.Ctx) error {
	var req types.ServiceRegisterReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	sl := logic.NewServiceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := sl.RegisterService(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// UpdateService 更新服务
func (h *ServiceHandler) UpdateService(fiberCtx *fiber.Ctx) error {
	var req types.ServiceUpdateReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	sl := logic.NewServiceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := sl.UpdateService(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// QueryService 查询服务
func (h *ServiceHandler) QueryService(fiberCtx *fiber.Ctx) error {
	var req types.ServiceQueryReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	sl := logic.NewServiceLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := sl.QueryService(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}
