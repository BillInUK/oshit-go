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
    record_id             public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    invite_code           character varying(16)       DEFAULT NULL::character varying,
    decimals              integer                                               NOT NULL,
    token_mint_account    character varying(64)                                 NOT NULL,
    reward_token_account  character varying(64)                                 NOT NULL,
    reward_native_account character varying(64)                                 NOT NULL,
    dex_native_account    character varying(64)                                 NOT NULL,
    amount                numeric(78, 0)                                        NOT NULL,
    invite_amount         numeric(78, 0)                                        NOT NULL,
    dex_fee_rate          numeric(78, 0)                                        NOT NULL,
    max_dex_fee           numeric(78, 0)                                        NOT NULL,
    interval              integer                                               NOT NULL,
    is_default            boolean                     DEFAULT false,
    reward_inviter        boolean                     DEFAULT true,
    invited               boolean                     DEFAULT true,
    created_at            timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at            timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_sol_official_give_token_record
DROP TABLE IF EXISTS public.t_take_token_record;
CREATE TABLE public.t_take_token_record
(
    record_id              public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    brand                  character varying(64)                                 NOT NULL,
    token_symbol           character varying(64)                                 NOT NULL,
    token_mint_account     character varying(64)                                 NOT NULL,
    reward_token_account   character varying(64)                                 NOT NULL,
    reward_native_account  character varying(64)                                 NOT NULL,
    receipt_token_account  character varying(64)                                 NOT NULL,
    receipt_native_account character varying(64)                                 NOT NULL,
    dex_native_account     character varying(64)                                 NOT NULL,
    reward_tx_id           character varying(128)                                NOT NULL,
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
    native_account  character varying(64)                                 NOT NULL,
    take_shit_date  date                                                  NOT NULL,
    take_shit_count integer                     DEFAULT 0                 NOT NULL,
    need_lottery    boolean                     DEFAULT false             NOT NULL,
    last_take_time  timestamp without time zone,
    total_lottery   numeric(78, 0)              DEFAULT 0                 NOT NULL,
    total_take      numeric(78, 0)              DEFAULT 0                 NOT NULL,
    lottery_count   integer                     DEFAULT 0                 NOT NULL,
    created_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at      timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

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
    token_mint_account    character varying(64) NOT NULL,
    decimal               integer               NOT NULL,
    reward_token_account  character varying(64) NOT NULL,
    reward_native_account character varying(64) NOT NULL,
    dex_native_account    character varying(64) NOT NULL,
    reward_rate           numeric(78, 0)        NOT NULL,
    max_valid_reward      numeric(78, 0)        NOT NULL,
    valid_rate            NUMERIC(10, 6)        NOT NULL,
    created_at            timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at            timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_sol_transfer_checked_record
DROP TABLE IF EXISTS public.t_give_token_record;
CREATE TABLE public.t_give_token_record
(
    record_id              public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    from_token_account     character varying(64)                                 NOT NULL,
    from_native_account    character varying(64)                                 NOT NULL,
    receipt_token_account  character varying(64)                                 NOT NULL,
    receipt_native_account character varying(64)                                 NOT NULL,
    tx_id                  character varying(128)                                NOT NULL,
    amount                 numeric(78, 0)                                        NOT NULL,
    state                  integer,
    created_at             timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at             timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

-- 旧工程 t_reward_key_config
DROP TABLE IF EXISTS public.t_reward_key_config;
CREATE TABLE public.t_reward_key_config
(
    service       character varying(64)   NOT NULL,
    encrypted_key character varying(1024) NOT NULL,
    created_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);










