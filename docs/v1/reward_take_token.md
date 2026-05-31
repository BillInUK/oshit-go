# 1. 业务概述

用户支付 SOL 成本费（打给 dex 账户），平台从奖励账户向用户发 token；若存在有效邀请关系，还按层级比例向各层邀请人发 token。

整体分四个阶段：**GetTxInfo → 前端构造交易 → CommitTx → 链上确认（Kafka）**

---

# 2. 数据库表

* t_take_token_config
```sql
-- take token 奖励配置表
DROP TABLE IF EXISTS public.t_take_token_config;
CREATE TABLE public.t_take_token_config
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    invite_code    character varying(16)       DEFAULT NULL::character varying,    -- 邀请码，默认规则为空
    reward_account character varying(64)                                 NOT NULL, -- 下发奖励的 solana 地址
    cost_account   character varying(64)                                 NOT NULL, -- 接收成本费的 solana 地址
    amount         numeric(78, 0)                                        NOT NULL, -- 奖励金额
    invite_amount  numeric(78, 0)                                        NOT NULL, -- 确定邀请关系时的奖励金额
    cost_fee_rate  numeric(78, 0)                                        NOT NULL, -- 成本费费率
    max_cost_fee   numeric(78, 0)                                        NOT NULL, -- 最大成本费
    is_default     boolean                     DEFAULT false,                      -- 是否是默认规则
    reward_inviter boolean                     DEFAULT true,                       -- 是否奖励邀请人
    invited        boolean                     DEFAULT true,                       -- 是否确定邀请关系
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
```

* t_take_token_record
```sql
-- take token 记录表
DROP TABLE IF EXISTS public.t_take_token_record;
CREATE TABLE public.t_take_token_record
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    reward_account  character varying(64)                                 NOT NULL, -- 下发奖励的 solana 地址
    receipt_account character varying(64)                                 NOT NULL, -- 接收奖励的地址
    cost_account    character varying(64)                                 NOT NULL, -- 接收成本费的地址
    tx_id           character varying(128)                                NOT NULL, -- 交易 id
    amount          numeric(78, 0)                                        NOT NULL, -- 奖励金额
    cost_fee        numeric(78, 0)                                        NOT NULL, -- 成本费
    use_invite_code boolean                                               NOT NULL, -- 是否使用邀请码
    invite_code     character varying(16),                                          -- 邀请码
    tx_state        integer,                                                        -- 交易状态：0=进行中 1=成功 -1=失败
    invited         boolean,                                                        -- 是否确定邀请关系
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
```

* t_daily_claim_stats

```sql
-- 当日 take token 和 lottery 统计表
DROP TABLE IF EXISTS public.t_daily_claim_stats;
CREATE TABLE public.t_daily_claim_stats
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL, -- 地址
    take_date      date                                                  NOT NULL, -- 日期
    take_count     integer                     DEFAULT 0                 NOT NULL, -- take token 次数
    need_lottery   boolean                     DEFAULT false             NOT NULL, -- 是否需要抽奖
    last_take_time timestamp without time zone,                                    -- 上次 take token 时间
    total_lottery  numeric(78, 0)              DEFAULT 0                 NOT NULL, -- 总计 lottery 金额
    total_take     numeric(78, 0)              DEFAULT 0                 NOT NULL, -- 总计 take token 金额
    lottery_count  integer                     DEFAULT 0                 NOT NULL, -- 抽奖次数
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE t_daily_claim_stats
    ADD CONSTRAINT uq_daily_claim_stats_account_date UNIQUE (native_account, take_date);
```

---

## 3. 数据结构

```go
// 奖励发放项
type RewardTokenItem struct {
    Index          int    `json:"index"`          // 排序索引
    ReceiptAccount string `json:"receiptAccount"` // 接收奖励的 solana 地址
    Amount         uint64 `json:"amount"`         // 奖励 token 金额（raw）
}
```

```go
// GetTxInfo 请求
type GetTakeTokenTxInfoReq struct {
    InviteCode     string  `json:"inviteCode"`     // 可选，邀请码
    Custom         bool    `json:"custom"`         // 是否自定义金额（暂未使用）
    CustomAmount   float64 `json:"customAmount"`   // 自定义金额（暂未使用）
    ReceiptAccount string  `json:"receiptAccount"` // 领取奖励的 native account
}
```

```go
// GetTxInfo 响应，前端据此打包交易
type TakeTokenTxInfo struct {
    RewardAccount   string  `json:"rewardAccount"`   // 下发奖励的 solana 地址
    Mint            string  `json:"mint"`            // token mint 地址
    CostAccount     string  `json:"costAccount"`     // 接收 SOL 成本费的地址
    CostFeeRate     float64 `json:"costFeeRate"`     // 成本费率
    MaxCostFee      float64 `json:"maxCostFee"`      // 最大成本费
    Decimals        int32   `json:"decimals"`        // token 精度
    InviteCode      string  `json:"inviteCode"`      // 邀请码
    InviteCodeValid bool    `json:"inviteCodeValid"` // 邀请码是否有效
    Invited         bool    `json:"invited"`         // 本次是否将确定邀请关系
    QuoteSOLPrice   float64 `json:"quoteSOLPrice"`   // token/SOL 价格
    TotalReward     float64 `json:"totalReward"`     // 整笔交易奖励的 token 总额
    QuotedSOLAmount float64 `json:"quotedSOLAmount"` // 成本费 SOL 金额
    CostFee         uint64  `json:"costFee"`         // SOL 成本费，单位 lamports
    Claims          []model.LevelRatio `json:"claims"`            // 各层邀请人奖励费率
    RewardInfo      RewardTokenItem    `json:"rewardInfo"`        // 领取人奖励项
    RewardInviterInfo []RewardTokenItem `json:"rewardInviterInfo"` // 各邀请人奖励项
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
    participant DB as PostgreSQL

    FE->>B: GET /base/fee/priority
    B-->>FE: FeeDetail{Low,Medium,High,Extreme}
    FE->>B: GET /base/fee/inst-units
    B-->>FE: InstUnits{TransferChecked,...}

    FE->>R: POST /reward/take/tx-info {receiptAccount, inviteCode}
    R->>DB: 查 t_native_account_info(inviteCode) 找邀请人
    R->>DB: 查 t_take_token_record 是否用过邀请码
    R->>DB: 查 t_invite_relation 邀请关系是否已存在
    R->>B: Dubbo GetTokenQuoteSOLPrice
    B-->>R: token/SOL 价格
    R-->>FE: TakeTokenTxInfo{RewardInfo, RewardInviterInfo, CostFee, InviteCodeValid, Invited,...}
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

    FE->>R: POST /reward/take/commit-tx {encodedTx(hex), inviteCode}
    R->>R: PreCheckEncodedTx（反序列化 + 验证签名）
    R->>R: getTxInfo（服务端重新计算，不信任客户端缓存）
    R->>R: decodeSOLTx（按 ProgramID 解析所有指令）
    R->>R: checkDecodedSOLTx（校验地址 / 金额 / 数量，见第 6 节）

    R->>B: Dubbo SendTransaction{encodedTx, "Reward", "TakeToken"}
    B->>B: 用服务私钥补签 Signatures[1]
    B->>DB: INSERT t_service_tx{state=0}
    B-->>B: goroutine 广播交易
    B->>SOL: SendTransaction
    B-->>R: txId（= Signatures[0]）

    R->>DB: INSERT t_service_tx{state=0}（唯一键冲突则跳过）
    R->>DB: INSERT t_take_token_record{state=0, invited}
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
        R->>DB: UPDATE t_take_token_record state=1
        R->>DB: INSERT t_sol_fund_flow（upsert：dex入账 + token出账×N）
        opt takeTokenRecord.Invited == true
            R->>DB: INSERT t_invite_relation（确定邀请层级）
        end
    else TxSig.Err != nil（链上失败）
        R->>DB: UPDATE t_take_token_record state=-1
    end

    Note over B: TxExpireTask 兜底（5min 后仍 state=0）
    B->>SOL: GetTransaction(txSig)
    alt 找到交易
        B->>K: NewScannedTransaction（同上路径）
    else ErrNotFound
        B->>K: NewExpiredTransaction
        K->>R: consume NewExpiredTransaction
        R->>DB: UPDATE t_take_token_record state=-1
    end

    loop 前端轮询
        FE->>R: POST /reward/take/record {txId}
        R->>DB: 查 t_take_token_record.state
        R-->>FE: state：0=进行中 / 1=成功 / -1=失败
    end
```

---

## 5. 前端构造交易指令

```mermaid
flowchart TD
    A([获得 TakeTokenTxInfo]) --> B["① SetComputeUnitPrice(Medium)"]
    B --> C["② SetComputeUnitLimit\n= (N × TransferChecked_CU + 500) × 1.2\nN = 1 + len(RewardInviterInfo)"]
    C --> D{领取人 ATA 不存在?}
    D -- 是 --> E["③ CreateAssociatedTokenAccount(领取人)"]
    D -- 否 --> F
    E --> F["④ TransferChecked\nrewardTokenAccount → receiptTokenAccount\namount = InviteAmount（有邀请）或 Amount（无邀请）"]
    F --> G["⑤ System.Transfer\nreceiptNativeAccount → costAccount\namount = CostFee（lamports）"]
    G --> H{有邀请人?}
    H -- 是 --> I["⑥ TransferChecked × N\nrewardTokenAccount → inviterTokenAccount[i]\namount = rewardAmount × Ratio[i]"]
    H -- 否 --> J
    I --> J(["用户私钥签名 Signatures[0]\nhex 序列化 → encodedTx"])
```

---

## 6. CommitTx 服务端处理

```mermaid
flowchart TD
    A([POST /reward/take/commit-tx]) --> B["PreCheckEncodedTx\nhex decode → 反序列化\n校验 Signatures[0] 有效\n提取 From / txId"]
    B --> C["getTxInfo\n服务端重新计算所有参数"]
    C --> D["decodeSOLTx\n按 ProgramID 分类解析：\nComputeBudget / SPLToken / System / LightHouse\n未知 Program → 直接报错拒绝"]
    D --> E{checkDecodedSOLTx}
    E --> F["数量校验\nTransferInstructions == 1\nTransferChecked == Claims数 + 1"]
    E --> G["SOL 转账校验\nto == CostAccount\nfrom == 用户 native\namount >= CostFee × (1 - MaxLessRate)"]
    E --> H["主奖励 TransferChecked 校验\nfrom == RewardTokenAccount\nmint == TokenMintAccount\nto == 用户 ATA\namount 精确匹配"]
    E --> I["邀请人 TransferChecked 校验（循环）\nowner == RewardNativeAccount\namount == rewardAmount × Ratio[i]"]
    F & G & H & I --> J{全部通过?}
    J -- 否 --> K([返回错误])
    J -- 是 --> L["Dubbo base.SendTransaction\n补签 + 写 t_service_tx + 异步广播"]
    L --> M["写 t_service_tx（唯一键冲突则跳过）\n写 t_take_token_record{state=0, invited}"]
    M --> N([返回 txId])
```

---

## 7. 邀请码有效性判断（`logic/take/logic.go`）

返回值：`(directInviter, codeValid, invited, err)`
- `codeValid=true, invited=true` → 邀请码有效，本次将确定邀请关系，奖励金额用 `InviteAmount`
- `codeValid=true, invited=false` → 邀请码有效但邀请关系已存在，奖励金额用 `Amount`
- `codeValid=false` → 邀请码无效

```mermaid
flowchart TD
    A([开始]) --> B{inviteCode 为空?}
    B -- 是 --> Z1([无效：codeValid=false, invited=false])
    B -- 否 --> C{t_native_account_info\n中能找到邀请人?}
    C -- 否 --> Z1
    C -- 是 --> D{邀请人 == 自己?}
    D -- 是 --> Z1
    D -- 否 --> E{t_take_token_record 中\nreceiptAccount 有过\nuse_invite_code=true AND state=1?}
    E -- 有 --> Z2([无效：codeValid=false, invited=false])
    E -- 没有 --> F{t_invite_relation 中\nreceiptAccount 已存在?}
    F -- 存在 --> Z3([有效但不确定关系：codeValid=true, invited=false])
    F -- 不存在 --> Z4([有效且确定关系：codeValid=true, invited=true])
```

---

## 8. Kafka 处理逻辑（`logic/take/kafka.go`）

```mermaid
flowchart TD
    subgraph S["HandleScannedTx"]
        direction TD
        A([收到 NewScannedTx]) --> B[查 t_take_token_record by txId]
        B --> C{TxSig.Err?}
        C -- 链上失败 --> D[UPDATE state=-1]
        C -- 链上成功 --> E[UPDATE state=1]
        E --> F[DecodeServiceTransaction]
        F --> G["INSERT t_sol_fund_flow（upsert，唯一键=tx_id+to+flow_type）\n① SOL 入账：dex fee\n② token 出账：奖励领取人\n③ token 出账×N：奖励各邀请人"]
        G --> H{takeTokenRecord.Invited?}
        H -- true --> I["INSERT t_invite_relation\nlevel = inviter.Level + 1（inviter 无记录则 level=1）\n幂等：invitee 已存在则跳过"]
        H -- false --> J([Commit])
        I --> J
        D --> J
    end

    subgraph E2["HandleExpiredTx"]
        direction TD
        A2([收到 NewExpiredTx]) --> B2[查 t_take_token_record by txId]
        B2 --> C2[UPDATE state=-1]
        C2 --> D2([Commit])
    end

    S ~~~ E2
```

---

## 9. 邀请关系管理（`logic/invite.go`）

| 方法 | 说明 |
|---|---|
| `GetAccountByInviteCode(code)` | 查 `t_native_account_info` by invite_code |
| `CheckInviteRecord(account)` | 查 `t_invite_relation`，account 是否作为 invitee 存在 |
| `GetUpInviterRecords(account, depth)` | `WITH RECURSIVE` 向上递归查（最大 depth=20） |
| `GetDownInviteeRecords(account, depth)` | `WITH RECURSIVE` 向下递归查 |
| `InviteRelationExist(account)` | 查是否作为 inviter 或 invitee 存在（用于前置检查） |
| `RecordDetermineInvitationHierarchy(...)` | 写 `t_invite_relation`（幂等：invitee 已存在则跳过） |
| `BuildSortedInviterItems(...)` | 构建排序好的邀请人奖励列表（供 getTxInfo 使用） |

---

## 10. HTTP API（`handler/take.go`）

所有路由挂载在 `/reward/take` 前缀下。

### POST `/reward/take/config`

获取 take token 业务规则配置。

- 请求：无 body
- 响应：`t_take_token_config` 记录

| 字段 | 类型 | 说明 |
|---|---|---|
| `rewardAccount` | string | 下发奖励的 solana 地址 |
| `costAccount` | string | 接收 SOL 成本费的地址 |
| `amount` | numeric | 无邀请码时的奖励金额（raw） |
| `inviteAmount` | numeric | 有邀请码且确定邀请关系时的奖励金额（raw） |
| `costFeeRate` | numeric | 成本费费率 |
| `maxCostFee` | numeric | 最大成本费 |
| `rewardInviter` | boolean | 是否奖励邀请人 |
| `invited` | boolean | 是否确定邀请关系 |

---

### POST `/reward/take/tx-info`

获取打包交易所需的全部参数，前端据此构造 Solana 交易。

- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `receiptAccount` | string | 是 | 领取奖励的 native account |
| `inviteCode` | string | 否 | 邀请码 |

- 响应：`TakeTokenTxInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `rewardAccount` | string | 下发奖励的 solana 地址 |
| `mint` | string | token mint 地址 |
| `costAccount` | string | 接收 SOL 成本费的地址 |
| `costFeeRate` | float64 | 成本费率 |
| `maxCostFee` | float64 | 最大成本费 |
| `decimals` | int32 | token 精度 |
| `inviteCode` | string | 邀请码（原样返回） |
| `inviteCodeValid` | boolean | 邀请码是否有效 |
| `invited` | boolean | 本次是否将确定邀请关系 |
| `quoteSOLPrice` | float64 | token/SOL 价格 |
| `totalReward` | float64 | 整笔交易奖励的 token 总额（raw） |
| `quotedSOLAmount` | float64 | 需支付的 SOL 成本费金额（lamports） |
| `costFee` | uint64 | 后端计算后的 SOL 成本费，单位 lamports；前端直接用于 System.Transfer |
| `claims` | []LevelRatio | 各层邀请人奖励费率配置 |
| `rewardInfo` | RewardTokenItem | 领取人奖励项（index/receiptAccount/amount） |
| `rewardInviterInfo` | []RewardTokenItem | 各层邀请人奖励项列表（按层级排序） |

---

### POST `/reward/take/commit-tx`

提交签名后的交易，服务端校验并广播到链上。

- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `encodedTx` | string | 是 | hex 编码的已签名 Solana 交易二进制 |
| `inviteCode` | string | 否 | 邀请码（需与 tx-info 时一致） |

- 响应：`txId`（string，链上交易 ID = `tx.Signatures[0]`）

- 失败场景：

| 场景 | 说明 |
|---|---|
| 未完成抽奖 | 当日需先完成 lottery 才能继续领取 |
| 同地址并发提交 | 分布式锁保护，同一地址同时只处理一笔 |
| 交易指令校验失败 | 地址/金额/数量不符合规则（见第 6 节） |
| base 广播失败 | Solana RPC 拒绝交易 |

---

### POST `/reward/take/record`

根据交易 ID 查询 take token 业务记录，用于前端轮询状态。

- 请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `txId` | string | 是 | commit-tx 返回的交易 ID |

- 响应：`t_take_token_record` 记录

| 字段 | 类型 | 说明 |
|---|---|---|
| `txId` | string | 交易 ID |
| `receiptAccount` | string | 领取奖励的地址 |
| `rewardAccount` | string | 下发奖励的地址 |
| `amount` | numeric | 奖励金额（raw） |
| `costFee` | numeric | 实际支付的成本费（raw） |
| `useInviteCode` | boolean | 是否使用了邀请码 |
| `inviteCode` | string | 邀请码 |
| `txState` | int | `0`=进行中 / `1`=成功 / `-1`=失败 |
| `invited` | boolean | 是否确定了邀请关系 |
