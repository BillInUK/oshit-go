package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/api")

	// 初始化所有handler
	authHandler := NewAuthHandler(srvCtx)
	configHandler := NewConfigHandler(srvCtx)
	feeHandler := NewFeeHandler(srvCtx)
	priceHandler := NewPriceHandler(srvCtx)
	inviteHandler := NewInviteHandler(srvCtx)
	serviceHandler := NewServiceHandler(srvCtx)

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)
		// TODO: 添加JWT中间件
		// auth.Post("/info", middleware.JWTProtected(), authHandler.QueryNativeAccountInfo)
	}

	// 配置路由
	config := api.Group("/config")
	{
		config.Get("/token", configHandler.GetTokenInfo)
		config.Get("/fee-tolerance", configHandler.GetFeeTolerance)
	}

	// 手续费路由
	fee := api.Group("/fee")
	{
		fee.Get("/priority", feeHandler.GetPriorityFee)
		fee.Get("/priority/on-chain", feeHandler.GetPriorityFeeOnBlockchain)
		fee.Get("/compute-units", feeHandler.GetComputeUnitConsumed)
	}
	
	// 价格路由
	price := api.Group("/price")
	{
		price.Get("/token/sol", priceHandler.GetTokenQuoteSOLPrice)
		price.Get("/token/usdt", priceHandler.GetTokenQuoteUSDTPrice)
		price.Get("/usdt/sol", priceHandler.GetUSDTQuoteSOLPrice)
		price.Post("/birdeye", priceHandler.GetBirdEyePrice)
	}

	// 邀请路由
	invite := api.Group("/invite")
	{
		invite.Post("/account-by-code", inviteHandler.GetAccountByInviteCode)
		invite.Post("/check-record", inviteHandler.CheckInviteRecord)
		invite.Post("/up-records", inviteHandler.GetUpInviterRecords)
		invite.Post("/down-records", inviteHandler.GetDownInviteeRecords)
		invite.Get("/reward-distribution", inviteHandler.GetRewardDistribution)
		invite.Get("/reward-claims", inviteHandler.GetRewardClaims)
		invite.Get("/token-holders", inviteHandler.GetTokenHolders)
	}

	// 服务注册路由
	service := api.Group("/service")
	{
		service.Post("/register", serviceHandler.RegisterService)
		service.Post("/update", serviceHandler.UpdateService)
		service.Post("/query", serviceHandler.QueryService)
	}
}
