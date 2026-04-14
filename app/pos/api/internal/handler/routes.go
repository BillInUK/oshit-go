package handler

import (
	"github.com/gofiber/fiber/v2"
	"oshit-go/app/pos/api/internal/svc"
)

func RegisterRoutes(fiberApp *fiber.App, srvCtx *svc.ServiceContext) {
	api := fiberApp.Group("/snap")
	posHandler := NewPosHandler(srvCtx)
	stakeHandler := NewStakeHandler(srvCtx)

	// pos 相关路由
	pos := api.Group("/pos")
	{
		// 快照相关接口
		pos.Post("/shot/take", posHandler.TakeSnapShot)   // 旧工程 	/sol/pos/takeStakeSnapShot
		pos.Post("/shot/reset", posHandler.ResetSnapShot) // 旧工程  	/sol/pos/resetStakeSnapShot

		pos.Post("/reward/group-info", posHandler.GetGroupInfo) // 旧工程 /sol/pos/queryGroupInfo
		pos.Post("/reward/stat", posHandler.GetRewardStat)      // 旧工程 /sol/pos/queryPosRewardDetail

		pos.Post("/reward/config", posHandler.GetConfig)            // 旧工程 /sol/pos/querySolPosRewardRule 查询pos奖励发放规则
		pos.Post("/reward/record", posHandler.GetRewardRecord)      // 旧工程 /sol/pos/queryPosRewards
		pos.Post("/reward/tx-info", posHandler.GetTxInfo)           // 旧工程 /sol/pos/getClaimPosRewardTxInfo
		pos.Post("/reward/claim-record", posHandler.GetClaimRecord) // 旧工程 /sol/pos/getClaimPosRewardTxInfo
		pos.Post("/reward/commit-tx", posHandler.CommitTx)          // 旧工程 /sol/pos/claimPosReward
	}

	// stake 相关路由
	stake := api.Group("/stake")
	{
		// 快照相关接口
		stake.Post("/shot/take", stakeHandler.TakeSnapShot)   // 旧工程 	/sol/pos/takeStakeSnapShot
		stake.Post("/shot/reset", stakeHandler.ResetSnapShot) // 旧工程  	/sol/pos/resetStakeSnapShot

		// 质押相关接口
		stake.Get("/token/min-amount/${type}", stakeHandler.MinAmount) // 旧工程		/sol/pos/queryMinStakeAmount
		stake.Post("/token/stake", stakeHandler.StakeToken)            // 旧工程 		/sol/pos/stakeToken
		stake.Post("/token/unstake", stakeHandler.UnStakeToken)        // 旧工程 		/sol/pos/unStakeToken
		stake.Post("/token/restake", stakeHandler.ReStakeToken)        // 旧工程 		/sol/pos/unStakeToken

		// 用户奖励相关接口
		stake.Post("/reward/star-level", stakeHandler.GetStarLevel) // 旧工程 		/sol/pos/fetchStakeStarLevel
		stake.Post("/reward/stat", stakeHandler.GetRewardStat)      // 旧工程 		/sol/pos/queryStakeRewardStat

		stake.Post("/reward/config", stakeHandler.GetConfig)                          // 旧工程 		/sol/pos/querySolStakeRewardRule
		stake.Post("/reward/record", JWTAuthMiddleware, stakeHandler.GetRewardRecord) // 旧工程 		/sol/pos/queryUnClaimedStakeRewards
		stake.Post("/reward/tx-info", JWTAuthMiddleware, stakeHandler.GetTxInfo)      // 旧工程 		/sol/pos/getClaimStakeRewardTxInfo
		stake.Post("/reward/claim-record", stakeHandler.GetClaimRecord)               // 旧工程			/sol/pos/queryStakeRewardClaimRecordByTxId
		stake.Post("/reward/commit-tx", stakeHandler.CommitTx)                        // 旧工程			/sol/pos/claimStakeReward

		stake.Post("/reward/leader/info", JWTAuthMiddleware, stakeHandler.GetLeaderInfo)
		stake.Post("/reward/leader/records", JWTAuthMiddleware, stakeHandler.GetLeaderRewardRecord) // 旧工程 		/sol/pos/querySolStakeRewardRule
		stake.Post("/reward/leader/tx-info", JWTAuthMiddleware, stakeHandler.GetLeaderTxInfo)       // 旧工程 		/sol/pos/queryUnClaimedStakeRewards
		stake.Post("/reward/leader/claim-record", stakeHandler.GetLeaderClaimRecord)                // 旧工程		/sol/pos/queryStakeRewardClaimRecordByTxId
		stake.Post("/reward/leader/commit-tx", stakeHandler.LeaderCommitTx)                         // 旧工程		/sol/pos/claimStakeReward
	}
}
