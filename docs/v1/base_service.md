# oshit-go / base 服务参考文档

> 面向 Claude Code 使用，涵盖 base 服务的所有关键细节。
> 技术栈、通用工具等见 `docs/overview.md`；Reward 服务见 `docs/reward_service.md`。

---

## 1. 服务概览

| 项目 | 值 |
|---|---|
| 服务名 | `base-api` |
| 监听端口 | `1100`（HTTP Fiber）/ `20880`（Dubbo Triple） |
| 工作目录（开发） | `app/base/api`（GoLand Run Configuration 须设置此路径，否则 `./etc/base.yaml` 读不到 dubbo 配置） |
| 配置文件 | `app/base/api/etc/base.yaml` |

---

## 2. 目录结构

```
app/base/api/
├── base.go                      # main 入口
├── etc/
│   └── base.yaml                # 应用配置（db/redis/kafka/dubbo）
├── internal/
│   ├── config/config.go         # 配置结构体 + viper 加载
│   ├── context/context.go       # CoreContext（所有服务共享的上下文字段）
│   ├── svc/
│   │   ├── context.go           # ServiceContext 初始化（启动流程入口）
│   │   └── kafka.go             # Kafka producer 初始化
│   ├── handler/
│   │   ├── routes.go            # 路由注册
│   │   ├── auth.go / fee.go / info.go / price.go / invite.go
│   ├── logic/
│   │   ├── auth.go              # 登录 + 注册逻辑
│   │   ├── fee.go               # 手续费查询
│   │   ├── info.go              # token 信息
│   │   ├── price.go             # 价格查询
│   │   ├── invite.go            # 邀请关系查询（含递归 CTE SQL）
│   │   └── tx.go                # SendTransaction（签名 + 广播 + 写 t_service_tx）
│   ├── server/
│   │   ├── dubbo.go             # Dubbo Triple 服务启动
│   │   └── server.go            # BaseRpcService（实现 dubbo handler 接口，委托 logic）
│   ├── task/
│   │   ├── context.go           # TaskContext + runPeriodic / runPeriodicWithWatchdog
│   │   ├── mgr.go               # TaskManager.StartAllTasks()
│   │   ├── fee.go               # FeeTask（4 个子任务）
│   │   ├── unit.go              # UnitTask（compute unit 模拟）
│   │   ├── price.go             # PriceTask（Raydium 价格）
│   │   ├── kline.go             # KLineTask（K 线爬虫）
│   │   ├── holders.go           # HoldersTask（token 持有者数量）
│   │   ├── scan.go              # TxScanTask（链上交易扫描，最核心）
│   │   ├── expire.go            # TxExpireTask（超时交易处理）
│   │   └── browser.go           # 浏览器辅助（供 KLineTask 使用）
│   └── types/
│       ├── types.go             # 通用数据结构
│       └── service.go           # HTTP 请求/响应类型
```

---

## 3. 启动流程（`svc/context.go`）

```
main()
 └── svc.NewServiceContext()
      ├── config.LoadConfig()           → viper 读 ./etc/base.yaml
      ├── initDatabase()                → GORM postgres 连接
      ├── initRedis()                   → go-redis UniversalClient（支持 Sentinel）
      ├── redsync.New(pool)             → 分布式锁
      ├── initRSAPublicKey()            → 硬编码 RSA 公钥（用于消息验证）
      ├── initDatabaseConfigs()         → 从 DB 加载全局配置表到 CoreContext：
      │    SystemConfig / ChainConfig / UserWalletRPCConfig
      │    TokenConfig / FeeTolerance / AwsConfig / LightHouseAddress
      ├── initServiceKeys()             → 从 t_service_key 加载加密私钥
      │    解密算法: PBEWithHMACSHA512AndAES_256，密码: fktYimwMl3OfUF3m
      │    存入 CoreContext.ServiceKeyMap[service][subService] = solana.PrivateKey
      ├── initSolanaRPC()               → rpc.New(ChainConfig.RPCURL)，分别初始化
      │    RpcClient（服务端）和 UserWalletRpcClient（给前端用）
      ├── initKafkaProducer()           → kafka-go writer
      └── startTasks()                  → goroutine: TaskManager.StartAllTasks()

main()
 ├── go server.StartDubboServer(svcCtx)   → Dubbo Triple on :20880 + Nacos 注册
 └── handler.RegisterRoutes(app, svcCtx)
     app.Listen(":1100")
```

---

## 4. CoreContext 结构（`internal/context/context.go`）

```go
type CoreContext struct {
    Config  *config.Config
    DB      *gorm.DB
    Redis   redis.UniversalClient
    RedSync redsync.Redsync
    Ctx     context.Context

    // Solana 客户端
    RpcClient           *rpc.Client    // 服务端 RPC（发送交易、读链上数据）
    UserWalletRpcClient *rpc.Client    // 给前端/APP 使用的 RPC
    LightHouseAddress   solana.PublicKey

    // 全局变量
    AppGlobalPublicKey *rsa.PublicKey
    TokenDecimal       float64        // 10^decimals
    KafkaProducer      interface{}    // 实际类型 *kafka.Writer（避免循环依赖）

    // 配置表数据（启动时从 DB 加载，全程只读）
    SystemConfig        *model.SystemConfig
    ChainConfig         *model.ChainConfig
    UserWalletRPCConfig *model.UserWalletRpcConfig
    TokenConfig         *model.TokenConfig
    FeeTolerance        model.FeeTolerance
    AwsConfig           *model.AwsConfig

    // 服务签名私钥
    ServiceKeyMap ServiceKey  // ServiceKeyMap[service][subService] = privateKey
}

type ServiceKey = map[string]map[string]solana.PrivateKey
```

---

## 5. HTTP API（`handler/routes.go`）

所有路由挂载在 `/base` 前缀下。

### 认证
| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| POST | `/base/auth/login` | `AuthHandler.Login` | Solana 消息签名验证登录，首次登录注册地址 |

**login 逻辑（`logic/auth.go`）**

1. 校验 `nativeAccount`（base58 PublicKey）、`sign`（base58 Signature）
2. 拼接待验证消息：
   - 无邀请码：`"I am login {Brand} for token {Symbol} with my address {Account} with nonce {Nonce}"`
   - 有邀请码：末尾追加 `inviteCode {InviteCode}`
3. `nativeAccount.Verify([]byte(msg), sign)` 验签
4. 查 `t_native_account_info`，不存在则调用 `RegisterNativeAccount`：
   - 计算 `tokenAccount = FindAssociatedTokenAddress(nativeAccount, mint)`
   - 生成随机 8 位 `inviteCode`
   - 如传入 `inviteCode`，写入 `t_invite_relation`（channel="InviteCode"，level 自动递增）
5. 生成 JWT token 返回（`record_id` 作为 subject）

### 信息
| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| GET | `/base/info/token-info` | `InfoHandler.GetTokenInfo` | 读 `t_token_config`，返回 name/symbol/decimals/mint |
| GET | `/base/info/token-holders` | `InfoHandler.GetTokenHolders` | token 持有者数量 |

### 手续费
| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| GET | `/base/fee/priority` | `FeeHandler.GetPriorityFee` | 从 Redis 读 4 档优先费（Low/Medium/High/Extreme），分 PerComputeUnit / PerTransaction 两组 |
| GET | `/base/fee/inst-units` | `FeeHandler.GetInstUnits` | 从 Redis 读各指令 compute unit 消耗量 |
| GET | `/base/fee/fee-tolerance` | `FeeHandler.GetFeeTolerance` | 读 `t_fee_tolerance.max_less_rate` |

### 价格
| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| GET | `/base/price/token/sol` | `PriceHandler.GetTokenQuoteSOLPrice` | token/SOL 价格（Redis） |
| GET | `/base/price/token/usdt` | `PriceHandler.GetTokenQuoteUSDTPrice` | token/USDT 价格（Redis） |
| GET | `/base/price/usdt/sol` | `PriceHandler.GetUSDTQuoteSOLPrice` | USDT/SOL 价格（Redis） |
| POST | `/base/price/kline` | `PriceHandler.GetBirdEyePrice` | token K 线数据（Redis） |

### 邀请关系
| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| POST | `/base/invite/account-by-code` | `InviteHandler.GetAccountByInviteCode` | 根据 `inviteCode` 查 `t_native_account_info` |
| POST | `/base/invite/check-record` | `InviteHandler.CheckInviteRecord` | 查 `nativeAccount` 是否作为被邀请人存在于 `t_invite_relation` |
| POST | `/base/invite/up-records` | `InviteHandler.GetUpInviterRecords` | 递归向上查邀请链（PostgreSQL CTE，`WITH RECURSIVE`，最大 depth=20） |
| POST | `/base/invite/down-records` | `InviteHandler.GetDownInviteeRecords` | 递归向下查被邀请链（同上） |

---

## 6. Dubbo Triple API（`server/server.go`）

服务接口：`base.BaseService`（proto 定义在 `app/pb/base/`）
注册中心：Nacos `127.0.0.1:8848`，协议：Triple

所有方法均委托给对应 `logic` 层实现：

| 方法 | 对应 logic | 返回 |
|---|---|---|
| `GetFeeTolerance` | `FeeLogic.GetFeeTolerance` | `MaxFeeLess float64` |
| `GetPriorityFee` | `FeeLogic.GetPriorityFee` | `PerComputeUnit / PerTransaction FeeDetail{Low,Medium,High,Extreme uint64}` |
| `GetInstUnits` | `FeeLogic.GetInstUnits` | `MiniRent / AssociatedAccount / TransferChecked / Memo uint64` |
| `GetTokenQuoteSOLPrice` | `PriceLogic.GetTokenQuoteSOLPrice` | `Price float64` |
| `GetTokenQuoteUSDTPrice` | `PriceLogic.GetTokenQuoteUSDTPrice` | `Price float64` |
| `GetUSDTQuoteSOLPrice` | `PriceLogic.GetUSDTQuoteSOLPrice` | `Price float64` |
| `SendTransaction` | `TxLogic.SendTransaction` | `RecordId / TxId string` |

### SendTransaction 完整流程（`logic/tx.go`）

```
1. base64 解码 EncodedTx → []byte
2. TransactionFromDecoder() → solana.Transaction
3. 从 ServiceKeyMap[Service][SubService] 取私钥
4. tx.Message.MarshalBinary() → privateKey.Sign() → tx.Signatures[1] = sig
   （索引0 = 用户签名，索引1 = 服务签名）
5. txID = tx.Signatures[0].String()（链上交易 ID 始终是用户签名）
6. 写 t_service_tx: state=0, retry_count=0, max_retries=5
7. goroutine: broadcastTx() → rpcClient.SendTransaction()
   失败则指数退避重试（每轮等 retry_count * 30s），超过 max_retries → state=-1
8. 同步返回 record_id 和 tx_id
```

---

## 7. 后台任务（`task/`）

所有周期任务使用统一工具函数，内置分布式锁防止多实例重复执行：

```go
// 常规任务（执行时间短）
runPeriodic(rs, interval, lockKey, lockTTL, fn)
// 长任务（执行时间不可预测，使用 watchdog 自动续期锁）
runPeriodicWithWatchdog(rs, interval, lockKey, initTTL, fn)
```

### 7.1 FeeTask（`task/fee.go`）

4 个子任务，各自独立的 redsync 锁，每 5s 触发，锁 TTL 2min：

| 子任务 | 锁 key | 说明 |
|---|---|---|
| `readPriorityFee` | `base:sol:fee:stat-chain-fee:lock` | 调用 `getBlock` RPC 获取最新 slot 区块，解析每笔交易的 `SetComputeUnitPrice` / `SetComputeUnitLimit` 指令，写入 `t_fee_statistics`（上限 10000 条，超出则滚动覆盖最旧记录） |
| `updatePerUnitFee` | `base:sol:fee:update-per-unit:lock` | 对 `t_fee_statistics` 按 `compute_unit_price` 4 分位（NTILE）加权平均，结果写 Redis `base:sol:fee:priority-per-unit` |
| `updatePerTxFee` | `base:sol:fee:update-per-tx:lock` | 按百分比排名（PERCENT_RANK）分 4 组，取中位数（PERCENTILE_CONT 0.5），写 Redis `base:sol:fee:priority-per-tx` |
| `estimateWeightAvgFee` | `base:sol:fee:est-weight-avg:lock` | 调用 QuickNode `qn_estimatePriorityFees` API，写 `t_qn_fee`（滚动 20 条），再均值写 Redis |

Redis key 格式：
- `base:sol:fee:priority-per-unit` / `base:sol:fee:priority-per-tx`
- `base:sol:fee:weight-avg-per-unit` / `base:sol:fee:weight-avg-per-tx`

Redis 值为 `entity.FeeDetail` 的 JSON：`{Low, Medium, High, Extreme uint64}`

### 7.2 UnitTask（`task/unit.go`）

构造典型 Solana 交易（TransferChecked / Memo / AssociatedAccount / MiniRent）并模拟执行，统计各指令类型实际消耗的 compute unit，写入 Redis。

### 7.3 PriceTask（`task/price.go`）

周期性请求 Raydium API，获取 token/SOL、token/USDT、USDT/SOL 价格，写入 Redis。

### 7.4 KLineTask（`task/kline.go`）

使用 `runPeriodicWithWatchdog`，通过浏览器（`browser.go`）爬取 Raydium token K 线数据，写入 Redis。

### 7.5 HoldersTask（`task/holders.go`）

周期统计 token 持有者数量，写入 Redis。

### 7.6 TxScanTask（`task/scan.go`）⭐ 最核心任务

**职责**：扫描链上与各业务地址相关的新交易，找到后发 Kafka 通知 reward 等下游服务。

**初始化**：读取 `t_tx_scan_info` 表全部记录，每条记录启动独立 goroutine。

**扫描循环**（每 3s 一轮）：
```
1. 获取 redsync 分布式锁（base:sol:tx-scan:{service}-{subService}:lock，WithTries(1)）
2. DB 事务 + FOR UPDATE NOWAIT 行级锁定 t_tx_scan_info 记录
3. 调用 getSignaturesForAddress(pdaAccount, until=untilTxId, limit=300, Finalized)
4. 按 slot 降序排序，取最新交易
5. 条件更新 t_tx_scan_info.until_tx_id（仅当 slot > 当前记录 slot 时更新）
6. 提交 DB 事务（释放行锁）
7. 对 slot > 旧 slot 的每笔交易，goroutine: handleServiceTx()
```

**handleServiceTx 流程**：
```
1. Redis 分布式限流：AcquireDistributedRateLimit("handleServiceTx", 200/s)
2. redsync 锁：base:sol:tx-handle:{txSig}:lock（防止重复处理同一交易）
3. GetTransactionResult（指数退避重试，最多 5 次，处理 ErrNotFound + RPC 限流错误）
4. TransactionFromDecoder() → DecodeSolanaTransaction()（解析指令、账户、金额）
5. 查 t_service_tx（txId）：不存在则跳过（非本系统业务交易）
6. MarkTxFetchState(txId, TxFetchSuccess=1)
7. 发送 Kafka 消息：
   Topic: "ServiceTransaction"
   MsgType: "NewScannedTransaction"
   MsgContent: NewScannedTx{Service, SubService, TxSig, DecodedTx}
```

### 7.7 TxExpireTask（`task/expire.go`）

**职责**：兜底任务，处理 TxScanTask 可能遗漏的交易。每 5s，redsync 锁 TTL 2min。

```
1. 查 t_service_tx: WHERE state=0 AND created_at <= NOW()-5min
2. 对每笔记录：
   a. Redis 限流（200/s）+ redsync 锁（HandlePosExpiredTx-{txId}）
   b. GetTransaction(txSig, Finalized)
      - ErrNotFound → MarkTxFetchState(-1) + 发 Kafka "NewExpiredTransaction"
      - 找到 → DecodeSolanaTransaction() + MarkTxFetchState(1) + 发 Kafka "NewScannedTransaction"
      - 其他错误 → 跳过（等下轮重试）
```

---

## 8. Kafka 消息生产格式（`common/pkg/entity/kafka.go`）

Kafka topic：`ServiceTransaction`

### NewScannedTransaction
```json
{
  "MsgType": "NewScannedTransaction",
  "MsgContent": {
    "Service": "Reward",
    "SubService": "TakeToken",
    "TxSig": { "Signature": "...", "Slot": 12345, "Err": null },
    "DecodedTx": { ... }
  }
}
```

### NewExpiredTransaction
```json
{
  "MsgType": "NewExpiredTransaction",
  "MsgContent": {
    "Service": "Reward",
    "SubService": "TakeToken",
    "TxID": "..."
  }
}
```

---

## 9. 数据库表（base 服务）

SQL 定义：`repositories/base_structure.sql`

| 表名 | 主键 | 核心字段 | 用途 |
|---|---|---|---|
| `t_system_config` | — | `env`(0=mainnet,1=testnet) | 全局环境配置 |
| `t_chain_config` | — | `chain`,`rpc_url`,`wss_url`,`decimals`,`symbol` | Solana 链 RPC 配置 |
| `t_user_wallet_rpc_config` | ULID | `chain`,`rpc_url`,`wss_url` | 给前端/APP 的 RPC |
| `t_token_config` | — | `name`,`symbol`,`decimals`,`mint` | 单 token 配置 |
| `t_fee_tolerance` | — | `max_less_rate` | 成本费容错比例 |
| `t_fee_statistics` | ULID | `slot`,`tx_index`,`tx_id`,`price`,`unit_limit`,`units_consumed`,`fee` | 链上手续费统计（上限 10000 条，滚动覆盖） |
| `t_qn_fee` | `id`(int) | `slot`,`low_avg`,`medium_avg`,`high_avg` | QuickNode 费用估算（滚动 20 条，id=slot%20+1） |
| `t_native_account_info` | ULID | `native_account`,`token_account`,`invite_code` | 用户地址注册信息 |
| `t_invite_relation` | ULID | `inviter`,`invitee`,`channel`,`level`,`tx_id` | 邀请关系链（level 由登录时自动推算） |
| `t_service_info` | — | `service`,`sub_service`,`address`,`hook_type`,`mq_group`,`mq_topic` | 业务服务注册（扫描地址、Kafka 配置） |
| `t_service_key` | — | `service`,`sub_service`,`encrypted_key` | 业务服务私钥（jasypt 加密） |
| `t_tx_scan_info` | — | `service`,`sub_service`,`pda_account`,`until_tx_id`,`slot` | 扫描游标（每个业务一行） |
| `t_service_tx` | ULID | `service`,`sub_service`,`tx_id`,`state`(0/1/-1),`retry_count`,`max_retries` | 业务交易生命周期跟踪 |
| `t_fund_flow` | ULID | `is_token`,`from_account`,`to_account`,`tx_id`,`direction`,`service_type`,`flow_type`,`amount` | 资金流水账（唯一索引：tx_id+to_account+flow_type） |

**t_service_tx state 含义**：
- `0` = 初始化（已提交，待链上确认）
- `1` = 已被扫描/确认
- `-1` = 超时/失败

---

## 10. base ↔ reward 交互总览

```
reward 服务
  │
  │  (1) 用户提交交易 → reward 验证后调用 Dubbo
  ▼
base.SendTransaction(EncodedTx, Service="Reward", SubService="TakeToken")
  │  (2) base 用服务私钥签名 tx.Signatures[1]
  │  (3) 写 t_service_tx{txId, state=0}
  │  (4) goroutine: 广播到 Solana RPC
  ▼
Solana 链上确认
  │
TxScanTask（每 3s）
  │  (5) getSignaturesForAddress(reward_token_account)
  │  (6) 发现 txId → 查 t_service_tx → 存在
  │  (7) state → 1，发 Kafka "NewScannedTransaction"
  ▼
reward.KafkaConsumerTask 消费，处理业务逻辑

TxExpireTask（兜底，每 5s 检查 state=0 且超过 5min 的记录）
  │  → ErrNotFound: state → -1，发 Kafka "NewExpiredTransaction"
  │  → 找到: state → 1，发 Kafka "NewScannedTransaction"
```

---

## 11. 注意事项 / 易踩坑点

1. **工作目录**：GoLand 须将工作目录设为 `app/base/api`，否则 dubbo/nacos 配置读不到，启动 panic。

2. **txID 来源**：链上交易 ID 始终是 `tx.Signatures[0]`（用户签名），`Signatures[1]` 是 base 服务签名，仅用于授权。

3. **t_fee_statistics 容量控制**：超过 10000 条时不删除，而是覆盖最旧的 N 条（以 `record_id` 定位更新）。

4. **redsync 锁策略**：
   - 周期任务用 `WithTries(1)`，抢不到直接跳过（不阻塞）
   - `handleServiceTx` 的 per-tx 锁无 `WithTries` 限制，确保同一交易只处理一次

5. **KafkaProducer 类型断言**：`CoreContext.KafkaProducer` 声明为 `interface{}`，使用时须断言为 `*kafka.Writer`（避免循环依赖）。

6. **私钥解密**：jasypt 算法 `PBEWithHMACSHA512AndAES_256`，固定密码 `fktYimwMl3OfUF3m`，实现在 `common/utils/jsypt_util.go`。
