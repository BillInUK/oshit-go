# 日志体系演进规划

## 当前状态（阶段 1 + 2 已完成）

### 阶段 1：日志轮转（已完成）

使用 `lumberjack` 实现自动日志轮转，在三个服务的 `main()` 中初始化：

| 配置项 | 值 | 说明 |
|---|---|---|
| MaxSizeMB | 100 | 单文件超过 100MB 自动轮转 |
| MaxAgeDays | 30 | 归档文件保留 30 天 |
| MaxBackups | 10 | 最多保留 10 个归档文件 |
| Compress | true | 归档自动 gzip 压缩 |

轮转后的文件命名：`server-2026-06-03T12-00-00.000.log.gz`

### 阶段 2：JSON 结构化输出（已完成）

使用 Go 标准库 `slog` 的 JSONHandler 作为底层，通过 fiber adapter 替换全局 logger，业务代码零改动。

输出格式：
```json
{"time":"2026-06-03T12:00:00.000+08:00","level":"INFO","source":{"function":"...","file":"scan.go","line":161},"msg":"启动扫描任务 [reward:take token:EHEu...]","service":"base-api"}
```

代码位置：`common/pkg/logging/`

| 文件 | 职责 |
|---|---|
| `logger.go` | Setup 函数：初始化 slog + lumberjack，替换标准库 log 和 fiber log |
| `fiber_adapter.go` | 实现 fiber v2 的 AllLogger 接口，将 Infof/Errorf 等调用委托给 slog |

---

## 阶段 3：日志集中收集与查询（待实施）

### 3.1 目标

- 三个服务的日志统一收集到一个平台，支持全文搜索、过滤、告警
- 支持按 service、level、时间范围查询
- 为后续链路追踪（tracing）做准备

### 3.2 推荐方案：OpenObserve

选择 OpenObserve 而非 ELK 的理由：

| 对比项 | OpenObserve | ELK |
|---|---|---|
| 部署复杂度 | 单二进制，1 条命令启动 | 需要 Elasticsearch + Logstash + Kibana 三组件 |
| 内存消耗 | ~100MB 起步 | Elasticsearch 至少需要 2GB+ |
| 存储 | 支持本地磁盘或 S3 | 依赖 Elasticsearch 索引 |
| 数据摄入 | 兼容 Elasticsearch API | 原生 |
| 功能覆盖 | 日志 + 指标 + 链路追踪 | 日志为主，追踪需额外组件 |
| 适合场景 | 中小团队，3-10 个服务 | 大规模，50+ 服务 |

### 3.3 架构设计

```
                                    ┌──────────────────┐
base-api  → ./log/server.log  ──┐   │                  │
                                │   │   OpenObserve    │
reward-api → ./log/server.log ──┼─→ │                  │ → Web UI 查询/告警
                                │   │  (单机部署)       │
pos-api  → ./log/server.log  ──┘   │                  │
                                    └──────────────────┘
         ↑                      ↑
    JSON 日志文件          Vector (采集 agent)
    (阶段 2 已完成)         读取日志文件并推送
```

### 3.4 实施步骤

#### 步骤 1：部署 OpenObserve

```bash
# 在日志服务器上（可以和 Nacos 同一台，或独立部署）
curl -L https://raw.githubusercontent.com/openobserve/openobserve/main/download.sh | sh

# 启动（首次会创建 admin 账户）
ZO_ROOT_USER_EMAIL=admin@oshit.io \
ZO_ROOT_USER_PASSWORD=<password> \
./openobserve

# 默认监听 5080 端口，Web UI: http://<host>:5080
```

建议配置：
- 存储：本地磁盘（小规模）或 S3（长期保留）
- 数据保留：30 天（与日志文件轮转一致）

#### 步骤 2：部署 Vector（日志采集 agent）

在每台应用服务器上部署 Vector，配置读取 JSON 日志并推送到 OpenObserve：

```toml
# /etc/vector/vector.toml

# 数据源：读取三个服务的 JSON 日志
[sources.base_logs]
type = "file"
include = ["/data/dist/oshit/base/backend/log/server.log"]
read_from = "end"

[sources.reward_logs]
type = "file"
include = ["/data/dist/oshit/reward/backend/log/server.log"]
read_from = "end"

[sources.pos_logs]
type = "file"
include = ["/data/dist/oshit/pos/backend/log/server.log"]
read_from = "end"

# 解析 JSON
[transforms.parse_json]
type = "remap"
inputs = ["base_logs", "reward_logs", "pos_logs"]
source = '''
. = parse_json!(.message)
'''

# 输出到 OpenObserve（兼容 Elasticsearch bulk API）
[sinks.openobserve]
type = "elasticsearch"
inputs = ["parse_json"]
endpoints = ["http://<openobserve-host>:5080/api/default/"]
auth.strategy = "basic"
auth.user = "admin@oshit.io"
auth.password = "<password>"
bulk.index = "oshit-logs"
```

#### 步骤 3：配置告警规则

在 OpenObserve Web UI 中配置：

| 告警 | 条件 | 通知方式 |
|---|---|---|
| 服务错误率突增 | `level = "ERROR"` 5分钟内超过 50 条 | Webhook / 邮件 |
| 交易广播失败 | `msg LIKE "%广播%失败%"` | Webhook |
| Kafka 消费异常 | `msg LIKE "%Kafka%错误%"` | Webhook |
| 服务宕机 | 某 service 10 分钟无日志 | Webhook |

#### 步骤 4（可选）：链路追踪

OpenObserve 原生支持 OpenTelemetry traces。后续可在 Go 代码中接入 OTEL SDK：

1. 在 HTTP handler 中注入 trace context
2. 在 Dubbo RPC 调用中传递 trace ID
3. 在 Kafka 消息中携带 trace ID
4. 全链路查询：用户请求 → reward 校验 → base 广播 → Kafka → reward 确认

### 3.5 资源评估

| 组件 | 服务器 | 配置建议 |
|---|---|---|
| OpenObserve | 独立部署或与 Nacos 同机 | 2 核 4GB（日志量 <10GB/天） |
| Vector | 每台应用服务器 | 几乎无额外资源消耗（~50MB 内存） |

### 3.6 前置条件

- [x] 阶段 1：日志轮转（已完成）
- [x] 阶段 2：JSON 结构化输出（已完成）
- [ ] 确认 OpenObserve 部署服务器
- [ ] 确认 AWS 安全组开放 5080 端口（OpenObserve Web UI）
- [ ] 部署 Vector 并测试日志摄入
