# Common模块到Base模块迁移指南

## 概述

本文档记录了从旧工程`micro-meme`的`common`模块迁移到新工程`oshit-go`的`base`模块的主要变化，包括表结构变更和API接口变更。

## 一、表结构对比

### 1.1 旧工程表结构（micro-meme/common）

| 表名 | 字段名（驼峰） | 类型 | 说明 |
|------|---------------|------|------|
| `t_system_config` | `Env` | INT | 环境: 0.主网 1.测试网 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_chain_config` | `Chain` | VARCHAR(64) | 链，例如,SOL,ETH |
| | `RpcUrl` | VARCHAR(1024) | rpc链接地址 |
| | `WssUrl` | VARCHAR(1024) | websocket链接地址 |
| | `Decimal` | INT | 原生币种的进制 |
| | `Symbol` | VARCHAR(64) | 原生币的符号 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_sol_token_config` | `Brand` | VARCHAR(64) | 项目方 |
| | `TokenSymbol` | VARCHAR(64) | token的符号 |
| | `Decimal` | INT | token金额的进制 |
| | `TokenMintAccount` | VARCHAR(64) | token合约地址 |
| | `MinterNativeAccount` | VARCHAR(64) | 发行token合约的native account |
| | `AirDropTokenAccount` | VARCHAR(64) | 空投token的token account |
| | `AirDropNativeAccount` | VARCHAR(64) | 空投token的native account |
| | `RewardTokenAccount` | VARCHAR(64) | 奖励token的token account |
| | `RewardNativeAccount` | VARCHAR(64) | 奖励token的native account |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_fee_tolerance` | `Brand` | VARCHAR(64) | 项目方 |
| | `TokenSymbol` | VARCHAR(64) | token符号 |
| | `MaxLessRate` | NUMERIC(5,2) | 允许的最大误差率 |
| `t_reward_exclude` | `Chain` | VARCHAR(64) | 链 |
| | `Account` | VARCHAR(64) | 地址 |
| | `Brand` | VARCHAR(64) | 项目方 |
| | `Remark` | VARCHAR(128) | 备注 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_sol_native_account_info` | `RecordId` | ULID | 记录Id |
| | `Brand` | VARCHAR(64) | 项目方 |
| | `TokenSymbol` | VARCHAR(64) | 币种符号 |
| | `NativeAccount` | VARCHAR(64) | native account |
| | `TokenAccount` | VARCHAR(64) | token account |
| | `InviteCode` | VARCHAR(8) | 邀请码 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_sol_fee_statistics` | `RecordId` | ULID | 记录Id |
| | `Slot` | BIGINT | 区块所在的slot |
| | `TransactionIndex` | INT | 交易在区块当中的index |
| | `BlockHash` | VARCHAR(64) | 区块哈希 |
| | `TransactionId` | VARCHAR(128) | 交易Id |
| | `ComputeUnitPrice` | NUMERIC(78,0) | ComputeUnitPrice指令 |
| | `ComputeUnitLimit` | NUMERIC(78,0) | ComputeUnitLimit指令 |
| | `UnitsConsumed` | NUMERIC(78,0) | 交易消耗的ComputeUnit |
| | `Fee` | NUMERIC(78,0) | 交易最终消耗的手续费 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_sol_qn_fee` | `Id` | INT | 主键 |
| | `Slot` | BIGINT | 统计的slot |
| | `LowAvg` | NUMERIC(78,0) | P50-P60 |
| | `MediumAvg` | NUMERIC(78,0) | P65-P85 |
| | `HighAvg` | NUMERIC(78,0) | P90-P95 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_sol_determine_invite_record` | `RecordId` | ULID | 记录Id |
| | `Brand` | VARCHAR(64) | 项目方 |
| | `TokenSymbol` | VARCHAR(64) | token符号 |
| | `InviterTokenAccount` | VARCHAR(64) | 邀请人token account |
| | `InviterNativeAccount` | VARCHAR(64) | 邀请人native account |
| | `InviteeTokenAccount` | VARCHAR(64) | 被邀请人token account |
| | `InviteeNativeAccount` | VARCHAR(64) | 被邀请人native account |
| | `TransferTxId` | VARCHAR(128) | 交易ID |
| | `InviteChannel` | VARCHAR(64) | 邀请关系确定渠道 |
| | `Level` | INT | 邀请人等级 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_sol_fund_flow` | `RecordId` | ULID | 记录Id |
| | `Brand` | VARCHAR(64) | 项目方 |
| | `TokenSymbol` | VARCHAR(64) | token符号 |
| | `IsToken` | BOOL | 是否是token |
| | `FromNativeAccount` | VARCHAR(64) | 流水发起地址 |
| | `ToNativeAccount` | VARCHAR(64) | 流水接收地址 |
| | `TxId` | VARCHAR(128) | 链上交易Id |
| | `Direction` | INT | 流水方向 0.入账 1.出账 |
| | `ServiceType` | INT | 业务类型 |
| | `FlowType` | INT | 流水类型 |
| | `Decimals` | INT | 资金精度 |
| | `Amount` | NUMERIC(78,0) | 资金金额 |
| | `CreateTime` | TIMESTAMP | 创建时间 |
| | `UpdateTime` | TIMESTAMP | 更新时间 |
| `t_user_wallet_rpc_config` | `RecordId` | ULID | 记录Id |
| | `Chain` | VARCHAR(1024) | 链 |
| | `RpcUrl` | VARCHAR(1024) | rpc链接地址 |
| | `WssUrl` | VARCHAR(1024) | websocket链接地址 |
| | `CreateTime` | TIMESTAMP | 创建时间 |

### 1.2 新工程表结构（oshit-go/base）

| 表名 | 字段名（下划线） | 类型 | 说明 | 对应旧表字段 |
|------|-----------------|------|------|-------------|
| `t_system_config` | `env` | INTEGER | 环境 | `Env` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |
| | `updated_at` | TIMESTAMP | 更新时间 | `UpdateTime` |
| `t_chain_config` | `chain` | VARCHAR(64) | 链 | `Chain` |
| | `rpc_url` | VARCHAR(1024) | rpc链接地址 | `RpcUrl` |
| | `wss_url` | VARCHAR(1024) | websocket链接地址 | `WssUrl` |
| | `decimal` | INTEGER | 原生币种的进制 | `Decimal` |
| | `symbol` | VARCHAR(64) | 原生币的符号 | `Symbol` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |
| | `updated_at` | TIMESTAMP | 更新时间 | `UpdateTime` |
| `t_token_config` | `name` | VARCHAR(64) | token名称 | `Brand` |
| | `symbol` | VARCHAR(64) | token符号 | `TokenSymbol` |
| | `decimal` | INTEGER | token金额的进制 | `Decimal` |
| | `mint` | VARCHAR(64) | token合约地址 | `TokenMintAccount` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |
| | `updated_at` | TIMESTAMP | 更新时间 | `UpdateTime` |
| `t_fee_tolerance` | `max_less_rate` | NUMERIC(5,2) | 允许的最大误差率 | `MaxLessRate` |
| `t_native_account_info` | `record_id` | ULID | 记录Id | `RecordId` |
| | `native_account` | VARCHAR(64) | native account | `NativeAccount` |
| | `token_account` | VARCHAR(64) | token account | `TokenAccount` |
| | `invite_code` | VARCHAR(16) | 邀请码 | `InviteCode` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |
| | `updated_at` | TIMESTAMP | 更新时间 | `UpdateTime` |
| `t_fee_statistics` | `record_id` | ULID | 记录Id | `RecordId` |
| | `slot` | BIGINT | 区块所在的slot | `Slot` |
| | `transaction_index` | INTEGER | 交易在区块当中的index | `TransactionIndex` |
| | `block_hash` | VARCHAR(64) | 区块哈希 | `BlockHash` |
| | `transaction_id` | VARCHAR(128) | 交易Id | `TransactionId` |
| | `compute_unit_price` | NUMERIC(78,0) | ComputeUnitPrice指令 | `ComputeUnitPrice` |
| | `compute_unit_limit` | NUMERIC(78,0) | ComputeUnitLimit指令 | `ComputeUnitLimit` |
| | `units_consumed` | NUMERIC(78,0) | 交易消耗的ComputeUnit | `UnitsConsumed` |
| | `fee` | NUMERIC(78,0) | 交易最终消耗的手续费 | `Fee` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |
| | `updated_at` | TIMESTAMP | 更新时间 | `UpdateTime` |
| `t_qn_fee` | `id` | INTEGER | 主键 | `Id` |
| | `slot` | BIGINT | 统计的slot | `Slot` |
| | `low_avg` | NUMERIC(78,0) | P50-P60 | `LowAvg` |
| | `medium_avg` | NUMERIC(78,0) | P65-P85 | `MediumAvg` |
| | `high_avg` | NUMERIC(78,0) | P90-P95 | `HighAvg` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |
| | `updated_at` | TIMESTAMP | 更新时间 | `UpdateTime` |
| `t_user_wallet_rpc_config` | `record_id` | ULID | 记录Id | `RecordId` |
| | `chain` | VARCHAR(1024) | 链 | `Chain` |
| | `rpc_url` | VARCHAR(1024) | rpc链接地址 | `RpcUrl` |
| | `wss_url` | VARCHAR(1024) | websocket链接地址 | `WssUrl` |
| | `created_at` | TIMESTAMP | 创建时间 | `CreateTime` |

### 1.3 新增表（新工程特有）

| 表名 | 字段名 | 类型 | 说明 |
|------|--------|------|------|
| `t_service_info` | `service` | VARCHAR(64) | 业务服务名称 |
| | `native_account` | VARCHAR(64) | 原生Solana地址 |
| | `pda_account` | VARCHAR(64) | PDA地址（通常是token_account） |
| | `webhook` | VARCHAR(1024) | Webhook回调URL |
| | `mq_group` | VARCHAR(64) | Kafka消费者组（可选） |
| | `mq_topic` | VARCHAR(64) | Kafka主题 |
| | `hook_type` | INTEGER | 通知类型: 0=Kafka, 1=Webhook |
| | `created_at` | TIMESTAMP | 创建时间 |
| | `updated_at` | TIMESTAMP | 更新时间 |
| `t_tx_scan_info` | `service` | VARCHAR(64) | 业务服务名称 |
| | `native_account` | VARCHAR(64) | 原生Solana地址 |
| | `pda_account` | VARCHAR(64) | PDA地址（通常是token_account） |
| | `until_tx_id` | VARCHAR(128) | 扫描截止的交易ID |
| | `before_tx_id` | VARCHAR(128) | 扫描起始的交易ID（可选） |
| | `slot` | NUMERIC(78,0) | 最后扫描的slot |
| `t_service_tx` | `record_id` | ULID | 记录ID（ULID） |
| | `service` | VARCHAR(64) | 业务服务名称 |
| | `tx_id` | VARCHAR(128) | 交易ID |
| | `state` | INTEGER | 交易状态: 0=初始化, <0=失败, >0=成功 |
| | `retry_count` | INTEGER | 重试次数 |
| | `next_retry_time` | TIMESTAMP | 下次重试时间 |
| | `max_retries` | INTEGER | 最大重试次数 |
| | `created_at` | TIMESTAMP | 创建时间 |
| | `updated_at` | TIMESTAMP | 更新时间 |

### 1.4 表结构主要变化

1. **命名规范**：
   - 旧工程：驼峰命名法（如 `CreateTime`, `NativeAccount`）
   - 新工程：小写下划线命名法（如 `created_at`, `native_account`）

2. **字段简化**：
   - `t_token_config`表移除了多个账户字段（`MinterNativeAccount`, `AirDropTokenAccount`, `RewardTokenAccount`等）
   - `t_native_account_info`表移除了`Brand`和`TokenSymbol`字段
   - `t_fee_tolerance`表移除了`Brand`和`TokenSymbol`字段，改为全局配置

3. **新增功能表**：
   - 新增交易扫描相关表：`t_service_info`, `t_tx_scan_info`, `t_service_tx`
   - 支持Kafka消息队列和Webhook通知

4. **移除的表**：
   - `t_micro_service_lb`（微服务网关配置）
   - `t_reward_exclude`（奖励排除地址表）
   - `t_sol_determine_invite_record`（邀请关系表）
   - `t_sol_fund_flow`（资金流水表）
   - `t_special_invite_code`（特殊邀请码表）
   - `t_ecommerce_order`（电商订单表）
   - `t_aws_config`（AWS配置表）

## 二、API接口对比

### 2.1 旧工程API接口（micro-meme/common）

#### 公共路由（PublicRoutes）
| 路径 | 方法 | 处理器 | 参数 |
|------|------|--------|------|
| `/external/api/v1/sol/config/qnStreamHook` | POST | `common_handlers.QnStreamHook` | - |
| `/external/api/v1/sol/config/queryTokenInfo` | POST | `common_handlers.QueryTokenInfo` | `brand`, `tokenSymbol` |
| `/external/api/v1/sol/config/querySOLPriorityFee` | POST | `common_handlers.QuerySOLPriorityFee` | - |
| `/external/api/v1/sol/config/querySOLPriorityFeeOnBlockChain` | POST | `common_handlers.QuerySOLPriorityFeeOnBlockChain` | - |
| `/external/api/v1/sol/config/querySOLConsumedUnits` | POST | `common_handlers.QuerySOLConsumedUnits` | - |
| `/external/api/v1/sol/config/queryTokenQuoteSOLPrice` | POST | `common_handlers.QueryTokenQuoteSOLPrice` | `brand`, `tokenSymbol` |
| `/external/api/v1/sol/config/queryTokenQuoteUSDTPrice` | POST | `common_handlers.QueryTokenQuoteUSDTPrice` | `brand`, `tokenSymbol` |
| `/external/api/v1/sol/config/queryUSDTQuoteSOLPrice` | POST | `common_handlers.QueryUSDTQuoteSOLPrice` | - |
| `/external/api/v1/sol/config/getBirdEyePriceData` | POST | `common_handlers.GetBirdEyePriceData` | - |
| `/external/api/v1/sol/auth/loginWithNativeAccount` | POST | `common_handlers.LoginWithNativeAccount` | `brand`, `tokenSymbol`, `nativeAccount`, `nonce`, `inviteCode`, `sign` |
| `/external/api/v1/sol/invite/queryNativeAccountInfoByInviteCode` | POST | `common_handlers.QueryNativeAccountInfoByInviteCode` | `inviteCode` |
| `/external/api/v1/sol/invite/querySolTransferRewardDistribution` | POST | `common_handlers.QuerySolTransferRewardDistribution` | - |
| `/external/api/v1/sol/invite/querySolTransferRewardClaim` | POST | `common_handlers.QuerySolTransferRewardClaim` | - |
| `/external/api/v1/sol/invite/recursiveQueryClaimsAndInviters` | POST | `common_handlers.RecursiveQueryClaimsAndInviters` | - |
| `/external/api/v1/sol/invite/recursiveQueryUpInviterRecords` | POST | `common_handlers.RecursiveQueryUpInviterRecords` | - |
| `/external/api/v1/sol/invite/recursiveQueryDownInviteeRecords` | POST | `common_handlers.RecursiveQueryDownInviteeRecords` | - |
| `/external/api/v1/sol/invite/queryTokenHoldersNumber` | POST | `common_handlers.QueryTokenHoldersNumber` | - |

#### 私有路由（PrivateRoutes）
| 路径 | 方法 | 处理器 | 参数 |
|------|------|--------|------|
| `/external/api/v1/sol/auth/queryNativeAccountInfo` | POST | `common_handlers.QueryNativeAccountInfo` | JWT Token |
| `/external/api/v1/sol/config/rpc` | POST | `common_handlers.SolanaRpc` | RPC请求参数 |

#### 内部路由（InternalRoutes）
| 路径 | 方法 | 处理器 | 参数 |
|------|------|--------|------|
| `/internal/api/v1/sol/config/queryTokenQuoteSOLPrice` | POST | `common_handlers.QueryTokenQuoteSOLPrice` | `brand`, `tokenSymbol` |
| `/internal/api/v1/sol/config/queryTokenQuoteUSDTPrice` | POST | `common_handlers.QueryTokenQuoteUSDTPrice` | `brand`, `tokenSymbol` |
| `/internal/api/v1/sol/config/queryUSDTQuoteSOLPrice` | POST | `common_handlers.QueryUSDTQuoteSOLPrice` | - |
| `/internal/api/v1/sol/config/querySOLPriorityFee` | POST | `common_handlers.QuerySOLPriorityFee` | - |
| `/internal/api/v1/sol/config/querySOLPriorityFeeOnBlockChain` | POST | `common_handlers.QuerySOLPriorityFeeOnBlockChain` | - |
| `/internal/api/v1/sol/config/querySOLConsumedUnits` | POST | `common_handlers.QuerySOLConsumedUnits` | - |
| `/internal/api/v1/sol/config/queryHackerByNativeAccount` | POST | `common_handlers.QueryHackerByNativeAccount` | - |
| `/internal/api/v1/sol/config/queryFeeTolerance` | POST | `common_handlers.QueryFeeTolerance` | - |
| `/internal/api/v1/sol/config/queryRewardExcludeAccount` | POST | `common_handlers.QueryRewardExcludeAccount` | - |
| `/internal/api/v1/sol/invite/queryNativeAccountInfoByInviteCode` | POST | `common_handlers.QueryNativeAccountInfoByInviteCode` | `inviteCode` |
| `/internal/api/v1/sol/invite/queryInviteRecordExistByToNativeAccount` | POST | `common_handlers.QueryInviteRecordExistByToNativeAccount` | - |
| `/internal/api/v1/sol/invite/recursiveQueryUpInviterRecords` | POST | `common_handlers.InternalRecursiveQueryUpInviterRecords` | - |
| `/internal/api/v1/sol/invite/recursiveQueryDownInviteeRecords` | POST | `common_handlers.InternalRecursiveQueryDownInviteeRecords` | - |
| `/internal/api/v1/sol/invite/recursiveQueryUpInviteRecordAndClaim` | POST | `common_handlers.InternalRecursiveQueryUpInviteRecordAndClaim` | - |
| `/internal/api/v1/sol/invite/recordDetermineInvitationHierarchy` | POST | `common_handlers.InternalRecordDetermineInvitationHierarchy` | - |

### 2.2 新工程API接口（oshit-go/base）

#### 认证路由（/api/auth）
| 路径 | 方法 | 处理器 | 请求参数（JSON） |
|------|------|--------|-----------------|
| `/api/auth/login` | POST | `authHandler.Login` | `{"brand": "string", "symbol": "string", "account": "string", "sign": "string", "nonce": "uint64", "invite_code": "string"}` |
| `/api/auth/info` | POST | `authHandler.QueryNativeAccountInfo` | JWT Token（TODO） |

#### 配置路由（/api/config）
| 路径 | 方法 | 处理器 | 请求参数 |
|------|------|--------|---------|
| `/api/config/token` | POST | `configHandler.GetTokenInfo` | `{"brand": "string", "symbol": "string"}` |
| `/api/config/fee-tolerance` | GET | `configHandler.GetFeeTolerance` | - |

#### 手续费路由（/api/fee）
| 路径 | 方法 | 处理器 | 请求参数 |
|------|------|--------|---------|
| `/api/fee/priority` | GET | `feeHandler.GetPriorityFee` | - |
| `/api/fee/priority/on-chain` | GET | `feeHandler.GetPriorityFeeOnBlockchain` | - |
| `/api/fee/compute-units` | GET | `feeHandler.GetComputeUnitConsumed` | - |

#### 价格路由（/api/price）
| 路径 | 方法 | 处理器 | 请求参数 |
|------|------|--------|---------|
| `/api/price/token/sol` | GET | `priceHandler.GetTokenQuoteSOLPrice` | Query: `brand`, `symbol` |
| `/api/price/token/usdt` | GET | `priceHandler.GetTokenQuoteUSDTPrice` | Query: `brand`, `symbol` |
| `/api/price/usdt/sol` | GET | `priceHandler.GetUSDTQuoteSOLPrice` | - |
| `/api/price/birdeye` | POST | `priceHandler.GetBirdEyePrice` | `{"interval": "string"}` |

#### 邀请路由（/api/invite）
| 路径 | 方法 | 处理器 | 请求参数 |
|------|------|--------|---------|
| `/api/invite/account-by-code` | POST | `inviteHandler.GetAccountByInviteCode` | `{"invite_code": "string"}` |
| `/api/invite/check-record` | POST | `inviteHandler.CheckInviteRecord` | `{"native_account": "string"}` |
| `/api/invite/up-records` | POST | `inviteHandler.GetUpInviterRecords` | `{"depth": "int", "native_account": "string"}` |
| `/api/invite/down-records` | POST | `inviteHandler.GetDownInviteeRecords` | `{"depth": "int", "native_account": "string"}` |
| `/api/invite/reward-distribution` | GET | `inviteHandler.GetRewardDistribution` | - |
| `/api/invite/reward-claims` | GET | `inviteHandler.GetRewardClaims` | - |
| `/api/invite/token-holders` | GET | `inviteHandler.GetTokenHolders` | - |

#### 服务注册路由（/api/service）
| 路径 | 方法 | 处理器 | 请求参数 |
|------|------|--------|---------|
| `/api/service/register` | POST | `serviceHandler.RegisterService` | TODO |
| `/api/service/update` | POST | `serviceHandler.UpdateService` | TODO |
| `/api/service/query` | POST | `serviceHandler.QueryService` | TODO |

### 2.3 API接口主要变化

1. **URL结构简化**：
   - 旧工程：`/external/api/v1/sol/config/queryTokenInfo`
   - 新工程：`/api/config/token`

2. **HTTP方法规范化**：
   - 旧工程：全部使用POST方法
   - 新工程：根据语义使用GET/POST方法
     - 查询类接口：GET方法
     - 操作类接口：POST方法

3. **参数传递方式**：
   - 旧工程：Form表单参数
   - 新工程：JSON请求体（POST）或Query参数（GET）

4. **接口合并与重构**：
   - 合并了多个查询接口
   - 移除了内部路由，统一为公共API
   - 新增服务注册相关接口

5. **响应格式**：
   - 旧工程：使用`response.FailWithError`等封装
   - 新工程：直接返回JSON响应

## 三、数据迁移指南

### 3.1 表数据迁移脚本

```sql
-- 1. 系统配置表
INSERT INTO public.t_system_config (env, created_at, updated_at)
SELECT "Env", "CreateTime", "UpdateTime" FROM "public"."t_system_config";

-- 2. 链配置表
INSERT INTO public.t_chain_config (chain, rpc_url, wss_url, decimal, symbol, created_at, updated_at)
SELECT "Chain", "RpcUrl", "WssUrl", "Decimal", "Symbol", "CreateTime", "UpdateTime" FROM "public"."t_chain_config";

-- 3. Token配置表（需要转换）
INSERT INTO public.t_token_config (name, symbol, decimal, mint, created_at, updated_at)
SELECT "Brand", "TokenSymbol", "Decimal", "TokenMintAccount", "CreateTime", "UpdateTime" 
FROM "public"."t_sol_token_config";

-- 4. 手续费容错表（简化处理）
INSERT INTO public.t_fee_tolerance (max_less_rate)
SELECT "MaxLessRate" FROM "public"."t_fee_tolerance" LIMIT 1;

-- 5. 原生账户信息表（移除brand和token_symbol字段）
INSERT INTO public.t_native_account_info (record_id, native_account, token_account, invite_code, created_at, updated_at)
SELECT "RecordId", "NativeAccount", "TokenAccount", "InviteCode", "CreateTime", "UpdateTime" 
FROM "public"."t_sol_native_account_info";

-- 6. 手续费统计表
INSERT INTO public.t_fee_statistics (record_id, slot, transaction_index, block_hash, transaction_id, compute_unit_price, compute_unit_limit, units_consumed, fee, created_at, updated_at)
SELECT "RecordId", "Slot", "TransactionIndex", "BlockHash", "TransactionId", "ComputeUnitPrice", "ComputeUnitLimit", "UnitsConsumed", "Fee", "CreateTime", "UpdateTime" 
FROM "public"."t_sol_fee_statistics";

-- 7. QN手续费表
INSERT INTO public.t_qn_fee (id, slot, low_avg, medium_avg, high_avg, created_at, updated_at)
SELECT "Id", "Slot", "LowAvg", "MediumAvg", "HighAvg", "CreateTime", "UpdateTime" 
FROM "public"."t_sol_qn_fee";

-- 8. 用户钱包RPC配置表
INSERT INTO public.t_user_wallet_rpc_config (record_id, chain, rpc_url, wss_url, created_at)
SELECT "RecordId", "Chain", "RpcUrl", "WssUrl", "CreateTime" 
FROM "public"."t_user_wallet_rpc_config";
```

### 3.2 注意事项

1. **字段映射**：
   - 注意驼峰命名转下划线命名
   - 注意字段类型匹配

2. **数据丢失**：
   - 以下表的数据不会迁移，需要业务方重新配置：
     - `t_reward_exclude`（奖励排除地址）
     - `t_sol_determine_invite_record`（邀请关系）
     - `t_sol_fund_flow`（资金流水）
     - `t_special_invite_code`（特殊邀请码）

3. **新增表初始化**：
   - `t_service_info`：需要业务方注册服务
   - `t_tx_scan_info`：交易扫描任务会自动创建
   - `t_service_tx`：交易处理表会自动使用

## 四、前端对接指南

### 4.1 API基础信息

- **Base URL**: `http://localhost:8888`（开发环境）
- **API前缀**: `/api`
- **Content-Type**: `application/json`

### 4.2 主要接口变更

#### 4.2.1 登录接口

**旧接口**：
```javascript
POST /external/api/v1/sol/auth/loginWithNativeAccount
Content-Type: application/x-www-form-urlencoded

brand=xxx&tokenSymbol=xxx&nativeAccount=xxx&nonce=123&inviteCode=xxx&sign=xxx
```

**新接口**：
```javascript
POST /api/auth/login
Content-Type: application/json

{
  "brand": "xxx",
  "symbol": "xxx",
  "account": "xxx",
  "sign": "xxx",
  "nonce": 123,
  "invite_code": "xxx"
}
```

#### 4.2.2 查询Token信息

**旧接口**：
```javascript
POST /external/api/v1/sol/config/queryTokenInfo
Content-Type: application/x-www-form-urlencoded

brand=xxx&tokenSymbol=xxx
```

**新接口**：
```javascript
POST /api/config/token
Content-Type: application/json

{
  "brand": "xxx",
  "symbol": "xxx"
}
```

#### 4.2.3 查询手续费

**旧接口**：
```javascript
POST /external/api/v1/sol/config/querySOLPriorityFee
```

**新接口**：
```javascript
GET /api/fee/priority
```

#### 4.2.4 查询价格

**旧接口**：
```javascript
POST /external/api/v1/sol/config/queryTokenQuoteSOLPrice
Content-Type: application/x-www-form-urlencoded

brand=xxx&tokenSymbol=xxx
```

**新接口**：
```javascript
GET /api/price/token/sol?brand=xxx&symbol=xxx
```

### 4.3 错误处理

**旧响应格式**：
```json
{
  "error": true,
  "msg": "error message"
}
```

**新响应格式**：
```json
{
  "error": "error message"
}
```

或成功响应：
```json
{
  "data": {...}
}
```

### 4.4 前端适配建议

1. **创建API客户端**：
   ```javascript
   class BaseAPI {
     constructor(baseURL = 'http://localhost:8888') {
       this.baseURL = baseURL;
     }
     
     async request(endpoint, method = 'GET', data = null) {
       const url = `${this.baseURL}/api${endpoint}`;
       const options = {
         method,
         headers: {
           'Content-Type': 'application/json',
         },
       };
       
       if (data) {
         if (method === 'GET') {
           // 处理GET请求的query参数
           const params = new URLSearchParams(data).toString();
           url = `${url}?${params}`;
         } else {
           options.body = JSON.stringify(data);
         }
       }
       
       const response = await fetch(url, options);
       return response.json();
     }
     
     // 登录
     async login(data) {
       return this.request('/auth/login', 'POST', data);
     }
     
     // 获取token信息
     async getTokenInfo(brand, symbol) {
       return this.request('/config/token', 'POST', { brand, symbol });
     }
     
     // 获取手续费
     async getPriorityFee() {
       return this.request('/fee/priority', 'GET');
     }
     
     // 获取价格
     async getTokenQuoteSOLPrice(brand, symbol) {
       return this.request(`/price/token/sol?brand=${brand}&symbol=${symbol}`, 'GET');
     }
   }
   ```

2. **更新所有API调用**：
   - 更新URL路径
   - 更新参数传递方式（Form → JSON）
   - 更新HTTP方法（POST → GET 对于查询接口）
   - 更新错误处理逻辑

3. **测试迁移**：
   - 先在小范围测试新接口
   - 逐步替换旧接口调用
   - 监控错误日志

## 五、总结

本次迁移主要变化：

1. **表结构**：
   - 命名规范：驼峰 → 下划线
   - 字段简化：移除冗余字段
   - 新增功能：交易扫描相关表

2. **API接口**：
   - URL简化：去除`/external/api/v1/sol`前缀
   - 方法规范化：合理使用GET/POST
   - 参数标准化：统一使用JSON格式

3. **架构改进**：
   - 消息队列：RocketMQ → Kafka
   - 模块化设计：清晰的handler-logic分层
   - 错误处理：统一的错误响应格式

建议按照以下步骤进行迁移：
1. 先执行数据迁移脚本
2. 前端逐步适配新API
3. 测试所有核心功能
4. 监控系统运行状态