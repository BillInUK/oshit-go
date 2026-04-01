package handler

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/reward/api/internal/logic"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/response"
	"oshit-go/common/utils"
)

type GiveTokenHandler struct {
	srvCtx *svc.ServiceContext
}

func NewGiveTokenHandler(srvCtx *svc.ServiceContext) *GiveTokenHandler {
	return &GiveTokenHandler{srvCtx}
}

// GetConfig 获取 give token 配置信息
func (h *GiveTokenHandler) GetConfig(fiberCtx *fiber.Ctx) error {
	l := logic.NewGiveTokenLogic(fiberCtx.Context(), h.srvCtx)
	rule, err := l.GetDefaultConfig()
	if err != nil {
		log.Errorf("获取take token配置错误: %v", err)
		return response.FailWithError(fiberCtx, "get give token config error:", err)
	}

	return response.OkWithData(fiberCtx, rule)
}

// GetRecord 根据 交易Id 获取 give token 记录
func (h *GiveTokenHandler) GetRecord(fiberCtx *fiber.Ctx) error {
	var req types.GetByTxIdReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	// 检查交易Id是否正确
	if _, err := solana.SignatureFromBase58(req.TxId); err != nil {
		return response.FailWithError(fiberCtx, "malformed tx id", err)
	}
	l := logic.NewGiveTokenLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetRecord(fiberCtx.Context(), req.TxId)
	if err != nil {
		log.Errorf("获取take token 交易信息错误: %v", err)
		return response.FailWithError(fiberCtx, "get give token tx info error:", err)
	}

	return response.OkWithData(fiberCtx, txInfo)
}

// GetTxInfo 根据 native account 获取打包 give token 交易的参数
func (h *GiveTokenHandler) GetTxInfo(fiberCtx *fiber.Ctx) error {
	// 检查jwt token，以及jwt token当中的from account
	claims, err := utils.ExtractTokenMetadata(fiberCtx)
	if err != nil {
		return response.UnAuthorizedError(fiberCtx, "unauthorized")
	}
	fromAccount := claims.Credentials["nativeAccount"].(string)
	if _, err := solana.PublicKeyFromBase58(fromAccount); err != nil {
		return response.FailWithError(fiberCtx, "malformed from account", err)
	}

	// 检查请求当中的to account
	var req types.GetGiveTokenTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	if _, err := solana.PublicKeyFromBase58(req.To); err != nil {
		return response.FailWithError(fiberCtx, "malformed to account", err)
	}

	// 获取打包take token的交易信息
	l := logic.NewGiveTokenLogic(fiberCtx.Context(), h.srvCtx)
	txInfo, err := l.GetTxInfo(fiberCtx.Context(), fromAccount, req.To, req.Amount)
	if err != nil {
		log.Errorf("获取take token 交易信息错误: %v", err)
		return response.FailWithError(fiberCtx, "get give token tx info error:", err)
	}

	return response.OkWithData(fiberCtx, txInfo)
}

// CommitTx 提交打包好的交易，等待服务器签名
func (h *GiveTokenHandler) CommitTx(fiberCtx *fiber.Ctx) error {
	ctx := fiberCtx.Context()
	// 检查请求当中的to account
	var req types.CommitGiveTokenTxInfoReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		return response.BadRequest(fiberCtx, "invalid request body")
	}
	encodedTx := req.EncodedTx

	// 检查to地址是否正确
	toNativeAccount, err := solana.PublicKeyFromBase58(req.To)
	if err != nil {
		return response.FailWithError(fiberCtx, "malformed to address", err)
	}

	// 解析交易，确定手续费支付地址，成本费支付地址，以及签名是否正确
	preCheckedTx, err := app_utils.PreCheckEncodedTx(encodedTx)
	if err != nil {
		log.Errorf("官方转账获取奖励 - 检查打包的交易错误: %v", err)
		return response.FailWithError(fiberCtx, "check transaction error: %v", err)
	}

	// 处理交易
	l := logic.NewGiveTokenLogic(ctx, h.srvCtx)
	txId, err := l.ProcessCommitTx(ctx, preCheckedTx, toNativeAccount)
	if err != nil {
		return response.FailWithError(fiberCtx, "process transaction error:", err)
	}

	// 返回交易Id
	return response.OkWithData(fiberCtx, txId)
}
