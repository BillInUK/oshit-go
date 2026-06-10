-- CREATE DATABASE "oshit_db" WITH OWNER "meme_server" ENCODING 'UTF8' TEMPLATE template0;
CREATE EXTENSION IF NOT EXISTS ulid WITH SCHEMA public;

-- 创建通用触发器函数：实现updated_at字段自动更新（PostgreSQL替代ON UPDATE CURRENT_TIMESTAMP）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 系统配置表
-- 对应旧工程表 t_system_config
DROP TABLE IF EXISTS public.t_system_config;
CREATE TABLE public.t_system_config
(
    env        integer NOT NULL, -- 0.主网 1.测试网 对应旧工程表字段 Env
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- aws api key配置表
-- 对应旧工程表 t_aws_config
DROP TABLE IF EXISTS public.t_aws_config;
CREATE TABLE public.t_aws_config
(
    access_key_id     character varying(128) NOT NULL, -- aws api key的 access key id, 对应旧工程表字段 AccessKeyId
    secret_access_key character varying(128) NOT NULL, -- aws api key的 secret access key, 对应旧工程表字段 SecretAccessKey
    region            character varying(32)  NOT NULL, -- aws api key的 region 属性, 对应旧工程表字段 Region
    created_at        timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at        timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 新工程表 rpc 端点配置
DROP TABLE IF EXISTS public.t_rpc_endpoint;
CREATE TABLE public.t_rpc_endpoint
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    scope        varchar(32)            NOT NULL, -- 'env' = 跟随环境(devnet/mainnet), 'mainnet' = 固定主网
    provider     varchar(32)            NOT NULL, -- 'quicknode', 'helius', 'alchemy', 'custom'
    endpoint     varchar(512)           NOT NULL, -- 不含 api key 的 base URL
    api_key      varchar(256) DEFAULT '',
    wss_endpoint varchar(512) DEFAULT '',
    wss_api_key  varchar(256) DEFAULT '',
    weight       integer      DEFAULT 1 NOT NULL, -- 轮询权重, 0 表示禁用
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_rpc_endpoint_scope_provider_endpoint ON public.t_rpc_endpoint (scope, provider, endpoint);

-- 链配置表
-- 对应旧工程 t_chain_config
DROP TABLE IF EXISTS public.t_chain_config;
CREATE TABLE public.t_chain_config
(
    chain_name character varying(64) NOT NULL, -- 链名称, 对应旧工程 Chain
    decimals   integer               NOT NULL,
    symbol     character varying(64) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- token配置表
-- 对应旧工程 t_sol_token_config
DROP TABLE IF EXISTS public.t_token_config;
CREATE TABLE public.t_token_config
(
    token_name   character varying(64) NOT NULL, -- token名称，对应 Brand
    token_symbol character varying(64) NOT NULL, -- token符号，对应 TokenSymbol
    decimals     integer               NOT NULL, -- token精度，对应 Decimal
    mint         character varying(64) NOT NULL, -- token地址，对应 TokenMintAccount
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 手续费容错表
-- 对应旧工程 t_fee_tolerance
DROP TABLE IF EXISTS public.t_fee_tolerance;
CREATE TABLE public.t_fee_tolerance
(
    max_less_rate numeric(5, 2) NOT NULL, -- 最大手续费容错，对应 MaxLessRate
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 手续费统计表
-- 对应旧工程 t_sol_fee_statistics
DROP TABLE IF EXISTS public.t_fee_statistics;
CREATE TABLE public.t_fee_statistics
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    slot           bigint,                 -- 交易所在slot, 对应 Slot
    tx_index       integer,                -- 交易所在区块索引, 对应 TransactionIndex
    block_hash     character varying(64),  -- 区块哈希,对应 BlockHash
    tx_id          character varying(128), -- 交易id，对应 TransactionId
    price          numeric(78, 0),         -- 优先费用,对应 ComputeUnitPrice
    unit_limit     numeric(78, 0),         -- 计算单元限制，对应 ComputeUnitLimit
    units_consumed numeric(78, 0),         -- 计算单元消耗, 对应 UnitsConsumed
    fee            numeric(78, 0),         -- 交易消耗手续费，对应 Fee
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE ONLY public.t_fee_statistics
    ADD CONSTRAINT fee_slot_tx_index UNIQUE (slot, tx_index);
ALTER TABLE ONLY public.t_fee_statistics
    ADD CONSTRAINT fee_tx_id UNIQUE (tx_id);
CREATE INDEX idx_fee_statistics_ttl ON public.t_fee_statistics (created_at);

-- quicknode接口手续费统计表
-- 对应旧工程 t_sol_qn_fee
DROP TABLE IF EXISTS public.t_qn_fee;
CREATE TABLE public.t_qn_fee
(
    id         integer NOT NULL, -- 主键id, 对应 Id
    slot       bigint,           -- 交易所在slot, 对应 Slot
    low_avg    numeric(78, 0),   -- 最低平均优先费用，对应 LowAvg
    medium_avg numeric(78, 0),   -- 中等平均优先费用，对应 MediumAvg
    high_avg   numeric(78, 0),   -- 高平均优先费用，对应 HighAvg
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE ONLY public.t_qn_fee
    ADD CONSTRAINT t_qn_fee_pkey PRIMARY KEY (id);
CREATE INDEX idx_qn_fee_ttl ON public.t_qn_fee (created_at);

-- 地址信息表
-- 对应旧工程 t_sol_native_account_info
DROP TABLE IF EXISTS public.t_native_account_info;
CREATE TABLE public.t_native_account_info
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64) NOT NULL, -- 用户solana地址，对应 NativeAccount
    token_account  character varying(64) NOT NULL, -- 用户token account，对应 TokenAccount
    invite_code    character varying(16) NOT NULL, -- 用户邀请码，对应 InviteCode
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 邀请关系表
-- 对应旧工程 t_sol_determine_invite_record
DROP TABLE IF EXISTS public.t_invite_relation;
CREATE TABLE public.t_invite_relation
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    inviter       character varying(64)  NOT NULL, -- 邀请人solana地址，对应 InviterTokenAccount
    invitee       character varying(64)  NOT NULL, -- 被邀请人solana地址，对应 InviteeNativeAccount
    channel       character varying(64)  NOT NULL, -- 邀请渠道，对应 InviteChannel
    inviter_level integer DEFAULT 1      NOT NULL, -- 邀请人级别，对应 Level
    tx_id         character varying(128) NOT NULL, -- 邀请时的交易id，对应 TransferTxId
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_invite_relation_invitee ON public.t_invite_relation (invitee);

-- 业务信息表
-- 新工程独有表
DROP TABLE IF EXISTS public.t_service_info;
CREATE TABLE public.t_service_info
(
    service     character varying(64)                NOT NULL, -- 微服务业务模块，目前只有reward,pos
    sub_service character varying(64)                NOT NULL, -- 微服务下面的子服务，例如: reward下面的take token,give token等
    address     character varying(64)                NOT NULL, -- 发放奖励的地址
    webhook     character varying(1024) DEFAULT NULL,          -- webhook url
    mq_group    character varying(64)   DEFAULT NULL,          -- kafka消息队列的group
    mq_topic    character varying(64)   DEFAULT NULL,          -- kafka消息队列的topic
    hook_type   integer                 DEFAULT 0    NOT NULL, -- 交易消息回调通知类型 0.通过kafka 1.通过webhook
    tx_source   integer                 DEFAULT 0    NOT NULL, -- 交易来源 0.由服务器签发发到区块链 1.直接从区块链监听到的交易
    confirm     bool                    DEFAULT true NOT NULL, -- 是否确认
    multi_sign  bool                    DEFAULT true NOT NULL, -- 是否多签
    created_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_service_info_service_sub_service ON public.t_service_info (service, sub_service);

-- 业务私钥
-- 新工程整合旧工程里面 t_reward_key_config,t_pos_reward_key_config 两张表过来
-- 需要手动导入
DROP TABLE IF EXISTS public.t_service_key;
CREATE TABLE public.t_service_key
(
    service       character varying(64)   NOT NULL, -- 微服务业务名称，新工程重新定义
    sub_service   character varying(64)   NOT NULL, -- 子业务名称，新工程重新定义
    encrypted_key character varying(1024) NOT NULL, -- 被加密后的私钥，对应 EncryptedKey
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_service_key_service_sub_service ON public.t_service_key (service, sub_service);
CREATE UNIQUE INDEX IF NOT EXISTS uq_service_info_service_sub_service ON public.t_service_info (service, sub_service);
CREATE UNIQUE INDEX IF NOT EXISTS uq_service_key_service_sub_service ON public.t_service_key (service, sub_service);
CREATE UNIQUE INDEX IF NOT EXISTS uq_tx_scan_info_service_sub_service_pda ON public.t_tx_scan_info (service, sub_service, pda_account);

-- 业务交易扫描表
-- 新工程整合旧工程里面的 t_reward_scan_info, t_pos_scan_info等几张表
-- 需要手动导入
DROP TABLE IF EXISTS public.t_tx_scan_info;
CREATE TABLE public.t_tx_scan_info
(
    service        character varying(64)    NOT NULL, -- 微服务业务名称，新工程重新定义
    sub_service    character varying(64)    NOT NULL, -- 子业务名称，新工程重新定义
    native_account character varying(64)    NOT NULL, -- 业务主地址，一般指发放奖励的地址
    pda_account    character varying(64)    NOT NULL, -- pda account，绝大多数情况下是 token account
    until_tx_id    character varying(128)   NOT NULL, -- getSignaturesForAddress 的untilTxId参数
    before_tx_id   character varying(128),            -- getSignaturesForAddress 的beforeTxId参数
    slot           numeric(78, 0) DEFAULT 0 NOT NULL, -- until_tx_id所在的slot
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_tx_scan_info_service_sub_service_pda ON public.t_tx_scan_info (service, sub_service, pda_account);

-- 业务交易表
-- 新工程整合旧工程里面的 t_reward_tx,t_pos_tx等几张表
-- 预演只需要导出 tx_state=0的数据1000条
-- 后续需要增加 ttl 机制减少表体积
DROP TABLE IF EXISTS public.t_service_tx;
CREATE TABLE public.t_service_tx
(
    record_id       ULID      DEFAULT gen_ulid() NOT NULL PRIMARY KEY,
    service         VARCHAR(64)                  NOT NULL, -- 微服务业务名称，新工程重新定义
    sub_service     VARCHAR(64)                  NOT NULL, -- 子业务名称，新工程重新定义
    tx_id           VARCHAR(128)                 NOT NULL, -- 交易id
    tx_state        INTEGER   DEFAULT 0          NOT NULL, -- 交易状态
    retry_count     INTEGER   DEFAULT 0          NOT NULL, -- 尝试重新查询次数
    next_retry_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,   -- 下次重试查询时间
    max_retries     INTEGER   DEFAULT 5          NOT NULL, -- 最大重试次数
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_service_tx_ttl ON public.t_service_tx (created_at);

-- 旧工程 t_sol_fund_flow
-- 需要导入并且搞分表，减少单表体积
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_fund_flow;
CREATE TABLE public.t_fund_flow
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    is_token     boolean                NOT NULL, -- 流水资金是否是token
    from_account character varying(64)  NOT NULL, -- 流水资金发起地址
    to_account   character varying(64)  NOT NULL, -- 流水地址金接收地址
    tx_id        character varying(128) NOT NULL, -- 交易id
    direction    character varying(8)   NOT NULL, -- 流水方向
    service_type character varying(32)  NOT NULL, -- 业务类型(一般指sub service,例如: take token ,give token等)
    flow_type    character varying(32)  NOT NULL, -- 流水类型
    decimals     smallint               NOT NULL, -- 资金的金额精度
    amount       numeric(78, 0)         NOT NULL, -- 流水资金金额，原始值
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_fund_flow_tx_to_flow ON public.t_fund_flow (tx_id, to_account, flow_type);

-- 奖励黑客表
-- 对应旧工程 t_reward_hacker
DROP TABLE IF EXISTS public.t_hacker_account;
CREATE TABLE public.t_hacker_account
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64) NOT NULL, -- solana地址,对应 NativeAccount
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 奖励排除名单表
-- 对应旧工程 t_reward_exclude
DROP TABLE IF EXISTS public.t_exclude_account;
CREATE TABLE public.t_exclude_account
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64) NOT NULL, -- solana地址,对应 NativeAccount
    remark         character varying(128),
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

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
    rate numeric(3, 2) NOT NULL -- 对应 Rate
);

-- take token 奖励配置表
-- 旧工程 t_sol_official_give_token_reward_rule
DROP TABLE IF EXISTS public.t_take_token_config;
CREATE TABLE public.t_take_token_config
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    invite_code    character varying(16) DEFAULT NULL:: character varying, -- 邀请码，默认规则为空，对应 InviteCode
    reward_account character varying(64) NOT NULL,                         -- 下发奖励的solana地址，对应 RewardNativeAccount
    cost_account   character varying(64) NOT NULL,                         -- 接收成本费的solana地址，对应 DexNativeAccount
    amount         numeric(78, 0)        NOT NULL,                         -- 奖励金额，对应 Amount
    invite_amount  numeric(78, 0)        NOT NULL,                         -- 确定邀请关系奖励金额，对应 InviteAmount
    cost_fee_rate  numeric(78, 0)        NOT NULL,                         -- 成本费费率，对应 DexFeeRate
    invited_rate   numeric(78, 0)        NOT NULL,                         -- 新工程新增字段，确定邀请关系时的成本费率
    max_cost_fee   numeric(78, 0)        NOT NULL,                         -- 最大成本费，对应 MaxDexFee
    is_default     boolean               DEFAULT false,                    -- 是否是默认规则，对应 Default
    reward_inviter boolean               DEFAULT true,                     -- 是否奖励邀请人，对应 RewardInviter
    invited        boolean               DEFAULT true,                     -- 是否确定邀请关系，对应 DetermineInvite
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- take token 记录表
-- 旧工程 t_sol_official_give_token_record
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_take_token_record;
CREATE TABLE public.t_take_token_record
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    reward_account  character varying(64)  NOT NULL, -- 下发奖励的solana地址，对应 RewardNativeAccount
    receipt_account character varying(64)  NOT NULL, -- 接收奖励的地址，对应 ReceiptNativeAccount
    cost_account    character varying(64)  NOT NULL, -- 接收成本费的地址，对应 DexNativeAccount
    tx_id           character varying(128) NOT NULL, -- 获取奖励交易id，对应 RewardTxId
    amount          numeric(78, 0)         NOT NULL, -- 奖励金额，对应 Amount
    cost_fee        numeric(78, 0)         NOT NULL, -- 成本费，对应 DexFee
    use_invite_code boolean                NOT NULL, -- 是否使用邀请码，UseInviteCode
    invite_code     character varying(16),           -- 邀请码，InviteCode
    tx_state        integer,                         -- 交易状态，State
    invited         boolean,                         -- 是否确定邀请关系，DetermineInvite
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_take_token_record_ttl ON public.t_take_token_record (created_at);

-- 当日take token和lottery统计表
-- 旧工程 t_daily_claim_stats
DROP TABLE IF EXISTS public.t_daily_claim_stats;
CREATE TABLE public.t_daily_claim_stats
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)        NOT NULL, -- take token和lottery的地址，对应 NativeAccount
    take_date      date                         NOT NULL, -- take token的日期，对应 TakeShitDate
    take_count     integer        DEFAULT 0     NOT NULL, -- take token的次数，对应 TakeShitCount
    need_lottery   boolean        DEFAULT false NOT NULL, -- 是否需要抽奖，对应 NeedLottery
    last_take_time timestamp without time zone,           -- 上次take token的时间，对应 LastTakeTime
    total_lottery  numeric(78, 0) DEFAULT 0     NOT NULL, -- 总计 take token 的金额，对应 TotalLottery
    total_take     numeric(78, 0) DEFAULT 0     NOT NULL, -- 总计 lottery 的金额，对应 TotalTake
    lottery_count  integer        DEFAULT 0     NOT NULL, -- 抽奖次数，对应 LotteryCount
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_daily_claim_stats_ttl ON public.t_daily_claim_stats (created_at);

ALTER TABLE t_daily_claim_stats
    ADD CONSTRAINT uq_daily_claim_stats_account_date UNIQUE (native_account, take_date);

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

-- 抽奖奖励发放表
-- 旧工程 t_reward_lottery
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_lottery_reward;
CREATE TABLE public.t_lottery_reward
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64) NOT NULL, -- lottery的地址，对应 NativeAccount
    reward_amount  numeric(78, 0),                 -- 奖励金额，对应 RewardAmount
    reward_type    integer,                        -- 奖励类型，对应 RewardType
    reward_state   integer DEFAULT 0,              -- 奖励状态，对应 State
    pending        boolean DEFAULT false,          -- 奖励是否被处理中，对应 Pending
    reward_day     date                  NOT NULL, -- 奖励当天，对应 Day
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lottery_reward_ttl ON public.t_lottery_reward (created_at);

-- 领取抽奖奖励记录表
-- 旧工程 t_reward_lottery_claim_record
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_lottery_claim;
CREATE TABLE public.t_lottery_claim
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    reward_ids public.ulid[] NOT NULL, -- 领取的抽奖奖励记录id，对应 RewardIds
    tx_id      character varying(128), -- 抽奖的交易id，对应 TxId
    tx_state   integer DEFAULT 0,      -- 奖励状态,对应 State
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
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

-- give token 记录表
-- 旧工程 t_sol_transfer_checked_record
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_give_token_record;
CREATE TABLE public.t_give_token_record
(
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    from_account    character varying(64)  NOT NULL, -- 发送 token 的 solana 地址， 对应 FromNativeAccount
    receipt_account character varying(64)  NOT NULL, -- 接收 token 的 solana 地址，对应 ReceiptNativeAccount
    tx_id           character varying(128) NOT NULL, -- give token 活动的solana交易id，对应 TransferTxId
    amount          numeric(78, 0)         NOT NULL, -- give token 的金额，对应 TransferAmount
    tx_state        integer,                         -- 交易状态，对应 State
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_give_token_record_ttl ON public.t_give_token_record (created_at);

-- 奖励码奖励规则
-- 旧工程 t_reward_code_rule
DROP TABLE IF EXISTS public.t_reward_code_config;
CREATE TABLE public.t_reward_code_config
(
    record_id      ulid DEFAULT gen_ulid() NOT NULL,
    reward_account character varying(64)   NOT NULL, -- 下发奖励的地址，对应 RewardNativeAccount
    cost_account   character varying(64)   NOT NULL, -- 接收成本费的地址，对应 DexNativeAccount
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- 奖励码
-- 旧工程 t_reward_code
DROP TABLE IF EXISTS public.t_reward_code;
CREATE TABLE public.t_reward_code
(
    record_id      ulid                   DEFAULT gen_ulid() NOT NULL,
    reward_code    character varying(6)                      NOT NULL,      -- 奖励码，对应 RewardCode
    reward_amount  numeric(78, 0)                            NOT NULL,      -- 奖励金额，对应 RewardAmount
    reward_state   integer                DEFAULT 0          NOT NULL,      -- 奖励状态，对应 State
    native_account character varying(64)  DEFAULT NULL,                     -- 领取奖励码的solana地址i，对应 NativeAccount
    tx_id          character varying(128) DEFAULT NULL:: character varying, -- 交易id，对应 TxId
    tx_state       integer                DEFAULT 0          NOT NULL,
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
    record_id          ulid                           DEFAULT gen_ulid() NOT NULL,
    reward_account     character varying(64) NOT NULL,                   -- 发放奖励的地址，对应 RewardNativeAccount
    cost_account       character varying(64) NOT NULL,                   -- 接收成本费的地址，对应 DexNativeAccount
    quote_rate         numeric(5, 2)         NOT NULL,                   -- 兑换费率，对应 Rate
    cost_rate          numeric(5, 2)         NOT NULL,                   -- 成本费率，对应 CostRate
    global_daily_limit numeric(78, 0)        NOT NULL DEFAULT 500000000, -- 全局每session默认兑换限额（原始值）
    user_daily_limit   numeric(78, 0)        NOT NULL DEFAULT 30000000,  -- 用户每日默认兑换限额（原始值）
    created_at         timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 全局每日兑换限额表（每天分2个session，SGT 08:00/20:00 各重置一次）
-- 对应旧工程 t_global_daily_exchange_limit
DROP TABLE IF EXISTS t_campaign_quote_limit;
CREATE TABLE t_campaign_quote_limit
(
    id          BIGSERIAL PRIMARY KEY,
    daily_limit NUMERIC(78, 0) NOT NULL  DEFAULT 500000000,    -- 每个session全局可兑换量（原始值，500000 token * 1000 = 500000000）
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

-- 兑换社交媒体积分为token的记录
-- 旧工程表 t_sol_exchange_campaign_score_to_token_record
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_campaign_quote_record;
CREATE TABLE public.t_campaign_quote_record
(
    record_id       ulid           NOT NULL DEFAULT gen_ulid(),   -- 记录Id
    reward_account  VARCHAR(64)    NOT NULL,                      -- 发放奖励的native account，对应 RewardNativeAccount
    receipt_account VARCHAR(64)    NOT NULL,                      -- 接收奖励的native account，对应 ReceiptNativeAccount
    provider        VARCHAR(64)    NOT NULL,                      -- 社交媒体，对应 Provider
    user_id         VARCHAR(64)    NOT NULL,                      -- 用户ID，对应 UserId
    tx_id           VARCHAR(128)   NOT NULL,                      -- 奖励token的txId，对应 ExchangeTxId
    score_flow_id   INT            NOT NULL,                      -- 积分流水Id，对应 ScoreFlowId
    score_tx_id     VARCHAR(128)   NOT NULL,                      -- 积分交易Id，对应 ScoreTxId
    amount          NUMERIC(78, 0) NOT NULL,                      -- 获取的token额度，对应 Amount
    score           NUMERIC(78, 0) NOT NULL,                      -- 兑换的积分额度，对应 Score
    quote_state     INT,                                          -- 状态,-1.失败 0.初始化 1.成功，对应 State
    session         SMALLINT       NOT NULL DEFAULT 0,            -- 下单时的全局场次（0=早/1=晚），用于Kafka回调时精确恢复对应session额度
    user_quota_date DATE           NOT NULL DEFAULT CURRENT_DATE, -- 下单时用户所在的SGT自然日，用于Kafka回调时精确恢复用户额度
    created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX idx_campaign_quote_record_ttl ON public.t_campaign_quote_record (created_at);
CREATE INDEX ON public.t_campaign_quote_record (receipt_account);
CREATE INDEX ON public.t_campaign_quote_record (provider, user_id);
CREATE INDEX ON public.t_campaign_quote_record (tx_id);

-- POS奖励发放规则
-- 旧工程 t_sol_pos_reward_rule
DROP TABLE IF EXISTS public.t_pos_reward_config;
CREATE TABLE public.t_pos_reward_config
(
    record_id          ulid           NOT NULL DEFAULT gen_ulid(), -- 记录Id
    reward_account     VARCHAR(64)    NOT NULL,                    -- 奖励token的native account，对应 RewardNativeAccount
    cost_account       VARCHAR(64)    NOT NULL,                    -- 收取dex手续费的native account，对应 DexNativeAccount
    cost_fee_rate      NUMERIC(78, 0) NOT NULL,                    -- 领取时支付给dex的手续费费率，对应 DexFeeRate
    max_cost_fee       NUMERIC(78, 0) NOT NULL,                    -- 领取时最多支付给dex的费用，对应 MaxDexFee
    quote_token_amount NUMERIC(78, 0) NOT NULL,                    -- 领取时最多支付给dex的费用，对应 QuoteTokenAmount
    created_at         TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- POS 星级用户评定规则
-- 旧工程 t_sol_pos_star_level_rule
DROP TABLE IF EXISTS public.t_pos_star_level_rule;
CREATE TABLE public.t_pos_star_level_rule
(
    record_id    ulid           NOT NULL DEFAULT gen_ulid(), -- 记录Id
    amount       NUMERIC(78, 2) NOT NULL,                    -- 持币金额，对应 Amount
    group_amount NUMERIC(78, 2) NOT NULL,                    -- 团队持币数量，对应 GroupAmount
    star_level   INT            NOT NULL,                    -- 星级数量，对应 StarLevel
    rate         NUMERIC(5, 2)  NOT NULL,                    -- 费率，对应 Rate
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- POS 固定星级用户配置
-- 旧工程 t_sol_pos_star_level_config
DROP TABLE IF EXISTS public.t_pos_star_whitelist;
CREATE TABLE public.t_pos_star_whitelist
(
    record_id      ulid        NOT NULL DEFAULT gen_ulid(), -- 记录Id
    native_account VARCHAR(64) NOT NULL,                    -- 地址，对应 NativeAccount
    star_level     INT         NOT NULL,                    -- 持币地址的星级，对应 StarLevel
    rate           NUMERIC(5, 2),                           -- 星级用户的极差费率，对应 Rate
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE UNIQUE INDEX ON public.t_pos_star_whitelist (native_account);

-- POS 任务奖励配置
-- 旧工程 t_sol_pos_mission_config
DROP TABLE IF EXISTS public.t_pos_mission_config;
CREATE TABLE public.t_pos_mission_config
(
    record_id   ulid NOT NULL DEFAULT gen_ulid(), -- 记录Id
    reward_type INT  NOT NULL,                    -- 奖励类型 0.固定收益 1.加推特奖励 2.推特转贴奖励 3.推特点赞或回复，对应 RewardType
    starred     BOOL NOT NULL,                    -- 任务奖励费率，对应 Starred
    rate        NUMERIC(5, 2),                    -- 任务奖励费率，对应 Rate
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_pos_mission_config (reward_type);

-- POS 每日快照表
-- 旧工程 t_sol_pos_snap_shot
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_pos_snap_shot;
CREATE TABLE public.t_pos_snap_shot
(
    record_id      ulid        NOT NULL DEFAULT gen_ulid(), -- 记录Id
    native_account VARCHAR(64) NOT NULL,                    -- native account 地址，对应 NativeAccount
    amount         NUMERIC(78, 0),                          -- 持币地址的金额，对应 Amount
    star_level     INT                  DEFAULT 0,          -- 持币地址的星级，对应 StarLevel
    rate           NUMERIC(5, 2),                           -- 极差费率，对应 Rate
    range_base     NUMERIC(5, 2),                           -- 极差基数，对应 RangeBase
    snap_day       DATE        NOT NULL,                    -- 快照的日期，对应 Day
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE UNIQUE INDEX ON public.t_pos_snap_shot (native_account, snap_day);
CREATE INDEX idx_pos_snap_shot_ttl ON public.t_pos_snap_shot (created_at);

-- POS 奖励明细表
-- 旧工程 t_sol_pos_reward
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_pos_reward;
CREATE TABLE public.t_pos_reward
(
    record_id      ulid        NOT NULL DEFAULT gen_ulid(), -- 记录Id
    group_id       VARCHAR(64) NOT NULL,                    -- 计算极差时的团队根地址，对应 GroupId
    native_account VARCHAR(64) NOT NULL,                    -- 持币地址，对应 NativeAccount
    star_level     INT                  DEFAULT 0,          -- 持币地址的星级，对应 StarLevel
    base           NUMERIC(78, 0),                          -- pos持币的Base(StarLevel=0时base=持币金额,StarLevel=1时Base=(持币金额+极差))，对应 Base
    rate           NUMERIC(5, 2),                           -- 奖励费率，对应 Rate
    reward_amount  NUMERIC(78, 0),                          -- 奖励金额，对应 RewardAmount
    reward_type    INT,                                     -- 奖励类型 1.固定收益 2.加推特奖励 3.推特转贴奖励 4.推特点赞或回复，对应 RewardType
    reward_state   INT                  DEFAULT 0,          -- 状态,-1.过期 0.初始化 1.已经领取，对应 State
    starred        BOOL        NOT NULL,                    -- 是否是星级用户奖励，对应 Starred
    pending        BOOL                 DEFAULT false,      -- 是否正在被处理，对应 Pending
    snap_day       DATE        NOT NULL,                    -- 快照的日期，对应 Day
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_pos_reward (group_id);
CREATE INDEX ON public.t_pos_reward (native_account, reward_state, pending);
CREATE UNIQUE INDEX ON public.t_pos_reward (native_account, snap_day, reward_type, starred);
CREATE INDEX idx_pos_reward_ttl ON public.t_pos_reward (created_at);

-- POS 奖励领取表
-- 旧工程 t_sol_pos_reward_claim_record
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_pos_reward_claim;
CREATE TABLE public.t_pos_reward_claim
(
    record_id  ulid NOT NULL DEFAULT gen_ulid(), -- 记录Id
    reward_ids ulid[] NOT NULL,                  -- 奖励Id，对应 RewardIds
    tx_id      VARCHAR(128),                     -- 交易Id，对应 TxId
    tx_state   INT           DEFAULT 0,          -- 状态,-2.过期 -1.失败 0.初始化 1.成功，对应 State
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON "public"."t_pos_reward_claim" (tx_id);
CREATE INDEX ON "public"."t_pos_reward_claim" (tx_state, created_at);

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
    record_id       ulid NOT NULL DEFAULT gen_ulid(),                      -- 记录Id
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
    record_id    ulid           NOT NULL DEFAULT gen_ulid(),            -- 记录Id
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
    record_id      ulid        NOT NULL DEFAULT gen_ulid(),               -- 记录Id
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
    record_id          ulid           not null default gen_ulid(),            -- 记录Id
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
    record_id      ulid        NOT NULL DEFAULT gen_ulid(),               -- 记录Id
    group_id       varchar(64) NOT NULL,                                  -- 根地址
    native_account varchar(64) NOT NULL,                                  -- 质押地址
    star_level     int                  DEFAULT 0,                        -- 质押地址的星级
    base           numeric(78, 0),                                        -- 质押数量的Base
    rate           numeric(5, 2),                                         -- 奖励费率
    reward_amount  numeric(78, 0),                                        -- 奖励金额
    stake_type     int         NOT NULL,                                  -- 质押类型 0.180天质押 1.360天质押
    reward_type    int         NOT NULL,                                  -- 奖励类型 0.质押每日固定利息 1.质押邀请奖励 2.质押激励奖励 -  个人奖励 3.质押激励奖励 - 星级奖励 4.质押激励奖励 - 团队奖励
    reward_state   int                  DEFAULT 0,                        -- 状态,-1.过期 0.初始化 1.已经领取
    starred        bool        NOT NULL,                                  -- 是否是星级用户奖励
    tx_id          varchar(128),                                          -- 交易Id
    pending        bool                 DEFAULT false,                    -- 是否正在被处理
    snap_day       date        NOT NULL,                                  -- 快照的日期
    created_at     timestamp without time zone DEFAULT current_timestamp, -- 记录创建时间
    updated_at     timestamp without time zone DEFAULT current_timestamp, -- 记录更新时间
    primary key (record_id)
);
create index on public.t_stake_reward (group_id);
create index on public.t_stake_reward (native_account, reward_state, pending);
create unique index on public.t_stake_reward (native_account, snap_day, reward_type, starred);
create index idx_stake_reward_ttl on public.t_stake_reward (created_at) where not (reward_type = 0 and reward_state = 0);

-- 质押奖励领取记录表
-- 旧工程 t_sol_stake_reward_claim_record
drop table if exists public.t_stake_reward_claim;
create table public.t_stake_reward_claim
(
    record_id  ulid not null default gen_ulid(),                      -- 记录Id
    reward_ids ulid[] not null,                                       -- 奖励Id
    tx_id      varchar(128),                                          -- 交易Id
    tx_state   int           default 0,                               -- 状态,-2.过期 -1.失败 0.初始化 1.成功
    created_at timestamp without time zone default current_timestamp, -- 记录创建时间
    updated_at timestamp without time zone default current_timestamp, -- 记录更新时间
    primary key (record_id)
);
create index on public.t_stake_reward_claim (tx_id);
create index on public.t_stake_reward_claim (tx_state, created_at);

-- stake 每日快照表
-- 旧工程 t_sol_stake_snap_shot
DROP TABLE IF EXISTS public.t_stake_snap_shot;
CREATE TABLE public.t_stake_snap_shot
(
    record_id      ulid        not null default gen_ulid(),
    native_account varchar(64) not null,                                  -- native account 地址
    amount         numeric(78, 0),                                        -- 质押金额
    stake_type     int         not null,                                  -- 质押类型 0.180天 1.360天
    snap_day       date        not null,                                  -- 快照的日期
    created_at     timestamp without time zone default current_timestamp, -- 记录创建时间
    updated_at     timestamp without time zone default current_timestamp, -- 记录更新时间
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_stake_snap_shot (native_account);
CREATE UNIQUE INDEX ON public.t_stake_snap_shot (native_account, stake_type, snap_day);
CREATE INDEX idx_stake_snap_shot_ttl ON public.t_stake_snap_shot (created_at);

-- 质押记录表
drop table if exists public.t_stake_record;
create table public.t_stake_record
(
    record_id     ulid not null default gen_ulid(),-- 记录Id
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
DROP TABLE IF EXISTS public.t_stake_buy_token;
CREATE TABLE public.t_stake_buy_token
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
    expired          bool                                        not null default false,
    created_at       timestamp without time zone default current_timestamp,
    expired_at       timestamp without time zone default current_timestamp
);
CREATE INDEX idx_stake_buy_token_locked ON public.t_stake_buy_token (locked);
CREATE INDEX idx_stake_buy_token_locked_at ON public.t_stake_buy_token (locked_at);
CREATE INDEX idx_stake_buy_token_locked_by ON public.t_stake_buy_token (locked_by);
CREATE INDEX idx_stake_buy_token_slot ON public.t_stake_buy_token (slot);
ALTER TABLE t_stake_buy_token
    ADD CONSTRAINT uq_stake_buy_token_tx_id UNIQUE (tx_id);
CREATE INDEX idx_stake_buy_token_ttl ON public.t_stake_buy_token (created_at);

-- 区域经理奖励配置
drop table if exists public.t_stake_leader_reward_config;
create table public.t_stake_leader_reward_config
(
    record_id      ulid not null default gen_ulid(),-- 记录Id
    reward_account varchar(64),-- 奖励发放地址
    created_at     timestamp without time zone default current_timestamp,-- 记录创建时间
    updated_at     timestamp without time zone default current_timestamp,-- 记录更新时间
    primary key (record_id)
);

-- 总区域经理表
drop table if exists public.t_stake_total_leader;
create table public.t_stake_total_leader
(
    record_id      ulid not null default gen_ulid(),-- 记录Id
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
    record_id      ulid     not null default gen_ulid(),-- 记录Id
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
    record_id      ulid        not null default gen_ulid(),
    native_account varchar(64) not null,               -- 区域经理地址
    staker         varchar(64) not null,               -- 触发奖励的质押者
    reward_type    int         not null,               -- 0=直接区域经理10% 1=区域经理7% 2=上级leader3% 3=总区域经理
    base_amount    numeric(78, 0),                     -- 基础金额(min(购买量,质押量))
    stake_share    numeric(5, 2),                      -- 奖励费率
    reward_amount  numeric(78, 0),                     -- 奖励金额
    reward_state   int         not null default 0,     -- -1=过期 0=初始化 1=已领取
    pending        bool        not null default false, -- 是否正在处理中
    tx_id          varchar(128),                       -- 领取交易Id
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
create index on public.t_stake_leader_reward (native_account, reward_state, pending);

-- 区域经理奖励领取表
drop table if exists public.t_stake_leader_reward_claim;
create table public.t_stake_leader_reward_claim
(
    record_id      ulid        not null default gen_ulid(),
    native_account varchar(64) not null,           -- 领取者（区域经理）地址
    reward_ids     ulid[] not null,                -- 本次领取的奖励Id列表
    tx_id          varchar(128),
    tx_state       int                  default 0, -- -2=过期 -1=失败 0=初始化 1=成功
    created_at     timestamp without time zone default current_timestamp,
    updated_at     timestamp without time zone default current_timestamp,
    primary key (record_id)
);
create index on public.t_stake_leader_reward_claim (tx_id);
create index on public.t_stake_leader_reward_claim (tx_state, created_at);
create index on public.t_stake_leader_reward_claim (native_account, tx_state);


-- 授权阶段
-- 1. 把现有所有表的 owner 改成 meme_server
DO $$
DECLARE r RECORD;
BEGIN
FOR r IN
SELECT tablename
FROM pg_tables
WHERE schemaname = 'public' LOOP
          EXECUTE 'ALTER TABLE public.' || quote_ident(r.tablename) || ' OWNER TO meme_server';
END LOOP;
END $$;

-- 2. 让 meme_server 可以在 public schema 下建表（分区表需要）
GRANT CREATE ON SCHEMA public TO meme_server;

-- 3. 设置默认权限：以后 postgres 用户建的表自动授权给 meme_server
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO meme_server;
