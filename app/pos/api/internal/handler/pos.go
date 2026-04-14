package handler

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/pos/api/internal/logic/pos"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	app_utils "oshit-go/app/utils"
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

// TakeSnapShot 手动快照
func (h *PosHandler) TakeSnapShot(fiberCtx *fiber.Ctx) error {
	if h.srvCtx.SystemConfig.Env == 0 {
		return response.FailWithMsg(fiberCtx, "can not take snap shot manually on mainnet")
	}

	l := pos.NewPosSnapShotLogic(fiberCtx.Context(), h.srvCtx)
	if err := l.TakeStakeSnapShot(); err != nil {
		return response.FailWithMsg(fiberCtx, "take snapshot error")
	}
	return response.Ok(fiberCtx)
}

// ResetSnapShot 手动重置快照
func (h *PosHandler) ResetSnapShot(fiberCtx *fiber.Ctx) error {
	if h.srvCtx.SystemConfig.Env == 0 {
		return response.FailWithMsg(fiberCtx, "can not reset snap shot manually on mainnet")
	}
	l := pos.NewPosSnapShotLogic(fiberCtx.Context(), h.srvCtx)
	if err := l.ResetStakeSnapShot(); err != nil {
		return response.FailWithMsg(fiberCtx, "reset snapshot error")
	}
	return response.Ok(fiberCtx)
}

// GetConfig 获取下发奖励配置
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
	nativeAccount := fiberCtx.Locals("nativeAccount").(string)

	l := pos.NewPosRewardLogic(fiberCtx.Context(), h.srvCtx)
	detail, err := l.GetRewardStat(nativeAccount)
	if err != nil {
		log.Errorf("%s 查询奖励明细错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "query sol pos rewards error")
	}

	return response.OkWithData(fiberCtx, detail)
}

// GetRewardRecord 获取 pos 奖励明细
func (h *PosHandler) GetRewardRecord(fiberCtx *fiber.Ctx) error {
	nativeAccount := fiberCtx.Locals("nativeAccount").(string)
	// 获取固定奖励
	l := pos.NewPosRewardLogic(fiberCtx.Context(), h.srvCtx)
	records, err := l.GetRewards(nativeAccount)
	if err != nil {
		log.Errorf("%s 查询pos固定奖励错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "query sol pos rewards error")
	}

	return response.OkWithData(fiberCtx, records)
}

// GetTxInfo 获取领取pos奖励交易信息
func (h *PosHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	nativeAccount := fiberCtx.Locals("nativeAccount").(string)
	// 获取固定奖励
	l := pos.NewPosRewardLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetTxInfo(nativeAccount)
	if err != nil {
		log.Errorf("%s 获取领取奖励交易信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get transaction info error")
	}
	return response.OkWithData(fiberCtx, txInfo)
}

// CommitTx 提交领取pos奖励
func (h *PosHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 处理领取pos奖励提交请求 -", h.prefix)
	ctx := fiberCtx.Context()

	var req types.CommitPosRewardTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.EncodedTx == "" {
		return response.BadRequest(fiberCtx, "encodedTx is required")
	}

	// 预检查：hex 解码、反序列化、校验用户签名
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预检查交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error", err)
	}

	prefix = fmt.Sprintf("%s 发起地址 %v txId %v", prefix, preCheckedTx.From, preCheckedTx.TxId)
	log.Infof("%s 提交pos奖励领取交易", prefix)

	l := pos.NewPosRewardLogic(ctx, h.srvCtx)
	txId, err := l.ProcessCommitTx(ctx, preCheckedTx)
	if err != nil {
		log.Errorf("%s 处理交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "process commit stake reward tx error", err)
	}
	return response.OkWithData(fiberCtx, txId)
}

// GetClaimRecord 根据交易Id 获取 pos 奖励领取记录
func (h *PosHandler) GetClaimRecord(fiberCtx *fiber.Ctx) error {
	return nil
}

func (h *PosHandler) GetGroupInfo(fiberCtx *fiber.Ctx) error {
	nativeAccount := fiberCtx.Locals("nativeAccount").(string)

	l := pos.NewPosRewardLogic(fiberCtx.Context(), h.srvCtx)
	groupInfo, err := l.GetGroupInfo(nativeAccount)
	if err != nil {
		log.Errorf("%s - 根据地址信息查看团队信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "query group info in snap shot error")
	}
	return response.OkWithData(fiberCtx, groupInfo)
}
