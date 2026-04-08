package handler

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/app/pos/api/internal/logic/stake"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/pkg/response"
	"oshit-go/common/utils"
)

type StakeHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewStakeHandler(srvCtx *svc.ServiceContext) *StakeHandler {
	return &StakeHandler{
		prefix: "Stake业务 - 处理前端请求 -",
		srvCtx: srvCtx,
	}
}

// GetConfig 获取 stake 下发奖励配置
func (h *StakeHandler) GetConfig(fiberCtx *fiber.Ctx) error {
	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)
	config, err := l.GetConfig()
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get stake reward config error")
	}
	return response.OkWithData(fiberCtx, config)
}

// GetClaimRecord 根据交易id获取领取stake奖励记录
func (h *StakeHandler) GetClaimRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}
	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)
	record, err := l.GetClaimRecord(req.TxId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.OkWithData(fiberCtx, nil)
		}
		return response.FailWithMsg(fiberCtx, "get stake claim reward record by tx id error")
	}
	return response.OkWithData(fiberCtx, record)
}

// GetRewardRecord 获取当天的 stake 奖励记录
func (h *StakeHandler) GetRewardRecord(fiberCtx *fiber.Ctx) error {
	claims, err := utils.ExtractTokenMetadata(fiberCtx)
	if err != nil {
		return response.UnAuthorizedError(fiberCtx, "unauthorized")
	}
	nativeAccountString := claims.Credentials["nativeAccount"].(string)
	if _, err := solana.PublicKeyFromBase58(nativeAccountString); err != nil {
		return response.FailWithError(fiberCtx, "malformed native account", err)
	}

	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}
	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)

	record, err := l.GetRewardRecord(req.TxId)
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get stake reward record error")
	}
	return response.OkWithData(fiberCtx, record)
}

// GetTxInfo 获取领取 stake 奖励的交易参数
func (h *StakeHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	claims, err := utils.ExtractTokenMetadata(fiberCtx)
	if err != nil {
		return response.UnAuthorizedError(fiberCtx, "unauthorized")
	}
	nativeAccountString := claims.Credentials["nativeAccount"].(string)
	if _, err := solana.PublicKeyFromBase58(nativeAccountString); err != nil {
		return response.FailWithError(fiberCtx, "malformed native account", err)
	}

	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetTxInfo(nativeAccountString)
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get claim stake reward transacion info error")
	}

	return response.OkWithData(fiberCtx, txInfo)
}

// CommitTx 提交领取奖励的交易
func (h *StakeHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	return nil
}

// TakeSnapShot 手动快照
func (h *StakeHandler) TakeSnapShot(fiberCtx *fiber.Ctx) error {
	l := stake.NewStakeSnapShotLogic(fiberCtx.Context(), h.srvCtx)
	if err := l.TakeStakeSnapShot(); err != nil {
		return response.FailWithMsg(fiberCtx, "take snapshot error")
	}
	return response.Ok(fiberCtx)
}

// ResetSnapShot 手动消除快照
func (h *StakeHandler) ResetSnapShot(fiberCtx *fiber.Ctx) error {
	l := stake.NewStakeSnapShotLogic(fiberCtx.Context(), h.srvCtx)
	if err := l.TakeStakeSnapShot(); err != nil {
		return response.FailWithMsg(fiberCtx, "reset snapshot error")
	}
	return response.Ok(fiberCtx)
}
