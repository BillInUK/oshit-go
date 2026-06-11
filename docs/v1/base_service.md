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
│   │   ├── auth.go / fee.go / info.go / price.go / invite.go / middleware.go
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
│   │   ├── fee.go               # FeeTask（3 个子任务）
│   │   ├── ttl_cleanup.go      # TTLCleanupTask（TTL 数据清理）
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

## 4. 数据库表

* 数据库表结构在 repositories/structures.sql 里面定义

---

## 5. 核心数据结构

* base 服务的 前端的 api请求和回复 数据结构在 `internal/types/types.go` 里面定义

* 核心上下文

base 服务的核心上下文在 `internal/context/context.go` 中定义

```go
type CoreContext struct {
    Config  *config.Config
    DB      *gorm.DB
    Redis   redis.UniversalClient
    RedSync redsync.Redsync
    Ctx     context.Context

    // 全局变量
    AppGlobalPublicKey  *rsa.PublicKey
    RpcClient           *rpc.Client    // 服务端 RPC（发送交易、读链上数据）
    UserWalletRpcClient *rpc.Client    // 给前端/APP 使用的 RPC
    LightHouseAddress   solana.PublicKey
    TokenDecimal        float64        // 10^decimals
    KafkaProducer       interface{}    // 实际类型 *kafka.Writer（避免循环依赖）
    NacosConfigClient   config_client.IConfigClient

    // 运行配置（启动时先从 DB 兜底加载，再优先用 Nacos 覆盖；部分配置支持 Nacos listener 热更新）
    ConfigMu            sync.RWMutex
    SystemConfig        *model.SystemConfig
    ChainConfig         *model.ChainConfig
    UserWalletRPCConfig *model.UserWalletRpcConfig
    MainnetRPCConfig    *model.MainnetRpcConfig    // 主网 RPC，供 MarketBuyToken 等需要主网的业务使用
    TokenConfig         *model.TokenConfig
    FeeTolerance        model.FeeTolerance
    AwsConfig           *model.AwsConfig

    // 服务签名私钥
    ServiceKeyMap  ServiceKey                          // ServiceKeyMap[service][subService] = privateKey
    // 服务配置（multiSign / confirm 等开关）
    ServiceInfoMap map[string]map[string]model.ServiceInfo  // ServiceInfoMap[service][subService] = ServiceInfo
}
```

* 业务上下文

base 服务的业务上下文在 `internal/svc/context.go` 中定义

```go
type ServiceContext struct {
	core_context.CoreContext
	TaskMgr *task.TaskManager
}
```

业务上下文继承自核心上下文，当然如果有其他需要也可以增加额外的字段或者属性

* 任务上下文

base 服务的任务上下文在 `internal/task/context.go` 中定义

```go
type TaskContext struct {
	core_context.CoreContext
}
```

任务上下文继承核心上下文，当然如果有其他需要也可以增加额外的字段或者属性

---

## 6. 启动流程（`internal/svc/context.go`）

```
main()
 └── svc.NewServiceContext()
      ├── config.LoadConfig()           → viper 读 ./etc/application.yaml
      ├── LoadConfigDecryptKey()        → KMS 信封解密，获取 AES 密钥（仅存内存）
      ├── JasyptDecode(cfg)             → 解密 application.yaml 中的 ENC~ 字段（DB/Redis/Nacos 密码）
      ├── initDatabase()                → GORM postgres 连接
      ├── initRedis()                   → go-redis UniversalClient（支持 Sentinel）
      ├── redsync.New(pool)             → 分布式锁
      ├── initRSAPublicKey()            → 硬编码 RSA 公钥（用于消息验证）
      ├── initDatabaseConfigs()         → 从 DB 兜底加载全局配置表到 CoreContext：
      │    SystemConfig / ChainConfig / UserWalletRPCConfig / MainnetRPCConfig
      │    TokenConfig / FeeTolerance / AwsConfig / LightHouseAddress
      │    ServiceInfoMap（从 t_service_info 加载，二维索引 [service][subService]）
      │    t_rpc_endpoint / t_aws_config 读取后 JasyptDecode 解密 ENC~ 字段
      ├── initServiceKeys()             → 从 DB 兜底加载 t_service_key 加密私钥
      │    解密算法: PBEWithHMACSHA512AndAES_256，密钥来自 KMS
      │    存入 CoreContext.ServiceKeyMap[service][subService] = solana.PrivateKey
      ├── initNacosConfigClient()       → 初始化 Nacos config client
      ├── initNacosRuntimeAndRegistry() → 优先从 Nacos 覆盖 DB 兜底配置：
      │    base-runtime.yaml: chain/token/system/fee_tolerance/aws/lighthouse
      │    base-service-registry.yaml: service_info/service_key/scan 静态注册配置
      │    YAML 解析后 JasyptDecode 解密 ENC~ 字段
      ├── initSolanaRPC()               → rpc.New(ChainConfig.RPCURL)，分别初始化
      │    RpcClient（服务端）和 UserWalletRpcClient（给前端用）
      ├── initKafkaProducer()           → kafka-go writer
      ├── utils.InitDTokenManager()     → 初始化 dtoken JWT 鉴权管理器
      ├── startTasks()                  → TaskManager.StartAllTasks()
      ├── ReconcileScanConfigs()        → 若 Nacos 有 service registry，按 Nacos 静态配置对齐扫描任务
      └── listenNacosConfigs()          → 监听 Nacos runtime / service registry 变更

main()
 ├── go server.StartDubboServer(svcCtx)   → Dubbo Triple on :20880 + Nacos 注册
 └── handler.RegisterRoutes(app, svcCtx)
     app.Listen(":1100")
```

### 6.1 Nacos 配置

base 当前使用两个 Nacos dataId：

| dataId | 说明 | 是否热更新 |
|---|---|---|
| `base-runtime.yaml` | 全局基础配置：`system`、`chain`、`user_wallet_rpc`、`mainnet_rpc`、`token`、`fee_tolerance`、`aws`、`lighthouse_address` | 是。会更新内存配置，并重建 `RpcClient` / `UserWalletRpcClient` |
| `base-service-registry.yaml` | 业务注册配置：`service_info` 静态字段、加密服务私钥、scan 静态字段 | 是。会原子替换 `ServiceInfoMap` / `ServiceKeyMap`，并 reconcile 扫描任务 |

Nacos 加载失败或 dataId 为空时，base 保留 DB 兜底配置，保证本地开发和配置未迁移环境仍可启动。

`base-service-registry.yaml` 中的 `scan.initial_until_tx_id` 和 `scan.initial_slot` 只在 DB 中不存在对应扫描检查点时生效；已有检查点不会被 Nacos 覆盖。

`base-runtime.yaml` 主要用于稳定基础配置。虽然 listener 会更新 ServiceContext 中的配置与 RPC client，但已启动的后台任务在构造时会持有自己的 RPC client，链配置类变更仍建议通过重启服务生效。

---

## 7. HTTP API（`internal/handler/routes.go`）

所有路由挂载在 `/base` 前缀下。

### 认证
| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| POST | `/base/auth/login` | `AuthHandler.Login` | Solana 消息签名验证登录，首次登录注册地址 |
| POST | `/base/auth/info` | `AuthHandler.QueryNativeAccountInfo` | 查询当前登录地址的账户信息（需 JWT） |

**login 逻辑（`logic/auth.go`）**

1. 校验 `nativeAccount`（base58 PublicKey）、`sign`（base58 Signature）
2. 拼接待验证消息：
   - 无邀请码：`"I am login {Brand} for token {Symbol} with my address {Account} with nonce {Nonce}"`
   - 有邀请码：`"I am login {Brand} for token {Symbol} with my address {Account} with nonce {Nonce} inviteCode {InviteCode}"`
3. `nativeAccount.Verify([]byte(msg), sign)` 验签
4. 查 `t_native_account_info`，不存在则调用 `RegisterNativeAccount`：
   - 计算 `tokenAccount = FindAssociatedTokenAddress(nativeAccount, mint)`
   - 生成随机 8 位 `inviteCode`
   - 如传入 `inviteCode`，写入 `t_invite_relation`（channel="InviteCode"，level 自动递增）
5. 调用 `dtoken.Login(ctx, nativeAccount)` 颁发 JWT，以 `nativeAccount` 作为 loginID，返回 token

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
| ~~GET~~ | ~~`/base/fee/fee-tolerance`~~ | ~~`FeeHandler.GetFeeTolerance`~~ | ~~读 `t_fee_tolerance.max_less_rate`~~（已注释，当前未启用） |

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

## 8. Dubbo Triple API（`server/server.go`）

服务接口：`base.BaseService`（proto 定义在 `app/pb/base/`）
注册中心：Nacos `127.0.0.1:8848`，协议：Triple

Dubbo 注册中心使用接口级注册（`registry-type=interface` / `registry.WithRegisterInterface()`），并关闭 registry 作为 config center 与 metadata report 的默认用途。这样 base 只向 Nacos naming 注册 RPC 服务，不会向 Nacos config 写入 `base.BaseService` / mapping 类配置。reward、pos 中连接 base 的客户端也使用相同的接口级发现方式。

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
1. 查 ServiceInfoMap[Service][SubService]，读取 multiSign / confirm 两个开关
2. base64 解码 EncodedTx → []byte
3. TransactionFromDecoder() → solana.Transaction
4. txID = tx.Signatures[0].String()（链上交易 ID 始终是用户签名，索引0）
5. 若 confirm=true：写 t_service_tx（state=TxStateInit）并记录 recordID
6. 若 multiSign=true：
   从 ServiceKeyMap[Service][SubService] 取私钥
   tx.Message.MarshalBinary() → privateKey.Sign() → tx.Signatures[1] = sig
7. 同步模拟执行交易（SimulateTransactionWithOpts，10s 超时）
   - SigVerify=false（服务端刚签完，签名一定正确）
   - ReplaceRecentBlockhash=true（避免因 blockhash 过期导致模拟误报失败）
   - 模拟链上失败（simResult.Value.Err != nil）→ 标记 t_service_tx 失败 → dubbo 返回错误
   - 模拟 RPC 网络层错误（connection refused/reset 等）→ 标记失败 → dubbo 返回错误
   - 模拟 RPC HTTP 层错误（400/500/timeout 等）→ 跳过模拟，继续广播（保守策略）
8. goroutine: broadcastTx()（60s 超时，单次广播，不重试）
   - 网络传输层错误 → 标记 t_service_tx 失败（交易根本没发出去）
   - HTTP 层/RPC 层错误 → 仅打日志，不标记失败（交易可能已到达节点，留给 TxScanTask/TxExpireTask 兜底）
9. 同步返回 record_id 和 tx_id
```

**错误分类辅助函数 `isNetworkError`**：优先用 `errors.As` 匹配 `net.OpError` / `net.DNSError`，兜底用关键字匹配（`connection refused`、`connection reset by peer`、`no such host`、`network is unreachable`、`i/o timeout`、`dial tcp`、`eof`）。

---

## 9. 后台任务（`task/`）

所有周期任务使用统一工具函数，内置分布式锁防止多实例重复执行：

```go
// 常规任务（执行时间短）
runPeriodic(rs, interval, lockKey, lockTTL, fn)
// 长任务（执行时间不可预测，使用 watchdog 自动续期锁）
runPeriodicWithWatchdog(rs, interval, lockKey, initTTL, fn)
```

### 9.1 FeeTask（`task/fee.go`）

3 个子任务，各自独立的 redsync 锁，每 5s 触发，锁 TTL 2min：

| 子任务 | 锁 key | 说明 |
|---|---|---|
| `readPriorityFee` | `base:sol:fee:stat-chain-fee:lock` | 调用 `getBlock` RPC 获取最新 slot 区块，解析每笔交易的 `SetComputeUnitPrice` / `SetComputeUnitLimit` 指令，写入 `t_fee_statistics`（上限 10000 条，超出则滚动覆盖最旧记录） |
| `updatePerUnitFee` | `base:sol:fee:update-per-unit:lock` | 对 `t_fee_statistics` 按 `compute_unit_price` 4 分位（NTILE）加权平均，结果写 Redis `base:sol:fee:priority-per-unit` |
| `updatePerTxFee` | `base:sol:fee:update-per-tx:lock` | 按百分比排名（PERCENT_RANK）分 4 组，取中位数（PERCENTILE_CONT 0.5），写 Redis `base:sol:fee:priority-per-tx` |

> 旧版的 `estimateWeightAvgFee`（调用 QuickNode `qn_estimatePriorityFees` API）已删除，前端统一使用 `readPriorityFee` 链路的 `priority-per-unit` / `priority-per-tx` 数据。`t_qn_fee` 表已废弃。

Redis key 格式：
- `base:sol:fee:priority-per-unit` / `base:sol:fee:priority-per-tx`

Redis 值为 `entity.FeeDetail` 的 JSON：`{Low, Medium, High, Extreme uint64}`

### 9.2 UnitTask（`task/unit.go`）

构造典型 Solana 交易（TransferChecked / Memo / AssociatedAccount / MiniRent）并模拟执行，统计各指令类型实际消耗的 compute unit，写入 Redis。

### 9.3 PriceTask（`task/price.go`）

周期性请求 Raydium API，获取 token/SOL、token/USDT、USDT/SOL 价格，写入 Redis。

### 9.4 KLineTask（`task/kline.go`）

使用 `runPeriodicWithWatchdog`，通过浏览器（`browser.go`）爬取 Raydium token K 线数据，写入 Redis。

### 9.5 HoldersTask（`task/holders.go`）

周期统计 token 持有者数量，写入 Redis。

### 9.6 TxScanTask（`task/scan.go`）⭐ 最核心任务

**职责**：扫描链上与各业务地址相关的新交易，找到后发 Kafka 通知 reward 等下游服务。

**初始化与动态变更**：
- 启动时先读取 `t_tx_scan_info` 表已有检查点，每条记录启动独立 goroutine。
- 若 Nacos `base-service-registry.yaml` 存在，则按 Nacos 的 scan 静态配置执行 `ReconcileScanConfigs`：
  - `scan.enabled=true` 且任务未运行：确保 DB 检查点存在，不存在时用 `initial_until_tx_id` / `initial_slot` 创建，然后启动 goroutine。
  - `scan.enabled=false` 或 Nacos 删除该 scan：取消对应 goroutine，但保留 DB 检查点。
  - 已存在 DB 检查点时，不覆盖 `until_tx_id` / `slot`。

**扫描循环**（每 3s 一轮）：
```
1. 获取 redsync 分布式锁（base:sol:tx-scan:{service}-{subService}-{pdaAccount}:lock，WithTries(1)）
2. DB 事务 + FOR UPDATE NOWAIT 行级锁定 t_tx_scan_info 记录
3. 调用 getSignaturesForAddress(pdaAccount, until=untilTxId, limit=300, Finalized)
4. 按 slot 降序排序，取最新交易
5. 条件更新 t_tx_scan_info.until_tx_id（仅当 slot > 当前记录 slot 时更新）
6. 提交 DB 事务（释放行锁）
7. 对 slot > 旧 slot 的每笔交易，goroutine: handleServiceTx()
```

**扫描检查点边界**：
- Nacos 保存静态扫描定义：`service` / `sub_service` / `native_account` / `pda_account` / `enabled` / 初始化检查点。
- DB 的 `t_tx_scan_info` 保存运行进度：`until_tx_id` / `slot`。
- 若需要回滚检查点，必须通过显式运维操作更新 DB，不应通过修改 Nacos 的 `initial_until_tx_id` / `initial_slot` 隐式覆盖运行进度。

**handleServiceTx 流程**：
```
1. Redis 分布式限流：AcquireDistributedRateLimit("handleServiceTx", 200/s)
2. redsync 锁：base:sol:tx-handle:{txSig}:lock（防止重复处理同一交易）
3. 根据 subService 选择 RPC 客户端（MarketBuyToken 使用 mainnetRpcClient，其余用默认 rpcClient）
4. GetTransactionResult（指数退避重试，最多 5 次，处理 ErrNotFound + RPC 限流错误）
5. 若 subService == "MarketBuyToken"：走独立的 handleMarketBuyTokenTx 流程（见下）
6. 否则：TransactionFromDecoder() → DecodeSolanaTransaction()（解析指令、账户、金额）
7. 查 t_service_tx（txId）：不存在则跳过（非本系统业务交易）
8. MarkTxFetchState(txId, TxFetchSuccess)
9. 发送 Kafka 消息：
   Topic: "ServiceTransaction"
   MsgType: "NewScannedTransaction"
   MsgContent: NewScannedTx{Service, SubService, TxSig, DecodedTx}
```

**MarketBuyToken 独立处理流程（handleMarketBuyTokenTx）**：
```
MarketBuyToken 是 DEX 购买行为，交易不经过 CommitTx，故不存在 t_service_tx 记录。
1. 跳过链上失败的交易（txSig.Err != nil）
2. 构建完整账户列表：静态 account keys + ALT 动态加载的地址（v0 交易）
3. 检查账户列表是否包含目标程序地址（HtNfUbD...FNT7i）
4. 遍历 inner instructions，找到 Token Program 发出、
   source == Raydium Pool source account（GjkvqFp...LXbQ）的 TransferChecked 指令
5. 解析 amount / decimals，组装 DecodedSolTransferCheckedInst
6. 直接发 Kafka NewScannedTransaction，不更新 t_service_tx
```

### 9.7 TxExpireTask（`task/expire.go`）

**职责**：兜底任务，处理 TxScanTask 可能遗漏的交易。每 5s，redsync 锁 TTL 2min。

```
1. 查 t_service_tx: WHERE tx_state=TxStateInit AND created_at <= NOW()-5min
2. 对每笔记录：
   a. Redis 限流（200/s）+ redsync 锁（HandlePosExpiredTx-{txId}）
   b. GetTransaction(txSig, Finalized)
      - ErrNotFound → MarkTxFetchState(TxFetchFailed) + 发 Kafka "NewExpiredTransaction"
      - 找到但解析失败 → MarkTxFetchState(TxFetchFailed) + 发 Kafka "NewExpiredTransaction"
      - 找到且解析成功 → MarkTxFetchState(TxFetchSuccess) + 发 Kafka "NewScannedTransaction"
      - 其他错误 → 跳过（等下轮重试）
```

### 9.8 TTLCleanupTask（`task/ttl_cleanup.go`）

**职责**：定期清理过期业务数据，防止表膨胀导致磁盘不足。每天 SGT 04:00 执行，Redsync 分布式锁（`lock:ttl_cleanup`，30min TTL，WithTries(1)）。

**清理方式**：批量 DELETE（每批 5000 行） + VACUUM FULL 回收磁盘空间。

| 表 | TTL | 清理条件 |
|---|---|---|
| `t_fee_statistics` | 2天 | `created_at < cutoff` |
| `t_service_tx` | 7天 | `created_at < cutoff` |
| `t_take_token_record` | 1月 | `created_at < cutoff` |
| `t_give_token_record` | 1月 | `created_at < cutoff` |
| `t_lottery_claim` | 1月 | `created_at < cutoff` |
| `t_daily_claim_stats` | 1月 | `created_at < cutoff` |
| `t_lottery_reward` | 1月 | `created_at < cutoff` |
| `t_campaign_quote_record` | 1月 | `created_at < cutoff` |
| `t_reward_code` | 1月 | `reward_state = -2 AND created_at < cutoff`（只清理已超时的奖励码） |
| `t_pos_snap_shot` | 3月 | `created_at < cutoff` |
| `t_pos_reward` | 3月 | `created_at < cutoff` |
| `t_pos_reward_claim` | 3月 | `created_at < cutoff` |
| `t_stake_snap_shot` | 3月 | `created_at < cutoff` |
| `t_stake_reward_claim` | 3月 | `created_at < cutoff` |
| `t_stake_buy_token` | 3月 | `created_at < cutoff` |
| `t_stake_reward` | 3月 | `created_at < cutoff AND NOT (reward_type=0 AND reward_state=0)`（保留未领取的固定利息） |

> VACUUM FULL 会锁全表，但凌晨执行且数据量小，锁表时间通常在秒级。

---
