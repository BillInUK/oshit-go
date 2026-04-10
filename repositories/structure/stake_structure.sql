-- 质押AMM配置
DROP TABLE IF EXISTS public.t_stake_amm_config;
CREATE TABLE public.t_stake_amm_config
(
    quote_token character varying(64) not null,
    public_key  character varying(64) not null
);

-- 质押池配置
DROP TABLE IF EXISTS public.t_stake_token_pool;
CREATE TABLE public.t_stake_token_pool
(
    source_account     character varying(64) not null,
    from_token_account character varying(64) not null
);

-- 质押每日固定利息
-- 旧工程 t_sol_stake_fix_interest_config
DROP TABLE IF EXISTS public.t_stake_fix_rate_config;
CREATE TABLE public.t_stake_fix_rate_config
(
    record_id       ulid NOT NULL               DEFAULT gen_ulid(),        -- 记录Id
    min_amount      NUMERIC(78, 0),                                        -- 持币地址的金额
    stake_type      INT  NOT NULL,                                         -- 质押类型 0.180天 1.360天
    fix_rate        NUMERIC(5, 2),                                         -- 每日固定利息费率
    individual_rate NUMERIC(5, 2),                                         -- 星级奖励当中个人费率 每日固定利息 * 个人费率=质押激励奖励个人部分
    created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 记录创建时间
    updated_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 记录更新时间
    PRIMARY KEY (record_id)
);

-- 邀请奖励级别
-- 旧工程 t_sol_stake_invite_dist
DROP TABLE IF EXISTS public.t_stake_invite_dist;
CREATE TABLE public.t_stake_invite_dist
(
    dist_level INTEGER NOT NULL -- 发放奖励往上追溯的级别
);

-- 邀请奖励每个级别的奖励费率
-- 旧工程 t_sol_stake_invite_rate
DROP TABLE IF EXISTS public.t_stake_invite_rate;
CREATE TABLE public.t_stake_invite_rate
(
    dist_level INTEGER       NOT NULL, -- 奖励级别
    rate       NUMERIC(5, 2) NOT NULL, -- 奖励费率从被邀请人的 质押每日固定利息 抽取的费率
    PRIMARY KEY (dist_level)
);

-- 质押星级配置
-- 旧工程 t_sol_stake_star_level_rule
DROP TABLE IF EXISTS public.t_stake_star_level_rule;
CREATE TABLE public.t_stake_star_level_rule
(
    record_id    ulid           NOT NULL     DEFAULT gen_ulid(),        -- 记录Id
    amount       NUMERIC(78, 2) NOT NULL,                               -- 个人质押金额
    group_amount NUMERIC(78, 2) NOT NULL,                               -- 团队质押金额
    star_level   INT            NOT NULL,                               -- 星级
    rate         NUMERIC(5, 2)  NOT NULL,                               -- 星级费率
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 记录创建时间
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 记录更新时间
    PRIMARY KEY (record_id)
);

-- 质押白名单地址
-- 旧工程 t_sol_stake_star_level_config
DROP TABLE IF EXISTS public.t_stake_star_whitelist;
CREATE TABLE public.t_stake_star_whitelist
(
    record_id      ulid        NOT NULL        DEFAULT gen_ulid(),        -- 记录Id
    native_account VARCHAR(64) NOT NULL,                                  -- 地址
    star_level     INT         NOT NULL,                                  -- 质押地址的星级
    rate           NUMERIC(5, 2),                                         -- 质押地址的极差费率
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 记录创建时间
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 记录更新时间
    PRIMARY KEY (record_id)
);
CREATE UNIQUE INDEX ON public.t_stake_star_whitelist (native_account);

-- 质押奖励发放配置表
-- 旧工程 t_sol_stake_reward_rule
drop table if exists public.t_stake_reward_config;
create table public.t_stake_reward_config
(
    record_id          ulid           not null     default gen_ulid(),        -- 记录Id
    program_id         varchar(64)    not null,                               -- stake质押池地址
    reward_account     varchar(64)    not null,                               -- 奖励token的native account
    cost_account       varchar(64)    not null,                               -- 收取dex手续费的native account
    quote_token_amount numeric(78, 0) not null,                               -- 折算token的费用
    cost_fee_rate      int            not null,                               -- 领取时支付给dex的手续费费率
    created_at         timestamp without time zone default current_timestamp, -- 记录创建时间
    updated_at         timestamp without time zone default current_timestamp, -- 记录更新时间
    primary key (record_id)
);

-- stake奖励明细表
-- 旧工程 t_sol_stake_reward
DROP TABLE IF EXISTS public.t_stake_reward;
CREATE TABLE public.t_stake_reward
(
    record_id      ulid        NOT NULL        DEFAULT gen_ulid(),        -- 记录Id
    group_id       varchar(64) NOT NULL,                                  -- 根地址
    native_account varchar(64) NOT NULL,                                  -- 质押地址
    star_level     int                         DEFAULT 0,                 -- 质押地址的星级
    base           numeric(78, 0),                                        -- 质押数量的Base
    rate           numeric(5, 2),                                         -- 奖励费率
    reward_amount  numeric(78, 0),                                        -- 奖励金额
    stake_type     int         NOT NULL,                                  -- 质押类型 0.180天质押 1.360天质押
    reward_type    int         NOT NULL,                                  -- 奖励类型 0.质押每日固定利息 1.质押邀请奖励 2.质押激励奖励 -  个人奖励 3.质押激励奖励 - 星级奖励 4.质押激励奖励 - 团队奖励
    reward_state   int                         DEFAULT 0,                 -- 状态,-1.过期 0.初始化 1.已经领取
    starred        bool        NOT NULL,                                  -- 是否是星级用户奖励
    tx_id          varchar(128),                                          -- 交易Id
    pending        bool                        DEFAULT false,             -- 是否正在被处理
    snap_day       date        NOT NULL,                                  -- 快照的日期
    created_at     timestamp without time zone DEFAULT current_timestamp, -- 记录创建时间
    updated_at     timestamp without time zone DEFAULT current_timestamp, -- 记录更新时间
    primary key (record_id)
);
create index on public.t_stake_reward (group_id);
create index on public.t_stake_reward (native_account, reward_state, pending);
create unique index on public.t_stake_reward (native_account, snap_day, reward_type, starred);

-- 质押奖励领取记录表
-- 旧工程 t_sol_stake_reward_claim_record
drop table if exists public.t_stake_reward_claim;
create table public.t_stake_reward_claim
(
    record_id  ulid   not null             default gen_ulid(),        -- 记录Id
    reward_ids ulid[] not null,                                       -- 奖励Id
    tx_id      varchar(128),                                          -- 交易Id
    tx_state   int                         default 0,                 -- 状态,-2.过期 -1.失败 0.初始化 1.成功
    created_at timestamp without time zone default current_timestamp, -- 记录创建时间
    updated_at timestamp without time zone default current_timestamp, -- 记录更新时间
    primary key (record_id)
);
create index on public.t_stake_reward_claim (tx_id);
create index on public.t_stake_reward_claim (tx_state, created_at);

-- stake 每日快照表
-- 旧工程 t_sol_stake_snap_shot
drop table if exists public.t_stake_snap_shot;
create table public.t_stake_snap_shot
(
    record_id      ulid        not null        default gen_ulid(),        -- 记录Id
    native_account varchar(64) not null,                                  -- native account 地址
    amount         numeric(78, 0),                                        -- 质押金额
    stake_type     int         not null,                                  -- 质押类型 0.180天 1.360天
    snap_day       date        not null,                                  -- 快照的日期
    created_at     timestamp without time zone default current_timestamp, -- 记录创建时间
    updated_at     timestamp without time zone default current_timestamp, -- 记录更新时间
    primary key (record_id)
);
create index on public.t_stake_snap_shot (native_account);
create unique index on public.t_stake_snap_shot (native_account, stake_type, snap_day);

-- 质押记录表
drop table if exists public.t_stake_record;
create table public.t_stake_record
(
    record_id     ulid not null               default gen_ulid(),-- 记录Id
    staker        varchar(64),-- 质押用户地址
    stake_amount  numeric(78, 0),-- 质押金额
    status        varchar(16),-- 质押状态
    stake_tx_hash varchar(128),-- 使用交易Id
    used_tx_ids   text,-- 使用交易Id
    locked_tx_ids text,-- 锁定的交易Id
    created_at    timestamp without time zone default current_timestamp,-- 记录创建时间
    updated_at    timestamp without time zone default current_timestamp,-- 记录更新时间
    primary key (record_id)
);

-- 购买token记录表
drop table if exists public.t_stake_buy_token;
create table public.t_stake_buy_token
(
    tx_id            varchar(128) collate "pg_catalog"."default" not null,
    slot             numeric(78, 0)                              not null,
    from_account     varchar(64) collate "pg_catalog"."default"  not null,
    to_account       varchar(64) collate "pg_catalog"."default"  not null,
    amount           numeric(78, 0)                              not null,
    locked           bool                                        not null default false,
    locked_by        varchar(128) collate "pg_catalog"."default",
    locked_at        timestamptz(6),
    staked_amount    numeric(78, 0)                              not null default 0,
    remaining_amount numeric(78, 0)                              not null default 0,
    created_at       timestamptz(6)                                       default now()
);

create index idx_stake_buy_token_locked on public.t_stake_buy_token (locked);
create index idx_stake_buy_token_locked_at on public.t_stake_buy_token (locked_at);
create index idx_stake_buy_token_locked_by on public.t_stake_buy_token (locked_by);
create index idx_stake_buy_token_slot on public.t_stake_buy_token (slot);

-- 区域经理奖励配置
drop table if exists public.t_stake_leader_reward_config;
create table public.t_stake_leader_reward_config
(
    record_id      ulid     not null           default gen_ulid(),-- 记录Id
    reward_account varchar(64),-- 奖励发放地址
    created_at     timestamp without time zone default current_timestamp,-- 记录创建时间
    updated_at     timestamp without time zone default current_timestamp,-- 记录更新时间
    primary key (record_id)
);

-- 总区域经理表
drop table if exists public.t_stake_total_leader;
create table public.t_stake_total_leader
(
    record_id      ulid not null               default gen_ulid(),-- 记录Id
    native_account varchar(64),-- 区域领导地址
    stake_share    numeric(5, 2), -- 用户质押时奖励总区域经理的分成费率
    created_at     timestamp without time zone default current_timestamp,-- 记录创建时间
    updated_at     timestamp without time zone default current_timestamp,-- 记录更新时间
    primary key (record_id)
);

-- 区域经理表
drop table if exists public.t_stake_leader;
create table public.t_stake_leader
(
    record_id      ulid     not null           default gen_ulid(),-- 记录Id
    native_account varchar(64),-- 区域经理地址
    leader_level   smallint not null, -- 区域经理等级
    up_leader      varchar(64),-- 区域经理的上级
    created_at     timestamp without time zone default current_timestamp,-- 记录创建时间
    updated_at     timestamp without time zone default current_timestamp,-- 记录更新时间
    primary key (record_id)
);

-- 区域经理奖励明细表
drop table if exists public.t_stake_leader_reward;
create table public.t_stake_leader_reward
(
    record_id      ulid        not null        default gen_ulid(),
    native_account varchar(64) not null,                      -- 区域经理地址
    staker         varchar(64) not null,                      -- 触发奖励的质押者
    reward_type    int         not null,                      -- 0=直接区域经理10% 1=区域经理7% 2=上级leader3% 3=总区域经理
    base_amount    numeric(78, 0),                            -- 基础金额(min(购买量,质押量))
    stake_share    numeric(5, 2),                             -- 奖励费率
    reward_amount  numeric(78, 0),                            -- 奖励金额
    reward_state   int         not null        default 0,     -- -1=过期 0=初始化 1=已领取
    pending        bool        not null        default false, -- 是否正在处理中
    tx_id          varchar(128),                              -- 领取交易Id
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
create index on public.t_stake_leader_reward (native_account, reward_state, pending);

-- 区域经理奖励领取表
drop table if exists public.t_stake_leader_reward_claim;
create table public.t_stake_leader_reward_claim
(
    record_id      ulid        not null        default gen_ulid(),
    native_account varchar(64) not null,                  -- 领取者（区域经理）地址
    reward_ids     ulid[]      not null,                  -- 本次领取的奖励Id列表
    tx_id          varchar(128),
    tx_state       int                         default 0, -- -2=过期 -1=失败 0=初始化 1=成功
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
create index on public.t_stake_leader_reward_claim (tx_id);
create index on public.t_stake_leader_reward_claim (tx_state, created_at);
create index on public.t_stake_leader_reward_claim (native_account, tx_state);