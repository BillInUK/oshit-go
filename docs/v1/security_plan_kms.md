# 安全加固计划（KMS 版）

> 基于 `security_plan.md` 的 KMS 信封加密版本。
> 用 AWS KMS 替代 appuser + 文件权限方案，简化部署流程，增强密钥安全性。

---

## 1. 安全原则

- Nacos 禁止公网访问，通过 AWS SSM Session Manager 或 VPC 内网访问
- Nacos 不存放任何明文密码，只存放被加密后的内容（`ENC~` 前缀）
- 应用解密敏感配置的 AES 密钥由 KMS 保护，非授权 IAM 角色无法解密
- 解密后的明文只保留在内存中，不写入日志、环境变量或临时文件
- 禁用 core dump，防止崩溃时明文密码写入磁盘

---

## 2. KMS 信封加密原理

### 2.1 什么是信封加密

信封加密（Envelope Encryption）是一种两层加密方案：

- **第一层**：用本地生成的 AES-256 密钥加密业务数据（DB 密码、API Key 等）
- **第二层**：用 AWS KMS 托管的对称密钥加密 AES-256 密钥本身

好处：KMS 不直接接触业务数据，只负责保护 AES 密钥。AES 密钥被 KMS 加密后变成密文，即使泄露也无法使用。

### 2.2 KMS 对称密钥

AWS KMS 创建的是**对称密钥**（不是公钥/私钥对），加密和解密用同一把密钥。这把密钥始终存储在 KMS 服务内部的硬件安全模块（HSM）中，不会被导出。

### 2.3 KMS 密文自带 Key ID

KMS 加密时会把 Key ID **嵌入到密文元数据**中。因此解密时：

- 应用**不需要**知道是哪把 KMS 密钥
- 应用只需要把密文原样发给 KMS Decrypt API
- KMS 服务端自动解析密文头部，识别出对应的密钥并解密
- 应用只需要有该密钥的 `kms:Decrypt` IAM 权限即可

### 2.4 加密流程图

```
运维人员（本地）
│
│  ① openssl rand -hex 32 → AES-256 明文密钥
│                              │
│         ┌────────────────────┼─────────────────────┐
│         ▼                    ▼                     │
│  ② encrypt_config 工具    ③ aws kms encrypt        │
│     用 AES 密钥加密          用 KMS 对称密钥         │
│     Nacos YAML 敏感字段      加密 AES 密钥本身       │
│         │                    │                     │
│         ▼                    ▼                     │
│  Nacos 里存的是           encrypted-dek.b64        │
│  ENC~xxx 密文             （KMS 密文，含 Key ID）    │
│                              │                     │
│                    ④ 部署到服务器                    │
│                    /etc/credentials/               │
│                    oshit-config-key.enc             │
│                              │                     │
│  ⑤ AES 明文密钥 → 存密码管理器，本地安全删除          │
└────────────────────────────────────────────────────┘
```

### 2.5 运行时解密流程图

```
EC2 实例（应用启动时）
│
│  ① 读取 /etc/credentials/oshit-config-key.enc
│     （这是一块 KMS 密文 blob，含 Key ID 元数据）
│         │
│         ▼
│  ② 原样发给 KMS Decrypt API
│     （EC2 通过 Instance Role 自动获取 IAM 临时凭据）
│         │
│         ▼
│  ③ KMS 服务端解析密文头部 → 识别 Key ID → 用对应密钥解密
│     → 返回 AES-256 明文密钥（仅存内存，不落盘）
│         │
│         ▼
│  ④ 用 AES 密钥调用 JasyptDecode 解密配置中的 ENC~ 字段
│     → 得到 DB 密码、Redis 密码、API Key 等明文
│         │
│         ▼
│  ⑤ 用解密后的明文连接 DB / Redis / Nacos / Solana RPC
```

### 2.6 安全保证

| 攻击场景 | 防护 |
|---|---|
| `.enc` 文件泄露 | 密文无法解密，需要 IAM 权限 + KMS 密钥 |
| Nacos 配置泄露 | `ENC~` 密文无法解密，需要 AES 密钥 |
| EC2 被入侵但没有 IAM Role | 无法调用 KMS API，无法获取 AES 密钥 |
| IAM 凭据泄露 | KMS 策略限定具体 Key + CloudTrail 审计 |

---

## 3. IAM 配置

### 3.1 创建 IAM Role（给 EC2 用）

1. AWS Console → IAM → Roles → Create role
2. Trusted entity type: **AWS service → EC2**
3. 附加以下内联策略（最小权限）：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "kms:Decrypt",
      "Resource": "arn:aws:kms:ap-southeast-1:<account-id>:key/<key-id>"
    }
  ]
}
```

4. Role 命名如 `oshit-ec2-kms-role`

### 3.2 绑定 Role 到 EC2 实例

1. EC2 Console → 选中实例 → Actions → Security → Modify IAM role
2. 选择 `oshit-ec2-kms-role`
3. 如果 EC2 已有 Instance Role，直接在现有 Role 上添加 KMS 权限策略即可

> EC2 通过 Instance Role 自动获取**临时凭据**（自动轮换），不需要在服务器上存 Access Key。

### 3.3 IAM User（仅本地 CLI 用）

本地执行 `aws kms encrypt` 等操作时使用的 IAM User，**不部署到服务器**：

```bash
# 本地配置（区分测试网/主网两个 AWS 账号）
aws configure --profile oshit-testnet   # 测试网 Access Key
aws configure --profile oshit-mainnet   # 主网 Access Key

# 使用时指定 profile
aws kms encrypt --profile oshit-testnet ...
aws kms encrypt --profile oshit-mainnet ...
```

---

## 4. 需要加密的敏感配置

### 4.1 配置密码类（用 `encrypt_config` 工具加密，`ENC~` 前缀）

| 敏感项 | 当前位置 | 说明 |
|---|---|---|
| 数据库密码 | `application.yaml` 明文 | PostgreSQL 连接密码 |
| Redis 密码 | `application.yaml` 明文 | Redis 连接密码 |
| Nacos 连接凭据 | `application.yaml` 明文 | 应用连接 Nacos 的认证账号密码 |
| RPC API Key | Nacos `base-runtime.yaml` | QuickNode/Helius 等 API Key |
| AWS 凭据 | Nacos `base-runtime.yaml` | AccessKey / SecretKey |

### 4.2 Solana 私钥类（用 `reencrypt_keys` 工具迁移）

这些字段**已经是** jasypt 加密的密文，但加密时用的是旧的硬编码密码 `fktYimwMl3OfUF3m`。
切换到新 AES 密钥后，必须用新密码**重新加密**，否则运行时解密会失败。

| 敏感项 | 当前位置 | 说明 |
|---|---|---|
| 服务私钥 (Nacos) | `base-service-registry.yaml` → `services.*.encrypted_key` | 各业务的 Solana 签名私钥 |
| 服务私钥 (数据库) | `t_service_key` 表 → `encrypted_key` 列 | 同上，数据库中的备份 |

> **注意**：这两类加密方式不同，不能混用工具。
> - 配置密码类：明文 → `encrypt_config` 加密 → `ENC~xxx` 密文
> - 私钥类：旧密文 → `reencrypt_keys` 用旧密码解密 → 用新密码重新加密 → 新密文（无 `ENC~` 前缀）

---

## 5. 与旧方案对比

| 项 | 旧方案（文件密钥） | KMS 方案 |
|---|---|---|
| appuser 用户 | 需要 | 不需要 |
| `/etc/credentials/` 明文密钥 | 需要 | 不需要（存的是 KMS 密文） |
| systemd LoadCredential | 需要 | 不需要 |
| setup-security.sh | 需要 | 不需要 |
| 密钥安全边界 | Linux 文件权限 | IAM 策略（更强、可审计、可跨区域） |
| 密钥轮换 | 手动替换文件 | KMS 原生支持，或手动重新 encrypt |
| 额外依赖 | 无 | AWS SDK v2（`~2MB` 编译体积） |
| 额外费用 | 无 | ~$1/月 |
| 启动延迟 | 无 | +100~200ms（一次 KMS API 调用） |
| 本地开发 | 无影响（fallback） | 无影响（fallback） |

---

## 6. 应用运行时安全

| 项 | 说明 |
|---|---|
| 禁用 core dump | systemd unit 加 `LimitCORE=0` |
| 明文不留痕 | 解密后只保留在内存变量中，不写入日志、环境变量或临时文件 |
| 配置变更校验 | Nacos 监听到变更时，验证密文格式和解密结果合法性 |

> 注：appuser、LoadCredential、ptrace_scope 在 KMS 方案中不再是必需项。
> 如需纵深防御，可在第二期按需添加。

---

## 7. 密钥轮换

- 周期：每 90 天轮换一次 AES 密钥
- 流程：
  1. 生成新明文 AES 密钥
  2. 用 `encrypt_config` 工具重新加密配置密码类字段
  3. 用 `reencrypt_keys` 工具重新加密 Solana 私钥（YAML + 数据库）
  4. 更新 Nacos 中的密文配置
  5. 用 KMS 加密新 AES 密钥 → 新的 `encrypted-dek.b64`
  6. 部署新 `encrypted-dek.b64` 到所有服务器
  7. 优雅重启所有服务（或等 Nacos 热更新 + 下次重启生效）
- KMS Key 本身可开启自动轮换（AWS 托管，对应用透明）
- 应急：怀疑密钥泄露时，执行上述 1~6 步 + 强制重启所有实例

---

## 8. 实施步骤

### 第一期：KMS 信封加密 + Nacos 配置加密

#### A. AWS 准备（一次性）

| 步骤 | 操作 | 说明 |
|---|---|---|
| A1 | AWS Console → KMS → Create key | 类型: Symmetric, 用途: Encrypt/Decrypt, 别名: `oshit-config-key` |
| A2 | 创建 IAM Role (Trusted entity: EC2) | 附加 `kms:Decrypt` 权限，仅限指定 Key ARN |
| A3 | EC2 Console → 绑定 Role 到实例 | 三台服务器都绑定同一个 Role（或在已有 Role 上加权限） |

> 测试网和主网是不同的 AWS 账号，各自执行 A1~A3。

#### B. 加密配置（本地执行）

```bash
# B1. 生成 AES-256 密钥（保存到密码管理器）
openssl rand -hex 32 > /tmp/oshit-key.txt

# B2. 编译工具
cd oshit-go
go build -o encrypt_config ./cmd/encrypt_config/
go build -o reencrypt_keys ./cmd/reencrypt_keys/
```

**B3. 加密配置密码类字段**（DB 密码、Redis 密码、API Key 等）

```bash
# 先 dry-run 预览
./encrypt_config --key "$(cat /tmp/oshit-key.txt)" --file migrate/testnet/base-runtime.yaml --dry-run
./encrypt_config --key "$(cat /tmp/oshit-key.txt)" --file migrate/testnet/reward-runtime.yaml --dry-run
./encrypt_config --key "$(cat /tmp/oshit-key.txt)" --file migrate/testnet/pos-runtime.yaml --dry-run

# 确认无误后正式加密
./encrypt_config --key "$(cat /tmp/oshit-key.txt)" --file migrate/testnet/base-runtime.yaml
./encrypt_config --key "$(cat /tmp/oshit-key.txt)" --file migrate/testnet/reward-runtime.yaml
./encrypt_config --key "$(cat /tmp/oshit-key.txt)" --file migrate/testnet/pos-runtime.yaml
```

> 注：`encrypt_config` 不处理 `base-service-registry.yaml` 中的 `encrypted_key`，
> 那些是 jasypt 加密的 Solana 私钥，由下面的 `reencrypt_keys` 工具单独处理。

**B4. 重新加密 Solana 私钥**（用新密码替换旧硬编码密码）

```bash
# Nacos YAML 中的私钥（先 dry-run 预览）
./reencrypt_keys \
  --old-key "fktYimwMl3OfUF3m" \
  --new-key "$(cat /tmp/oshit-key.txt)" \
  --file migrate/testnet/base-service-registry.yaml \
  --dry-run

# 确认无误后正式执行
./reencrypt_keys \
  --old-key "fktYimwMl3OfUF3m" \
  --new-key "$(cat /tmp/oshit-key.txt)" \
  --file migrate/testnet/base-service-registry.yaml
```

```bash
# 数据库 t_service_key 表中的私钥（逐条处理）
# 对每条记录：
./reencrypt_keys \
  --old-key "fktYimwMl3OfUF3m" \
  --new-key "$(cat /tmp/oshit-key.txt)" \
  --value "<数据库中的 encrypted_key 值>"
# 输出新密文，然后更新数据库：
# UPDATE t_service_key SET encrypted_key = '<新密文>'
#   WHERE service = '...' AND sub_service = '...';
```

> 完成后，旧硬编码密码 `fktYimwMl3OfUF3m` 不再被任何生产环境依赖（仅保留为本地开发 fallback）。

**B5. 更新 Nacos 配置**

把 B3 和 B4 处理后的 YAML 内容更新到 Nacos Console。

**B6. 用 KMS 加密 AES 密钥**（CloudShell 或本地 CLI）

```bash
aws kms encrypt \
  --key-id alias/oshit-config-key \
  --plaintext fileb:///tmp/oshit-key.txt \
  --region ap-southeast-1 \
  --output text --query CiphertextBlob > encrypted-dek.b64
```

**B7. 安全删除本地明文密钥**

```bash
rm -P /tmp/oshit-key.txt   # macOS
# 或 shred -u /tmp/oshit-key.txt  # Linux
```

#### C. 部署到服务器

```bash
# C1. 部署 KMS 密文文件到每台主机（三台用同一个文件）
scp encrypted-dek.b64 <base-host>:/etc/credentials/oshit-config-key.enc
scp encrypted-dek.b64 <reward-host>:/etc/credentials/oshit-config-key.enc
scp encrypted-dek.b64 <snap-host>:/etc/credentials/oshit-config-key.enc

# C2. 在服务器上创建目录（如不存在）
ssh <host> "sudo mkdir -p /etc/credentials && sudo chmod 755 /etc/credentials"

# C3. 部署新版代码（包含 KMS 解密逻辑的新二进制）

# C4. 重启服务
ssh <host> "cd /data/dist/oshit/base/backend/bin && ./server.sh restart"
```

#### D. 验证

```bash
# 查看日志，确认 KMS 解密成功
ssh <host> "tail -50 /data/dist/oshit/base/backend/log/server.log"
# 应该看到: [credential] AES key loaded via KMS Decrypt
# 不应该看到解密错误

# 验证服务正常运行
curl http://<host>:1100/health   # base
curl http://<host>:1200/health   # reward
```

#### E. 本地开发

不需要任何改动。本地没有 `/etc/credentials/oshit-config-key.enc` 文件时，自动 fallback 到代码中的旧密码，开发体验完全不变。

---

### 代码改动总结

| 文件 | 改动 | 状态 |
|---|---|---|
| `common/utils/credential.go` | 新增：KMS 信封解密，本地 fallback | 已完成 |
| `go.mod` | 新增：`aws-sdk-go-v2/service/kms` | 已完成 |
| `app/base/api/internal/svc/context.go` | 硬编码密码 → `LoadConfigDecryptKey()` | 已完成 |
| `app/reward/api/internal/svc/context.go` | 同上 | 已完成 |
| `app/pos/api/internal/svc/context.go` | 同上 | 已完成 |
| `app/base/api/internal/svc/nacos_config.go` | YAML 解析后 `JasyptDecode()` 解密 | 已完成 |
| `app/reward/api/internal/svc/nacos_config.go` | 同上 | 已完成 |
| `app/pos/api/internal/svc/nacos_config.go` | 同上 | 已完成 |
| `cmd/encrypt_config/main.go` | 新增：YAML 敏感字段加密工具（配置密码类） | 已完成 |
| `cmd/reencrypt_keys/main.go` | 新增：Solana 私钥重新加密迁移工具 | 已完成 |
| `deploy/systemd/*.service` | 简化：去掉 appuser/LoadCredential | 已完成 |

---

### 第二期：Nacos 加固

| 步骤 | 任务 |
|---|---|
| 1 | Nacos 开启认证，创建角色账号 |
| 2 | 安全组限制 Nacos 8848 端口仅 VPC 内网 |
| 3 | Go 代码：Nacos 客户端改用认证账号 |

### 第三期：增强（按需）

| 任务 | 说明 |
|---|---|
| KMS Key 自动轮换 | 开启 AWS 托管轮换（每年） |
| Nacos TLS | 配置证书，全链路 HTTPS |
| AES 密钥自动轮换 | 脚本化 90 天重新加密流程 |
| 审计日志 | KMS CloudTrail + Nacos 操作日志对接 CloudWatch |
| 纵深防御 | 按需添加 appuser、ptrace_scope 等 |
