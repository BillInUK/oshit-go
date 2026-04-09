package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/pos/api/internal/logic/pos"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/common/pkg/response"
)

type PosHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewPosHandler(srvCtx *svc.ServiceContext) *PosHandler {
	return &PosHandler{
		prefix: "Pos业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h *PosHandler) GetConfig(fiberCtx *fiber.Ctx) error {
	l := pos.NewPosRewardLogic(fiberCtx.Context(), h.srvCtx)
	config, err := l.GetConfig()
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get pos reward config error")
	}
	return response.OkWithData(fiberCtx, config)
}

// GetRewardStat 获取 pos 奖励统计
func (h *PosHandler) GetRewardStat(fiberCtx *fiber.Ctx) error {
	return nil
}

// GetRewardRecord 获取 pos 奖励明细
func (h *PosHandler) GetRewardRecord(fiberCtx *fiber.Ctx) error {
	return nil
}

// GetTxInfo 获取领取pos奖励交易信息
func (h *PosHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	return nil
}

// CommitTx 提交领取pos奖励
func (h *PosHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	return nil
}

// GetClaimRecord 获取 pos 奖励领取记录
func (h *PosHandler) GetClaimRecord(fiberCtx *fiber.Ctx) error {
	return nil
}
