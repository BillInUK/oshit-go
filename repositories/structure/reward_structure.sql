-- 旧工程 t_sol_transfer_reward_distribution
DROP TABLE IF EXISTS public.t_level_dist;
CREATE TABLE public.t_level_dist
(
    level integer NOT NULL
);

-- 旧工程 t_sol_transfer_reward_claim
DROP TABLE IF EXISTS public.t_level_ratio;
CREATE TABLE public.t_level_ratio
(
    level integer       NOT NULL,
    ratio numeric(5, 2) NOT NULL
);

-- 旧工程 t_reward_discount_rate
DROP TABLE IF EXISTS public.t_discount_rate;
CREATE TABLE public.t_discount_rate
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    rate      numeric(3, 2)                         NOT NULL
);

-- 旧工程 t_sol_official_give_token_reward_rule
DROP TABLE IF EXISTS public.t_take_token_config;
CREATE TABLE public.t_take_token_config
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    invite_code    character varying(16)       DEFAULT NULL::character varying,
    reward_account character varying(64)                                 NOT NULL,
    cost_account   character varying(64)                                 NOT NULL,
    amount         numeric(78, 0)                                        NOT NULL,
    invite_amount  numeric(78, 0)                                        NOT NULL,
    dex_fee_rate   numeric(78, 0)                                        NOT NULL,
    max_dex_fee    numeric(78, 0)                                        NOT NULL,
    interval       integer                                               NOT NULL,
    is_default     boolean                     DEFAULT false,
    reward_inviter boolean                     DEFAULT true,
    invited        boolean                     DEFAULT true,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_sol_official_give_token_record
DROP TABLE IF EXISTS public.t_take_token_record;
CREATE TABLE public.t_take_token_record
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    reward_account  character varying(64)                                 NOT NULL,
    receipt_account character varying(64)                                 NOT NULL,
    cost_account    character varying(64)                                 NOT NULL,
    tx_id           character varying(128)                                NOT NULL,
    amount                 numeric(78, 0)                                        NOT NULL,
    dex_fee                numeric(78, 0)                                        NOT NULL,
    use_invite_code        boolean                                               NOT NULL,
    invite_code            character varying(16),
    state                  integer,
    invited                boolean,
    created_at             timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at             timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_daily_claim_stats
DROP TABLE IF EXISTS public.t_daily_claim_stats;
CREATE TABLE public.t_daily_claim_stats
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL,
    take_date      date                                                  NOT NULL,
    take_count     integer                     DEFAULT 0                 NOT NULL,
    need_lottery    boolean                     DEFAULT false             NOT NULL,
    last_take_time  timestamp without time zone,
    total_lottery   numeric(78, 0)              DEFAULT 0                 NOT NULL,
    total_take      numeric(78, 0)              DEFAULT 0                 NOT NULL,
    lottery_count   integer                     DEFAULT 0                 NOT NULL,
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE t_daily_claim_stats
    ADD CONSTRAINT uq_daily_claim_stats_account_date UNIQUE (native_account, take_date);

-- 旧工程 t_reward_lottery
DROP TABLE IF EXISTS public.t_lottery_reward;
CREATE TABLE public.t_lottery_reward
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL,
    reward_amount  numeric(78, 0),
    reward_type    integer,
    state          integer                     DEFAULT 0,
    pending        boolean                     DEFAULT false,
    reward_day     date                                                  NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_reward_lottery_claim_record
DROP TABLE IF EXISTS public.t_lottery_claim_record;
CREATE TABLE public.t_lottery_claim_record
(
    record_id  public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    reward_ids public.ulid[]                                         NOT NULL,
    tx_id      character varying(128),
    state      integer                     DEFAULT 0,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


-- 旧工程 t_sol_transfer_token_reward_rule
DROP TABLE IF EXISTS public.t_give_token_config;
CREATE TABLE public.t_give_token_config
(
    reward_account   character varying(64) NOT NULL,
    cost_account     character varying(64) NOT NULL,
    reward_rate      numeric(78, 0)        NOT NULL,
    max_valid_reward numeric(78, 0)        NOT NULL,
    valid_rate       NUMERIC(10, 6)        NOT NULL,
    created_at       timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_sol_transfer_checked_record
DROP TABLE IF EXISTS public.t_give_token_record;
CREATE TABLE public.t_give_token_record
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    from_account    character varying(64)                                 NOT NULL,
    receipt_account character varying(64)                                 NOT NULL,
    tx_id           character varying(128)                                NOT NULL,
    amount          numeric(78, 0)                                        NOT NULL,
    state           integer,
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 奖励码发放奖励配置
DROP TABLE IF EXISTS public.t_reward_code_config;
CREATE TABLE public.t_reward_code_config
(
    record_id      ulid                        DEFAULT gen_ulid() NOT NULL,
    reward_account character varying(64)                          NOT NULL,
    cost_account   character varying(64)                          NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- 兑换奖励码兑换费率表
DROP TABLE IF EXISTS public.t_reward_code_fee;
CREATE TABLE public.t_reward_code_fee
(
    amount     numeric(78, 0) NOT NULL,
    cost_rate  numeric(78, 0) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (amount)
);

-- 奖励码
DROP TABLE IF EXISTS public.t_reward_code;
CREATE TABLE public.t_reward_code
(
    record_id     ulid                        DEFAULT gen_ulid() NOT NULL,
    reward_code   character varying(6)                           NOT NULL,
    reward_amount numeric(78, 0)                                 NOT NULL,
    tx_id         character varying(128)      DEFAULT NULL:: character varying,
    tx_state      integer                     DEFAULT 0          NOT NULL,
    expire_time   timestamp without time zone DEFAULT (CURRENT_TIMESTAMP + '24:00:00':: interval),
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id),
    CONSTRAINT unique_reward_code UNIQUE (reward_code)
);


-- campaign积分兑换
DROP TABLE IF EXISTS public.t_campaign_exchange_config;
CREATE TABLE public.t_campaign_exchange_config
(
    record_id      ulid                        DEFAULT gen_ulid() NOT NULL,
    reward_account character varying(64)                          NOT NULL,
    cost_account   character varying(64)                          NOT NULL,
    rate           numeric(5, 2)                                  NOT NULL,
    cost_rate      numeric(5, 2)                                  NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


-- 全局每日兑换限额表
DROP TABLE IF EXISTS t_global_daily_exchange_limit;
CREATE TABLE t_global_daily_exchange_limit
(
    id          BIGSERIAL PRIMARY KEY,
    daily_limit NUMERIC(78, 0) NOT NULL  DEFAULT 0,            -- 每日全局最大兑换量（单位：1/1000 token）
    quota_date  DATE           NOT NULL  DEFAULT CURRENT_DATE, -- 日期
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (quota_date)
);

-- 用户每日兑换额度表
DROP TABLE IF EXISTS t_user_daily_exchange_quota;
CREATE TABLE t_user_daily_exchange_quota
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         VARCHAR(64)    NOT NULL,                   -- 用户ID（与t_sol_exchange_campaign_score_to_token_record中的UserId对应）
    quota_date      DATE           NOT NULL,                   -- 日期（天）
    max_quota       NUMERIC(78, 0) NOT NULL  DEFAULT 50000000, -- 最大兑换额度（默认500000，token decimals=3，数据库存50000000）
    frozen_quota    NUMERIC(78, 0) NOT NULL  DEFAULT 0,        -- 冻结兑换额度
    available_quota NUMERIC(78, 0) NOT NULL  DEFAULT 50000000, -- 可用兑换额度（初始等于max_quota）
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, quota_date)
);

-- 用户每日兑换额度表
DROP TABLE IF EXISTS t_user_daily_exchange_quota;
CREATE TABLE t_user_daily_exchange_quota
(
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL, -- 用户ID（与t_sol_exchange_campaign_score_to_token_record中的UserId对应）
    quota_date DATE NOT NULL, -- 日期（天）
    max_quota NUMERIC(78, 0) NOT NULL DEFAULT 50000000, -- 最大兑换额度（默认500000，token decimals=3，数据库存50000000）
    frozen_quota NUMERIC(78, 0) NOT NULL DEFAULT 0, -- 冻结兑换额度
    available_quota NUMERIC(78, 0) NOT NULL DEFAULT 50000000, -- 可用兑换额度（初始等于max_quota）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, quota_date)
);

-- 兑换社交媒体积分为token的记录
DROP TABLE IF EXISTS public.t_campaign_exchange_record;
CREATE TABLE public.t_campaign_exchange_record
(
    record_id       ulid           NOT NULL     DEFAULT gen_ulid(),-- 记录Id
    reward_account  VARCHAR(64)    NOT NULL,-- 发放奖励的native account
    receipt_account VARCHAR(64)    NOT NULL,-- 接收奖励的native account
    provider        VARCHAR(64)    NOT NULL,-- 社交媒体
    user_id         VARCHAR(64)    NOT NULL,-- 用户ID
    tx_id           VARCHAR(128)   NOT NULL,-- 奖励token的txId
    score_flow_id   INT            NOT NULL,-- 积分流水Id
    score_tx_id     VARCHAR(128)   NOT NULL,-- 积分交易Id
    amount          NUMERIC(78, 0) NOT NULL,-- 获取的token额度
    score           NUMERIC(78, 0) NOT NULL,-- 兑换的积分额度
    state           INT,-- 状态,-1.失败 0.初始化 1.成功
    created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_campaign_exchange_record (receipt_account);
CREATE INDEX ON public.t_campaign_exchange_record (provider, user_id);
CREATE INDEX ON public.t_campaign_exchange_record (tx_id);














