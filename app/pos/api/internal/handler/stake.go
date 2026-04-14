package handler

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"oshit-go/app/pos/api/internal/logic/stake"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/response"
	"strconv"
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

// GetRewardStat 获取质押奖励信息
func (h *StakeHandler) GetRewardStat(fiberCtx *fiber.Ctx) error {
	return nil
}

// GetStarLevel 获取质押星级
func (h *StakeHandler) GetStarLevel(fiberCtx *fiber.Ctx) error {
	return nil
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
	nativeAccountString := fiberCtx.Locals("nativeAccount").(string)

	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}
	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)

	record, err := l.GetRewardRecord(nativeAccountString)
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get stake reward record error")
	}
	return response.OkWithData(fiberCtx, record)
}

// GetTxInfo 获取领取 stake 奖励的交易参数
func (h *StakeHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	nativeAccountString := fiberCtx.Locals("nativeAccount").(string)

	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetTxInfo(nativeAccountString)
	if err != nil {
		return response.FailWithMsg(fiberCtx, "get claim stake reward transaction info error")
	}

	return response.OkWithData(fiberCtx, txInfo)
}

// CommitTx 提交领取 stake 奖励的交易
func (h *StakeHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 处理领取stake奖励提交请求 -", h.prefix)
	ctx := fiberCtx.Context()

	var req types.CommitStakeRewardTxReq
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
	log.Infof("%s 提交stake奖励领取交易", prefix)

	l := stake.NewStakeRewardLogic(ctx, h.srvCtx)
	txId, err := l.ProcessCommitTx(ctx, preCheckedTx)
	if err != nil {
		log.Errorf("%s 处理交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "process commit stake reward tx error", err)
	}
	return response.OkWithData(fiberCtx, txId)
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
	if err := l.ResetStakeSnapShot(); err != nil {
		return response.FailWithMsg(fiberCtx, "reset snapshot error")
	}
	return response.Ok(fiberCtx)
}

// MinAmount 获取最小质押奖励
func (h *StakeHandler) MinAmount(fiberCtx *fiber.Ctx) error {
	stakeTypeStr := fiberCtx.Params("type")
	stakeType, err := strconv.ParseInt(stakeTypeStr, 10, 64)
	if err != nil || (stakeType != 0 && stakeType != 1) {
		return response.FailWithMsg(fiberCtx, "stake type must 0 or 1")
	}
	l := stake.NewStakeLogic(fiberCtx.Context(), h.srvCtx)
	minStakeAmount := l.GetStakeMinAmount(stakeType)
	return response.OkWithData(fiberCtx, minStakeAmount)
}

// StakeToken 提交质押token交易
func (h *StakeHandler) StakeToken(fiberCtx *fiber.Ctx) error {
	var err error
	var txId *solana.Signature
	var prefix = fmt.Sprintf("Stake业务 - 质押token -")

	// 1. 检查交易参数
	ctx := fiberCtx.Context()
	var req types.EncodedTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化请求错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "invalid request")
	}

	// 2. 预处理交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预处理交易错误: %v", prefix, err)
		return response.FailWithMsg(fiberCtx, "pre check transaction error")
	}

	log.Infof("%s 质押地址 %v, 交易id: %v", prefix, preCheckedTx.From, preCheckedTx.TxId)

	// 3. 处理交易主逻辑
	l := stake.NewStakeLogic(ctx, h.srvCtx)
	if err = l.ProcessStakeToken(ctx, preCheckedTx); err != nil {
		log.Errorf("%s 处理地址 %v 交易id %v 错误: %v", prefix, preCheckedTx.From, preCheckedTx.TxId, err)
		return response.FailWithError(fiberCtx, "process transaction error:", err)
	}

	return response.OkWithData(fiberCtx, txId)
}

// UnStakeToken 解除质押
func (h *StakeHandler) UnStakeToken(fiberCtx *fiber.Ctx) error {
	var err error
	var txId *solana.Signature
	var prefix = fmt.Sprintf("Stake业务 - 解除质押token -")

	// 1. 检查交易参数
	ctx := fiberCtx.Context()
	var req types.EncodedTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化请求错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "invalid request")
	}

	// 2. 预处理交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预处理交易错误: %v", prefix, err)
		return response.FailWithMsg(fiberCtx, "pre check transaction error")
	}

	log.Infof("%s 质押地址 %v, 交易id: %v", prefix, preCheckedTx.From, preCheckedTx.TxId)

	// 3. 处理交易主逻辑
	l := stake.NewStakeLogic(ctx, h.srvCtx)
	if err = l.ProcessUnStakeToken(ctx, preCheckedTx); err != nil {
		log.Errorf("%s 处理地址 %v 交易id %v 错误: %v", prefix, preCheckedTx.From, preCheckedTx.TxId, err)
		return response.FailWithError(fiberCtx, "process transaction error:", err)
	}

	return response.OkWithData(fiberCtx, txId)
}

// ReStakeToken 重新质押
func (h *StakeHandler) ReStakeToken(fiberCtx *fiber.Ctx) error {
	var err error
	var txId *solana.Signature
	var prefix = fmt.Sprintf("Stake业务 - 重新质押token -")

	// 1. 检查交易参数
	ctx := fiberCtx.Context()
	var req types.EncodedTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 反序列化请求错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "invalid request")
	}
	// 2. 预处理交易
	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预处理交易错误: %v", prefix, err)
		return response.FailWithMsg(fiberCtx, "pre check transaction error")
	}

	log.Infof("%s 质押地址 %v, 交易id: %v", prefix, preCheckedTx.From, preCheckedTx.TxId)

	// 3. 处理交易主逻辑
	l := stake.NewStakeLogic(ctx, h.srvCtx)
	if err = l.ProcessReStakeToken(ctx, preCheckedTx); err != nil {
		log.Errorf("%s 处理地址 %v 交易id %v 错误: %v", prefix, preCheckedTx.From, preCheckedTx.TxId, err)
		return response.FailWithError(fiberCtx, "process transaction error:", err)
	}

	return response.OkWithData(fiberCtx, txId)
}

// GetLeaderInfo 获取当前登录用户的区域经理信息
func (h *StakeHandler) GetLeaderInfo(fiberCtx *fiber.Ctx) error {
	nativeAccountString := fiberCtx.Locals("nativeAccount").(string)

	l := stake.NewStakeLogic(fiberCtx.Context(), h.srvCtx)
	leader, err := l.GetLeaderInfo(nativeAccountString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.OkWithData(fiberCtx, nil)
		}
		return response.FailWithMsg(fiberCtx, "get leader info error")
	}
	return response.OkWithData(fiberCtx, leader)
}

// GetLeaderRewardRecord 获取区域经理奖励发放记录
func (h *StakeHandler) GetLeaderRewardRecord(fiberCtx *fiber.Ctx) error {
	nativeAccountString := fiberCtx.Locals("nativeAccount").(string)
	nativeAccount, err := solana.PublicKeyFromBase58(nativeAccountString)
	if err != nil {
		return response.FailWithError(fiberCtx, "malformed native account", err)
	}

	l := stake.NewStakeLogic(fiberCtx.Context(), h.srvCtx)
	records, err := l.GetLeaderRewards(nativeAccount.String())
	if err != nil {
		log.Errorf("Stake业务 - 查询区域经理未领取奖励错误: %v", err)
		return response.FailWithMsg(fiberCtx, "query area leader rewards error")
	}
	return response.OkWithData(fiberCtx, records)
}

// GetLeaderTxInfo 获取区域经理奖励信息
func (h *StakeHandler) GetLeaderTxInfo(fiberCtx *fiber.Ctx) error {
	nativeAccountString := fiberCtx.Locals("nativeAccount").(string)
	nativeAccount, err := solana.PublicKeyFromBase58(nativeAccountString)
	if err != nil {
		return response.FailWithError(fiberCtx, "malformed native account", err)
	}

	l := stake.NewStakeLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetLeaderTxInfo(nativeAccount)
	if err != nil {
		log.Errorf("%s - 获取区域经理奖励交易信息错误: %v", h.prefix, err)
		return response.FailWithMsg(fiberCtx, "get claim area leader reward tx info failed")
	}
	return response.OkWithData(fiberCtx, txInfo)
}

// GetLeaderClaimRecord 获取区域经理奖励领取记录
func (h *StakeHandler) GetLeaderClaimRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.FailWithMsg(fiberCtx, "invalid request body")
	}
	l := stake.NewStakeRewardLogic(fiberCtx.Context(), h.srvCtx)
	record, err := l.GetLeaderClaimRecord(req.TxId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.OkWithData(fiberCtx, nil)
		}
		return response.FailWithMsg(fiberCtx, "get leader reward claim record error")
	}
	return response.OkWithData(fiberCtx, record)
}

// LeaderCommitTx 区域经理领取奖励
func (h *StakeHandler) LeaderCommitTx(fiberCtx *fiber.Ctx) error {
	prefix := fmt.Sprintf("%s 处理区域经理领取奖励提交请求 -", h.prefix)
	ctx := fiberCtx.Context()

	var req types.CommitStakeRewardTxReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if req.EncodedTx == "" {
		return response.BadRequest(fiberCtx, "encodedTx is required")
	}

	preCheckedTx, err := app_utils.PreCheckEncodedTx(req.EncodedTx)
	if err != nil {
		log.Errorf("%s 预检查交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "pre check encoded transaction error", err)
	}

	prefix = fmt.Sprintf("%s 发起地址 %v txId %v", prefix, preCheckedTx.From, preCheckedTx.TxId)
	log.Infof("%s 提交区域经理领取奖励交易", prefix)

	l := stake.NewStakeRewardLogic(ctx, h.srvCtx)
	txId, err := l.ProcessLeaderCommitTx(ctx, preCheckedTx)
	if err != nil {
		log.Errorf("%s 处理交易错误: %v", prefix, err)
		return response.FailWithError(fiberCtx, "process leader commit tx error", err)
	}
	return response.OkWithData(fiberCtx, txId)
}
