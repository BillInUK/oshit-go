# 1. 业务概述

用户将 token 转给其他地址（to），平台根据转账金额和 to 地址是否为"有效地址"，按比例奖励转账人（from）token；若 from 存在上级邀请关系，还按层级比例向各层邀请人发 token。

整体分四个阶段：**GetTxInfo → 前端构造交易 → CommitTx → 链上确认（Kafka）**

> **与 TakeToken 的核心差异**：
> - GetTxInfo 需要 JWT，`from` 地址从 token 中提取
> - 奖励金额取决于 `to` 是否为有效地址（两档费率）
> - 无邀请码机制，不写 `t_invite_relation`

---

# 2. 数据库表

* t_give_token_config

```sql
-- give token 奖励配置
DROP TABLE IF EXISTS public.t_give_token_config;
CREATE TABLE public.t_give_token_config
(
    reward_account   character varying(64) NOT NULL, -- 发放奖励的 solana 地址
    cost_account     character varying(64) NOT NULL, -- 接收成本费的 solana 地址
    reward_rate      numeric(5, 2)         NOT NULL, -- 普通地址奖励费率（to 有 token account 时）
    max_valid_reward numeric(78, 0)        NOT NULL, -- 奖励金额上限
    valid_rate       numeric(10, 6)        NOT NULL, -- 有效地址奖励费率（to 无 token account 时）
    created_at       timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
```

* t_give_token_record

```sql
-- give token 记录表
DROP TABLE IF EXISTS public.t_give_token_record;
CREATE TABLE public.t_give_token_record
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    from_account    character varying(64)                                 NOT NULL, -- 发送 token 的 solana 地址
    receipt_account character varying(64)                                 NOT NULL, -- 接收 token 的 solana 地址
    tx_id           character varying(128)                                NOT NULL, -- 交易 id
    amount          numeric(78, 0)                                        NOT NULL, -- give token 的金额
    tx_state        integer,                                                        -- 交易状态：0=进行中 1=成功 -1=失败
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
```

---

## 3. 数据结构

```go
// GetTxInfo 请求（from 从 JWT 提取，不在请求体中）
type GetGiveTokenTxInfoReq struct {
    To     string  `json:"to"`     // 接收 token 的 solana 地址
    Amount float64 `json:"amount"` // 用户转出 token 数量（UI 单位）
}
```

```go
// GetTxInfo 响应，前端据此打包交易
type GiveTokenTxInfo struct {
    RewardAccount string  `json:"rewardAccount"` // 服务端下发奖励的地址
    Mint          string  `json:"mint"`          // token mint 地址
    CostAccount   string  `json:"costAccount"`   // 接收 SOL 成本费的地址
    CostFeeRate   float64 `json:"costFeeRate"`   // 成本费率（实际为 RewardRate）
    MaxCostFee    float64 `json:"maxCostFee"`    // 奖励金额上限（实际为 MaxValidReward）
    Decimals      int32   `json:"decimals"`      // token 精度
    QuoteSOLPrice float64 `json:"quoteSOLPrice"` // token/SOL 价格

    TotalReward     float64 `json:"totalReward"`     // 整笔交易下发的 token 奖励总额（raw）
    QuotedSOLAmount float64 `json:"quotedSOLAmount"` // 需支付的 SOL 成本费金额（lamports）
    CostFee         uint64  `json:"costFee"`         // 需支付的 SOL 成本费，lamports 整数

    Claims            []model.LevelRatio `json:"claims"`            // 各层邀请人奖励费率
    GiveInfo          RewardTokenItem    `json:"giveInfo"`          // from → to 的转账项
    RewardInfo        RewardTokenItem    `json:"rewardInfo"`        // 平台奖励 from 的项
    RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"` // 平台奖励各层邀请人的项
}
```

---

## 4. 业务流程

### 4.1 阶段一：获取交易参数（GetTxInfo）

```mermaid
sequenceDiagram
    actor FE as 前端
    participant R as reward-api
    participant B as base-api
    participant SOL as Solana RPC
    participant DB as PostgreSQL

    FE->>B: GET /base/fee/priority
    B-->>FE: FeeDetail{Low,Medium,High,Extreme}
    FE->>B: GET /base/fee/inst-units
    B-->>FE: InstUnits{TransferChecked,...}

    FE->>R: POST /reward/give/tx-info {to, amount}<br/>Authorization: Bearer JWT
    R->>R: 从 JWT 提取 fromAccount
    R->>SOL: GetAccountInfo(to 的 tokenAccount)
    Note over R: toTokenAccountExists 决定奖励费率<br/>存在→RewardRate / 不存在→ValidRate
    R->>DB: 查 t_invite_relation 获取 from 的上级邀请人
    R->>B: Dubbo GetTokenQuoteSOLPrice
    B-->>R: token/SOL 价格
    R-->>FE: GiveTokenTxInfo{GiveInfo, RewardInfo, RewardInviterInfo, CostFee,...}
```

### 4.2 阶段二：提交交易（CommitTx）

```mermaid
sequenceDiagram
    actor FE as 前端
    participant R as reward-api
    participant B as base-api
    participant SOL as Solana RPC
    participant DB as PostgreSQL

    Note over FE: 构造交易并用自己私钥签名 Signatures[0]（见第 5 节）

    FE->>R: POST /reward/give/commit-tx {encodedTx(hex), to}
    R->>R: PreCheckEncodedTx（反序列化 + 验证签名，提取 from）
    R->>R: 获取分布式锁（key: give-token:process:commit-tx:{from}，1h）
    R->>DB: accountValid(to)（校验 to 是否为有效地址，见第 6 节）
    R->>DB: getUpInviters(from)（查 from 的上级邀请人列表）
    R->>R: decodeSOLTx（按 ProgramID 解析所有指令）
    R->>B: Dubbo GetTokenQuoteSOLPrice（用于 SOL 成本费校验）
    R->>R: checkSOLTx（校验地址 / 金额 / 数量，见第 7 节）

    R->>B: Dubbo SendTransaction{encodedTx, "Reward", "GiveToken"}
    B->>B: 用服务私钥补签 Signatures[1]
    B->>DB: INSERT t_service_tx{state=0}
    B-->>B: goroutine 广播交易
    B->>SOL: SendTransaction
    B-->>R: txId（= Signatures[0]）

    R->>DB: INSERT t_service_tx{state=0}（唯一键冲突则跳过）
    R->>DB: INSERT t_give_token_record{state=0}
    R-->>FE: txId
```

### 4.3 阶段三：链上确认与 Kafka 处理

```mermaid
sequenceDiagram
    participant B as base-api
    participant SOL as Solana RPC
    participant K as Kafka
    participant R as reward-api
    participant DB as PostgreSQL
    actor FE as 前端

    loop TxScanTask 每 3s
        B->>SOL: getSignaturesForAddress(rewardTokenAccount)
        SOL-->>B: []TransactionSignature
        B->>DB: 查 t_service_tx(txId) 确认是本系统交易
        B->>DB: UPDATE t_service_tx tx_state=TxFetchSuccess
        B->>K: NewScannedTransaction{TxSig, DecodedTx}
    end

    K->>R: consume NewScannedTransaction
    alt TxSig.Err == nil（链上成功）
        R->>DB: UPDATE t_give_token_record state=1
        R->>DB: INSERT t_fund_flow（upsert：SOL入账 + token出账×(1+N)）
    else TxSig.Err != nil（链上失败）
        R->>DB: UPDATE t_give_token_record state=-1
    end

    Note over B: TxExpireTask 兜底（5min 后仍 state=0）
    B->>SOL: GetTransaction(txSig)
    alt 找到交易
        B->>K: NewScannedTransaction（同上路径）
    else ErrNotFound
        B->>K: NewExpiredTransaction
        K->>R: consume NewExpiredTransaction
        R->>DB: UPDATE t_give_token_record state=-1
    end

    loop 前端轮询
        FE->>R: POST /reward/give/record {txId}
        R->>DB: 查 t_give_token_record.state
        R-->>FE: state：0=进行中 / 1=成功 / -1=失败
    end
```

---

## 5. 前端构造交易指令

```mermaid
flowchart TD
    A([获得 GiveTokenTxInfo]) --> B["① SetComputeUnitPrice(Medium)"]
    B --> C["② SetComputeUnitLimit\n= (N × TransferChecked_CU + 500) × 1.2\nN = 2 + len(RewardInviterInfo)"]
    C --> D{to 没有 token account?}
    D -- 是 --> E["③ CreateAssociatedTokenAccount(to)\n由 from 付租金"]
    D -- 否 --> F
    E --> F["④ TransferChecked\nfromTokenAccount → toTokenAccount\namount = GiveInfo.Amount，owner = from"]
    F --> G["⑤ TransferChecked\nrewardTokenAccount → fromTokenAccount\namount = RewardInfo.Amount，owner = rewardNativeAccount"]
    G --> H["⑥ System.Transfer\nfrom → costAccount\namount = CostFee（lamports）"]
    H --> I{有邀请人?}
    I -- 是 --> J["⑦ TransferChecked × N\nrewardTokenAccount → inviterTokenAccount[i]\namount = RewardInfo.Amount × Ratio[i]，owner = rewardNativeAccount"]
    I -- 否 --> K
    J --> K(["from 私钥签名 Signatures[0]\nhex 序列化 → encodedTx"])
```

---

## 6. 有效地址判断（`logic/give/logic.go:accountValid`）

`to` 地址满足以下全部条件才认为是**有效地址**（valid=true），使用 `ValidRate` 计算奖励；否则为普通地址，使用 `RewardRate`。

```mermaid
flowchart TD
    A([accountValid to]) --> B{to 是合法 PublicKey?}
    B -- 否 --> Z([无效：valid=false])
    B -- 是 --> C{t_take_token_record 中\nreceipt_account=to\ntx_state >= 0?}
    C -- 存在 --> Z
    C -- 不存在 --> D{t_give_token_record 中\nreceipt_account=to\ntx_state >= 0?}
    D -- 存在 --> Z
    D -- 不存在 --> E{链上 to 的 token account\n是否存在?}
    E -- 存在 --> Z
    E -- 不存在 --> V([有效：valid=true，使用 ValidRate])
```

---

## 7. CommitTx 服务端处理

```mermaid
flowchart TD
    A([POST /reward/give/commit-tx]) --> B["PreCheckEncodedTx\nhex decode → 反序列化\n校验 Signatures[0] 有效\n提取 from / txId"]
    B --> C["获取分布式锁\ngive-token:process:commit-tx:{from}\n防止同一地址并发重复提交"]
    C --> D["accountValid(to)（见第 6 节）"]
    D --> E["getUpInviters(from)\n查 t_invite_relation 向上最多 levelDist 层"]
    E --> F["decodeSOLTx\n按 ProgramID 解析：ComputeBudget / SPLToken / System\n未知 Program → 直接报错拒绝"]
    F --> G{checkSOLTx}
    G --> G1["数量校验\nTransferInstructions == 1\nTransferChecked == upInviters数 + 2"]
    G --> G2["SOL 转账校验\nto == CostAccount\nfrom == 交易发起人\namount >= TotalReward/decimal × price × LAMPORTS × (1-MaxLessRate)"]
    G --> G3["用户转账指令校验\nmint == 配置 mint\nfrom == FindATA(txFrom, mint)\nowner == txFrom\namount > 0，decimals 匹配"]
    G --> G4["奖励 from 指令校验\nfrom == RewardTokenAccount\nto == FindATA(txFrom, mint)\nowner == RewardAccount\namount ≤ min(MaxValidReward, transferAmount × Rate)\ndecimals 匹配"]
    G --> G5["奖励邀请人指令校验（循环）\nfrom == RewardTokenAccount\nowner == RewardAccount\namount == rewardFromAmount × Ratio[i]\ndecimals 匹配"]
    G1 & G2 & G3 & G4 & G5 --> H{全部通过?}
    H -- 否 --> I([返回错误])
    H -- 是 --> J["Dubbo base.SendTransaction\n补签 + 写 t_service_tx + 异步广播"]
    J --> K["写 t_service_tx（唯一键冲突则跳过）\n写 t_give_token_record{state=0}"]
    K --> L([返回 txId])
```

---

## 8. Kafka 处理逻辑（`logic/give/kafka.go`）

```mermaid
flowchart TD
    subgraph S["HandleScannedTx"]
        direction TD
        A([收到 NewScannedTx]) --> B[查 t_give_token_record by txId]
        B --> C{TxSig.Err?}
        C -- 链上失败 --> D[UPDATE state=-1]
        C -- 链上成功 --> E[UPDATE state=1]
        E --> F[DecodeServiceTransaction]
        F --> G["INSERT t_fund_flow（upsert，唯一键=tx_id+to_account+flow_type）\n① SOL 入账：成本费（FlowCost）\n② token 出账：奖励 from（FlowReceipt）\n③ token 出账×N：奖励各邀请人（FlowInviter）"]
        G --> H([Commit])
        D --> H
    end

    subgraph E2["HandleExpiredTx"]
        direction TD
        A2([收到 NewExpiredTx]) --> B2[查 t_give_token_record by txId]
        B2 --> C2[UPDATE state=-1]
        C2 --> D2([Commit])
    end

    S ~~~ E2
```

> GiveToken 成功后**不**写 `t_invite_relation`，邀请关系由 TakeToken 在领取时确定。

---

## 9. HTTP API（`handler/give.go`）

所有路由挂载在 `/reward/give` 前缀下。

### POST `/reward/give/config`

获取 give token 业务规则配置。

- 请求：无 body
- 响应：`t_give_token_config` 记录

| 字段 | 类型 | 说明 |
|---|---|---|
| `rewardAccount` | string | 发放奖励的 solana 地址 |
| `costAccount` | string | 接收 SOL 成本费的地址 |
| `rewardRate` | numeric | 普通地址奖励费率（to 有 token account 时） |
| `maxValidReward` | numeric | 奖励金额上限（raw） |
| `validRate` | numeric | 有效地址奖励费率（to 无 token account 时） |

---

### POST `/reward/give/tx-info` 🔒 需要 JWT

获取打包交易所需的全部参数，`from` 地址由服务端从 JWT 中提取。

- 请求头：`Authorization: Bearer <token>`
- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `to` | string | 是 | 接收 token 的 native account |
| `amount` | float64 | 是 | 转出 token 数量（UI 单位） |

- 响应：`GiveTokenTxInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `rewardAccount` | string | 服务端下发奖励的地址 |
| `mint` | string | token mint 地址 |
| `costAccount` | string | 接收 SOL 成本费的地址 |
| `decimals` | int32 | token 精度 |
| `quoteSOLPrice` | float64 | token/SOL 价格 |
| `totalReward` | float64 | 整笔交易奖励的 token 总额（raw） |
| `quotedSOLAmount` | float64 | 需支付的 SOL 成本费（lamports） |
| `costFee` | uint64 | 需支付的 SOL 成本费，lamports 整数；前端构造 `System.Transfer` 时直接使用 |
| `claims` | []LevelRatio | 各层邀请人奖励费率配置 |
| `giveInfo` | RewardTokenItem | from → to 的转账项（index/receiptAccount/amount） |
| `rewardInfo` | RewardTokenItem | 平台奖励 from 的项 |
| `rewardInviterInfo` | []RewardTokenItem | 平台奖励各层邀请人的项列表 |

---

### POST `/reward/give/commit-tx`

提交签名后的交易，服务端校验并广播到链上。

- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `encodedTx` | string | 是 | hex 编码的已签名 Solana 交易二进制 |
| `to` | string | 是 | 接收 token 的 native account（与构造交易时一致） |

- 响应：`txId`（string，链上交易 ID = `tx.Signatures[0]`）

- 失败场景：

| 场景 | 说明 |
|---|---|
| 同地址并发提交 | 分布式锁保护，同一地址同时只处理一笔 |
| 交易指令校验失败 | 地址/金额/数量不符合规则（见第 7 节） |
| base 广播失败 | Solana RPC 拒绝交易 |

---

### POST `/reward/give/record`

根据交易 ID 查询 give token 业务记录，用于前端轮询状态。

- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `txId` | string | 是 | commit-tx 返回的交易 ID |

- 响应：`t_give_token_record` 记录

| 字段 | 类型 | 说明 |
|---|---|---|
| `txId` | string | 交易 ID |
| `fromAccount` | string | 转出 token 的地址 |
| `receiptAccount` | string | 接收 token 的地址 |
| `amount` | numeric | 转出 token 金额（raw） |
| `txState` | int | `0`=进行中 / `1`=成功 / `-1`=失败 |
