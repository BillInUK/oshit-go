package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
)

type InfoHandler struct {
	srvCtx *svc.ServiceContext
}

func NewInfoHandler(svc *svc.ServiceContext) *InfoHandler {
	return &InfoHandler{svc}
}

// GetTokenInfo 获取token信息
func (h *InfoHandler) GetTokenInfo(fiberCtx *fiber.Ctx) error {
	cl := logic.NewInfoLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := cl.GetTokenInfo()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return fiberCtx.JSON(resp)
}

// GetTokenHolders 获取token持有者数量
func (h *InfoHandler) GetTokenHolders(fiberCtx *fiber.Ctx) error {
	il := logic.NewInfoLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := il.GetTokenHolders()
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}
