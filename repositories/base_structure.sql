CREATE EXTENSION IF NOT EXISTS ulid WITH SCHEMA public;

DROP TABLE IF EXISTS public.t_system_config;
CREATE TABLE public.t_system_config
(
    env        integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_aws_config;
CREATE TABLE public.t_aws_config
(
    access_key_id     character varying(128) NOT NULL,
    secret_access_key character varying(128) NOT NULL,
    region            character varying(32)  NOT NULL,
    created_at        timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at        timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_user_wallet_rpc_config;
CREATE TABLE public.t_user_wallet_rpc_config
(
    record_id  public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    chain      character varying(1024)                               NOT NULL,
    rpc_url    character varying(1024)                               NOT NULL,
    wss_url    character varying(1024)                               NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_chain_config;
CREATE TABLE public.t_chain_config
(
    chain      character varying(64)   NOT NULL,
    rpc_url    character varying(1024) NOT NULL,
    wss_url    character varying(1024) NOT NULL,
    decimals   integer                 NOT NULL,
    symbol     character varying(64)   NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_token_config;
CREATE TABLE public.t_token_config
(
    name       character varying(64) NOT NULL,
    symbol     character varying(64) NOT NULL,
    decimals   integer               NOT NULL,
    mint       character varying(64) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_fee_tolerance;
CREATE TABLE public.t_fee_tolerance
(
    max_less_rate numeric(5, 2) NOT NULL,
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_fee_statistics;
CREATE TABLE public.t_fee_statistics
(
    record_id  public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    slot       bigint,
    tx_index   integer,
    block_hash character varying(64),
    tx_id      character varying(128),
    price          numeric(78, 0),
    unit_limit     numeric(78, 0),
    units_consumed numeric(78, 0),
    fee        numeric(78, 0),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.t_fee_statistics ADD CONSTRAINT fee_slot_tx_index UNIQUE (slot, tx_index);
ALTER TABLE ONLY public.t_fee_statistics ADD CONSTRAINT fee_tx_id UNIQUE (tx_id);

DROP TABLE IF EXISTS public.t_qn_fee;
CREATE TABLE public.t_qn_fee
(
    id         integer NOT NULL,
    slot       bigint,
    low_avg    numeric(78, 0),
    medium_avg numeric(78, 0),
    high_avg   numeric(78, 0),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.t_qn_fee
    ADD CONSTRAINT t_sol_qn_fee_pkey PRIMARY KEY (id);

DROP TABLE IF EXISTS public.t_native_account_info;
CREATE TABLE public.t_native_account_info
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL,
    token_account  character varying(64)                                 NOT NULL,
    invite_code    character varying(16)                                 NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_invite_relation;
CREATE TABLE public.t_invite_relation
(
    record_id  public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    inviter    character varying(64)                                 NOT NULL,
    invitee    character varying(64)                                 NOT NULL,
    channel    character varying(64)                                 NOT NULL,
    level      integer                     DEFAULT 1                 NOT NULL,
    tx_id      character varying(128)                                NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_service_info;
CREATE TABLE public.t_service_info
(
    service     character varying(64)                    NOT NULL,
    sub_service character varying(64)                    NOT NULL,
    address     character varying(64)                    NOT NULL,
    webhook     character varying(1024)     DEFAULT NULL,
    mq_group    character varying(64)       DEFAULT NULL,
    mq_topic    character varying(64)       DEFAULT NULL,
    hook_type   integer                     DEFAULT 0    NOT NULL,
    confirm     bool                        DEFAULT true NOT NULL,
    multi_sign  bool                        DEFAULT true NOT NULL,
    created_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 业务私钥
DROP TABLE IF EXISTS public.t_service_key;
CREATE TABLE public.t_service_key
(
    service       character varying(64)   NOT NULL,
    sub_service   character varying(64)   NOT NULL,
    encrypted_key character varying(1024) NOT NULL,
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_tx_scan_info;
CREATE TABLE public.t_tx_scan_info
(
    service        character varying(64)                 NOT NULL,
    sub_service    character varying(64)                 NOT NULL,
    native_account character varying(64)                 NOT NULL,
    pda_account    character varying(64)                 NOT NULL,
    until_tx_id    character varying(128)                NOT NULL,
    before_tx_id   character varying(128),
    slot           numeric(78, 0)              DEFAULT 0 NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_service_tx;
CREATE TABLE public.t_service_tx
(
    record_id       ULID      DEFAULT gen_ulid() NOT NULL PRIMARY KEY,
    service         VARCHAR(64)                  NOT NULL,
    sub_service     VARCHAR(64)                  NOT NULL,
    tx_id           VARCHAR(128)                 NOT NULL,
    state           INTEGER   DEFAULT 0          NOT NULL,
    retry_count     INTEGER   DEFAULT 0          NOT NULL,
    next_retry_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    max_retries     INTEGER   DEFAULT 5          NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_sol_fund_flow
-- 需要导入并且搞分表
CREATE TABLE public.t_fund_flow
(
    record_id    public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    brand        character varying(64)                                 NOT NULL,
    token_symbol character varying(64)                                 NOT NULL,
    is_token     boolean                                               NOT NULL,
    from_account character varying(64)                                 NOT NULL,
    to_account   character varying(64)                                 NOT NULL,
    tx_id        character varying(128)                                NOT NULL,
    direction    smallint                                              NOT NULL,
    service_type smallint                                              NOT NULL,
    flow_type    smallint                                              NOT NULL,
    decimals     smallint                                              NOT NULL,
    amount       numeric(78, 0)                                        NOT NULL,
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_fund_flow_tx_to_flow ON public.t_fund_flow (tx_id, to_account, flow_type);

DROP TABLE IF EXISTS public.t_hacker_account;
CREATE TABLE public.t_hacker_account
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS public.t_exclude_account;
CREATE TABLE public.t_exclude_account
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    native_account character varying(64)                                 NOT NULL,
    remark         character varying(128),
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
