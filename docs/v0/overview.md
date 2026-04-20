# oshit-go 项目总览

> 面向 Claude Code 使用的项目全局参考文档。详细服务文档请参阅各服务独立文档。

---

## 1. 项目简介

oshit-go 是基于 Solana 区块链的 Token 奖励平台后端，采用 Go 微服务架构。前端用户完成特定行为（如领取 Token、赠出 Token）后，通过链上交易获得奖励，并通过邀请链机制将部分奖励分配给各层邀请人。

---

## 2. 技术栈

| 层次 | 组件 | 库 |
|---|---|---|
| Web 框架 | HTTP 服务 | `github.com/gofiber/fiber/v2` |
| ORM | 数据库访问 | `gorm.io/gorm` + `gorm.io/driver/postgres` |
| 缓存 | Redis | `github.com/redis/go-redis/v9` |
| 分布式锁 | 基于 Redis | `github.com/go-redsync/redsync/v4` |
| 消息队列 | Kafka | `github.com/segmentio/kafka-go` |
| 服务间 RPC | Dubbo Triple 协议 | `dubbo.apache.org/dubbo-go/v3` |
| 注册中心 | Nacos | `127.0.0.1:8848` |
| 区块链 | Solana SDK | `github.com/gagliardetto/solana-go` |
| 区块链节点 | QuickNode | RPC/WSS |
| 配置加载 | Viper | `github.com/spf13/viper` |
| JSON | json-iterator | `github.com/json-iterator/go`（禁用 6 位小数截断） |
| 数据库 | PostgreSQL | 主键使用 ULID |

---

## 3. 目录结构

```
oshit-go/
├── app/
│   ├── base/api/         # Base 服务（端口 1100 HTTP + 20880 Dubbo）
│   ├── reward/api/       # Reward 服务（端口 1200 HTTP）
│   └── utils/            # 应用层工具（tx.go, rpc.go）
├── common/
│   ├── constants/        # 业务常量（交易状态、资金流向类型等）
│   ├── pkg/
│   │   ├── dal/          # GORM 自动生成的 model + query
│   │   ├── entity/       # 跨服务共享数据结构（Kafka 消息、解码后交易）
│   │   ├── pb/base/      # Protobuf 生成代码（Dubbo Triple 协议，base 服务 RPC 接口）
│   │   └── response/     # HTTP 响应结构
│   └── utils/            # 通用工具（amount, jwt, jasypt, solana_util）
├── repositories/
│   ├── base_structure.sql          # Base 服务所有表的 DDL
│   └── reward_structure.sql        # Reward 服务所有表的 DDL
└── docs/                 # 参考文档（本文件所在目录）
```

---

## 4. 服务划分

| 服务 | 目录 | 端口 | 职责 |
|---|---|---|---|
| base-api | `app/base/api` | HTTP 1100 / Dubbo 20880 | 基础能力：链上交易扫描、Kafka 推送、费率/价格/邀请等 HTTP API、Dubbo 对内服务 |
| reward-api | `app/reward/api` | HTTP 1200 | 业务层：TakeToken / GiveToken，消费 Kafka 消息，写奖励记录、资金流向、邀请关系 |

两个服务共享同一个 PostgreSQL 实例和 Kafka 集群。服务间通信使用 Dubbo Triple（注册在 Nacos）。

---

## 5. `common/utils/` 工具包

| 文件 | 主要函数 | 说明 |
|---|---|---|
| `amount.go` | `ToAmountRaw(amount, decimals) uint64` | 字面值金额 → 链上最小单位（用 `big.Float` 避免浮点误差） |
| | `ToAmountUI(raw, decimals) float64` | 链上最小单位 → 字面值金额 |
| `jwt_generator.go` | `GenerateNewTokens(id, credentials)` | 生成 Access Token（HS256，5天有效）+ Refresh Token（SHA256，5小时有效） |
| | `GenerateRandomString(length)` | 生成随机字母数字字符串（用于生成邀请码） |
| | `ParseRsaPublicKeyFromPemStr(pem)` | 从 PEM 字符串解析 RSA 公钥 |
| `jwt_parser.go` | `ParseToken(token)` | 解析并验证 JWT，提取 claims |
| `jsypt_util.go` | `Decrypt(encrypted, password)` | jasypt PBEWithHMACSHA512AndAES_256 解密（用于解密 DB 中存储的私钥） |
| `solana_util.go` | `CalcGasFee(...)` | 计算 Solana 交易 gas 费 |
| | `GetSPLTokenAccountOwner(rpc, tokenAccount)` | 从链上 RPC 查 SPL Token Account 对应的 native owner |
| | `VerifySolanaSignedMessage(pubKey, msg, sig)` | 验证 Solana 钱包签名的消息（用于登录） |

---

## 6. `app/utils/` 应用工具包

| 文件 | 主要函数 | 说明 |
|---|---|---|
| `rpc.go` | `AcquireDistributedRateLimit(rdb, key, limit)` | 基于 Redis INCR 的分布式限流（每秒 N 次），用于 TxScanTask 控制 RPC 调用频率 |
| | `IsRpcRateLimitedError(err)` | 判断错误是否为 RPC 请求限制（含 "request limit reached"） |
| `tx.go` | `PreCheckEncodedTx(encodedTx)` | hex 解码 → 反序列化 `solana.Transaction` → 校验 `Signatures[0]` 签名有效性，返回 `PreCheckedTx{From, TxId, SOLTx}` |
| | `DecodeSolanaTransaction(rpc, db, tx, txID)` | 解析 Solana 交易所有指令（ComputeBudget / ATA / TransferChecked / System.Transfer），返回 `DecodedSolanaTransaction` |
| | `DecodeServiceTransaction(decodedTx)` | 将底层解码结果转换为业务语义结构（发起人转 token、奖励领取人、奖励邀请人） |
| | `CalDecodedTxFee(sigNum, decodedTx)` | 计算交易 gas 费 |
| | `QueryNativeAccountByTokenAccount(rpc, db, tokenAccount)` | 先查 DB，再查链上 RPC，获取 token account 对应的 native account |

---

## 7. `common/pkg/entity/` 跨服务数据结构

| 文件 | 类型 | 说明 |
|---|---|---|
| `kafka.go` | `KafkaTxMsg` | Kafka 消息：`MsgType`, `Service`, `SubService`, `MsgContent`（JSON） |
| | `NewScannedTx` | `MsgContent` 的具体类型：`TxSig`（含 Err）、`DecodedTx` |
| | `NewExpiredTx` | 超时消息：`TxID` |
| `tx.go` | `DecodedSolanaTransaction` | 底层解码结果：`TransferCheckedInstructions[]`、`TransferInstructions[]`、`ComputeUnitPrice/Limit` |
| | `DecodedServiceTransaction` | 业务层解码结果：`TransferTokenInst`（用户转 token）、`RewardInst`（奖励领取人）、`RewardInviterInst[]`（奖励各邀请人）、`ToDexInst`（SOL 成本费） |

---

## 8. `common/constants/` 业务常量

| 文件 | 常量 | 值 |
|---|---|---|
| `service_const.go` | `TxFetchInit` | 0（t_service_tx 初始态） |
| | `TxFetchSuccess` | 1 |
| | `TxFetchFailed` | -1 |
| | `FlowInput` / `FlowOutput` | 资金流向（入账/出账） |
| | `ServiceTakeToken` | 1（业务类型 - 官方领取奖励） |
| | `ServiceGiveToken` | 3（业务类型 - 官方转账） |
| | `ServiceTransferToken` | 2（业务类型 - 非官方转账） |
| | `FlowTakTokenCost` | TakeToken DEX 成本费入账 flow_type |
| | `FlowTakeTokenReceipt` | TakeToken Token 奖励出账（领取人） |
| | `FlowTakeTokenInviter` | TakeToken Token 奖励出账（邀请人） |
| | `FlowGiveTokenCost` | GiveToken DEX 成本费入账 flow_type |
| | `FlowGiveTokenReceipt` | GiveToken Token 奖励出账（领取人） |
| | `FlowGiveTokenInviter` | GiveToken Token 奖励出账（邀请人） |
