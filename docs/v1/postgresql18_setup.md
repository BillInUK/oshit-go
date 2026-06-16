# PostgreSQL 18 测试环境安装与配置

> 记录测试环境从 PostgreSQL 16 迁移到 PostgreSQL 18 的完整过程。

---

## 1. 安装 PostgreSQL 18

Debian 12 环境，通过 apt 安装（需先添加 PostgreSQL 官方源）：

```bash
sudo apt install postgresql-18
```

安装完成后系统会自动创建 `postgresql@18-main` systemd 服务和默认数据目录。

---

## 2. 自定义数据目录和日志目录

将数据和日志分别放到 `/data/postgresql/data` 和 `/data/postgresql/log`。

### 2.1 创建目录

```bash
sudo mkdir -p /data/postgresql/data /data/postgresql/log
sudo chown -R postgres:postgres /data/postgresql
```

### 2.2 初始化数据目录

```bash
sudo -u postgres /usr/lib/postgresql/18/bin/initdb -D /data/postgresql/data
```

### 2.3 修改 systemd 管理的配置文件

编辑 `/etc/postgresql/18/main/postgresql.conf`，修改以下配置项：

```ini
# 数据目录
data_directory = '/data/postgresql/data'

# 监听地址（允许远程连接）
listen_addresses = '*'
port = 5432

# 日志配置
logging_collector = on
log_directory = '/data/postgresql/log'
log_filename = 'postgresql-%Y-%m-%d.log'
log_rotation_age = 1d
log_rotation_size = 100MB
```

### 2.4 启动服务

```bash
sudo systemctl start postgresql@18-main
sudo systemctl enable postgresql@18-main
```

### 2.5 验证

```bash
sudo -u postgres psql -c "SHOW data_directory;"
#  /data/postgresql/data

sudo -u postgres psql -c "SHOW log_directory;"
#  /data/postgresql/log

sudo ss -tlnp | grep 5432
# 应看到 0.0.0.0:5432
```

---

## 3. 安装 ULID 扩展

oshit-go 项目依赖 `pgx_ulid` 扩展提供 `ulid` 类型和 `gen_ulid()` 函数。

```bash
wget https://github.com/pksunkara/pgx_ulid/releases/download/v0.2.3/pgx_ulid-v0.2.3-pg18-amd64-linux-gnu.deb
sudo dpkg -i pgx_ulid-v0.2.3-pg18-amd64-linux-gnu.deb
```

> 注意：`pgx_ulid` 是普通扩展，**不需要**在 `postgresql.conf` 中配置 `shared_preload_libraries`。添加该配置会导致启动失败（`FATAL: could not access file "ulid": No such file or directory`）。

在需要的数据库中创建扩展：

```bash
sudo -u postgres psql -d oshit_db -c "CREATE EXTENSION IF NOT EXISTS ulid;"
```

---

## 4. 数据迁移

### 4.1 环境信息

| 项目 | 值 |
|---|---|
| 旧实例 (PG16) | 172.31.3.251:5432 |
| 新实例 (PG18) | 172.31.50.118:5432 |

### 4.2 迁移计划

| 旧库 (PG16) | 新库 (PG18) | 说明 |
|---|---|---|
| pay_db_web3_mainnet | pay_db_web3 | 改名迁移 |
| customer_db | customer_db | 同名迁移 |
| card_db | card_db | 同名迁移 |
| peg_db | peg_db | 同名迁移 |
| meme_db_mgr | meme_db_mgr | 同名迁移 |
| pay_db_web3 | - | 不迁移 |
| meme_db | - | 不迁移，后续用 db_migrate.py 迁移到 oshit_db |

角色迁移：payserver、ob_meme、meme_server（保持密码一致）。

### 4.3 导出并导入角色（先于数据迁移）

```bash
# 导出指定角色（带加密密码）
sudo -u postgres /usr/lib/postgresql/18/bin/pg_dumpall \
  -h 172.31.3.251 -p 5432 -U postgres --roles-only | \
  grep -E '(payserver|ob_meme|meme_server)' > /tmp/roles.sql

# 在新实例导入角色
sudo -u postgres psql -f /tmp/roles.sql
```

> 必须先导入角色再迁移数据，否则 pg_dump 中的 GRANT/OWNER 语句会因角色不存在而报错。

### 4.4 逐库迁移

```bash
# pay_db_web3_mainnet -> pay_db_web3（改名）
sudo -u postgres psql -c "CREATE DATABASE pay_db_web3;"
sudo -u postgres /usr/lib/postgresql/18/bin/pg_dump \
  -h 172.31.3.251 -p 5432 -U postgres pay_db_web3_mainnet | \
  sudo -u postgres psql -d pay_db_web3

# customer_db
sudo -u postgres psql -c "CREATE DATABASE customer_db;"
sudo -u postgres /usr/lib/postgresql/18/bin/pg_dump \
  -h 172.31.3.251 -p 5432 -U postgres customer_db | \
  sudo -u postgres psql -d customer_db

# card_db
sudo -u postgres psql -c "CREATE DATABASE card_db;"
sudo -u postgres /usr/lib/postgresql/18/bin/pg_dump \
  -h 172.31.3.251 -p 5432 -U postgres card_db | \
  sudo -u postgres psql -d card_db

# peg_db
sudo -u postgres psql -c "CREATE DATABASE peg_db;"
sudo -u postgres /usr/lib/postgresql/18/bin/pg_dump \
  -h 172.31.3.251 -p 5432 -U postgres peg_db | \
  sudo -u postgres psql -d peg_db

# meme_db_mgr
sudo -u postgres psql -c "CREATE DATABASE meme_db_mgr;"
sudo -u postgres /usr/lib/postgresql/18/bin/pg_dump \
  -h 172.31.3.251 -p 5432 -U postgres meme_db_mgr | \
  sudo -u postgres psql -d meme_db_mgr
```

### 4.5 验证迁移完整性

在新旧实例分别执行以下查询，对比结果是否一致：

```sql
-- 表数量
SELECT count(*) FROM information_schema.tables WHERE table_schema='public';

-- 每张表行数（估算值）
SELECT schemaname, relname, n_live_tup
FROM pg_stat_user_tables
ORDER BY relname;

-- 索引、序列、函数数量
SELECT
  (SELECT count(*) FROM pg_indexes WHERE schemaname='public') AS indexes,
  (SELECT count(*) FROM information_schema.sequences WHERE sequence_schema='public') AS sequences,
  (SELECT count(*) FROM information_schema.routines WHERE routine_schema='public') AS functions;
```

---

## 5. 配置 pg_hba.conf

编辑 `/etc/postgresql/18/main/pg_hba.conf`，添加业务访问规则：

```ini
# 业务数据库
host    pay_db_web3         payserver       0.0.0.0/0               md5
host    customer_db         all             0.0.0.0/0               md5
host    card_db             all             0.0.0.0/0               md5
host    peg_db              all             0.0.0.0/0               md5
host    meme_db_mgr         all             0.0.0.0/0               md5
host    oshit_db            all             0.0.0.0/0               md5

# ob_meme 用户
local   meme_db_mgr         ob_meme                                 md5
host    meme_db_mgr         ob_meme         127.0.0.1/32            md5
host    meme_db_mgr         ob_meme         ::1/128                 md5
```

生效：

```bash
sudo systemctl reload postgresql@18-main
```

---

## 6. db_migrate.py 运行环境

Debian 12 自带 Python 3.11.2，使用 venv 运行 db_migrate.py：

```bash
# 安装 venv 模块
sudo apt install -y python3-venv

# 创建虚拟环境
python3 -m venv /data/migrate-env

# 激活并安装依赖
source /data/migrate-env/bin/activate
pip install psycopg2-binary

# 执行迁移脚本
python db_migrate.py <参数>
```

---

## 7. 踩坑记录

| 问题 | 原因 | 解决方案 |
|---|---|---|
| 启动报 `could not bind IPv4 address: Address already in use` | 旧 PG 实例或残留进程占用 5432 端口 | `sudo systemctl stop postgresql@16-main` 或 `pg_ctl stop` 杀掉残留进程 |
| 启动报 `FATAL: could not access file "ulid": No such file or directory` | 将 `ulid` 加入了 `shared_preload_libraries` | 移除该配置，`pgx_ulid` 不需要预加载；同时检查 `/etc/postgresql/18/main/postgresql.conf` 和 `/data/postgresql/data/postgresql.conf` 两份配置 |
| systemd 报 `Cluster is already running` | 手动 `pg_ctl start` 启动的进程未停止，与 systemd 冲突 | 先 `pg_ctl -D /data/postgresql/data stop`，再用 systemd 管理 |
| pg_dump 导入报 `role "payserver" does not exist` | 先迁移了数据，未提前导入角色 | 必须先 `pg_dumpall --roles-only` 导入角色，再迁移数据 |
| 远程 telnet 5432 Connection refused | `listen_addresses` 默认为 `localhost` | 改为 `listen_addresses = '*'` 并 reload |
| oshit-go 服务启动卡在 `AES key loaded via KMS Decrypt` | 配置文件中数据库地址填写错误 | 检查 `etc/reward.yaml` 中的数据库连接地址 |
