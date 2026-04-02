package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/common/pkg/response"
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
		return response.FailWithMsg(fiberCtx, err.Error())
	}
	return response.OkWithData(fiberCtx, resp)
}

// GetTokenHolders 获取token持有者数量
func (h *InfoHandler) GetTokenHolders(fiberCtx *fiber.Ctx) error {
	il := logic.NewInfoLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := il.GetTokenHolders()
	if err != nil {
		return response.ServerError(fiberCtx, "GetTokenHolders failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}
