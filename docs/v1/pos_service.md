
# 1. 服务简介

pos 服务主要是通过快照用户持有token的余额，或者是快照用户质押的金额

# 2. 文档索引
### 1.1 文档索引
| 文件名称                  | 文档功能                  |
|-----------------------|-----------------------|
| pos_pos_snapshot.md   | pos snapshot 流程详细介绍   |
| pos_pos_reward.md     | pos 奖励业务              |
| pos_stake_token.md    | stake token 业务的详细介绍   |
| pos_stake_reward.md   | stake reward 业务详细介绍   |
| pos_stake_snapshot.md | stake snapshot 流程详细介绍 |
| pos_stake_leader.md   | stake leader 业务详细介绍   |

# 3. 目录结构

```
├── etc
├── internal
│   ├── config
│   ├── context
│   ├── handler
│   ├── logic
│   │   ├── pos
│   │   └── stake
│   ├── rpc
│   ├── svc
│   └── task
├── test
└── types
```

## 3.1 目录结构说明

| 目录 | 目录名称| 详细说明 |
|--------| --------| -------- |
| etc | 配置目录 | 该目录放置了启动 pos 服务需要的参数，包括数据库，redis，kafka,dubbo等配置 |
| internal | 内部业务实现目录 | pos 模块具体的内部业务逻辑实现 |
| internal/config | 配置类 | 整个 pos 服务的业务配置定义，包括 数据库，redis等配置定义 |
| internal/context  | 核心上下文 | pos 服务启动后，会将数据库句柄，redis句柄，以及核心的token配置，链配置等数据结构保存在核心的上下文里面 |
| internal/svc  | 业务上下文 |  |
| internal/handler  | 与前端的api接口以及路由 | xxxxx |
| internal/logic  | 业务逻辑实现 | pos 服务所有的业务逻辑实现，包括: pos,stake |
| internal/logic/pos  | pos 业务逻辑实现 | 包括前端请求处理，领取奖励的交易校验和发送，base 服务回调的kafka消息处理等  |
| internal/logic/stake  | stake 业务逻辑实现 | 包括前端请求处理，领取奖励的交易校验和发送，base 服务回调的kafka消息处理等  |
| internal/rpc  | 与其他服务的rpc交互 | 包括 通过 dubbo 与base 模块交互，获取手续费信息，地址信息 |
| internal/task  | 后台任务 | 包括: 周期性的扫描用户持币金额，扫描用户质押在智能合约内的金额等 |
| test  | 测试用例 | pos,stake 等业务的接口测试 |
| types  | 模块公共的数据结构 | 包括前端请求和回复，以及一些业务的数据结构等 |

# 4. 核心数据结构

* 核心上下文
  pos 服务的核心上下文在 `internal/context/context.go` 中定义

```go
type CoreContext struct {
	Config  *config.Config        // 从配置文件读取的基础配置
	DB      *gorm.DB              // postgres 数据库句柄
	Redis   redis.UniversalClient // redis 句柄
	RedSync redsync.Redsync       // redlock 句柄
	Ctx     context.Context       // 上下文句柄

	// 全局变量
	RpcClient             *rpc.Client        // solana rpc 客户端
	LightHouseAddress     solana.PublicKey   // light house 指令
	TokenDecimal          float64            // token 精度基数
	KafkaProducer         interface{}        // *kafka.Writer，在kafka.go中定义
	KafkaConsumer         interface{}        // *kafka.Reader，消费 ServiceTransaction（base 模块）
	SnapShotKafkaConsumer interface{}        // *kafka.Reader，消费 PosTopic + StakeTopic
	BaseClient            *posrpc.BaseClient // base 模块dubbo 客户端

	// 配置表数据
	SystemConfig *model.SystemConfig // 系统配置
	ChainConfig  *model.ChainConfig  // 链配置
	TokenConfig  *model.TokenConfig  // token配置
	FeeTolerance *model.FeeTolerance // 手续费容错配置
}
```

* 业务上下文
  pos 服务的业务上下文在 `internal/svc/context.go` 中定义

```go
type ServiceContext struct {
	core_context.CoreContext
	LightHouseAddress solana.PublicKey
	TaskMgr           *task.TaskManager

	// pos业务配置
	PosStarLevelRule map[int32]model.PosStarLevelRule  // pos 星级用户等级配置
	PosRewardConfig  *model.PosRewardConfig            // pos 奖励发放配置
	PosWhiteListMap  map[string]model.PosStarWhitelist // pos 星级用户白名单

	// stake业务配置
	StakeAmmConfig     *model.StakeAmmConfig               // stake amm 做市地址
	StakeRewardConfig  *model.StakeRewardConfig            // stake 普通用户奖励发放配置
	LeaderRewardConfig *model.StakeLeaderRewardConfig      // stake 区域经理(领导)奖励发放配置
	StakeFixConfig     map[int32]model.StakeFixRateConfig  // stake 每日固定利息配置
	StakeInviteRate    map[int32]model.StakeInviteRate     // stake 邀请人奖励配置
	StakeStarLevelRule map[int32]model.StakeStarLevelRule  // stake 星级用户配置
	TotalAreaLeaders   []model.StakeTotalLeader            // stake 总区域经理(领导)配置
	StakeTokenPoolMap  map[string]model.StakeTokenPool     // stake token 池配置
	StakeDistLevel     int32                               // stake 奖励层级配置
	StakeStarWhitelist map[string]model.StakeStarWhitelist // stake 星级用户白名单
}
```

* 任务上下文
  pos 服务的任务上下文在 `internal/task/context.go` 中定义

```go
type TaskContext struct {
	core_context.CoreContext
	SnapShotHandlers map[string]SnapShotHandler
	// ScannedHandlers 按 SubService 注册的已确认交易处理器，key 为 SubService 名称
	ScannedHandlers map[string]ScannedTxHandler
	// ExpiredHandlers 按 SubService 注册的超时交易处理器，key 为 SubService 名称
	ExpiredHandlers       map[string]ExpiredTxHandler
	RewardConfig          *model.StakeRewardConfig
	SnapShotKafkaConsumer interface{} // *kafka.Reader，消费 PosTopic + StakeTopic
}
```

# 5. 启动流程

```
main()
 └── svc.NewServiceContext()
      ├── config.LoadConfig()              → viper 读 ./etc/application.yaml
      ├── initDatabase()                   → GORM postgres 连接
      ├── initRedis()                      → go-redis UniversalClient（支持 Sentinel）
      ├── redsync.New(pool)                → 分布式锁
      ├── initDatabaseConfigs()            → 从 DB 兜底加载全局配置：
      │    SystemConfig / ChainConfig / TokenConfig / FeeTolerance / LightHouseAddress
      ├── initNacosConfigClient()          → 初始化 Nacos config client
      ├── initNacosRuntimeConfigs()        → 优先从 Nacos 覆盖 DB 兜底配置：
      │    base-runtime.yaml  → SystemConfig / ChainConfig / TokenConfig / FeeTolerance / LightHouseAddress
      │    pos-runtime.yaml   → PosRewardConfig / PosStarLevelRule / PosWhiteListMap
      │                         StakeRewardConfig / LeaderRewardConfig / StakeFixConfig
      │                         StakeInviteRate / StakeDistLevel / StakeStarLevelRule
      │                         StakeStarWhitelist / StakeAmmConfig / TotalAreaLeaders / StakeTokenPoolMap
      ├── initSolanaRPC()                  → rpc.New(ChainConfig.RPCURL)
      ├── initKafkaProducer()              → kafka-go writer
      ├── initKafkaConsumer()              → kafka-go reader（消费 ServiceTransaction topic）
      ├── initSnapShotKafkaConsumer()      → kafka-go reader（消费 PosTopic + StakeTopic 快照消息）
      ├── initBaseClient()                 → 通过 Nacos 初始化 Dubbo Triple BaseClient
      ├── initPosConfig()                  → 从 DB 兜底加载 pos 业务配置（Nacos 已覆盖时为二次确认）
      │    PosStarLevelRule / PosRewardConfig / PosWhiteListMap
      ├── initStakeConfig()                → 从 DB 兜底加载 stake 业务配置（Nacos 已覆盖时为二次确认）
      │    StakeAmmConfig / StakeRewardConfig / LeaderRewardConfig
      │    StakeFixConfig / StakeInviteRate / StakeStarLevelRule
      │    TotalAreaLeaders / StakeTokenPoolMap / StakeDistLevel / StakeStarWhitelist
      ├── utils.InitDTokenManager()        → 初始化 dtoken JWT 鉴权管理器
      ├── startTasks()                     → 创建 TaskManager（尚未启动任务）
      └── listenNacosConfigs()             → 监听 base-runtime.yaml / pos-runtime.yaml 变更并热更新

main()
 ├── registerTasks(srvCtx)                → 注册 Kafka 消息处理器并启动后台任务
 │    ├── RegisterSnapShotHandler("Pos")   → NewPosSnapShot / NewStakeSnapShot 按 MsgType 路由
 │    ├── RegisterScannedTxHandler("Pos")  → 按 SubService 路由：
 │    │    PosReward / StakeToken / MarketBuyToken / StakeReward / StakeLeaderReward
 │    ├── RegisterExpiredTxHandler("Pos")  → 按 SubService 路由：
 │    │    PosReward / StakeReward / StakeLeaderReward
 │    └── go TaskMgr.StartAllTasks()
 │         ├── KafkaConsumerTask.Start()          → 消费 ServiceTransaction，派发已确认/超时交易
 │         ├── SnapShotConsumerTask.Start()        → 消费快照 topic，派发快照消息
 │         └── StakeBuyTokenExpireTask.Start()     → 定期清理过期的购买记录
 ├── fiber.New(...)                        → 配置 json-iterator（禁用 6 位小数截断）
 ├── handler.RegisterRoutes(app, srvCtx)
 └── app.Listen(":1300")
```

## 5.1 Nacos 配置

pos 当前订阅两个 Nacos dataId，Group 均为 `oshit-go`：

| Data ID | 用途 | 来源 |
|---|---|---|
| `base-runtime.yaml` | 全局链、Token、环境、手续费容错、LightHouse 地址配置 | 与 base / reward 服务共用 |
| `pos-runtime.yaml` | pos/stake 自身业务规则配置 | `migrate/pos-runtime.yaml` |

`base-runtime.yaml` 中 pos 使用的字段：

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

`pos-runtime.yaml` 包含以下业务规则（详见 `migrate/pos-runtime.yaml`）：

| 配置项 | 说明 |
|---|---|
| `pos_reward` | pos 奖励发放配置（reward_account / cost_account / cost_fee_rate 等） |
| `pos_star_level_rule` | pos 星级评定规则（1~5 星，个人持币 + 团队持币双重门槛） |
| `pos_star_whitelist` | pos 星级白名单（运营手动配置，优先级高于规则） |
| `stake_reward` | 质押奖励发放配置（program_id / reward_account / cost_account 等） |
| `stake_leader_reward` | 区域经理奖励发放配置（reward_account） |
| `stake_fix_rate` | 每日固定利息配置（按 stake_type 区分 180天/360天） |
| `stake_invite_dist` | 邀请奖励最大追溯层级 |
| `stake_invite_rate` | 各层级邀请奖励费率 |
| `stake_star_level_rule` | 质押星级评定规则（1~6 星） |
| `stake_star_whitelist` | 质押星级白名单 |
| `stake_amm` | AMM 做市地址配置 |
| `stake_total_leaders` | 总区域经理配置（固定 2 人） |
| `stake_token_pool` | token 池地址映射 |

启动时 pos 会先从 DB 读取这些配置作为兜底，再读取 Nacos 覆盖内存配置。Nacos 读取失败或 dataId 内容为空时，服务继续使用 DB 兜底配置启动。

Nacos listener 会热更新内存中的全局配置和 pos 业务规则。已经进入执行过程的单次请求或 Kafka 消息处理可能继续使用创建逻辑对象时持有的旧配置；新的请求和新的处理流程会使用更新后的配置。

---

# 6. pos 业务整体介绍

pos 业务通过每日快照用户持有 token 的余额，按照星级规则发放奖励给符合条件的持币用户。

**流程概述**：
1. 每天新加坡时间中午 12 点，系统对链上持币地址进行快照，写入数据库
2. 写入完成后发送 Kafka `NewPosSnapShot` 消息通知 pos 服务处理
3. pos 服务根据快照数据和星级规则计算每位用户的奖励，写入 `t_pos_reward`
4. 用户打包交易领取奖励（流程同 TakeToken：GetTxInfo → CommitTx → Kafka 确认）

详见 `pos_pos.md`。

---

# 7. stake 业务整体介绍

stake 业务分为四个子业务，各自有独立的文档：

| 子业务 | 文档 | 简述 |
|---|---|---|
| StakeToken | `pos_stake_token.md` | 用户质押/解质押/重质押 token 到智能合约 |
| StakeLeaderReward | `pos_stake_leader.md` | 用户质押时按比例奖励其上级区域经理 |
| StakeSnapshot | `pos_stake_token.md` | 每日快照智能合约内质押数据，计算极差奖励 |
| StakeReward | `pos_stake_reward.md` | 用户领取快照后生成的质押奖励 |

**stake 业务的智能合约接口**（Anchor 程序，`contract/lib.rs`）：

```rust
// 质押 token（stakeType: 0=180天 1=360天）
pub fn stake(ctx: Context<Stake>, amount: u64, stake_type: u8) -> Result<()>

// 解除质押（stakeIndex: 质押槽位索引）
pub fn unstake(ctx: Context<DeStake>, stake_index: u8) -> Result<()>

// 重新质押（不提取，直接续期）
pub fn restake(ctx: Context<ReStake>, stake_index: u8, stake_type: u8) -> Result<()>
```

**指令数据布局**（Anchor discriminator = `SHA256("global:<name>")` 前 8 字节）：

| 指令 | 数据长度 | 布局 |
|---|---|---|
| stake | 17 bytes | disc(8) + amount(8, LE) + stakeType(1) |
| unstake | 9 bytes | disc(8) + stakeIndex(1) |
| restake | 10 bytes | disc(8) + stakeIndex(1) + stakeType(1) |

**锁仓时间**（以 Solana slot 计算，假设每个 slot 400ms）：
- stakeType=0：180 天
- stakeType=1：360 天

---

# 8. 其他































