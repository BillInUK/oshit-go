
# 1. 服务简介

reward 服务主要用来做一些相对简单固定的奖励领取服务，在reward服务里面，用户可以根据一定的业务规则来领取或者兑换token.根据规则，用户需要根据兑换的token数量来支付一定的成本费，也就是solana.

# 2. 文档索引
### 1.1 文档索引
| 文件名称 | 文档功能 |
|---|---|
|reward_take_token.md|take token 业务的详细介绍|
|reward_give_token.md|give token 业务的详细介绍|
|reward_lottery.md|lottery 业务详细介绍|
|reward_code.md|reward code 业务详细介绍|
|reward_campaign.md|campaign 业务详细介绍|

# 3. 目录结构

reward服务的目录结构如下

```
├── etc
├── internal
│   ├── config
│   ├── context
│   ├── handler
│   ├── logic
│   │   ├── campaign
│   │   ├── give
│   │   ├── lottery
│   │   ├── redeem
│   │   ├── rewardcode
│   │   └── take
│   ├── rpc
│   ├── svc
│   └── task
├── test
└── types

```

## 3.1 目录结构说明

| 目录 | 目录名称| 详细说明 |
|--------| --------| -------- |
| etc | 配置目录 | 该目录放置了启动 reward 服务需要的参数，包括数据库，redis，kafka,dubbo等配置 |
| internal | 内部业务实现目录 | reward模块具体的内部业务逻辑实现 |
| internal/config | 配置类 | 整个 reward 服务的业务配置定义，包括 数据库，redis等配置定义 |
| internal/context  | 核心上下文 | reward 服务启动后，会将数据库句柄，redis句柄，以及核心的token配置，链配置等数据结构保存在核心的上下文里面 |
| internal/svc  | 业务上下文 |  |
| internal/handler  | 与前端的api接口以及路由 | xxxxx |
| internal/logic  | 业务逻辑实现 | reward 服务所有的业务逻辑实现，包括: take token,give token等业务 |
| internal/logic/take  | take token 业务逻辑实现 | 包括前端请求处理，交易发送，base 服务回调的kafka消息处理等  |
| internal/logic/give  | give token 业务逻辑实现 | 包括前端请求处理，交易发送，base 服务回调的kafka消息处理等 |
| internal/logic/lottery  | lottery 业务逻辑实现 | 包括前端请求处理，交易发送，base 服务回调的kafka消息处理等 |
| internal/logic/rewardcode  | reward code 业务逻辑实现 | 包括前端请求处理，交易发送，base 服务回调的kafka消息处理等 |
| internal/logic/campaign  | campaign业务逻辑实现 | 包括前端请求处理，交易发送，base 服务回调的kafka消息处理等 |
| internal/rpc  | 与其他服务的rpc交互 | 包括 通过 dubbo 与base 模块交互，获取手续费信息，地址信息，以及通过 http内部接口跟campaign服务交互等 |
| internal/task  | 后台任务 | 包括 奖励码超时扫描任务等 |
| test  | 测试用例 | take token,give token等业务的接口测试 |
| types  | 模块公共的数据结构 | 包括前端请求和回复，以及一些业务的数据结构等 |


# 4. 数据库表

* reward 服务的数据库表结构在 repositories/structure/reward_structure.sql 里面定义

# 5. 核心数据结构

* base 服务的 前端的 api请求和回复 数据结构在 `internal/types/types.go` 里面定义

* 核心上下文

reward 服务的核心上下文在 `internal/context/context.go` 中定义

```go
type CoreContext struct {
	Config  *config.Config
	DB      *gorm.DB
	Redis   redis.UniversalClient
	RedSync redsync.Redsync
	Ctx     context.Context

	// 全局变量
	RpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
	TokenDecimal      float64
	KafkaProducer     interface{} // *kafka.Writer，在kafka.go中定义
	KafkaConsumer     interface{} // *kafka.Reader，在kafka.go中定义
	BaseClient        *rewardrpc.BaseClient

	NacosConfigClient config_client.IConfigClient
	ConfigMu          sync.RWMutex

	// 配置表数据
	SystemConfig *model.SystemConfig
	ChainConfig  *model.ChainConfig
	TokenConfig  *model.TokenConfig
	FeeTolerance *model.FeeTolerance
}
```

* 业务上下文

base 服务的业务上下文在 `internal/svc/context.go` 中定义

```
type ServiceContext struct {
	core_context.CoreContext
	LevelDist           *model.LevelDist
	LevelRatio          []model.LevelRatio
	LevelRatioMap       map[int32]model.LevelRatio
	DiscountRate        *model.DiscountRate
	TakeTokenConfig     *model.TakeTokenConfig
	GiveTokenConfig     *model.GiveTokenConfig
	CampaignQuoteConfig *model.CampaignQuoteConfig
	RewardCodeConfig    *model.RewardCodeConfig
	RewardKeyMap        map[string]solana.PrivateKey
	LightHouseAddress   solana.PublicKey
	TaskMgr             *task.TaskManager
	CampaignClientV1    *rewardrpc.CampaignClient
}

```

业务上下文继承自核心上下文，另外业务上下文还包含 reward 服务公共的配置，包括: 奖励邀请级别配置，take token奖励配置，give token 奖励配置等

* 任务上下文

```go
type TaskContext struct {
	core_context.CoreContext
	// ScannedHandlers 按 SubService 注册的已确认交易处理器，key 为 SubService 名称
	ScannedHandlers map[string]ScannedTxHandler
	// ExpiredHandlers 按 SubService 注册的超时交易处理器，key 为 SubService 名称
	ExpiredHandlers map[string]ExpiredTxHandler
}
```

任务上下文继承自核心上下文，另外还添加了后台任务所需要的一些 handlers，包括： 扫描到新交易的 ScannedHandlers, ExpiredHandlers，base 服务扫描到新交易，或者发现交易超时之后会通过这些 handlers 通知到 reward 服务。

# 6. 公共部分

## 6.1 公共表

* 奖励层级配置表

```sql
-- 奖励层级配置表
-- 旧工程 t_sol_transfer_reward_distribution
DROP TABLE IF EXISTS public.t_level_dist;
CREATE TABLE public.t_level_dist
(
    dist_level integer NOT NULL -- 对应 Level
);
```

* 表说明

该表配置 向上奖励邀请人的层级，比方说: A->B->C->D，如果使用该表作为奖励配置参数, dist_level=2 ，则奖励D的同时，还会奖励C,B,因为是向上奖励2级

### 6.2 奖励层级费率配置表

```sql
-- 奖励层级费率配置表
-- 旧工程 t_sol_transfer_reward_claim
DROP TABLE IF EXISTS public.t_level_ratio;
CREATE TABLE public.t_level_ratio
(
    dist_level integer       NOT NULL, -- 对应 Level
    ratio      numeric(5, 2) NOT NULL  -- 对应 ClaimRatio
);
```

* 表说明

该表向上级别的奖励每级的费率，比方说: dist_level=1,ratio=10,dist_level=2,ratio=5，A->B->C->D，那么奖励D的同时，奖励C,B，奖励C 10%，奖励B 5%

### 6.3 奖励成本费扣减表
```sql
-- 奖励成本费扣减表
-- 旧工程 t_reward_discount_rate
DROP TABLE IF EXISTS public.t_discount_rate;
CREATE TABLE public.t_discount_rate
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    rate      numeric(3, 2)                         NOT NULL -- 对应 Rate
);
```

* 表说明

一些特殊业务需要对成本费进行答复减免时会用到该表，目前暂时没有用到该表。

# 7. 公共定义

app/reward/api/types/types.go 里面定义了请求类型和数据结构类型

* RewardTokenItem - 奖励发放项

| 字段 | 类型 | 说明 | 
| ----- | ----- | ----- | 
| Index | int | 索引，防止后端返回给前端后，json乱序，方便前端排序 | 
| ReceiptAccount | string | 接收token奖励的solana地址 | 
| Amount | uint64 | 接收token奖励的原始金额 | 

* GetByTxIdReq - 通过交易id请求查询

| 字段 | 类型 | 说明 | 
| ----- | ----- | ----- | 
| TxId | string | 业务记录对应的交易Id | 

# 8. 启动流程

```
main()
 └── svc.NewServiceContext()
      ├── config.LoadConfig()           → viper 读 ./etc/reward.yaml
      ├── initDatabase()                → GORM postgres 连接
      ├── initRedis()                   → go-redis UniversalClient（支持 Sentinel）
      ├── redsync.New(pool)             → 分布式锁
	├── initDatabaseConfigs()         → 从 DB 加载全局配置表到 CoreContext / ServiceContext：
	│    SystemConfig / ChainConfig / TokenConfig / FeeTolerance / LightHouseAddress
	│    LevelDist / LevelRatio / LevelRatioMap（奖励层级配置，构建二维索引）
	│    DiscountRate / TakeTokenConfig / GiveTokenConfig
	│    CampaignQuoteConfig / RewardCodeConfig
	│    以上 DB 配置作为 Nacos 不可用或配置为空时的兜底配置
      ├── initNacosConfigClient()      → 初始化 Nacos config client
      ├── initNacosRuntimeConfigs()    → 优先从 Nacos 覆盖 DB 兜底配置：
      │    base-runtime.yaml           → SystemConfig / ChainConfig / TokenConfig / FeeTolerance / LightHouseAddress
      │    reward-runtime.yaml         → LevelDist / LevelRatio / DiscountRate
      │                                  TakeTokenConfig / GiveTokenConfig / CampaignQuoteConfig / RewardCodeConfig
	├── initSolanaRPC()               → rpc.New(ChainConfig.RPCURL)，初始化 RpcClient
	├── initKafkaProducer()           → kafka-go writer
	├── initKafkaConsumer()           → kafka-go reader
      ├── initBaseClient()              → 通过 Nacos 地址初始化 Dubbo Triple BaseClient
      │    连接 base 服务（用于获取手续费、价格、发送交易等）
      ├── initCampaignClient()          → 按环境初始化 Campaign HTTP 内部客户端
      │    env=0（测试环境）: http://172.31.48.20:4000/internal/api/v1
	│    env=1（生产环境）: http://172.31.48.157:80/internal/api/v1
	├── utils.InitDTokenManager()     → 初始化 dtoken JWT 鉴权管理器
	├── startTasks()                  → 创建 TaskManager（此时尚未启动任务，只初始化）
      └── listenNacosConfigs()         → 监听 base-runtime.yaml / reward-runtime.yaml 变更并热更新内存配置

main()
 ├── registerTasks(srvCtx)             → 注册 Kafka 消息处理器并启动后台任务
 │    ├── TaskMgr.RegisterScannedTxHandler("Reward", ...)
 │    │    内部按 SubService 路由到各业务的 HandleScannedTx：
 │    │    TakeToken / GiveToken / Lottery / RewardCode
 │    ├── TaskMgr.RegisterExpiredTxHandler("Reward", ...)
 │    │    内部按 SubService 路由到各业务的 HandleExpiredTx：
 │    │    TakeToken / GiveToken / Lottery / RewardCode
 │    └── go TaskMgr.StartAllTasks()
 │         ├── KafkaConsumerTask.Start()     → 消费 ServiceTransaction topic，派发到已注册 handlers
 │         └── RewardCodeExpireTask.Start()  → 定期扫描超时的 reward code 记录
 ├── fiber.New(...)                    → 配置 json-iterator（禁用 6 位小数截断）
 ├── handler.RegisterRoutes(app, srvCtx)
 └── app.Listen(":1200")
```

# 9. 其他

## 9.1 Nacos 配置

reward 当前订阅两个 Nacos dataId，Group 均为 `oshit-go`：

| Data ID | 用途 | 来源 |
|---|---|---|
| `base-runtime.yaml` | 全局链、Token、环境、手续费容错、LightHouse 地址配置 | 与 base 服务共用 |
| `reward-runtime.yaml` | reward 自身奖励规则配置 | `migrate/reward-runtime.yaml` |

`base-runtime.yaml` 中 reward 使用的字段如下：

```yaml
system:
  env: 1
chain:
  chain_name: "solana"
  rpc_url: "..."
  wss_url: "..."
  decimals: 9
  symbol: "SOL"
token:
  token_name: "OShit"
  token_symbol: "OShit"
  decimals: 3
  mint: "..."
fee_tolerance:
  max_less_rate: 0.05
lighthouse_address: "..."
```

`reward-runtime.yaml` 中包含以下业务规则：

```yaml
level_dist:
  dist_level: 2
level_ratio:
  - dist_level: 1
    ratio: 10.00
discount_rate:
  rate: 1.25
take_token:
  reward_account: "..."
  cost_account: "..."
  amount: 500000
  invite_amount: 1500000
  cost_fee_rate: 200
  max_cost_fee: 400
  is_default: true
  reward_inviter: true
  invited: true
give_token:
  reward_account: "..."
  cost_account: "..."
  reward_rate: 200.00
  max_valid_reward: 4000000
  valid_rate: 300.000000
campaign_quote:
  reward_account: "..."
  cost_account: "..."
  quote_rate: 500.00
  cost_rate: 17.00
reward_code:
  reward_account: "..."
  cost_account: "..."
```

启动时 reward 会先从 DB 读取这些配置作为兜底，再读取 Nacos 覆盖内存配置。Nacos 读取失败或 dataId 内容为空时，服务继续使用 DB 兜底配置启动。

Nacos listener 会热更新内存中的全局配置和 reward 业务规则。已经进入执行过程的单次请求或 Kafka 消息处理可能继续使用创建逻辑对象时持有的旧配置；新的请求和新的处理流程会使用更新后的配置。

