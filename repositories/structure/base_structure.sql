CREATE EXTENSION IF NOT EXISTS ulid WITH SCHEMA public;

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

-- 用户钱包rpc配置表
-- 对应旧工程表 t_user_wallet_rpc_config
DROP TABLE IF EXISTS public.t_user_wallet_rpc_config;
CREATE TABLE public.t_user_wallet_rpc_config
(
    record_id  public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    chain_name character varying(64)                                 NOT NULL, -- 链名称, 对应旧工程 Chain
    rpc_url    character varying(1024)                               NOT NULL, -- rpc url,对应旧工程 RpcUrl
    wss_url    character varying(1024)                               NOT NULL, -- wss url,对应旧工程 WssUrl
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 主网钱包rpc配置
-- 新工程新增表
DROP TABLE IF EXISTS public.t_mainnet_rpc_config;
CREATE TABLE public.t_mainnet_rpc_config
(
    record_id  public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    chain_name character varying(64)                                 NOT NULL, -- 链名称, 对应旧工程 Chain
    rpc_url    character varying(1024)                               NOT NULL, -- rpc url,对应旧工程 RpcUrl
    wss_url    character varying(1024)                               NOT NULL, -- wss url,对应旧工程 WssUrl
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 链配置表
-- 对应旧工程 t_chain_config
DROP TABLE IF EXISTS public.t_chain_config;
CREATE TABLE public.t_chain_config
(
    chain_name character varying(64)   NOT NULL, -- 链名称, 对应旧工程 Chain
    rpc_url    character varying(1024) NOT NULL, -- rpc url,对应旧工程 RpcUrl
    wss_url    character varying(1024) NOT NULL, -- wss url,对应旧工程 WssUrl
    decimals   integer                 NOT NULL,
    symbol     character varying(64)   NOT NULL,
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
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
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
    ADD CONSTRAINT t_sol_qn_fee_pkey PRIMARY KEY (id);

-- 地址信息表
-- 对应旧工程 t_sol_native_account_info
DROP TABLE IF EXISTS public.t_native_account_info;
CREATE TABLE public.t_native_account_info
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL, -- 用户solana地址，对应 NativeAccount
    token_account  character varying(64)                                 NOT NULL, -- 用户token account，对应 TokenAccount
    invite_code    character varying(16)                                 NOT NULL, -- 用户邀请码，对应 InviteCode
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 邀请关系表
-- 对应旧工程 t_invite_relation
DROP TABLE IF EXISTS public.t_invite_relation;
CREATE TABLE public.t_invite_relation
(
    record_id     public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    inviter       character varying(64)                                 NOT NULL, -- 邀请人solana地址，对应 InviterTokenAccount
    invitee       character varying(64)                                 NOT NULL, -- 被邀请人solana地址，对应 InviteeNativeAccount
    channel       character varying(64)                                 NOT NULL, -- 邀请渠道，对应 InviteChannel
    inviter_level integer                     DEFAULT 1                 NOT NULL, -- 邀请人级别，对应 Level
    tx_id         character varying(128)                                NOT NULL, -- 邀请时的交易id，对应 TransferTxId
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 业务信息表
-- 新工程独有表
DROP TABLE IF EXISTS public.t_service_info;
CREATE TABLE public.t_service_info
(
    service     character varying(64)                    NOT NULL, -- 微服务业务模块，目前只有reward,pos
    sub_service character varying(64)                    NOT NULL, -- 微服务下面的子服务，例如: reward下面的take token,give token等
    address     character varying(64)                    NOT NULL, -- 发放奖励的地址
    webhook     character varying(1024)     DEFAULT NULL,          -- webhook url
    mq_group    character varying(64)       DEFAULT NULL,          -- kafka消息队列的group
    mq_topic    character varying(64)       DEFAULT NULL,          -- kafka消息队列的topic
    hook_type   integer                     DEFAULT 0    NOT NULL, -- 交易消息回调通知类型 0.通过kafka 1.通过webhook
    tx_source   integer                     DEFAULT 0    NOT NULL, -- 交易来源 0.由服务器签发发到区块链 1.直接从区块链监听到的交易
    confirm     bool                        DEFAULT true NOT NULL, -- 是否确认
    multi_sign  bool                        DEFAULT true NOT NULL, -- 是否多签
    created_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

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

-- 业务交易扫描表
-- 新工程整合旧工程里面的 t_reward_scan_info, t_pos_scan_info等几张表
-- 需要手动导入
DROP TABLE IF EXISTS public.t_tx_scan_info;
CREATE TABLE public.t_tx_scan_info
(
    service        character varying(64)                 NOT NULL, -- 微服务业务名称，新工程重新定义
    sub_service    character varying(64)                 NOT NULL, -- 子业务名称，新工程重新定义
    native_account character varying(64)                 NOT NULL, -- 业务主地址，一般指发放奖励的地址
    pda_account    character varying(64)                 NOT NULL, -- pda account，绝大多数情况下是 token account
    until_tx_id    character varying(128)                NOT NULL, -- getSignaturesForAddress 的untilTxId参数
    before_tx_id   character varying(128),                         -- getSignaturesForAddress 的beforeTxId参数
    slot           numeric(78, 0)              DEFAULT 0 NOT NULL, -- until_tx_id所在的slot
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

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

-- 旧工程 t_sol_fund_flow
-- 需要导入并且搞分表，减少单表体积
-- 预演导出只需要导出前 1000 条
DROP TABLE IF EXISTS public.t_fund_flow;
CREATE TABLE public.t_fund_flow
(
    record_id    public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    is_token     boolean                                               NOT NULL, -- 流水资金是否是token
    from_account character varying(64)                                 NOT NULL, -- 流水资金发起地址
    to_account   character varying(64)                                 NOT NULL, -- 流水地址金接收地址
    tx_id        character varying(128)                                NOT NULL, -- 交易id
    direction    character varying(8)                                  NOT NULL, -- 流水方向
    service_type character varying(32)                                 NOT NULL, -- 业务类型(一般指sub service,例如: take token ,give token等)
    flow_type    character varying(32)                                 NOT NULL, -- 流水类型
    decimals     smallint                                              NOT NULL, -- 资金的金额精度
    amount       numeric(78, 0)                                        NOT NULL, -- 流水资金金额，原始值
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_fund_flow_tx_to_flow ON public.t_fund_flow (tx_id, to_account, flow_type);

-- 奖励黑客表
-- 对应旧工程 t_reward_hacker
DROP TABLE IF EXISTS public.t_hacker_account;
CREATE TABLE public.t_hacker_account
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL, -- solana地址,对应 NativeAccount
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 奖励排除名单表
-- 对应旧工程 t_reward_exclude
DROP TABLE IF EXISTS public.t_exclude_account;
CREATE TABLE public.t_exclude_account
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL, -- solana地址,对应 NativeAccount
    remark         character varying(128),
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
