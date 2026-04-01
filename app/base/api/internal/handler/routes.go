package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/base")

	// 初始化所有handler
	authHandler := NewAuthHandler(srvCtx)
	infoHandler := NewInfoHandler(srvCtx)
	feeHandler := NewFeeHandler(srvCtx)
	priceHandler := NewPriceHandler(srvCtx)
	inviteHandler := NewInviteHandler(srvCtx)

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)
		// TODO: 添加JWT中间件
		// auth.Post("/info", middleware.JWTProtected(), authHandler.QueryNativeAccountInfo)
	}

	// 信息路由
	info := api.Group("/info")
	{
		info.Get("/token-info", infoHandler.GetTokenInfo)
		info.Get("/token-holders", infoHandler.GetTokenHolders)
	}

	// 手续费路由
	fee := api.Group("/fee")
	{
		fee.Get("/priority", feeHandler.GetPriorityFee)
		fee.Get("/inst-units", feeHandler.GetInstUnits)
		fee.Get("/fee-tolerance", feeHandler.GetFeeTolerance)
	}

	// 价格路由
	price := api.Group("/price")
	{
		price.Get("/token/sol", priceHandler.GetTokenQuoteSOLPrice)
		price.Get("/token/usdt", priceHandler.GetTokenQuoteUSDTPrice)
		price.Get("/usdt/sol", priceHandler.GetUSDTQuoteSOLPrice)
		price.Post("/kline", priceHandler.GetBirdEyePrice)
	}

	// 邀请路由
	invite := api.Group("/invite")
	{
		invite.Post("/account-by-code", inviteHandler.GetAccountByInviteCode)
		invite.Post("/check-record", inviteHandler.CheckInviteRecord)
		invite.Post("/up-records", inviteHandler.GetUpInviterRecords)
		invite.Post("/down-records", inviteHandler.GetDownInviteeRecords)
	}
}
