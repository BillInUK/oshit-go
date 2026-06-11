# AWS SSM Session Manager 端口转发指南

> 通过 AWS SSM Session Manager 将 VPC 内网服务端口映射到本地，用于本地开发调试。
> 无需开放公网端口，无需 SSH 密钥，所有流量走 AWS SSM 加密通道。

---

## 1. 前置条件

### 1.1 本地环境

- AWS CLI 已安装并配置好 profile（如 `oshit-testnet`）
- Session Manager 插件已安装：
  ```bash
  # macOS
  brew install --cask session-manager-plugin
  ```

### 1.2 IAM 用户权限

本地开发用的 IAM 用户（如 `oshit-local-dev`）需要以下权限：

```json
{
    "Sid": "SSMSession",
    "Effect": "Allow",
    "Action": [
        "ssm:StartSession",
        "ssm:TerminateSession",
        "ssm:ResumeSession"
    ],
    "Resource": [
        "arn:aws:ec2:ap-southeast-1:<account-id>:instance/<ec2-instance-id>",
        "arn:aws:ssm:ap-southeast-1::document/AWS-StartPortForwardingSessionToRemoteHost"
    ]
},
{
    "Sid": "SSMDescribe",
    "Effect": "Allow",
    "Action": "ssm:DescribeInstanceInformation",
    "Resource": "*"
}
```

> `ssm:DescribeInstanceInformation` 不支持资源级限制，必须单独声明 `Resource: "*"`。

### 1.3 EC2 实例

用于中转的 EC2 实例（VPC 内任意一台能访问目标服务的机器）需要：

**安装 SSM Agent（Debian 12）：**
```bash
wget https://s3.ap-southeast-1.amazonaws.com/amazon-ssm-ap-southeast-1/latest/debian_amd64/amazon-ssm-agent.deb
sudo dpkg -i amazon-ssm-agent.deb
sudo systemctl enable --now amazon-ssm-agent
```

> Amazon Linux 2 / Ubuntu 默认已预装，无需手动安装。
> ARM 架构实例将 `debian_amd64` 替换为 `debian_arm64`。

**EC2 IAM Role 附加策略：**

EC2 的 IAM Role 需要附加 AWS 托管策略 `AmazonSSMManagedInstanceCore`，否则 SSM Agent 无法注册。

验证 Agent 已上线：
```bash
aws ssm describe-instance-information \
  --profile oshit-testnet \
  --region ap-southeast-1 \
  --query "InstanceInformationList[?InstanceId=='<ec2-instance-id>'].PingStatus"
```

返回 `["Online"]` 表示就绪。

---

## 2. 端口转发命令

基本格式：
```bash
aws ssm start-session \
    --target <ec2-instance-id> \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters '{"host":["<内网IP>"],"portNumber":["<远程端口>"],"localPortNumber":["<本地端口>"]}' \
    --profile oshit-testnet \
    --region ap-southeast-1
```

---

## 3. 常用端口映射

### 3.1 Nacos（172.31.50.82）

Nacos 3.x 将 Console 和 API 端口分离：

| 用途 | 远程端口 | 本地端口 | 何时需要 |
|---|---|---|---|
| Console（浏览器） | 8080 | 18080 | 查看/编辑配置 |
| API（Go SDK） | 8848 | 18848 | 本地运行服务 |
| gRPC（Go SDK） | 9848 | 19848 | 本地运行服务 |

```bash
# Console（浏览器访问 http://localhost:18080/）
aws ssm start-session \
    --target <ec2-instance-id> \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters '{"host":["172.31.50.82"],"portNumber":["8080"],"localPortNumber":["18080"]}' \
    --profile oshit-testnet \
    --region ap-southeast-1

# API
aws ssm start-session \
    --target <ec2-instance-id> \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters '{"host":["172.31.50.82"],"portNumber":["8848"],"localPortNumber":["18848"]}' \
    --profile oshit-testnet \
    --region ap-southeast-1

# gRPC
aws ssm start-session \
    --target <ec2-instance-id> \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters '{"host":["172.31.50.82"],"portNumber":["9848"],"localPortNumber":["19848"]}' \
    --profile oshit-testnet \
    --region ap-southeast-1
```

### 3.2 PostgreSQL（172.31.3.251）

| 用途 | 远程端口 | 本地端口 |
|---|---|---|
| PostgreSQL | 5432 | 15432 |

```bash
aws ssm start-session \
    --target <ec2-instance-id> \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters '{"host":["172.31.3.251"],"portNumber":["5432"],"localPortNumber":["15432"]}' \
    --profile oshit-testnet \
    --region ap-southeast-1
```

### 3.3 Redis（172.31.3.251）

| 用途 | 远程端口 | 本地端口 |
|---|---|---|
| Redis | 6379 | 16379 |

```bash
aws ssm start-session \
    --target <ec2-instance-id> \
    --document-name AWS-StartPortForwardingSessionToRemoteHost \
    --parameters '{"host":["172.31.3.251"],"portNumber":["6379"],"localPortNumber":["16379"]}' \
    --profile oshit-testnet \
    --region ap-southeast-1
```

---

## 4. 本地开发 application.yaml 配置

通过 SSM 端口转发后，本地 `application.yaml` 中的地址需要改为 `localhost` + 映射的本地端口：

```yaml
database:
    host: localhost
    port: 15432
    # ...

redis:
    hosts:
        - localhost:16379
    # ...

nacos:
    host: localhost
    port: 18848
    # ...
```

---

## 5. 故障排查

| 问题 | 原因 | 解决 |
|---|---|---|
| `TargetNotConnected` | EC2 上 SSM Agent 未注册 | 检查 EC2 IAM Role 是否有 `AmazonSSMManagedInstanceCore`，重启 Agent |
| `AccessDeniedException` | IAM 用户缺少 SSM 权限 | 检查用户策略，注意 `DescribeInstanceInformation` 需要 `Resource: "*"` |
| `describe-instance-information` 返回空 | Agent 未上线 | 查看 Agent 日志：`sudo tail -30 /var/log/amazon/ssm/amazon-ssm-agent.log` |
| 策略修改后仍报权限错误 | IAM 策略传播延迟 | 等 1-2 分钟后重试 |
| Nacos Console 404 | 访问了错误端口 | Nacos 3.x Console 在 8080 端口，不是 8848 |
