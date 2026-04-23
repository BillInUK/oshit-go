# 前端接口迁移记录

> 记录 oshit-dapp 从旧工程迁移到 oshit-go 新工程的接口映射情况。
> 旧工程 base URL：`/meme/api/v1/sol`
> 新工程各服务 base URL 见下表。

| 服务 | 新 base URL |
|---|---|
| base | `/meme/base/api/v1` |
| reward | `/meme/reward/api/v1` |
| pos/snap | `/meme/snap/api/v1` |

---

## base 服务

> 涉及前端文件：`src/helper/api/auth.ts`、`src/helper/api/config.ts`、`src/helper/api/invite.ts`、`src/config.ts`

| 前端函数 | 旧路径（POST） | 新路径 | 方法 |
|---|---|---|---|
| `apiConfig.rpc`（Solana RPC endpoint） | `${rpcDomain}/meme/api/v1/sol/config/rpc` | `/meme/base/api/v1/rpc/solana` | POST（由 Solana SDK 内部调用，需 JWT） |
| `loginApi` | `/auth/loginWithNativeAccount` | `/auth/login` | POST |
| `loginApi`（内部 queryInfo） | `/auth/queryNativeAccountInfo` | `/auth/info` | POST |
| `querySOLPriorityFee` | `/config/querySOLPriorityFee` | `/fee/priority` | GET |
| `querySOLConsumedUnits` | `/config/QuerySOLConsumedUnits` | `/fee/inst-units` | GET |
| `queryTokenQuoteSOLPrice` | `/config/queryTokenQuoteSOLPrice` | `/price/token/sol` | GET |
| `queryTokenQuoteUSDTPrice` | `/config/queryTokenQuoteUSDTPrice` | `/price/token/usdt` | GET |
| `getBirdEyePriceData` | `/config/getBirdEyePriceData` | `/price/kline` | POST |
| `recursiveQueryDownInviteRecords` | `/invite/recursiveQueryDownInviteeRecords` | `/invite/down-records` | POST |
| `queryTokenHoldersNumber` | `/invite/queryTokenHoldersNumber` | `/info/token-holders` | GET |

---

## reward 服务

> 涉及前端文件：`src/helper/api/reward.ts`

### 已迁移

| 前端函数 | 旧路径 | 新路径 | 方法 |
|---|---|---|---|
| `officialGiveToken` | `/reward/officialGiveToken` | `/take/commit-tx` | POST |
| `getTokenRewardTransactionContent` | `/reward/getTokenRewardTransactionContent` | `/take/tx-info` | POST |
| `queryOfficialGiveTokenRewardRule` | `/reward/queryOfficialGiveTokenRewardRule` | `/take/config` | POST |
| `queryOfficialGivenTokenRewardRecordByTxId` | `/reward/queryOfficialGivenTokenRewardRecordByTxId` | `/take/record` | POST |
| `officialTransferTokenForReward` | `/reward/officialTransferTokenForReward` | `/give/commit-tx` | POST |
| `queryOfficialTransferTokenRewardRule`（主调用） | `/reward/queryOfficialTransferTokenRewardRule` | `/give/config` | POST |
| `queryOfficialTransferTokenRecordByTxId` | `/reward/queryOfficialTransferTokenRecordByTxId` | `/give/record` | POST |
| `getGiveTokenStats` | `/reward/getGiveTokenStats` | `/lottery/status` | POST |
| `queryUnRedeemedLotteryRecords` | `/reward/queryUnRedeemedLotteryRecords` | `/lottery/unclaimed` | POST |
| `lotteryOfficialGiveToken` | `/reward/lotteryOfficialGiveToken` | `/lottery/execute` | POST |
| `getClaimRewardLotteryTxInfo` | `/reward/getClaimRewardLotteryTxInfo` | `/lottery/tx-info` | POST |
| `claimRewardLottery` | `/reward/claimRewardLottery` | `/lottery/commit-tx` | POST |
| `getRedeemTxInfo` | `/reward/getRedeemTxInfo` | `/reward-code/tx-info` | POST |
| `redeemRewardByCode` | `/reward/redeemRewardByCode` | `/reward-code/commit-tx` | POST |
| `newExchangeScoreToToken` | `/reward/exchangeScoreToToken` | `/campaign/commit-tx` | POST |
| `queryCampaignExchangeLimit` | `/reward/queryCampaignExchangeLimit` | `/campaign/limit` | POST |
| `queryExchangeCampaignScoreRule` | `/reward/queryExchangeCampaignScoreRule` | `/campaign/config` | POST |

### 半迁移（`queryOfficialTransferTokenRewardRule`）

该函数内部共发起 5 个请求，主调用已迁移，其余 3 个 invite 子调用暂无新路由映射，仍请求旧服务：

| 子调用 | 旧路径 | 状态 |
|---|---|---|
| `queryAccountMatchUnofficialTransferRewardRule` | `/reward/queryAccountMatchUnofficialTransferRewardRule` | 保留旧服务，新路由无对应 |
| `querySolTransferRewardDistribution` | `/invite/querySolTransferRewardDistribution` | 保留旧服务，新路由无对应 |
| `querySolTransferRewardClaim` | `/invite/querySolTransferRewardClaim` | 保留旧服务，新路由无对应 |
| `recursiveQueryUpInviterRecords` | `/invite/recursiveQueryUpInviterRecords` | 保留旧服务，新路由无对应 |

### 保留旧服务（新路由无映射）

| 前端函数 | 旧路径 | 备注 |
|---|---|---|
| `queryRewardStatics` | `/reward/queryRewardStatics` | 新路由无对应接口 |
| `getRewardTransactionParam` | `/reward/getRewardTransactionParam` | 新路由无对应接口 |
| `querySwapTokenRule` | `/reward/querySwapTokenRule` | swap 功能未迁移 |
| `swapToToken` | `/reward/swapToToken` | swap 功能未迁移 |
| `queryExtraSOLFee` | `/reward/queryExtraSOLFee` | swap 功能未迁移 |
| `queryLightHouseInfo` | `/reward/queryLightHouseInfo` | swap/lighthouse 功能未迁移 |
| `getTotalSwapAmountByHouseId` | `/reward/getTotalSwapAmountByHouseId` | swap/lighthouse 功能未迁移 |
| `queryLightHouseTokenQuotaById` | `/reward/queryLightHouseTokenQuotaById` | swap/lighthouse 功能未迁移 |

---

## pos/snap 服务

> 涉及前端文件：`src/helper/api/pos.ts`、`src/helper/staking/staking.api.ts`

### 已迁移（pos 快照奖励）

| 前端函数 | 旧路径 | 新路径 | 方法 |
|---|---|---|---|
| `querySolPosRewardRule` | `/pos/querySolPosRewardRule` | `/pos/reward/config` | POST |
| `queryPosRewards` | `/pos/queryPosRewards` | `/pos/reward/record` | POST |
| `queryPosRewardDetail` | `/pos/queryPosRewardDetail` | `/pos/reward/stat` | POST |
| `queryGroupInfo` | `/pos/queryGroupInfo` | `/pos/reward/group-info` | POST |
| `claimPosReward` | `/pos/claimPosReward` | `/pos/reward/commit-tx` | POST |
| `queryClaimPosRewardByTxId` | `/pos/queryClaimPosRewardByTxId` | `/pos/reward/claim-record` | POST |

### 已迁移（stake 质押与奖励）

| 前端函数 | 旧路径 | 新路径 | 方法 |
|---|---|---|---|
| `querySolStakeRewardRule` | `/pos/querySolStakeRewardRule` | `/stake/reward/config` | POST |
| `queryUnClaimedStakeRewards` | `/pos/queryUnClaimedStakeRewards` | `/stake/reward/record` | POST |
| `queryStakeRewardStat` | `/pos/queryStakeRewardStat` | `/stake/reward/stat` | POST |
| `getClaimStakeRewardTxInfo` | `/pos/getClaimStakeRewardTxInfo` | `/stake/reward/tx-info` | POST |
| `stake` | `/pos/stakeToken` | `/stake/token/stake` | POST |
| `unstake` | `/pos/unStakeToken` | `/stake/token/unstake` | POST |
| `restake` | `/pos/reStakeToken` | `/stake/token/restake` | POST |
| `claimStakeReward` | `/pos/claimStakeReward` | `/stake/reward/commit-tx` | POST |
| `queryAreaLeaderInfo` | `/pos/areaLeaderInfo` | `/stake/reward/leader/info` | POST |
| `queryUnClaimedAreaLeaderRewards` | `/pos/queryUnClaimedAreaLeaderRewards` | `/stake/reward/leader/records` | POST |
| `getClaimAreaLeaderRewardTxInfo` | `/pos/getClaimAreaLeaderRewardTxInfo` | `/stake/reward/leader/tx-info` | POST |
| `claimAreaLeaderReward` | `/pos/claimAreaLeaderReward` | `/stake/reward/leader/commit-tx` | POST |

### 保留旧服务（新路由无映射）

| 前端函数 | 旧路径 | 备注 |
|---|---|---|
| `queryRankActivityExchangeRule` | `/pos/queryRankActivityExchangeRule` | 新路由无对应接口 |
| `queryExchangeScoreToTokenByTxId` | `/pos/queryExchangeScoreToTokenByTxId` | 新路由无对应接口 |
| `faucetStakeToken` | `/pos/faucetStakeToken` | 新路由无对应接口 |
| `getStakeAbleAmount` | `/pos/stakeAbleAmount` | 新路由无对应接口 |
| `queryMinStakeAmount` | `/pos/queryMinStakeAmount` | 新路由改为 `GET /snap/stake/token/min-amount/:type`，path param 值待确认 |

### 不迁移（外部服务）

| 前端函数 | 调用地址 | 备注 |
|---|---|---|
| `getRank` | `{dev}/oshit/activityRank/query` 等 | 外部 dev 服务，与新工程无关 |
| `getRankRule` | `{dev}/oshit/rankRule` | 外部 dev 服务，与新工程无关 |
| `getRankAllCount` | `{dev}/oshit/activityRank/twitterCount` | 外部 dev 服务，与新工程无关 |
