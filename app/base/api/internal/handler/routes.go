package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/base/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/base")

	/*
		旧工程前端接口
		route.Post("/sol/config/queryTokenInfo", common_handlers.QueryTokenInfo)
		route.Post("/sol/config/querySOLPriorityFee", common_handlers.QuerySOLPriorityFee)
		route.Post("/sol/config/querySOLPriorityFeeOnBlockChain", common_handlers.QuerySOLPriorityFeeOnBlockChain)	废弃统一使用 /fee/priority
		route.Post("/sol/config/querySOLConsumedUnits", common_handlers.QuerySOLConsumedUnits)
		route.Post("/sol/config/queryTokenQuoteSOLPrice", common_handlers.QueryTokenQuoteSOLPrice)
		route.Post("/sol/config/queryTokenQuoteUSDTPrice", common_handlers.QueryTokenQuoteUSDTPrice)
		route.Post("/sol/config/queryUSDTQuoteSOLPrice", common_handlers.QueryUSDTQuoteSOLPrice)
		route.Post("/sol/config/getBirdEyePriceData", common_handlers.GetBirdEyePriceData)

		route.Post("/sol/auth/loginWithNativeAccount", common_handlers.LoginWithNativeAccount)
		route.Post("/sol/invite/queryNativeAccountInfoByInviteCode", common_handlers.QueryNativeAccountInfoByInviteCode)
		route.Post("/sol/invite/querySolTransferRewardDistribution", common_handlers.QuerySolTransferRewardDistribution)
		route.Post("/sol/invite/querySolTransferRewardClaim", common_handlers.QuerySolTransferRewardClaim)
		route.Post("/sol/invite/recursiveQueryClaimsAndInviters", common_handlers.RecursiveQueryClaimsAndInviters)
		route.Post("/sol/invite/recursiveQueryUpInviterRecords", common_handlers.RecursiveQueryUpInviterRecords)
		route.Post("/sol/invite/recursiveQueryDownInviteeRecords", common_handlers.RecursiveQueryDownInviteeRecords)
		route.Post("/sol/invite/queryTokenHoldersNumber", common_handlers.QueryTokenHoldersNumber)

		route.Post("/sol/auth/queryNativeAccountInfo", middleware.JWTProtected(), common_handlers.QueryNativeAccountInfo)
	*/

	// 初始化所有handler
	authHandler := NewAuthHandler(srvCtx)
	infoHandler := NewInfoHandler(srvCtx)
	feeHandler := NewFeeHandler(srvCtx)
	priceHandler := NewPriceHandler(srvCtx)
	inviteHandler := NewInviteHandler(srvCtx)

	// 认证路由
	auth := api.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)                                    // 旧工程 /sol/auth/loginWithNativeAccount
		auth.Post("/info", JWTAuthMiddleware, authHandler.QueryNativeAccountInfo) // 旧工程 /sol/auth/queryNativeAccountInfo
	}

	// 信息路由
	info := api.Group("/info")
	{
		info.Get("/token-info", infoHandler.GetTokenInfo)       // 旧工程 /sol/config/queryTokenInfo
		info.Get("/token-holders", infoHandler.GetTokenHolders) // 旧工程 /sol/invite/queryTokenHoldersNumber
	}

	// 手续费路由
	fee := api.Group("/fee")
	{
		fee.Get("/priority", feeHandler.GetPriorityFee) // 旧工程 /sol/config/querySOLPriorityFee
		fee.Get("/inst-units", feeHandler.GetInstUnits) // 旧工程 /sol/config/querySOLConsumedUnits
		// fee.Get("/fee-tolerance", feeHandler.GetFeeTolerance)
	}

	// 价格路由
	price := api.Group("/price")
	{
		price.Get("/token/sol", priceHandler.GetTokenQuoteSOLPrice)   // 旧工程 /sol/config/queryTokenQuoteSOLPrice
		price.Get("/token/usdt", priceHandler.GetTokenQuoteUSDTPrice) // 旧工程 /sol/config/queryTokenQuoteUSDTPrice
		price.Get("/usdt/sol", priceHandler.GetUSDTQuoteSOLPrice)     // 旧工程 /sol/config/queryUSDTQuoteSOLPrice
		price.Post("/kline", priceHandler.GetBirdEyePrice)            // 旧工程 /sol/config/getBirdEyePriceData
	}

	// 邀请路由
	invite := api.Group("/invite")
	{
		invite.Post("/account-by-code", inviteHandler.GetAccountByInviteCode) // 旧工程 /sol/invite/queryNativeAccountInfoByInviteCode
		invite.Post("/check-record", inviteHandler.CheckInviteRecord)
		invite.Post("/up-records", inviteHandler.GetUpInviterRecords)     // 旧工程 /sol/invite/recursiveQueryUpInviterRecords
		invite.Post("/down-records", inviteHandler.GetDownInviteeRecords) // 旧工程 /sol/invite/recursiveQueryDownInviteeRecords
	}
}
