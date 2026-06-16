# 本地开发环境搭建指南

本文档指导新开发者从零搭建 oshit-go 本地开发环境。

---

## 1. 前置要求

| 工具 | 版本要求 | 安装方式 |
|---|---|---|
| Go | >= 1.21 | `brew install go` |
| Docker | >= 24.x | [Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| AWS CLI v2 | >= 2.x | `brew install awscli` |
| protoc | >= 3.x | `brew install protobuf`（仅修改 proto 文件时需要） |

确保 `$GOPATH/bin` 在 PATH 中：

```bash
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```

---

## 2. 启动 Docker 容器

依次启动以下 5 个容器。所有容器均使用 host 网络端口映射，无需额外网络配置。

### 2.1 PostgreSQL

```bash
docker run --name postgres16.2 \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=oshit_db \
    -p 5432:5432 \
    -d postgres:16.2
```

启动后验证：

```bash
docker exec -it postgres16.2 psql -U postgres -d oshit_db -c "SELECT version();"
```

### 2.2 Redis

```bash
docker run --name redis \
    -p 6379:6379 \
    -d redis:latest
```

验证：

```bash
docker exec -it redis redis-cli ping
# 预期输出: PONG
```

### 2.3 Kafka

使用 Apache 官方镜像（内置 KRaft，无需 Zookeeper）：

```bash
docker run --name kafka \
    -e KAFKA_NODE_ID=1 \
    -e KAFKA_PROCESS_ROLES=broker,controller \
    -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093 \
    -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
    -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
    -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
    -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
    -e CLUSTER_ID=MkU3OEVBNTcwNTJENDM2Qk \
    -p 9092:9092 \
    -d apache/kafka:latest
```

（可选）启动 Kafka UI 方便查看 topic 和消息：

```bash
docker run --name kafka-ui \
    -e DYNAMIC_CONFIG_ENABLED=true \
    -e KAFKA_CLUSTERS_0_NAME=local \
    -e KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS=host.docker.internal:9092 \
    -p 9080:8080 \
    -d provectuslabs/kafka-ui:latest
```

Kafka UI 访问地址：http://localhost:9080

### 2.4 Nacos

```bash
# 1. 生成 auth token 和 identity key（首次搭建时执行，记录输出值）
# openssl rand -base64 36    → 用作 NACOS_AUTH_TOKEN
# openssl rand -hex 8        → 用作 NACOS_AUTH_IDENTITY_KEY
# openssl rand -base64 24    → 用作 NACOS_AUTH_IDENTITY_VALUE

# 2. 启动 Nacos（standalone 模式，内置 Derby 数据库）
docker run --name nacos-standalone-derby \
    -e MODE=standalone \
    -e NACOS_AUTH_TOKEN=pHVWBPFlwQ9EYMntyyzIQPweJOdU7BerkjGFsV6aohFBGjxW \
    -e NACOS_AUTH_IDENTITY_KEY=4bb5bd15aa29f5e8 \
    -e NACOS_AUTH_IDENTITY_VALUE=+2a+Nwzd038BV5pEQDANara7fTwVbMXt \
    -p 8080:8080 \
    -p 8848:8848 \
    -p 9848:9848 \
    -d nacos/nacos-server:latest
```

> 如果团队已有统一的 auth token，请向项目负责人获取并替换上面的值。

Nacos 控制台：http://127.0.0.1:8080/index.html

首次登录会提示设置用户名密码，设置完成后请记录。`application.yaml` 中的 `nacos.username` / `nacos.password` 为加密后的值，需要与 Nacos 实际用户名密码对应。

### 2.5 XXL-Job（可选）

当前代码中 XXL-Job 尚未正式集成（仅有 TODO 标记），本地开发暂不需要启动。如后续集成，启动方式如下：

```bash
# 需要先在 PostgreSQL 或 MySQL 中初始化 xxl-job 的 schema
docker run --name xxl-job-admin \
    -e PARAMS="--spring.datasource.url=jdbc:mysql://host.docker.internal:3306/xxl_job?useUnicode=true&characterEncoding=UTF-8 \
               --spring.datasource.username=root \
               --spring.datasource.password=root" \
    -p 8888:8080 \
    -d xuxueli/xxl-job-admin:2.4.1
```

### 2.6 确认所有容器运行正常

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

预期输出（XXL-Job 可选）：

```
NAMES                      STATUS       PORTS
kafka-ui                   Up ...       0.0.0.0:9080->8080/tcp
kafka                      Up ...       0.0.0.0:9092->9092/tcp
nacos-standalone-derby     Up ...       0.0.0.0:8080->8080/tcp, 0.0.0.0:8848->8848/tcp, 0.0.0.0:9848->9848/tcp
redis                      Up ...       0.0.0.0:6379->6379/tcp
postgres16.2               Up ...       0.0.0.0:5432->5432/tcp
```

---

## 3. 配置 AWS CLI（用于 KMS 解密配置密钥）

服务启动时需要调用 AWS KMS 解密配置密钥（AES key），用于解密 `application.yaml` 和数据库中 `ENC~` 前缀的加密字段。

### 3.1 获取 IAM 用户凭证

向项目负责人申请一个具有 `kms:Decrypt` 权限的 IAM 用户，获取：

- `AWS Access Key ID`
- `AWS Secret Access Key`
- KMS Key 所在 Region（默认 `ap-southeast-1`）

### 3.2 配置 AWS Profile

```bash
aws configure --profile oshit-dev
```

按提示输入：

```
AWS Access Key ID [None]: <你的 Access Key ID>
AWS Secret Access Key [None]: <你的 Secret Access Key>
Default region name [None]: ap-southeast-1
Default output format [None]: json
```

验证配置：

```bash
aws sts get-caller-identity --profile oshit-dev
```

应返回你的 IAM 用户 ARN。

### 3.3 放置加密密钥文件

向项目负责人获取 `oshit-config-key.enc` 文件（base64 编码的 KMS 密文），放置到默认路径：

```bash
sudo mkdir -p /etc/credentials
sudo cp oshit-config-key.enc /etc/credentials/oshit-config-key.enc
sudo chmod 644 /etc/credentials/oshit-config-key.enc
```

> 也可以放到项目本地路径（如 `./etc/dev-config-key.enc`），启动时通过环境变量 `CONFIG_KEY_ENC_FILE` 指定。

验证 KMS 解密是否正常：

```bash
AWS_PROFILE=oshit-dev aws kms decrypt \
    --ciphertext-blob fileb://<(base64 -d /etc/credentials/oshit-config-key.enc) \
    --region ap-southeast-1 \
    --query Plaintext --output text | base64 -d
```

如果输出一段明文字符串（AES 密钥），说明配置正确。

---

## 4. 导入数据

### 4.1 创建数据库表（DDL）

```bash
# 安装 ULID 扩展和创建所有表
docker exec -i postgres16.2 psql -U postgres -d oshit_db < repositories/structures.sql
```

验证：

```bash
docker exec -it postgres16.2 psql -U postgres -d oshit_db -c "\dt"
```

应能看到 `t_system_config`、`t_chain_config`、`t_token_config`、`t_service_tx` 等表。

### 4.2 导入初始配置数据

```bash
docker exec -i postgres16.2 psql -U postgres -d oshit_db < migrate/testnet/config.sql
```

该脚本会初始化：
- `t_system_config` — 系统环境（env=1 表示测试网）
- `t_aws_config` — AWS 凭证（加密）
- `t_rpc_endpoint` — Solana RPC 节点地址（devnet + mainnet，加密）
- `t_chain_config` — 区块链配置（Solana）
- `t_token_config` — Token 配置（OShit token mint 地址）
- 以及其他业务初始配置表

### 4.3 导入 Nacos 配置

打开 Nacos 控制台 http://127.0.0.1:8080/index.html，登录后执行以下操作：

1. 创建命名空间（如果使用默认 `public` 可跳过）

2. 在 **配置管理 → 配置列表** 中，逐个创建以下配置，Group 统一填 `oshit-go`，格式选 `YAML`：

| Data ID | 对应本地文件 | 说明 |
|---|---|---|
| `base-runtime.yaml` | `migrate/testnet/base-runtime.yaml` | 全局链/Token/系统配置，三个服务共用 |
| `base-service-registry.yaml` | `migrate/testnet/base-service-registry.yaml` | base 服务的业务注册配置（含加密私钥和扫描配置） |
| `reward-runtime.yaml` | `migrate/testnet/reward-runtime.yaml` | reward 服务的奖励规则配置 |
| `pos-runtime.yaml` | `migrate/testnet/pos-runtime.yaml` | pos/stake 服务的业务规则配置 |

操作步骤（每个配置重复）：

```
1. 点击右上角 "+" 号
2. Data ID: 填写上表中的 Data ID
3. Group: oshit-go
4. 配置格式: YAML
5. 配置内容: 将对应本地文件的内容完整粘贴进去
6. 点击 "发布"
```

验证：在配置列表中能看到 4 个配置项，Group 均为 `oshit-go`。

---

## 5. 配置环境变量并启动服务

### 5.1 关键环境变量

| 环境变量 | 说明 | 示例值 |
|---|---|---|
| `AWS_PROFILE` | AWS 凭证 profile 名称 | `oshit-dev` |
| `AWS_REGION` | AWS KMS 所在区域 | `ap-southeast-1`（默认值，可不设） |
| `CONFIG_KEY_ENC_FILE` | 加密密钥文件路径 | `/etc/credentials/oshit-config-key.enc`（默认值，可不设） |

### 5.2 命令行启动

**重要**：每个服务的工作目录必须是其自身目录（`app/{service}/api`），否则读不到 `./etc/application.yaml` 和 dubbo 配置。

```bash
# 启动 base 服务（必须先启动，其他服务依赖它的 Dubbo RPC）
cd app/base/api && AWS_PROFILE=oshit-dev go run base.go

# 启动 reward 服务（新终端窗口）
cd app/reward/api && AWS_PROFILE=oshit-dev go run reward.go

# 启动 pos 服务（新终端窗口）
cd app/pos/api && AWS_PROFILE=oshit-dev go run pos.go
```

如果密钥文件不在默认路径：

```bash
cd app/base/api && AWS_PROFILE=oshit-dev CONFIG_KEY_ENC_FILE=./etc/dev-config-key.enc go run base.go
```

### 5.3 IDE 配置（GoLand / VS Code）

#### GoLand

对每个服务创建一个 Run Configuration：

1. **Run → Edit Configurations → + → Go Build**
2. 配置如下：

| 配置项 | base-api | reward-api | pos-api |
|---|---|---|---|
| Run kind | Package | Package | Package |
| Package path | `oshit-go/app/base/api` | `oshit-go/app/reward/api` | `oshit-go/app/pos/api` |
| Working directory | `$PROJECT_DIR$/app/base/api` | `$PROJECT_DIR$/app/reward/api` | `$PROJECT_DIR$/app/pos/api` |
| Environment | `AWS_PROFILE=oshit-dev` | `AWS_PROFILE=oshit-dev` | `AWS_PROFILE=oshit-dev` |

> Working directory 必须设置正确，否则 viper 读不到 `./etc/application.yaml`。

#### VS Code

在 `.vscode/launch.json` 中添加：

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "base-api",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/app/base/api/base.go",
            "cwd": "${workspaceFolder}/app/base/api",
            "env": {
                "AWS_PROFILE": "oshit-dev"
            }
        },
        {
            "name": "reward-api",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/app/reward/api/reward.go",
            "cwd": "${workspaceFolder}/app/reward/api",
            "env": {
                "AWS_PROFILE": "oshit-dev"
            }
        },
        {
            "name": "pos-api",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/app/pos/api/pos.go",
            "cwd": "${workspaceFolder}/app/pos/api",
            "env": {
                "AWS_PROFILE": "oshit-dev"
            }
        }
    ]
}
```

### 5.4 启动顺序与验证

**启动顺序**：base-api → reward-api / pos-api（后两个可并行）

base-api 必须先启动，因为：
- reward-api 和 pos-api 通过 Dubbo Triple 调用 base-api 的 RPC 接口
- base-api 向 Nacos 注册 Dubbo 服务，其他服务通过 Nacos 发现

**验证服务启动成功**：

```bash
# base-api 健康检查
curl http://localhost:1100/base/info/token-info

# reward-api 验证
curl http://localhost:1200/reward/take-token/config

# pos-api 验证
curl http://localhost:1300/pos/pos-reward/config
```

---

## 6. 运行测试

测试为集成测试，需要所有服务和中间件都处于运行状态：

```bash
# reward 服务测试
cd app/reward/api && go test ./test/... -v

# 单个测试用例
cd app/reward/api && go test ./test/... -v -run TestTakeToken

# base 服务测试
cd app/base/api && go test ./test/... -v -run TestLogin
```

---

## 7. 常用运维操作

### 7.1 重新生成 GORM Model

当数据库表结构变更后，需要重新生成 DAL 代码：

```bash
cd cmd && go run gen.go
```

生成文件位于 `common/pkg/dal/`，**不要手动编辑** `*.gen.go` 文件。

### 7.2 加密配置字段

使用内置工具加密 YAML 中的敏感字段：

```bash
cd cmd/encrypt_config && go run main.go \
    --key "<AES密钥明文>" \
    --file ../../migrate/testnet/base-runtime.yaml \
    --dry-run
```

去掉 `--dry-run` 会就地修改文件，将敏感字段替换为 `ENC~...` 格式。

### 7.3 重新加密 Solana 私钥

当需要更换加密密码时：

```bash
cd cmd/reencrypt_keys && go run main.go \
    --old-key "<旧密钥>" \
    --new-key "<新密钥>" \
    --file ../../migrate/testnet/base-service-registry.yaml
```

---

## 8. 常见问题

### Q: 启动时报 `read encrypted key file /etc/credentials/oshit-config-key.enc: no such file`

未放置加密密钥文件。参见第 3.3 节获取并放置文件，或通过 `CONFIG_KEY_ENC_FILE` 环境变量指定自定义路径。

### Q: 启动时报 `KMS Decrypt: operation error KMS: Decrypt`

AWS 凭证配置有误。检查：
1. `aws sts get-caller-identity --profile oshit-dev` 是否正常
2. IAM 用户是否有 `kms:Decrypt` 权限
3. `AWS_REGION` 是否与 KMS Key 所在 region 一致（默认 `ap-southeast-1`）

### Q: reward-api / pos-api 启动时报 Dubbo 连接失败

确保 base-api 已经先启动并在 Nacos 注册成功。在 Nacos 控制台的 **服务管理 → 服务列表** 中应能看到 `base.BaseService`。

### Q: Nacos 控制台登录提示鉴权失败

检查 docker 启动时的 `NACOS_AUTH_TOKEN` / `NACOS_AUTH_IDENTITY_KEY` / `NACOS_AUTH_IDENTITY_VALUE` 是否与团队统一值一致。

### Q: Kafka 消息消费不到

检查 Kafka 容器是否正常运行，以及 topic 是否已自动创建。可在 Kafka UI (http://localhost:9080) 中查看 topic 列表。服务首次启动时会自动创建所需 topic。

### Q: 编译报 `missing go.sum entry`

```bash
go mod tidy
```
