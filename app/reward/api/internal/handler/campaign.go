package handler

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic/campaign"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/response"
)

type CampaignHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewCampaignHandler(srvCtx *svc.ServiceContext) *CampaignHandler {
	return &CampaignHandler{
		prefix: "Campaign兑换积分 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

// GetExchangeConfig 查看通过官方转token奖励规则
func (h *CampaignHandler) GetExchangeConfig(fiberCtx *fiber.Ctx) error {
	return response.OkWithData(fiberCtx, h.srvCtx.CampaignExchangeConfig)
}

// GetExchangeLimit 查询campaign用户兑换token额度和全局额度
func (h *CampaignHandler) GetExchangeLimit(fiberCtx *fiber.Ctx) error {
	xAcJwt := fiberCtx.Get("x-ac-jwt")
	if xAcJwt == "" {
		return response.FailWithMsg(fiberCtx, "x-ac-jwt token is required")
	}

	// 1. 获取campaign的 userId 是否存在
	userId, err := h.srvCtx.CampaignClientV1.LoginTokenToUserID(xAcJwt)
	if err != nil {
		log.Errorf("%s 根据x-ac-jwt获取用户id错误: %v", h.prefix, err)
		return response.UnAuthorizedError(fiberCtx, "can not get user id")
	}

	// 2. 获取用户额度和全局额度
	l := campaign.NewCampaignLogic(fiberCtx.Context(), h.srvCtx)
	quotaInfo, globalLimit, err := l.GetExchangeQuotaInfo(fiberCtx.Context(), userId)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, err.Error())
	}

	// 3. 在handler层处理返回参数
	// 3.1 隐藏userId字段
	// 3.2 不返回userQuota.ID
	responseData := fiber.Map{
		// 全局额度信息
		"global_daily_limit": globalLimit.DailyLimit,
		"global_quota_date":  globalLimit.QuotaDate,
		// 用户额度信息
		"quota_date":      quotaInfo.QuotaDate,
		"max_quota":       quotaInfo.MaxQuota,
		"frozen_quota":    quotaInfo.FrozenQuota,
		"available_quota": quotaInfo.AvailableQuota,
	}

	return response.OkWithData(fiberCtx, responseData)
}

// GetTxInfo 获取兑换积分交易信息
func (h *CampaignHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 获取交易信息 -")
	ctx := fiberCtx.Context()
	xAcJwt := fiberCtx.Get("x-ac-jwt")
	if xAcJwt == "" {
		return response.FailWithMsg(fiberCtx, "x-ac-jwt token is required")
	}

	// 1. 获取campaign的 userId 是否存在
	userId, err := h.srvCtx.CampaignClientV1.LoginTokenToUserID(xAcJwt)
	if err != nil {
		log.Errorf("%s 根据x-ac-jwt获取用户id错误: %v", h.prefix, err)
		return response.UnAuthorizedError(fiberCtx, "can not get user id")
	}

	// 2. 反序列化请求
	var req types.CampaignExchangeTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}

	// 3. 获取交易信息
	l := campaign.NewCampaignLogic(ctx, h.srvCtx)
	txInfo, err := l.GetTxInfo(ctx, userId, req.Score)
	if err != nil {
		log.Errorf("%s 用户id: %d 兑换积分额度: %d", prefix, userId, req.Score)
		return response.FailWithMsg(fiberCtx, "get transaction info error")
	}
	return response.OkWithData(fiberCtx, txInfo)
}

// CommitTx 通过官网领取token
func (h *CampaignHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	var err error
	var txId *solana.Signature
	prefix := fmt.Sprintf("%s 处理钱包提交交易请求 -", h.prefix)
	ctx := fiberCtx.Context()

	// 1. 反序列化请求
	var req types.CampaignExchangeReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.EncodedTx == "" {
		return response.BadRequest(fiberCtx, "encodedTx is required")
	}
	// 2. 检查要兑换的积分额度是否符合要求
	if req.Score != 1 && req.Score != 5 && req.Score != 10 {
		return response.FailWithMsg(fiberCtx, "score value must be 1,5,10")
	}
	// 3. 获取campaign的 userId 是否存在
	userId, err := h.srvCtx.CampaignClientV1.LoginTokenToUserID(req.XAcJwt)
	if err != nil {
		log.Errorf("%s 根据x-ac-jwt获取用户id错误: %v", prefix, err)
		return response.UnAuthorizedError(fiberCtx, "can not get user id")
	}
	// 4. 初步检查提交上来的交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 检查打包的交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error", err)
	}

	// 4. 打印交易id以及业务信息
	prefix = fmt.Sprintf("%s 业务发起地址 %v 交易id %v 用户id: %d", prefix, preCheckedTx.From, preCheckedTx.TxId, userId)
	log.Infof("%s 兑换积分额度 %s", prefix, req.Score)

	// 5. 处理交易主逻辑
	l := campaign.NewCampaignLogic(ctx, h.srvCtx)
	if err := l.ProcessCommitTx(ctx, preCheckedTx, req, userId); err != nil {
		log.Errorf("%s 处理交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "process transaction error:", err)
	}
	return response.OkWithData(fiberCtx, txId)
}
