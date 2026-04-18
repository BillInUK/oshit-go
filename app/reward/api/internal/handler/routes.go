package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/reward/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/reward")
	takeTokenHandler := NewTakeTokenHandler(srvCtx)
	giveTokenHandler := NewGiveTokenHandler(srvCtx)
	lotteryHandler := NewLotteryHandler(srvCtx)
	campaignHandler := NewCampaignHandler(srvCtx)
	rewardCodeHandler := NewRewardCodeHandler(srvCtx)

	/*
		旧工程前端接口
		route.Post("/sol/reward/getRewardTransactionParam", reward_handlers.GetRewardTransactionParam)
		route.Post("/sol/reward/getRedeemTxInfo", reward_handlers.GetRedeemTxInfo)
		route.Post("/sol/reward/getRedeemRewardByTxId", middleware.JWTProtected(), reward_handlers.GetRedeemRewardByTxId)
		route.Post("/sol/reward/queryOfficialGiveTokenRewardRule", middleware.JWTProtected(), reward_handlers.QueryOfficialGiveTokenRewardRule)
		route.Post("/sol/reward/queryOfficialTransferTokenRewardRule", middleware.JWTProtected(), reward_handlers.QueryOfficialTransferTokenRewardRule)
		route.Post("/sol/reward/queryAccountMatchUnofficialTransferRewardRule", middleware.JWTProtected(), reward_handlers.QueryAccountMatchUnofficialTransferRewardRule)
		route.Post("/sol/reward/queryOfficialGivenTokenRewardRecordByTxId", reward_handlers.QueryOfficialGivenTokenRewardRecordByTxId)
		route.Post("/sol/reward/queryOfficialTransferTokenRecordByTxId", reward_handlers.QueryOfficialTransferTokenRecordByTxId)
		route.Post("/sol/reward/getTokenRewardTransactionContent", middleware.JWTProtected(), reward_handlers.GetTokenRewardTransactionContent)
		route.Post("/sol/reward/queryRewardStatics", middleware.JWTProtected(), reward_handlers.QueryRewardStatics)

		route.Post("/sol/reward/officialGiveToken", reward_prehandle.OfficialGiveLimit, reward_handlers.OfficialGiveToken)

		route.Post("/sol/reward/queryLotteryRecords", middleware.JWTProtected(), reward_handlers.QueryLotteryRecords)
		route.Post("/sol/reward/queryUnRedeemedLotteryRecords", middleware.JWTProtected(), reward_handlers.QueryUnClaimedLotteryRecords)
		route.Post("/sol/reward/lotteryOfficialGiveToken", middleware.JWTProtected(), reward_handlers.LotteryOfficialGiveToken)
		route.Post("/sol/reward/getClaimRewardLotteryTxInfo", middleware.JWTProtected(), reward_handlers.GetClaimRewardLotteryTxInfo)
		route.Post("/sol/reward/getGiveTokenStats", middleware.JWTProtected(), reward_handlers.GetGiveTokenStats)
		route.Post("/sol/reward/claimRewardLottery", middleware.JWTProtected(), reward_handlers.ClaimRewardLottery)

		route.Post("/sol/reward/officialTransferTokenForReward", reward_prehandle.OfficialTransferLimit, reward_handlers.OfficialTransferTokenForReward)
		route.Post("/sol/reward/queryCampaignUserScore", reward_handlers.QueryCampaignScore)
		route.Post("/sol/reward/queryExchangeCampaignScoreRule", reward_handlers.QueryExchangeCampaignScoreRule)
		route.Post("/sol/reward/queryCampaignExchangeLimit", reward_handlers.QueryCampaignExchangeLimit)
		route.Post("/sol/reward/exchangeScoreToToken", reward_handlers.ExchangeScoreToToken)
		route.Post("/sol/reward/redeemRewardByCode", reward_prehandle.OfficialGiveLimit, reward_handlers.RedeemRewardByCode)

		route.Post("/sol/reward/querySwapTokenRule", reward_handlers.QuerySwapTokenRule)
		route.Post("/sol/reward/queryExtraSOLFee", reward_handlers.QueryExtraSOLFee)

		route.Post("/sol/reward/querySwapTokenRecordByTxId", reward_handlers.QuerySwapTokenRecordByTxId)
		route.Post("/sol/reward/queryLightHouseInfo", reward_handlers.QueryLightHouseInfo)
		route.Post("/sol/reward/queryLightHouseTokenQuotaById", reward_handlers.QueryLightHouseTokenQuotaById)
		route.Post("/sol/reward/getTotalSwapAmountByHouseId", reward_handlers.GetTotalSwapAmountByHouseId)
		route.Post("/sol/reward/swapToToken", reward_handlers.SwapToToken)
	*/

	// take token 相关路由
	take := api.Group("/take")
	{
		take.Post("/config", takeTokenHandler.GetConfig)   // 旧工程 /sol/reward/queryOfficialGiveTokenRewardRule
		take.Post("/record", takeTokenHandler.GetRecord)   // 旧工程 /sol/reward/queryOfficialGivenTokenRewardRecordByTxId
		take.Post("/tx-info", takeTokenHandler.GetTxInfo)  // 旧工程 /sol/reward/getTokenRewardTransactionContent
		take.Post("/commit-tx", takeTokenHandler.CommitTx) // 旧工程 /sol/reward/officialGiveToken
	}

	// give token 相关路由
	give := api.Group("/give")
	{
		give.Post("/config", giveTokenHandler.GetConfig)                     // 旧工程 /sol/reward/queryOfficialTransferTokenRewardRule
		give.Post("/record", giveTokenHandler.GetRecord)                     // 旧工程 /sol/reward/queryOfficialTransferTokenRecordByTxId
		give.Post("/tx-info", JWTAuthMiddleware, giveTokenHandler.GetTxInfo) // 旧工程没有这个接口，新工程简化流程，这块需要新开发
		give.Post("/commit-tx", giveTokenHandler.CommitTx)                   // 旧工程 /sol/reward/officialTransferTokenForReward
	}

	// lottery token 相关路由
	lottery := api.Group("/lottery")
	{
		lottery.Post("/status", JWTAuthMiddleware, lotteryHandler.GetStatus)       // 旧工程 /sol/reward/getGiveTokenStats
		lottery.Post("/execute", JWTAuthMiddleware, lotteryHandler.ExecuteLottery) // 旧工程 /sol/reward/lotteryOfficialGiveToken
		lottery.Post("/unclaimed", lotteryHandler.GetUnClaimedRecord)              // 旧工程 /sol/reward/queryUnRedeemedLotteryRecords
		lottery.Post("/record", lotteryHandler.GetRecord)                          // 旧工程 /sol/reward/queryLotteryRecords
		lottery.Post("/tx-info", lotteryHandler.GetTxInfo)                         // 旧工程 /sol/reward/getClaimRewardLotteryTxInfo
		lottery.Post("/commit-tx", lotteryHandler.CommitTx)                        // 旧工程 /sol/reward/claimRewardLottery
	}

	// campaign 相关路由
	campaign := api.Group("/campaign")
	{
		campaign.Post("/score", campaignHandler.GetUserScore)       // 旧工程 /sol/reward/queryCampaignUserScore
		campaign.Post("/config", campaignHandler.GetExchangeConfig) // 旧工程 /sol/reward/queryExchangeCampaignScoreRule
		campaign.Post("/limit", campaignHandler.GetExchangeLimit)   // 旧工程 /sol/reward/queryCampaignExchangeLimit
		campaign.Post("/tx-info", campaignHandler.GetTxInfo)        // 旧工程没有这个接口，新工程简化流程，这块需要新开发
		campaign.Post("/commit-tx", campaignHandler.CommitTx)       // 旧工程 /sol/reward/exchangeScoreToToken
	}

	// reward code 相关路由
	rewardCode := api.Group("/reward-code")
	{
		rewardCode.Post("/info", rewardCodeHandler.GetInfo)       // 旧工程没有这个接口，该接口用来查询奖励码信息
		rewardCode.Post("/tx-info", rewardCodeHandler.GetTxInfo)  // 旧工程 /sol/reward/getRedeemTxInfo
		rewardCode.Post("/commit-tx", rewardCodeHandler.CommitTx) // 旧工程 /sol/reward/redeemRewardByCode
	}
}
