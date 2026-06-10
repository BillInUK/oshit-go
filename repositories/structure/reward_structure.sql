-- 奖励层级配置表
-- 旧工程 t_sol_transfer_reward_distribution
DROP TABLE IF EXISTS public.t_level_dist;
CREATE TABLE public.t_level_dist
(
    dist_level integer NOT NULL -- 对应 Level
);

-- 奖励层级费率配置表
-- 旧工程 t_sol_transfer_reward_claim
DROP TABLE IF EXISTS public.t_level_ratio;
CREATE TABLE public.t_level_ratio
(
    dist_level integer       NOT NULL, -- 对应 Level
    ratio      numeric(5, 2) NOT NULL  -- 对应 ClaimRatio
);

-- 奖励成本费扣减表
-- 旧工程 t_reward_discount_rate
DROP TABLE IF EXISTS public.t_discount_rate;
CREATE TABLE public.t_discount_rate
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    rate      numeric(3, 2)                         NOT NULL -- 对应 Rate
);

-- take token 奖励配置表
-- 旧工程 t_sol_official_give_token_reward_rule
DROP TABLE IF EXISTS public.t_take_token_config;
CREATE TABLE public.t_take_token_config
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    invite_code    character varying(16)       DEFAULT NULL::character varying,    -- 邀请码，默认规则为空，对应 InviteCode
    reward_account character varying(64)                                 NOT NULL, -- 下发奖励的solana地址，对应 RewardNativeAccount
    cost_account   character varying(64)                                 NOT NULL, -- 接收成本费的solana地址，对应 DexNativeAccount
    amount         numeric(78, 0)                                        NOT NULL, -- 奖励金额，对应 Amount
    invite_amount  numeric(78, 0)                                        NOT NULL, -- 确定邀请关系奖励金额，对应 InviteAmount
    cost_fee_rate  numeric(78, 0)                                        NOT NULL, -- 成本费费率，对应 DexFeeRate
    invited_rate   numeric(78, 0)                                        NOT NULL, -- 新工程新增字段，确定邀请关系时的成本费率
    max_cost_fee   numeric(78, 0)                                        NOT NULL, -- 最大成本费，对应 MaxDexFee
    is_default     boolean                     DEFAULT false,                      -- 是否是默认规则，对应 Default
    reward_inviter boolean                     DEFAULT true,                       -- 是否奖励邀请人，对应 RewardInviter
    invited        boolean                     DEFAULT true,                       -- 是否确定邀请关系，对应 DetermineInvite
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- take token 记录表 (TTL: 1月, DELETE+VACUUM清理)
-- 旧工程 t_sol_official_give_token_record
DROP TABLE IF EXISTS public.t_take_token_record;
CREATE TABLE public.t_take_token_record
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    reward_account  character varying(64)                                 NOT NULL, -- 下发奖励的solana地址，对应 RewardNativeAccount
    receipt_account character varying(64)                                 NOT NULL, -- 接收奖励的地址，对应 ReceiptNativeAccount
    cost_account    character varying(64)                                 NOT NULL, -- 接收成本费的地址，对应 DexNativeAccount
    tx_id           character varying(128)                                NOT NULL, -- 获取奖励交易id，对应 RewardTxId
    amount          numeric(78, 0)                                        NOT NULL, -- 奖励金额，对应 Amount
    cost_fee         numeric(78, 0)                                        NOT NULL, -- 成本费，对应 DexFee
    use_invite_code boolean                                               NOT NULL, -- 是否使用邀请码，UseInviteCode
    invite_code     character varying(16),                                          -- 邀请码，InviteCode
    tx_state        integer,                                                        -- 交易状态，State
    invited         boolean,                                                        -- 是否确定邀请关系，DetermineInvite
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_take_token_record_ttl ON public.t_take_token_record (created_at);

-- 当日take token和lottery统计表 (TTL: 1月, DELETE+VACUUM清理)
-- 旧工程 t_daily_claim_stats
DROP TABLE IF EXISTS public.t_daily_claim_stats;
CREATE TABLE public.t_daily_claim_stats
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL, -- take token和lottery的地址，对应 NativeAccount
    take_date      date                                                  NOT NULL, -- take token的日期，对应 TakeShitDate
    take_count     integer                     DEFAULT 0                 NOT NULL, -- take token的次数，对应 TakeShitCount
    need_lottery   boolean                     DEFAULT false             NOT NULL, -- 是否需要抽奖，对应 NeedLottery
    last_take_time timestamp without time zone,                                    -- 上次take token的时间，对应 LastTakeTime
    total_lottery  numeric(78, 0)              DEFAULT 0                 NOT NULL, -- 总计 take token 的金额，对应 TotalLottery
    total_take     numeric(78, 0)              DEFAULT 0                 NOT NULL, -- 总计 lottery 的金额，对应 TotalTake
    lottery_count  integer                     DEFAULT 0                 NOT NULL, -- 抽奖次数，对应 LotteryCount
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE t_daily_claim_stats
    ADD CONSTRAINT uq_daily_claim_stats_account_date UNIQUE (native_account, take_date);
CREATE INDEX idx_daily_claim_stats_ttl ON public.t_daily_claim_stats (created_at);

-- 抽奖奖励配置表
-- 新工程新增表，用于控制抽奖时的成本费
DROP TABLE IF EXISTS public.t_lottery_config;
CREATE TABLE public.t_lottery_config
(
    reward_account character varying(64)    NOT NULL, -- 发放奖励的 solana 地址
    cost_account   character varying(64)    NOT NULL, -- 接收成本费的 solana 地址
    cost_amount    numeric(78, 0) DEFAULT 0 NOT NULL, -- 抽奖时的成本token额度，用户抽奖可能抽了1000多个token，但是我们可能按照500个来计算成本费
    cost_fee_rate  numeric(5, 2)            NOT NULL, -- 抽奖时奖励金额的配置表
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 抽奖奖励发放表 (TTL: 1月, 按周分区, 分区键: reward_day)
-- 旧工程 t_reward_lottery
DROP TABLE IF EXISTS public.t_lottery_reward CASCADE;
CREATE TABLE public.t_lottery_reward
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL, -- lottery的地址，对应 NativeAccount
    reward_amount  numeric(78, 0),                                                 -- 奖励金额，对应 RewardAmount
    reward_type    integer,                                                        -- 奖励类型，对应 RewardType
    reward_state   integer                     DEFAULT 0,                          -- 奖励状态，对应 State
    pending        boolean                     DEFAULT false,                      -- 奖励是否被处理中，对应 Pending
    reward_day     date                                                  NOT NULL, -- 奖励当天，对应 Day
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lottery_reward_ttl ON public.t_lottery_reward (created_at);

-- 领取抽奖奖励记录表 (TTL: 1月, DELETE+VACUUM清理)
-- 旧工程 t_reward_lottery_claim_record
DROP TABLE IF EXISTS public.t_lottery_claim;
CREATE TABLE public.t_lottery_claim
(
    record_id    public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    reward_ids   public.ulid[]                                         NOT NULL, -- 领取的抽奖奖励记录id，对应 RewardIds
    tx_id        character varying(128),                                         -- 抽奖的交易id，对应 TxId
    tx_state integer                     DEFAULT 0,                          -- 奖励状态,对应 State
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lottery_claim_ttl ON public.t_lottery_claim (created_at);

-- give token 奖励配置
-- 旧工程 t_sol_transfer_token_reward_rule
DROP TABLE IF EXISTS public.t_give_token_config;
CREATE TABLE public.t_give_token_config
(
    reward_account   character varying(64) NOT NULL, -- 发放奖励的 solana 地址
    cost_account     character varying(64) NOT NULL, -- 接收成本费的 solana 地址
    max_reward       numeric(78, 0)        NOT NULL, -- 发送到有token account的地址的奖励金额上限
    max_valid_reward numeric(78, 0)        NOT NULL, -- 发送到没有token account的地址的奖励金额上限
    reward_rate      numeric(5, 2)         NOT NULL, -- 普通地址奖励费率（to 有 token account 时）
    valid_rate       numeric(10, 6)        NOT NULL, -- 有效地址奖励费率（to 无 token account 时）
    created_at       timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- give token 记录表 (TTL: 1月, DELETE+VACUUM清理)
-- 旧工程 t_sol_transfer_checked_record
DROP TABLE IF EXISTS public.t_give_token_record;
CREATE TABLE public.t_give_token_record
(
    record_id       public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    from_account    character varying(64)                                 NOT NULL, -- 发送 token 的 solana 地址， 对应 FromNativeAccount
    receipt_account character varying(64)                                 NOT NULL, -- 接收 token 的 solana 地址，对应 ReceiptNativeAccount
    tx_id           character varying(128)                                NOT NULL, -- give token 活动的solana交易id，对应 TransferTxId
    amount          numeric(78, 0)                                        NOT NULL, -- give token 的金额，对应 TransferAmount
    tx_state        integer,                                                        -- 交易状态，对应 State
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_give_token_record_ttl ON public.t_give_token_record (created_at);

-- 奖励码奖励规则
-- 旧工程 t_reward_code_rule
DROP TABLE IF EXISTS public.t_reward_code_config;
CREATE TABLE public.t_reward_code_config
(
    record_id      ulid                        DEFAULT gen_ulid() NOT NULL,
    reward_account character varying(64)                          NOT NULL, -- 下发奖励的地址，对应 RewardNativeAccount
    cost_account   character varying(64)                          NOT NULL, -- 接收成本费的地址，对应 DexNativeAccount
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- 奖励码 (TTL: 1月, DELETE+VACUUM清理)
-- 旧工程 t_reward_code
DROP TABLE IF EXISTS public.t_reward_code;
CREATE TABLE public.t_reward_code
(
    record_id      ulid                        DEFAULT gen_ulid() NOT NULL,
    reward_code    character varying(6)                           NOT NULL,      -- 奖励码，对应 RewardCode
    reward_amount  numeric(78, 0)                                 NOT NULL,      -- 奖励金额，对应 RewardAmount
    reward_state   integer                     DEFAULT 0          NOT NULL,      -- 奖励状态，对应 State
    native_account character varying(64)       DEFAULT NULL,                     -- 领取奖励码的solana地址i，对应 NativeAccount
    tx_id          character varying(128)      DEFAULT NULL:: character varying, -- 交易id，对应 TxId
    tx_state       integer                     DEFAULT 0          NOT NULL,
    expired_at     timestamp without time zone DEFAULT (CURRENT_TIMESTAMP + '24:00:00':: interval),
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_reward_code_ttl ON public.t_reward_code (created_at);

-- 奖励码兑换费率表
-- 旧工程 t_reward_code_fee
DROP TABLE IF EXISTS public.t_reward_code_fee;
CREATE TABLE public.t_reward_code_fee
(
    amount     numeric(78, 0) NOT NULL, -- 奖励码金额，对应 Amount
    fee_rate   numeric(78, 0) NOT NULL, -- 奖励码费率，对应 CostFeeRate
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- campaign积分兑换
-- 对应旧工程 t_sol_campaign_exchange_score_rule
DROP TABLE IF EXISTS public.t_campaign_quote_config;
CREATE TABLE public.t_campaign_quote_config
(
    record_id      ulid                        DEFAULT gen_ulid() NOT NULL,
    reward_account character varying(64)                          NOT NULL, -- 发放奖励的地址，对应 RewardNativeAccount
    cost_account   character varying(64)                          NOT NULL, -- 接收成本费的地址，对应 DexNativeAccount
    quote_rate        numeric(5, 2)                                  NOT NULL, -- 兑换费率，对应 Rate
    cost_rate         numeric(5, 2)                                  NOT NULL, -- 成本费率，对应 CostRate
    global_daily_limit numeric(78, 0) NOT NULL DEFAULT 500000000,              -- 全局每session默认兑换限额（原始值）
    user_daily_limit   numeric(78, 0) NOT NULL DEFAULT 30000000,               -- 用户每日默认兑换限额（原始值）
    created_at        timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at        timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 全局每日兑换限额表（每天分2个session，SGT 08:00/20:00 各重置一次）
-- 对应旧工程 t_global_daily_exchange_limit
DROP TABLE IF EXISTS t_campaign_quote_limit;
CREATE TABLE t_campaign_quote_limit
(
    id          BIGSERIAL PRIMARY KEY,
    daily_limit NUMERIC(78, 0) NOT NULL  DEFAULT 500000000,   -- 每个session全局可兑换量（原始值，500000 token * 1000 = 500000000）
    quota_date  DATE           NOT NULL  DEFAULT CURRENT_DATE, -- UTC自然日（与session共同定位唯一记录）
    session     SMALLINT       NOT NULL  DEFAULT 0,            -- 0=早上场次(UTC 00:00-11:59 / SGT 08:00-19:59), 1=晚上场次(UTC 12:00-23:59 / SGT 20:00-07:59)
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (quota_date, session)
);

-- 用户每日兑换额度表（按SGT自然日计算，不随session重置）
-- 对应旧工程 t_user_daily_exchange_quota
DROP TABLE IF EXISTS t_user_daily_quota;
CREATE TABLE t_user_daily_quota
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         VARCHAR(64)    NOT NULL,                   -- 用户ID
    quota_date      DATE           NOT NULL,                   -- SGT自然日（UTC+8），用于跨场次累计限额
    max_quota       NUMERIC(78, 0) NOT NULL  DEFAULT 30000000, -- 每日最大兑换额度（30000 token * 1000 = 30000000）
    frozen_quota    NUMERIC(78, 0) NOT NULL  DEFAULT 0,        -- 冻结兑换额度（交易待确认中）
    available_quota NUMERIC(78, 0) NOT NULL  DEFAULT 30000000, -- 可用兑换额度
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (user_id, quota_date)
);

-- 兑换社交媒体积分为token的记录 (TTL: 1月, DELETE+VACUUM清理)
-- 旧工程表 t_sol_exchange_campaign_score_to_token_record
DROP TABLE IF EXISTS public.t_campaign_quote_record;
CREATE TABLE public.t_campaign_quote_record
(
    record_id       ulid           NOT NULL     DEFAULT gen_ulid(), -- 记录Id
    reward_account  VARCHAR(64)    NOT NULL,                        -- 发放奖励的native account，对应 RewardNativeAccount
    receipt_account VARCHAR(64)    NOT NULL,                        -- 接收奖励的native account，对应 ReceiptNativeAccount
    provider        VARCHAR(64)    NOT NULL,                        -- 社交媒体，对应 Provider
    user_id         VARCHAR(64)    NOT NULL,                        -- 用户ID，对应 UserId
    tx_id           VARCHAR(128)   NOT NULL,                        -- 奖励token的txId，对应 ExchangeTxId
    score_flow_id   INT            NOT NULL,                        -- 积分流水Id，对应 ScoreFlowId
    score_tx_id     VARCHAR(128)   NOT NULL,                        -- 积分交易Id，对应 ScoreTxId
    amount          NUMERIC(78, 0) NOT NULL,                        -- 获取的token额度，对应 Amount
    score           NUMERIC(78, 0) NOT NULL,                        -- 兑换的积分额度，对应 Score
    quote_state     INT,                                            -- 状态,-1.失败 0.初始化 1.成功，对应 State
    session         SMALLINT       NOT NULL     DEFAULT 0,          -- 下单时的全局场次（0=早/1=晚），用于Kafka回调时精确恢复对应session额度
    user_quota_date DATE           NOT NULL     DEFAULT CURRENT_DATE, -- 下单时用户所在的SGT自然日，用于Kafka回调时精确恢复用户额度
    created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_campaign_quote_record (receipt_account);
CREATE INDEX ON public.t_campaign_quote_record (provider, user_id);
CREATE INDEX ON public.t_campaign_quote_record (tx_id);
CREATE INDEX idx_campaign_quote_record_ttl ON public.t_campaign_quote_record (created_at);