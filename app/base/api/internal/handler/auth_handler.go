package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/pkg/response"
)

type AuthHandler struct {
	srvCtx *svc.ServiceContext
}

func NewAuthHandler(svc *svc.ServiceContext) *AuthHandler {
	return &AuthHandler{svc}
}

func (auth *AuthHandler) Login(fiberCtx *fiber.Ctx) error {
	var req types.LoginReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "Invalid request body")
	}

	al := logic.NewAuthLogic(fiberCtx.Context(), auth.srvCtx)
	resp, err := al.Login(&req)
	if err != nil {
		return response.ServerError(fiberCtx, "Login failed", err)
	}

	return response.OkWithData(fiberCtx, resp)
}

func (auth *AuthHandler) QueryNativeAccountInfo(fiberCtx *fiber.Ctx) error {
	// TODO: 实现JWT验证和查询逻辑
	return response.FailWithStatus(fiberCtx, fiber.StatusNotImplemented)
}
