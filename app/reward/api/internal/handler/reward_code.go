package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic/rewardcode"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/response"
)

type RewardCodeHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewRewardCodeHandler(srvCtx *svc.ServiceContext) *RewardCodeHandler {
	return &RewardCodeHandler{
		prefix: "RewardCode业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

func (h *RewardCodeHandler) GetInfo(fiberCtx *fiber.Ctx) error {
	var req types.GetRewardCodeTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 查询奖励码信息 - 反序列化请求错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}

	l := rewardcode.NewRewardCodeLogic(fiberCtx.Context(), h.srvCtx)
	info, err := l.GetInfo(req.RewardCode)
	if err != nil {
		log.Errorf("%s 查询奖励码信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, err.Error())
	}

	return response.OkWithData(fiberCtx, info)
}

func (h *RewardCodeHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	var req types.GetRewardCodeTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 获取交易信息 - 反序列化请求错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}

	l := rewardcode.NewRewardCodeLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetTxInfo(fiberCtx.Context(), req.RewardCode)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, err.Error())
	}

	return response.OkWithData(fiberCtx, txInfo)
}

func (h *RewardCodeHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	ctx := fiberCtx.Context()
	prefix := fmt.Sprintf("%s 处理钱包提交交易请求 -", h.prefix)

	var req types.CommitRewardCodeTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化请求错误: %v", prefix, err)
		return response.BadRequest(fiberCtx, "invalid request body")
	}

	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预检查打包的交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error: %v", err)
	}

	prefix = fmt.Sprintf("%s 业务发起地址 %v 交易id %v rewardCode %s", prefix, preCheckedTx.From, preCheckedTx.TxId, req.RewardCode)
	log.Infof("%s 开始处理", prefix)

	l := rewardcode.NewRewardCodeLogic(ctx, h.srvCtx)
	if err := l.ProcessCommitTx(ctx, preCheckedTx, req.RewardCode); err != nil {
		log.Errorf("%s 处理交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "process commit tx error", err)
	}

	return response.OkWithData(fiberCtx, preCheckedTx.TxId)
}
