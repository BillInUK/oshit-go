package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/account/api/internal/logic"
	"oshit-go/app/account/api/internal/svc"
	"oshit-go/app/account/api/internal/types"
)

type AuthHandler struct {
	srvCtx *svc.ServiceContext
}

func NewAuthHandler(svc *svc.ServiceContext) *AuthHandler {
	return &AuthHandler{svc}
}

func (auth *AuthHandler) Login(fiberCtx *fiber.Ctx) error {
	var req types.LoginRequest
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	al := logic.NewAuthLogic(fiberCtx.Context(), auth.srvCtx)
	resp, err := al.Login(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}
