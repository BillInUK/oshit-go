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
func (h *LotteryHandler) GetStatus(fiberCtx *fiber.Ctx) error {
	nativeAccountString := fiberCtx.Locals("nativeAccount").(string)
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
func (h *LotteryHandler) GetUnClaimedRecord(fiberCtx *fiber.Ctx) error {
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
func (h *LotteryHandler) ExecuteLottery(fiberCtx *fiber.Ctx) error {
	nativeAccount := fiberCtx.Locals("nativeAccount").(string)

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
func (h *LotteryHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.TxId == "" {
		return response.BadRequest(fiberCtx, "txId is required")
	}

	var record model.LotteryClaim
	if err := h.srvCtx.DB.Table(model.TableNameLotteryClaim).
		Where("tx_id = ?", req.TxId).
		First(&record).Error; err != nil {
		log.Errorf("%s 查询领取记录错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get lottery claim record error")
	}
	return response.OkWithData(fiberCtx, record)
}

// GetTxInfo 获取抽奖领取的交易信息
func (h *LotteryHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
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
func (h *LotteryHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 处理钱包提交交易请求 -", h.prefix)
	ctx := fiberCtx.Context()

	// 1. 发序列化请求
	var req types.CommitLotteryTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}

	// 2. 检查交易参数
	if req.EncodedTx == "" {
		return response.BadRequest(fiberCtx, "encodedTx is required")
	}
	if req.RewardId == "" {
		return response.BadRequest(fiberCtx, "rewardId is required")
	}

	// 3. 预检查提交上来的交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预检查打包的交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error", err)
	}

	// 4. 打印交易id以及业务信息
	prefix = fmt.Sprintf("%s 业务发起地址 %v 交易id %v", prefix, preCheckedTx.From, preCheckedTx.TxId)
	log.Infof("%s 提交抽奖领取交易 rewardId %s", prefix, req.RewardId)

	// 5. 处理交易主体逻辑，同步等待链上确认
	l := lottery.NewLotteryLogic(ctx, h.srvCtx)
	txState, err := l.ProcessCommitTx(ctx, preCheckedTx, req.RewardId)
	if err != nil {
		log.Errorf("%s 处理交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "process commit lottery tx error", err)
	}

	// 6. 返回链上确认结果给前端
	return response.OkWithData(fiberCtx, types.CommitTxResult{
		TxId:    preCheckedTx.TxId.String(),
		TxState: txState,
	})
}
