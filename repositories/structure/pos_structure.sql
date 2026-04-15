-- POS奖励发放规则
-- 旧工程 t_sol_pos_reward_rule
DROP TABLE IF EXISTS public.t_pos_reward_config;
CREATE TABLE public.t_pos_reward_config
(
    record_id          ulid           NOT NULL     DEFAULT gen_ulid(), -- 记录Id
    reward_account     VARCHAR(64)    NOT NULL,                        -- 奖励token的native account
    cost_account       VARCHAR(64)    NOT NULL,                        -- 收取dex手续费的native account
    cost_fee_rate      NUMERIC(78, 0) NOT NULL,                        -- 领取时支付给dex的手续费费率
    max_cost_fee       NUMERIC(78, 0) NOT NULL,                        -- 领取时最多支付给dex的费用
    quote_token_amount NUMERIC(78, 0) NOT NULL,                        -- 领取时最多支付给dex的费用
    created_at         TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- POS 星级用户评定规则
-- 旧工程 t_sol_pos_star_level_rule
DROP TABLE IF EXISTS public.t_pos_star_level_rule;
CREATE TABLE public.t_pos_star_level_rule
(
    record_id    ulid           NOT NULL     DEFAULT gen_ulid(), -- 记录Id
    amount       NUMERIC(78, 2) NOT NULL,                        -- 持币金额
    group_amount NUMERIC(78, 2) NOT NULL,                        -- 团队持币数量
    star_level   INT            NOT NULL,                        -- 星级数量
    rate         NUMERIC(5, 2)  NOT NULL,                        -- 费率
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

-- POS 固定星级用户配置
-- 旧工程 t_sol_pos_star_level_config
DROP TABLE IF EXISTS public.t_pos_star_whitelist;
CREATE TABLE public.t_pos_star_whitelist
(
    record_id      ulid        NOT NULL        DEFAULT gen_ulid(), -- 记录Id
    native_account VARCHAR(64) NOT NULL,                           -- 地址
    star_level     INT         NOT NULL,                           -- 持币地址的星级
    rate           NUMERIC(5, 2),                                  -- 星级用户的极差费率
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
    record_id   ulid NOT NULL               DEFAULT gen_ulid(), -- 记录Id
    reward_type INT  NOT NULL,                                  -- 奖励类型 0.固定收益 1.加推特奖励 2.推特转贴奖励 3.推特点赞或回复
    starred     BOOL NOT NULL,                                  -- 任务奖励费率
    rate        NUMERIC(5, 2),                                  -- 任务奖励费率
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_pos_mission_config (reward_type);

-- POS 每日快照表
-- 旧工程 t_sol_pos_snap_shot
DROP TABLE IF EXISTS public.t_pos_snap_shot;
CREATE TABLE public.t_pos_snap_shot
(
    record_id      ulid        NOT NULL        DEFAULT gen_ulid(), -- 记录Id
    native_account VARCHAR(64) NOT NULL,                           -- native account 地址
    amount         NUMERIC(78, 0),                                 -- 持币地址的金额
    star_level     INT                         DEFAULT 0,          -- 持币地址的星级
    rate           NUMERIC(5, 2),                                  -- 极差费率
    range_base     NUMERIC(5, 2),                                  -- 极差基数
    snap_day       DATE        NOT NULL,                           -- 快照的日期，旧工程字段 Day
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE UNIQUE INDEX ON public.t_pos_snap_shot (native_account, snap_day);

-- POS 奖励明细表
-- 旧工程 t_sol_pos_reward
DROP TABLE IF EXISTS public.t_pos_reward;
CREATE TABLE public.t_pos_reward
(
    record_id      ulid        NOT NULL        DEFAULT gen_ulid(), -- 记录Id
    group_id       VARCHAR(64) NOT NULL,                           -- 根地址
    native_account VARCHAR(64) NOT NULL,                           -- 持币地址
    star_level     INT                         DEFAULT 0,          -- 持币地址的星级
    base           NUMERIC(78, 0),                                 -- pos持币的Base(StarLevel=0时base=持币金额,StarLevel=1时Base=(持币金额+极差))
    rate           NUMERIC(5, 2),                                  -- 奖励费率
    reward_amount  NUMERIC(78, 0),                                 -- 奖励金额
    reward_type    INT,                                            -- 奖励类型 1.固定收益 2.加推特奖励 3.推特转贴奖励 4.推特点赞或回复
    reward_state   INT                         DEFAULT 0,          -- 状态,-1.过期 0.初始化 1.已经领取
    starred        BOOL        NOT NULL,                           -- 是否是星级用户奖励
    pending        BOOL                        DEFAULT false,      -- 是否正在被处理
    snap_day       DATE        NOT NULL,                           -- 快照的日期
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON public.t_pos_reward (group_id);
CREATE INDEX ON public.t_pos_reward (native_account, reward_state, pending);
CREATE UNIQUE INDEX ON public.t_pos_reward (native_account, snap_day, reward_type, starred);

-- POS 奖励领取表
-- 旧工程 t_sol_pos_reward_claim_record
DROP TABLE IF EXISTS public.t_pos_reward_claim;
CREATE TABLE public.t_pos_reward_claim
(
    record_id  ulid   NOT NULL             DEFAULT gen_ulid(), -- 记录Id
    reward_ids ulid[] NOT NULL,                                -- 奖励Id
    tx_id      VARCHAR(128),                                   -- 交易Id
    tx_state   INT                         DEFAULT 0,          -- 状态,-2.过期 -1.失败 0.初始化 1.成功
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);
CREATE INDEX ON "public"."t_pos_reward_claim" (tx_id);
CREATE INDEX ON "public"."t_pos_reward_claim" (tx_state, created_at);

