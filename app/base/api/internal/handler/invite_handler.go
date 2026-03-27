package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
)

type InviteHandler struct {
	srvCtx *svc.ServiceContext
}

func NewInviteHandler(svc *svc.ServiceContext) *InviteHandler {
	return &InviteHandler{svc}
}

// GetAccountByInviteCode 根据邀请码查询账户信息
func (h *InviteHandler) GetAccountByInviteCode(fiberCtx *fiber.Ctx) error {
	var req types.GetAccountByInviteCodeReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	il := logic.NewInviteLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := il.GetAccountByInviteCode(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// CheckInviteRecord 检查是否存在邀请记录
func (h *InviteHandler) CheckInviteRecord(fiberCtx *fiber.Ctx) error {
	var req types.CheckInviteRecordReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	il := logic.NewInviteLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := il.CheckInviteRecord(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// GetUpInviterRecords 递归查询上级邀请人
func (h *InviteHandler) GetUpInviterRecords(fiberCtx *fiber.Ctx) error {
	var req types.RecursiveQueryReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	il := logic.NewInviteLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := il.GetUpInviterRecords(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}

// GetDownInviteeRecords 递归查询下级被邀请人
func (h *InviteHandler) GetDownInviteeRecords(fiberCtx *fiber.Ctx) error {
	var req types.RecursiveQueryReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return fiberCtx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	il := logic.NewInviteLogic(fiberCtx.Context(), h.srvCtx)
	resp, err := il.GetDownInviteeRecords(&req)
	if err != nil {
		return fiberCtx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return fiberCtx.JSON(resp)
}
