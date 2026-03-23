-- 服务信息表（更新版）
DROP TABLE IF EXISTS public.t_service_info;
CREATE TABLE public.t_service_info
(
    service        character varying(64)                 NOT NULL,
    native_account character varying(64)                 NOT NULL,
    pda_account    character varying(64)                 NOT NULL,
    webhook        character varying(1024)     DEFAULT NULL,
    mq_group       character varying(64)       DEFAULT NULL,
    mq_topic       character varying(64)       DEFAULT NULL,
    hook_type      integer                     DEFAULT 0 NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (service)
);

COMMENT ON TABLE public.t_service_info IS '服务注册信息表';
COMMENT ON COLUMN public.t_service_info.service IS '业务服务名称';
COMMENT ON COLUMN public.t_service_info.native_account IS '原生Solana地址';
COMMENT ON COLUMN public.t_service_info.pda_account IS 'PDA地址（通常是token_account）';
COMMENT ON COLUMN public.t_service_info.webhook IS 'Webhook回调URL';
COMMENT ON COLUMN public.t_service_info.mq_group IS 'Kafka消费者组（可选）';
COMMENT ON COLUMN public.t_service_info.mq_topic IS 'Kafka主题';
COMMENT ON COLUMN public.t_service_info.hook_type IS '通知类型: 0=Kafka, 1=Webhook';

-- 交易扫描信息表
DROP TABLE IF EXISTS public.t_tx_scan_info;
CREATE TABLE public.t_tx_scan_info (
    service        character varying(64) NOT NULL,
    native_account character varying(64) NOT NULL,
    pda_account    character varying(64) NOT NULL,
    until_tx_id    character varying(128) NOT NULL,
    before_tx_id   character varying(128),
    slot           numeric(78,0) DEFAULT 0 NOT NULL,
    PRIMARY KEY (service, pda_account)
);

COMMENT ON TABLE public.t_tx_scan_info IS '交易扫描信息表';
COMMENT ON COLUMN public.t_tx_scan_info.service IS '业务服务名称';
COMMENT ON COLUMN public.t_tx_scan_info.native_account IS '原生Solana地址';
COMMENT ON COLUMN public.t_tx_scan_info.pda_account IS 'PDA地址（通常是token_account）';
COMMENT ON COLUMN public.t_tx_scan_info.until_tx_id IS '扫描截止的交易ID';
COMMENT ON COLUMN public.t_tx_scan_info.before_tx_id IS '扫描起始的交易ID（可选）';
COMMENT ON COLUMN public.t_tx_scan_info.slot IS '最后扫描的slot';

-- 创建索引
CREATE INDEX idx_tx_scan_info_service ON public.t_tx_scan_info (service);
CREATE INDEX idx_tx_scan_info_pda_account ON public.t_tx_scan_info (pda_account);
CREATE INDEX idx_tx_scan_info_slot ON public.t_tx_scan_info (slot);