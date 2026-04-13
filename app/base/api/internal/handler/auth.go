package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/pkg/response"
)

type AuthHandler struct {
	prefix string
	srvCtx *svc.ServiceContext
}

func NewAuthHandler(svc *svc.ServiceContext) *AuthHandler {
	return &AuthHandler{
		prefix: "登录鉴权 -",
		srvCtx: svc,
	}
}

func (auth *AuthHandler) Login(fiberCtx *fiber.Ctx) error {
	var req types.LoginReq
	if err := fiberCtx.BodyParser(&req); err != nil {
		log.Errorf("%s 解析请求提错误: %v", auth.prefix, err)
		return response.BadRequest(fiberCtx, "Invalid request body")
	}

	al := logic.NewAuthLogic(fiberCtx.Context(), auth.srvCtx)
	resp, err := al.Login(&req)
	if err != nil {
		log.Errorf("%s 处理登录逻辑错误: %v", auth.prefix, err)
		return response.ServerError(fiberCtx, "login failed", err)
	}
	return response.OkWithData(fiberCtx, resp)
}

func (auth *AuthHandler) QueryNativeAccountInfo(fiberCtx *fiber.Ctx) error {
	fromAccount := fiberCtx.Locals("nativeAccount").(string)

	l := logic.NewAuthLogic(fiberCtx.Context(), auth.srvCtx)
	rsp, err := l.QueryNativeAccountInfo(fromAccount)
	if err != nil {
		log.Errorf("%s 查询地址 %s 信息错误: %v", auth.prefix, fromAccount, err)
		return response.FailWithMsg(fiberCtx, "query native account info failed")
	}

	return response.OkWithData(fiberCtx, rsp)
}
