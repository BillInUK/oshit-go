# 安全加固计划

> 本文档描述 oshit-go 后端服务的安全加固方案与实施进度。

---

## 1. 安全原则

- Nacos 禁止公网访问，通过 AWS SSM Session Manager 或 VPC 内网访问
- Nacos 不存放任何明文密码，只存放被加密后的内容
- 禁止用 root 运行应用，只允许 appuser 身份运行
- 应用解密 Nacos 加密配置的密钥必须只有 appuser 可以访问
- 限制 ptrace 和 /proc 访问

---

## 2. 密钥管理方案

### 2.1 密钥文件

应用启动需要一把 AES-256 密钥来解密 Nacos 中的敏感配置。密钥文件存放明文密钥，这是信任链的终点（与 Vault unseal key、KMS bootstrap key 同理）。安全性靠权限隔离保证。

```bash
# 创建密钥文件
mkdir -p /etc/credentials
echo -n "your-256-bit-aes-key" > /etc/credentials/oshit-config-key
chmod 600 /etc/credentials/oshit-config-key
chown root:root /etc/credentials/oshit-config-key
```

### 2.2 systemd LoadCredential

通过 systemd 的 `LoadCredential` 机制将密钥传递给应用进程，不经过环境变量、不出现在 `/proc/<pid>/environ`：

```ini
[Service]
User=appuser
LoadCredential=CONFIG_DECRYPT_KEY:/etc/credentials/oshit-config-key
LimitCORE=0
```

systemd 在 fork 前将文件内容映射到应用专属的内存目录，通过 `$CREDENTIALS_DIRECTORY` 环境变量告知应用读取路径。

### 2.3 应用读取密钥

```go
// Go 代码
key, err := os.ReadFile(os.Getenv("CREDENTIALS_DIRECTORY") + "/CONFIG_DECRYPT_KEY")
```

本地开发时无 `$CREDENTIALS_DIRECTORY` 环境变量，fallback 到旧方式（硬编码或本地配置文件）。

### 2.4 可选增强：AWS KMS 信封加密

密钥文件存 KMS 加密后的密文，应用启动时调 KMS Decrypt API 得到明文 AES 密钥。即使密钥文件泄露，没有 IAM 权限也解不开。增加了 AWS 依赖和启动延迟，按需选用。

---

## 3. 需要加密的敏感配置

| 敏感项 | 当前位置 | 说明 |
|---|---|---|
| jasypt 解密密码 | `svc/context.go` 硬编码 | 用于解密 `t_service_key` 中的 Solana 私钥 |
| 数据库密码 | `application.yaml` 明文 | PostgreSQL 连接密码 |
| Redis 密码 | `application.yaml` 明文 | Redis 连接密码 |
| Nacos 连接凭据 | `application.yaml` 明文 | 应用连接 Nacos 的认证账号密码 |
| Solana 私钥 | `t_service_key` 表（jasypt 加密） | 已加密，但解密密码需保护 |
| RPC API Key | `t_rpc_endpoint` 表 / Nacos | QuickNode/Helius 等 API Key |

---

## 4. Nacos 加固

| 项 | 说明 |
|---|---|
| 开启认证 | `nacos.core.auth.enabled=true`，删除或重命名默认管理员 nacos/nacos |
| 角色分离 | 应用服务账号（仅读配置）、运维只读账号、发布账号（最小写权限） |
| 网络隔离 | 安全组限制 8848 端口仅 VPC 内网访问 |
| TLS（后续） | Nacos 与客户端全部 HTTPS，防止内网嗅探 |
| 审计日志（后续） | 记录谁、何时、拉取/修改了什么配置 |

---

## 5. 应用运行时安全

| 项 | 说明 |
|---|---|
| 禁止 root 运行 | `useradd -r -s /sbin/nologin appuser`，systemd unit 指定 `User=appuser` |
| 限制 ptrace | `kernel.yama.ptrace_scope = 1`（写入 `/etc/sysctl.d/`），防止非 root 调试其他进程 |
| 禁用 core dump | systemd unit 加 `LimitCORE=0`，防止崩溃时明文密码写入磁盘 |
| 明文不留痕 | 解密后只保留在内存变量中，不写入日志、环境变量或临时文件 |
| 配置变更校验 | Nacos 监听到变更时，验证密文格式和解密结果合法性，对关键字段（IP、端口）做合法性检查 |

---

## 6. 密钥文件保护

| 项 | 说明 |
|---|---|
| 防备份泄露 | `/etc/credentials/` 目录从备份策略中显式排除（AWS 快照、CI/CD 脚本等） |
| 文件不可变 | `chattr +i /etc/credentials/oshit-config-key` 防止误删或篡改 |
| AppArmor（后续） | 限制只有标记过的应用进程才能读取密钥文件 |

---

## 7. 密钥轮换

- 周期：每 90 天轮换一次 AES 密钥
- 流程：
  1. 生成新 AES 密钥
  2. 用新密钥重新加密所有敏感配置
  3. 更新 Nacos 中的密文
  4. 更新 `/etc/credentials/oshit-config-key` 为新密钥
  5. 应用通过 Nacos 监听机制热加载新密文，或优雅重启
- 应急：准备一键吊销流程，怀疑密钥泄露时立即更换密钥、重加密、强制重启所有服务实例
- 第一期先手动轮换，后续再自动化

---

## 8. 实施进度

### 第一期：基础加固

| 步骤 | 任务 | 涉及文件 | 说明 |
|---|---|---|---|
| 1 | 创建 appuser 用户 | 部署脚本 | `useradd -r -s /sbin/nologin appuser` |
| 2 | 创建密钥文件 | `/etc/credentials/oshit-config-key` | AES-256 密钥，`root:root 600` |
| 3 | 编写 systemd unit 文件 | `deploy/` | 加入 `LoadCredential` + `LimitCORE=0` + `User=appuser` |
| 4 | 新增 credential 工具 | `common/utils/credential.go` | 从 `$CREDENTIALS_DIRECTORY` 读密钥；无此变量时 fallback 旧方式（兼容本地开发） |
| 5 | jasypt 密码迁移 | 三个服务的 `svc/context.go` | 从硬编码改为 credential 读取 |
| 6 | 编写加密工具 | `cmd/encrypt_config/` | 用 AES 密钥加密 Nacos yaml 中的敏感字段 |
| 7 | Nacos 配置解密 | 三个服务的 `svc/nacos_config.go` | 加载 Nacos 配置时对密文字段解密 |
| 8 | 测试验证 | — | 3 个服务正常启动、配置热更新正常 |

### 第二期：Nacos 加固

| 步骤 | 任务 |
|---|---|
| 1 | Nacos 开启认证，创建 3 个角色账号 |
| 2 | 安全组限制 Nacos 8848 端口仅 VPC 内网 |
| 3 | Go 代码：Nacos 客户端改用认证账号（从 credential 读取） |
| 4 | `ptrace_scope` 设为 1，写入 `/etc/sysctl.d/99-security.conf` |

### 第三期：增强（按需）

| 任务 | 说明 |
|---|---|
| AWS KMS 信封加密 | 密钥文件存 KMS 密文，启动时调 KMS API 解密 |
| Nacos TLS | 配置证书，全链路 HTTPS |
| 密钥自动轮换 | 脚本化 90 天轮换流程 |
| AppArmor profile | 限制应用进程文件访问范围 |
| 审计日志 | Nacos 操作日志对接 CloudWatch |
