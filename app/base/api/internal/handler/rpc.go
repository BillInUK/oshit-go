package handler

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"oshit-go/app/base/api/internal/svc"
	"time"
)

type RpcHandler struct {
	srvCtx *svc.ServiceContext
}

func NewRpcHandler(svc *svc.ServiceContext) *RpcHandler {
	return &RpcHandler{svc}
}

func (h *RpcHandler) SolanaRpc(fiberCtx *fiber.Ctx) error {
	// 1. 定义允许的方法白名单
	allowedMethods := map[string]bool{
		"getBalance":              true,
		"getTokenAccountsByOwner": true,
		"getLatestBlockhash":      true,
		"getAccountInfo":          true,
		"getSignatureStatus":      true,
		"getSlot":                 true,
		"getBlockTime":            true,
		"getSignatureStatuses":    true,
	}

	// 2. 解析前端请求
	var rpcReq struct {
		JSONRPC string        `json:"jsonrpc"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
		ID      interface{}   `json:"id"`
	}

	if err := fiberCtx.BodyParser(&rpcReq); err != nil {
		return fiberCtx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32600, "message": "Invalid request"},
			"id":      nil,
		})
	}

	// 3. 验证必需字段和方法权限
	if rpcReq.Method == "" {
		return fiberCtx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32600, "message": "Method is required"},
			"id":      rpcReq.ID,
		})
	}

	// 检查方法是否在白名单中
	if !allowedMethods[rpcReq.Method] {
		return fiberCtx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32601,
				"message": fmt.Sprintf("Method '%s' is not allowed", rpcReq.Method),
			},
			"id": rpcReq.ID,
		})
	}

	// 4. 创建响应结构
	var response interface{}

	// 5. 执行 RPC 调用
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.srvCtx.RpcClient.RPCCallForInto(ctx, &response, rpcReq.Method, rpcReq.Params)
	if err != nil {
		// 处理不同类型的错误
		var rpcErr *jsonrpc.RPCError
		if errors.As(err, &rpcErr) {
			return fiberCtx.Status(fiber.StatusOK).JSON(fiber.Map{
				"jsonrpc": "2.0",
				"error":   map[string]interface{}{"code": rpcErr.Code, "message": rpcErr.Message},
				"id":      rpcReq.ID,
			})
		}

		// 处理网络级错误
		return fiberCtx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32603, "message": err.Error()},
			"id":      rpcReq.ID,
		})
	}

	// 6. 返回成功响应
	return fiberCtx.Status(fiber.StatusOK).JSON(fiber.Map{
		"jsonrpc": "2.0",
		"result":  response,
		"id":      rpcReq.ID,
	})
}
