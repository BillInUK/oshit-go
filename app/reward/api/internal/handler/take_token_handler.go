package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/response"
)

type TakeTokenHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewTakeTokenHandler(srvCtx *svc.ServiceContext) *TakeTokenHandler {
	return &TakeTokenHandler{
		prefix: "TakeToken业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h *TakeTokenHandler) GetConfig(fiberCtx *fiber.Ctx) error {
	l := logic.NewTakeLogic(fiberCtx.Context(), h.srvCtx)
	config, err := l.GetConfig()
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get take token config error")
	}
	return response.OkWithData(fiberCtx, config)
}

func (h *TakeTokenHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	var req types.GetTakeTokenTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化参数错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}

	l := logic.NewTakeLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.ProcessGetTxInfo(fiberCtx.Context(), req)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get transaction info error")
	}

	return response.OkWithData(fiberCtx, txInfo)
}

func (h *TakeTokenHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}

	l := logic.NewTakeLogic(fiberCtx.Context(), h.srvCtx)
	record, err := l.GetRecordByTxId(req.TxId)
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get record error")
	}
	return response.OkWithData(fiberCtx, record)
}

func (h *TakeTokenHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 处理钱包提交交易请求 -", h.prefix)
	ctx := fiberCtx.Context()

	// 反序列化请求
	var req types.CommitTakeTokenTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	encodedTx := req.EncodedTx
	inviteCode := req.InviteCode

	// 初步检查提交上来的交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(encodedTx)
	if err != nil {
		log.Errorf("%s - 检查打包的交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error: %v", err)
	}

	log.Infof("%s 地址 %v 使用邀请码 %s 领取奖励", prefix, preCheckedTx.From, inviteCode)

	// 处理交易主逻辑
	l := logic.NewTakeLogic(ctx, h.srvCtx)
	rsp, err := l.ProcessCommitTx(ctx, preCheckedTx, inviteCode)
	if err != nil {
		return response.FailWithError(fiberCtx, "process commit tx error", err)
	}
	return response.OkWithData(fiberCtx, rsp)
}
