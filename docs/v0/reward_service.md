# oshit-go / reward 服务参考文档

> 面向 Claude Code 使用，涵盖 reward 服务的架构与启动流程。
> TakeToken 业务完整流程见 `docs/reward_take_token.md`；通用工具见 `docs/overview.md`。

---

## 1. 服务概览

| 项目 | 值 |
|---|---|
| 服务名 | `reward-api` |
| 监听端口 | `1200`（HTTP Fiber） |
| 工作目录（开发） | `app/reward/api` |
| 配置文件 | `app/reward/api/etc/reward.yaml` |
| 已完成业务 | TakeToken |
| 开发中业务 | GiveToken（commit-tx 未实装） |
| 规划业务 | Lottery、Reward Code、Campaign Exchange |

---

## 2. 目录结构

```
app/reward/api/
├── reward.go                         # main 入口
├── setup.go                          # registerTasks()：注册 Kafka handler
├── etc/reward.yaml                   # 配置（db/redis/kafka/dubbo/nacos）
├── types/types.go                    # HTTP 请求/响应数据类型
├── internal/
│   ├── config/config.go              # 配置结构体 + viper 加载
│   ├── context/context.go            # CoreContext
│   ├── svc/
│   │   ├── context.go                # ServiceContext 初始化
│   │   └── kafka.go                  # Kafka producer/consumer 初始化
│   ├── rpc/client.go                 # BaseClient（封装 Dubbo Triple 调用 base 模块）
│   ├── handler/
│   │   ├── routes.go                 # 路由注册
│   │   ├── take.go                   # TakeToken HTTP handler
│   │   └── give.go                   # GiveToken HTTP handler（部分实现）
│   ├── logic/
│   │   ├── invite.go                 # RewardInviteLogic（邀请关系查询/写入）
│   │   ├── take/
│   │   │   ├── logic.go              # TakeTokenLogic：核心业务逻辑
│   │   │   ├── decode.go             # 交易解析 + 规则校验
│   │   │   └── kafka.go              # HandleScannedTx / HandleExpiredTx
│   │   └── give/
│   │       ├── logic.go              # GiveTokenLogic（开发中）
│   │       ├── decode.go
│   │       └── kafka.go
│   └── task/
│       ├── context.go                # TaskContext + handler 注册表
│       ├── mgr.go                    # TaskManager（注册/启动）
│       └── kafka.go                  # KafkaConsumerTask（消费循环 + 分发）
└── test/reward_test.go               # 集成测试（TestTakeToken）
```

---

## 3. 启动流程（`reward.go` + `setup.go`）

```
main()
 └── svc.NewServiceContext()
      ├── config.LoadConfig()           → ./etc/reward.yaml
      ├── initDatabase()                → GORM postgres
      ├── initRedis()                   → go-redis + redsync
      ├── initDatabaseConfigs()         → 从 DB 加载到 ServiceContext：
      │    SystemConfig / ChainConfig / TokenConfig / FeeTolerance
      │    LevelDist / LevelRatio / LevelRatioMap / DiscountRate
      │    TakeTokenConfig / GiveTokenConfig
      │    RewardKeyMap（从 t_reward_key_config 加密读取，jasypt 解密）
      │    LightHouseAddress（硬编码 L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95）
      ├── initSolanaRPC()               → rpc.New(ChainConfig.RPCURL)
      ├── initKafkaProducer()           → kafka-go writer
      ├── initKafkaConsumer()           → kafka-go reader（消费 topic: ServiceTransaction）
      ├── initBaseClient()              → rpc/client.go: NewBaseClient(nacosAddr, appName)
      └── startTasks()                  → 仅创建 TaskManager，不启动（等 registerTasks）

 └── registerTasks(srvCtx)             ← setup.go
      ├── TaskMgr.RegisterScannedTxHandler("TakeToken", take.HandleScannedTx)
      ├── TaskMgr.RegisterScannedTxHandler("GiveToken", give.HandleScannedTx)
      ├── TaskMgr.RegisterExpiredTxHandler("TakeToken", take.HandleExpiredTx)
      ├── TaskMgr.RegisterExpiredTxHandler("GiveToken", give.HandleExpiredTx)
      └── go TaskMgr.StartAllTasks()    → KafkaConsumerTask.Start()

 └── handler.RegisterRoutes(app, srvCtx)
     app.Listen(":1200")
```

**关键设计**：`registerTasks` 在 `NewServiceContext` 之后、HTTP 启动之前调用，确保 Kafka handler 先于消息到达完成注册。

---

## 4. CoreContext 与 ServiceContext（`internal/context/` + `internal/svc/`）

### CoreContext（`internal/context/context.go`）

```go
type CoreContext struct {
    // 基础设施（与 base 服务相同）
    Config  *config.Config
    DB      *gorm.DB
    Redis   redis.UniversalClient
    RedSync redsync.Redsync
    Ctx     context.Context

    // Solana 客户端
    RpcClient         *rpc.Client
    LightHouseAddress solana.PublicKey

    // 配置表数据
    SystemConfig  *model.SystemConfig
    ChainConfig   *model.ChainConfig
    TokenConfig   *model.TokenConfig
    FeeTolerance  model.FeeTolerance
    KafkaProducer interface{}  // *kafka.Writer

    // Dubbo 客户端（调用 base 模块）
    BaseClient *rewardrpc.BaseClient
}
```

### ServiceContext 额外字段

```go
type ServiceContext struct {
    core_context.CoreContext

    // reward 专属配置
    LevelDist       *model.LevelDist              // 奖励向上层级数（t_level_dist.level）
    LevelRatio      []model.LevelRatio             // 每层级分成比例列表
    LevelRatioMap   map[int32]model.LevelRatio     // level → ratio 快查
    DiscountRate    *model.DiscountRate             // 折扣率（暂未使用）
    TakeTokenConfig *model.TakeTokenConfig          // TakeToken 业务规则配置
    GiveTokenConfig *model.GiveTokenConfig          // GiveToken 业务规则配置
    RewardKeyMap    map[string]solana.PrivateKey    // service → 私钥（reward 自己暂未用）
    TaskMgr         *task.TaskManager
}
```

---

## 5. Dubbo 客户端（`rpc/client.go`）

reward 通过 `BaseClient` 调用 base 模块的所有 Dubbo 接口：

| 方法 | 作用 |
|---|---|
| `GetFeeTolerance(ctx)` | 获取成本费容错 |
| `GetPriorityFee(ctx)` | 获取优先费（4 档 × 2 组） |
| `GetInstUnits(ctx)` | 获取各指令 compute unit |
| `GetTokenQuoteSOLPrice(ctx)` | token/SOL 价格（TakeToken 计算 dex fee 必用） |
| `GetTokenQuoteUSDTPrice(ctx)` | token/USDT 价格 |
| `GetUSDTQuoteSOLPrice(ctx)` | USDT/SOL 价格 |
| `SendTransaction(ctx, tx, service, subService)` | 将交易序列化为 base64 后调用 base 签名+广播 |

`SendTransaction` 内部：
```
solana.Transaction.MarshalBinary() → base64.StdEncoding.EncodeToString() → Dubbo SendTransactionReq → 返回 txId
```

---

## 6. HTTP API（`handler/routes.go`）

所有路由挂载在 `/reward` 前缀下。

### TakeToken（`/reward/take`）

| Method | 路径 | Handler | 说明 |
|---|---|---|---|
| POST | `/reward/take/config` | `GetConfig` | 返回 `t_take_token_config` 全部字段 |
| POST | `/reward/take/record` | `GetRecord` | 根据 `txId` 查 `t_take_token_record` |
| POST | `/reward/take/tx-info` | `GetTxInfo` | 计算打包交易所需参数（TakeTokenTxInfo） |
| POST | `/reward/take/commit-tx` | `CommitTx` | 接收前端签名交易，验证后发给 base 广播 |

### GiveToken（`/reward/give`，部分实现）

| Method | 路径 | 说明 |
|---|---|---|
| POST | `/reward/give/config` | GiveToken 配置 |
| POST | `/reward/give/record` | GiveToken 记录 |
| POST | `/reward/give/tx-info` | GiveToken 交易参数 |
| POST | `/reward/give/commit-tx` | **未实装**（路由已注释） |

---

## 7. Kafka 消费架构（`task/kafka.go`）

```
KafkaConsumerTask.consume()       ← 阻塞读 topic "ServiceTransaction"
  └── dispatch(msg)
       ├── MsgType="NewScannedTransaction" → handleScannedTx(msg.MsgContent)
       │    └── scannedHandlers[SubService](ctx, msg)
       │         └── take.HandleScannedTx(msg)  ← 注册在 setup.go
       └── MsgType="NewExpiredTransaction" → handleExpiredTx(msg.MsgContent)
            └── expiredHandlers[SubService](ctx, msg)
                 └── take.HandleExpiredTx(msg)
```

**注意**：`kafka-go` 的 `ReadMessage` 自动提交 offset，消息处理失败不会重试。业务需自行保证幂等。

handler 注册表存储在 `TaskContext` 中：
```go
type TaskContext struct {
    scannedHandlers map[string]ScannedTxHandler   // SubService → handler
    expiredHandlers map[string]ExpiredTxHandler   // SubService → handler
}
```

---

## 8. 数据库表（reward 服务）

SQL 定义：`repositories/reward_structure.sql`

| 表名 | 主键 | 核心字段 | 用途 |
|---|---|---|---|
| `t_level_dist` | — | `level` | 奖励向上层级数（如 2 表示奖励直接邀请人和二级邀请人） |
| `t_level_ratio` | — | `level`, `ratio` | 每层邀请人分成比例（如 level=1 ratio=0.10） |
| `t_discount_rate` | ULID | `rate` | 折扣率（暂未使用） |
| `t_take_token_config` | ULID | 见下 | TakeToken 规则配置（单条 is_default=true） |
| `t_take_token_record` | ULID | 见下 | 每次 TakeToken 的业务记录 |
| `t_give_token_config` | — | `reward_rate`, `max_valid_reward`, `valid_rate` | GiveToken 规则配置 |
| `t_give_token_record` | ULID | `from_token_account`, `receipt_native_account`, `tx_id`, `state` | GiveToken 记录 |
| `t_reward_key_config` | — | `service`, `encrypted_key` | reward 私钥（jasypt 加密，按 service 索引） |
| `t_daily_claim_stats` | ULID | `native_account`, `take_shit_date`, `take_shit_count` | 每日领取统计（Lottery 预留） |
| `t_lottery_reward` | ULID | `reward_amount`, `reward_type`, `state`, `reward_day` | 抽奖奖励记录（预留） |
| `t_lottery_claim_record` | ULID | `reward_ids[]`, `tx_id`, `state` | 抽奖领取记录（预留） |

### t_take_token_config 关键字段

| 字段 | 说明 |
|---|---|
| `reward_account` | 发放奖励的 native account（token account 由此动态推导） |
| `cost_account` | 收取 SOL 成本费的地址 |
| `amount` | 无邀请码奖励数量（含 decimals） |
| `invite_amount` | 有邀请码奖励数量（含 decimals，通常更多） |
| `dex_fee_rate` | SOL 成本费率 |
| `max_dex_fee` | 最大 SOL 成本费 |
| `reward_inviter` | 是否奖励邀请人（当前 true） |
| `invited` | 是否确定邀请关系（当前 true） |

> `decimals`/`token_mint_account` 统一从 `t_token_config` 读取；`reward_token_account` 由 `reward_account` + mint 动态推导 ATA。

### t_take_token_record 关键字段

| 字段 | 说明 |
|---|---|
| `tx_id` | Solana 交易 ID（主查询键） |
| `reward_account` | 发放奖励的 native account |
| `receipt_account` | 收款人 native account |
| `cost_account` | 收取 SOL 成本费的地址 |
| `state` | 0=初始化，1=成功，-1=失败 |
| `use_invite_code` | 是否使用了邀请码 |
| `invite_code` | 使用的邀请码 |
| `invited` | 是否需要在成功后确定邀请关系 |
| `dex_fee` | 实际支付的 SOL 成本费（lamports） |
