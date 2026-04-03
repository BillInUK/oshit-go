package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic/lottery"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/response"
	"oshit-go/common/utils"
)

type LotteryHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewLotteryHandler(srvCtx *svc.ServiceContext) *LotteryHandler {
	return &LotteryHandler{
		prefix: "Lottery业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

// GetStatus 查询今日抽奖状态
func (h LotteryHandler) GetStatus(fiberCtx *fiber.Ctx) error {
	claims, err := utils.ExtractTokenMetadata(fiberCtx)
	if err != nil {
		return response.UnAuthorizedError(fiberCtx, "unauthorized")
	}
	nativeAccountVal, ok := claims.Credentials["account"]
	if !ok {
		return response.UnAuthorizedError(fiberCtx, "missing account in token claims")
	}
	nativeAccountString, ok := nativeAccountVal.(string)
	if !ok || nativeAccountString == "" {
		return response.UnAuthorizedError(fiberCtx, "invalid account in token claims")
	}
	// 获取当日的领取token的次数
	l := lottery.NewLotteryLogic(fiberCtx.Context(), h.srvCtx)
	stats, err := l.GetStatus(nativeAccountString)
	if err != nil {
		log.Errorf("%s 查询抽奖状态错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get lottery status error")
	}
	return response.OkWithData(fiberCtx, stats)
}

// GetUnClaimedRecord 查询未领取的抽奖奖励
func (h LotteryHandler) GetUnClaimedRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetUnclaimedLotteryReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化参数错误: %v", h.prefix, err)
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.NativeAccount == "" {
		return response.BadRequest(fiberCtx, "nativeAccount is required")
	}

	l := lottery.NewLotteryLogic(fiberCtx.Context(), h.srvCtx)
	rewards, err := l.GetUnclaimedRewards(req.NativeAccount)
	if err != nil {
		log.Errorf("%s 查询未领取奖励错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get unclaimed lottery rewards error")
	}
	return response.OkWithData(fiberCtx, rewards)
}

// ExecuteLottery 执行抽奖
func (h LotteryHandler) ExecuteLottery(fiberCtx *fiber.Ctx) error {
	// 从 JWT 中提取 native account
	claims, err := utils.ExtractTokenMetadata(fiberCtx)
	if err != nil {
		return response.UnAuthorizedError(fiberCtx, "unauthorized")
	}
	nativeAccountVal, ok := claims.Credentials["account"]
	if !ok {
		return response.UnAuthorizedError(fiberCtx, "missing account in token claims")
	}
	nativeAccount, ok := nativeAccountVal.(string)
	if !ok || nativeAccount == "" {
		return response.UnAuthorizedError(fiberCtx, "invalid account in token claims")
	}

	log.Infof("%s 地址 %v 请求执行抽奖", h.prefix, nativeAccount)

	l := lottery.NewLotteryLogic(fiberCtx.Context(), h.srvCtx)
	reward, err := l.ExecuteLottery(nativeAccount)
	if err != nil {
		log.Errorf("%s 执行抽奖错误: %v", h.prefix, err)
		return response.FailWithError(fiberCtx, "execute lottery error", err)
	}
	return response.OkWithData(fiberCtx, reward)
}

// GetRecord 根据 txId 查询抽奖领取记录
func (h LotteryHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.TxId == "" {
		return response.BadRequest(fiberCtx, "txId is required")
	}

	var record model.LotteryClaimRecord
	if err := h.srvCtx.DB.Table(model.TableNameLotteryClaimRecord).
		Where("tx_id = ?", req.TxId).
		First(&record).Error; err != nil {
		log.Errorf("%s 查询领取记录错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get lottery claim record error")
	}
	return response.OkWithData(fiberCtx, record)
}

// GetTxInfo 获取抽奖领取的交易信息
func (h LotteryHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	var req types.GetLotteryTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化参数错误: %v", h.prefix, err)
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.RecordId == "" {
		return response.BadRequest(fiberCtx, "recordId is required")
	}

	l := lottery.NewLotteryLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetTxInfo(fiberCtx.Context(), req.RecordId)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", h.prefix, err)
		return response.FailWithError(fiberCtx, "get lottery tx info error", err)
	}
	return response.OkWithData(fiberCtx, txInfo)
}

// CommitTx 提交抽奖领取交易
func (h LotteryHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 处理钱包提交抽奖领取交易请求 -", h.prefix)
	ctx := fiberCtx.Context()

	var req types.CommitLotteryTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.EncodedTx == "" {
		return response.BadRequest(fiberCtx, "encodedTx is required")
	}
	if req.RewardId == "" {
		return response.BadRequest(fiberCtx, "rewardId is required")
	}

	// 初步检查提交上来的交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 检查打包的交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error", err)
	}

	log.Infof("%s 地址 %v 提交抽奖领取交易 rewardId %s", prefix, preCheckedTx.From, req.RewardId)

	l := lottery.NewLotteryLogic(ctx, h.srvCtx)
	txId, err := l.ProcessCommitTx(ctx, preCheckedTx, req.RewardId)
	if err != nil {
		return response.FailWithError(fiberCtx, "process commit lottery tx error", err)
	}
	return response.OkWithData(fiberCtx, txId)
}
