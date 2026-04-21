# 1. 业务概述

StakeLeaderReward 是 StakeToken 的附属奖励机制。当用户从 DEX 购买 token（`t_stake_buy_token`）后进行质押（stake）时，系统按比例向其邀请链上的区域经理（Area Leader）生成奖励记录。区域经理随后可自行打包交易领取累积奖励。

**整体流程**：
1. DEX 购买记录由 `HandleMarketBuyTx` 写入 `t_stake_buy_token`
2. 用户质押时（`HandleStakeTx`）计算区域经理奖励，写入 `t_stake_leader_reward`
3. 区域经理通过接口查询待领取奖励 → 构造交易 → CommitTx → 链上确认

> **与普通 StakeReward 的核心差异**：
> - 奖励**由质押行为触发**，而非快照触发
> - 领取时**无 SOL 成本费**，交易仅包含 1 条 TransferChecked（无 System.Transfer）
> - 奖励额度基于 `min(DEX购买剩余量, 质押量)`，按区域经理等级分配不同比例

---

## 2. 数据库表

* t_stake_buy_token（见 `pos_stake_token.md §2`，此处不重复）

* t_stake_leader_reward_config

```sql
-- 区域经理奖励发放配置
drop table if exists public.t_stake_leader_reward_config;
create table public.t_stake_leader_reward_config
(
    record_id      ulid not null               default gen_ulid(),
    reward_account varchar(64),                                               -- 发放奖励的 native account
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
```

* t_stake_leader

```sql
-- 区域经理表（由运营配置）
drop table if exists public.t_stake_leader;
create table public.t_stake_leader
(
    record_id      ulid     not null           default gen_ulid(),
    native_account varchar(64),                                               -- 区域经理地址
    leader_level   smallint not null,                                        -- 等级：1=普通区域经理 2=无上级的高级区域经理
    up_leader      varchar(64),                                               -- 上级 leader 地址（level=1 时有值）
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
```

* t_stake_total_leader

```sql
-- 总区域经理表（平台级别，按比例分配奖励）
drop table if exists public.t_stake_total_leader;
create table public.t_stake_total_leader
(
    record_id      ulid not null               default gen_ulid(),
    native_account varchar(64),                                               -- 总区域经理地址
    stake_share    numeric(5, 2),                                            -- 奖励分成比例（小数，如 0.07 = 7%）
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
```

* t_stake_leader_reward

```sql
-- 区域经理奖励明细表（质押时生成）
drop table if exists public.t_stake_leader_reward;
create table public.t_stake_leader_reward
(
    record_id      ulid        not null default gen_ulid(),
    native_account varchar(64) not null,                       -- 区域经理地址
    staker         varchar(64) not null,                       -- 触发奖励的质押者地址
    reward_type    int         not null,                       -- 奖励类型（见下方说明）
    base_amount    numeric(78, 0),                             -- 基础金额 = min(DEX购买剩余量, 质押量)
    stake_share    numeric(5, 2),                              -- 奖励费率（%）
    reward_amount  numeric(78, 0),                            -- 奖励金额（raw）= base_amount × stake_share/100
    reward_state   int         not null default 0,             -- 0=待领取 1=已领取 -1=过期
    pending        bool        not null default false,         -- true=已提交交易待链上确认
    tx_id          varchar(128),                               -- 领取交易 id（pending 时写入）
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
create index on public.t_stake_leader_reward (native_account, reward_state, pending);
```

**reward_type 枚举**：

| reward_type | 说明 | 比例 |
|---|---|---|
| 0（Direct） | level=2 区域经理直接获得 | 10% |
| 1（Level1） | level=1 区域经理获得 | 7% |
| 2（Leader） | level=1 的上级 Leader 获得 | 3% |
| 3（TotalLeader） | 总区域经理按各自 stake_share 获得 | 各自配置 |

* t_stake_leader_reward_claim

```sql
-- 区域经理奖励领取记录
drop table if exists public.t_stake_leader_reward_claim;
create table public.t_stake_leader_reward_claim
(
    record_id      ulid        not null default gen_ulid(),
    native_account varchar(64) not null,                   -- 领取者（区域经理）地址
    reward_ids     ulid[]      not null,                   -- 本次领取的 t_stake_leader_reward.record_id 列表
    tx_id          varchar(128),                           -- 链上交易 id
    tx_state       int                  default 0,         -- 0=待确认 1=成功 -1=失败
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
create index on public.t_stake_leader_reward_claim (tx_id);
create index on public.t_stake_leader_reward_claim (native_account, tx_state);
```

**t_stake_leader_reward 状态机**：

```
质押触发写入    → pending=false, reward_state=0（待领取）
CommitTx 提交  → pending=true,  reward_state=0，tx_id 写入
Kafka 链上成功  → pending=false, reward_state=1（RewardStateClaimed）
Kafka 链上失败  → pending=false, reward_state=0（重置，可重新领取）
Kafka 超时      → pending=false, reward_state=0（重置，可重新领取）
```

---

## 3. 数据结构

```go
// GetTxInfo 响应，前端据此打包领取交易
type LeaderRewardTxInfo struct {
    RewardAccount string  `json:"rewardAccount"` // 发放奖励的 native account
    Mint          string  `json:"mint"`          // token mint 地址
    Decimals      int32   `json:"decimals"`      // token 精度
    TotalReward   float64 `json:"totalReward"`   // 当前可领取奖励总额（raw）
}
```

---

## 4. 奖励生成机制

奖励由 `HandleStakeTx`（Kafka 消费 stake 交易）触发，详见 `pos_stake_token.md §6`。

```mermaid
flowchart TD
    A([用户质押成功\nHandleStakeTx]) --> B["baseAmount = min(DEX购买剩余量, 质押量)"]
    B --> C{baseAmount > 0?}
    C -- 否 --> Z([不生成奖励])
    C -- 是 --> D["findDirectLeader\nCTE 递归向上查邀请链\n找最近的 t_stake_leader"]
    D --> E{找到区域经理?}
    E -- 否 --> Z
    E -- 是 --> F{areaLeader.level?}

    F -- level=2 --> G["本人 +10%\ntype=Direct"]
    F -- level=1 --> H["本人 +7%（type=Level1）\n上级 up_leader +3%（type=Leader）"]

    G & H --> I["总区域经理各按 stake_share 分配\n(type=TotalLeader)"]
    I --> J["deductBuyRemaining\n按 slot asc 扣减 t_stake_buy_token.remaining_amount"]
    J --> K["INSERT t_stake_leader_reward\npending=false, reward_state=0"]
    K --> L([Commit])
```

---

## 5. 业务流程

### 5.1 阶段一：查询奖励信息

```mermaid
sequenceDiagram
    actor FE as 前端（区域经理）
    participant R as pos-api
    participant DB as PostgreSQL

    FE->>R: POST /snap/stake/reward/leader/info（JWT）
    R->>DB: 查 t_stake_leader by native_account
    R-->>FE: StakeLeader{nativeAccount, leaderLevel, upLeader}

    FE->>R: POST /snap/stake/reward/leader/records（JWT）
    R->>DB: 查 t_stake_leader_reward\nWHERE native_account=? AND reward_state=0 AND pending=false
    R-->>FE: []StakeLeaderReward（各笔奖励明细）

    FE->>R: POST /snap/stake/reward/leader/tx-info（JWT）
    R->>DB: SUM(reward_amount) FROM t_stake_leader_reward\nWHERE native_account=? AND state=0 AND pending=false
    R-->>FE: LeaderRewardTxInfo{TotalReward, RewardAccount, Mint,...}
```

### 5.2 阶段二：提交领取交易（CommitTx）

```mermaid
sequenceDiagram
    actor FE as 前端（区域经理）
    participant R as pos-api
    participant B as base-api
    participant SOL as Solana RPC
    participant DB as PostgreSQL

    Note over FE: 构造交易（见第 6 节）<br/>用户私钥签名 Signatures[0]

    FE->>R: POST /snap/stake/reward/leader/commit-tx {encodedTx}
    R->>R: PreCheckEncodedTx（hex decode + 验签）
    R->>R: 获取分布式锁（key: stake:leader-reward:commit-tx:{from}，1h）
    R->>R: getLeaderTxInfo（重新查询总额，不信任客户端）
    R->>DB: 查 t_stake_leader_reward（reward_state=0 AND pending=false）
    R->>R: decodeLeaderSOLTx（解析指令）
    R->>R: checkLeaderSOLTx（校验地址/金额）

    R->>B: Dubbo SendTransaction{encodedTx, "Pos", "StakeLeaderReward"}
    B->>B: 用服务私钥补签 Signatures[1]
    B->>DB: INSERT t_service_tx{state=0}
    B-->>B: goroutine 广播交易
    B->>SOL: SendTransaction
    B-->>R: txId

    R->>DB: recordLeaderClaim（事务）：
    R->>DB:   UPDATE t_stake_leader_reward SET pending=true, tx_id=txId
    R->>DB:   INSERT t_stake_leader_reward_claim{reward_ids, tx_id, state=0}
    R-->>FE: txId
```

### 5.3 阶段三：链上确认与 Kafka 处理

```mermaid
sequenceDiagram
    participant B as base-api
    participant SOL as Solana RPC
    participant K as Kafka
    participant R as pos-api
    participant DB as PostgreSQL
    actor FE as 前端

    loop TxScanTask 每 3s
        B->>SOL: getSignaturesForAddress(leaderRewardTokenAccount)
        B->>DB: UPDATE t_service_tx tx_state=TxFetchSuccess
        B->>K: NewScannedTransaction{SubService="StakeLeaderReward"}
    end

    K->>R: consume NewScannedTransaction
    alt TxSig.Err == nil（链上成功）
        R->>DB: UPDATE t_stake_leader_reward_claim tx_state=1
        R->>DB: UPDATE t_stake_leader_reward reward_state=1, pending=false
        R->>DB: INSERT t_fund_flow（token出账流水，upsert）
    else TxSig.Err != nil（链上失败）
        R->>DB: UPDATE t_stake_leader_reward_claim tx_state=-1
        R->>DB: UPDATE t_stake_leader_reward reward_state=0, pending=false, tx_id=nil
        Note over R,DB: 重置为可重新领取状态
    end

    Note over B: TxExpireTask 兜底（5min 后仍 state=0）
    B->>K: NewExpiredTransaction{SubService="StakeLeaderReward"}
    K->>R: consume NewExpiredTransaction
    R->>DB: UPDATE t_stake_leader_reward_claim tx_state=-1
    R->>DB: UPDATE t_stake_leader_reward reward_state=0, pending=false, tx_id=nil

    loop 前端轮询
        FE->>R: POST /snap/stake/reward/leader/claim-record {txId}
        R->>DB: 查 t_stake_leader_reward_claim.tx_state
        R-->>FE: tx_state：0=待确认 / 1=成功 / -1=失败
    end
```

---

## 6. 前端构造交易指令

```mermaid
flowchart TD
    A([获得 LeaderRewardTxInfo]) --> B["① SetComputeUnitPrice(Medium)"]
    B --> C["② SetComputeUnitLimit\n= (TransferChecked_CU) × 1.2"]
    C --> D{区域经理 token account 不存在?}
    D -- 是 --> E["③ CreateAssociatedTokenAccount(区域经理)"]
    D -- 否 --> F
    E --> F["④ TransferChecked\nleaderRewardTokenAccount → leaderTokenAccount\namount = TotalReward，owner = RewardAccount"]
    F --> G(["区域经理私钥签名 Signatures[0]\nhex 序列化 → encodedTx"])
```

> **注意**：领取区域经理奖励**无 SOL 成本费**，交易中**不含** System.Transfer 指令。

---

## 7. CommitTx 服务端处理

```mermaid
flowchart TD
    A([POST /snap/stake/reward/leader/commit-tx]) --> B["PreCheckEncodedTx\nhex decode + 验签 + 提取 from / txId"]
    B --> C["获取分布式锁\nstake:leader-reward:commit-tx:{from}，1h"]
    C --> D["getLeaderTxInfo(from)\n重新汇总 t_stake_leader_reward 未领取总额"]
    D --> E{TotalReward > 0?}
    E -- 否 --> F([返回：无可领取奖励])
    E -- 是 --> G["查询本次可领取奖励列表"]
    G --> H["decodeLeaderSOLTx\n解析指令（仅 TransferChecked，无 System.Transfer）"]
    H --> I{checkLeaderSOLTx}
    I --> I1["数量校验\nTransferInstructions == 0\nTransferChecked == 1"]
    I --> I2["TransferChecked 校验\nfrom == leaderRewardTokenAccount\nto == FindATA(from, Mint)\nmint == 配置 mint\namount == TotalReward（精确匹配）\ndecimals 匹配"]
    I1 & I2 --> J{全部通过?}
    J -- 否 --> K([返回错误])
    J -- 是 --> L["Dubbo base.SendTransaction\n补签 + 写 t_service_tx + 异步广播"]
    L --> M["recordLeaderClaim（事务）\nUPDATE t_stake_leader_reward pending=true, tx_id=txId\nINSERT t_stake_leader_reward_claim{state=0}"]
    M --> N([返回 txId])
```

---

## 8. Kafka 处理逻辑（`logic/stake/reward_kafka.go`）

```mermaid
flowchart TD
    subgraph S["HandleLeaderScannedTx"]
        direction TD
        A([收到 NewScannedTx\nSubService=StakeLeaderReward]) --> B{TxSig.Err?}
        B -- 链上失败 --> C["UPDATE t_stake_leader_reward_claim tx_state=-1\nUPDATE t_stake_leader_reward reward_state=0, pending=false, tx_id=nil\n（重置为可重新领取）"]
        B -- 链上成功 --> D["UPDATE t_stake_leader_reward_claim tx_state=1\nUPDATE t_stake_leader_reward reward_state=1, pending=false\nINSERT t_fund_flow（token出账，upsert）"]
        C --> E([Commit])
        D --> E
    end

    subgraph E2["HandleLeaderExpiredTx"]
        direction TD
        A2([收到 NewExpiredTx\nSubService=StakeLeaderReward]) --> B2["UPDATE t_stake_leader_reward_claim tx_state=-1\nUPDATE t_stake_leader_reward reward_state=0, pending=false, tx_id=nil"]
        B2 --> C2([Commit])
    end

    S ~~~ E2
```

---

## 9. HTTP API（`handler/stake.go`）

所有路由挂载在 `/snap/stake/reward/leader` 前缀下，除 `claim-record` 外均需 JWT。

### POST `/snap/stake/reward/leader/info` 🔒 需要 JWT

查询当前登录地址的区域经理信息。

- 响应：`model.StakeLeader`

| 字段 | 类型 | 说明 |
|---|---|---|
| `nativeAccount` | string | 区域经理地址 |
| `leaderLevel` | int16 | 等级：`1`=有上级 / `2`=无上级 |
| `upLeader` | string | 上级 leader 地址（level=1 时有值） |

---

### POST `/snap/stake/reward/leader/records` 🔒 需要 JWT

查询当前登录地址的待领取区域经理奖励明细列表。

- 响应：`[]model.StakeLeaderReward`

| 字段 | 类型 | 说明 |
|---|---|---|
| `nativeAccount` | string | 区域经理地址 |
| `staker` | string | 触发奖励的质押者地址 |
| `rewardType` | int | 奖励类型（0/1/2/3） |
| `baseAmount` | numeric | 基础金额（raw） |
| `stakeShare` | numeric | 奖励费率（%） |
| `rewardAmount` | numeric | 奖励金额（raw） |
| `rewardState` | int | `0`=待领取 / `1`=已领取 |

---

### POST `/snap/stake/reward/leader/tx-info` 🔒 需要 JWT

获取打包领取交易所需参数。

- 响应：`LeaderRewardTxInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `rewardAccount` | string | 发放奖励的 native account |
| `mint` | string | token mint 地址 |
| `decimals` | int32 | token 精度 |
| `totalReward` | float64 | 当前可领取奖励总额（raw） |

---

### POST `/snap/stake/reward/leader/commit-tx`

提交签名后的领取交易。

- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `encodedTx` | string | 是 | hex 编码的已签名 Solana 交易 |

- 响应：`txId`（string）

- 失败场景：

| 场景 | 说明 |
|---|---|
| 无可领取奖励 | `TotalReward <= 0` |
| 并发提交 | 分布式锁保护，同一地址同时只处理一笔 |
| 交易金额不匹配 | amount != TotalReward（精确匹配） |
| 交易含 SOL 转账 | 不应包含 System.Transfer |

---

### POST `/snap/stake/reward/leader/claim-record`

根据交易 ID 查询领取记录，用于前端轮询状态。

- 请求体：`{ "txId": "..." }`
- 响应：`model.StakeLeaderRewardClaim`

| 字段 | 类型 | 说明 |
|---|---|---|
| `nativeAccount` | string | 领取者地址 |
| `rewardIds` | ulid[] | 本次领取的奖励 ID 列表 |
| `txId` | string | 链上交易 ID |
| `txState` | int | `0`=待确认 / `1`=成功 / `-1`=失败 |
