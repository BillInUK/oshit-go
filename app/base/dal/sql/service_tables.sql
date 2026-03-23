-- 服务信息表
DROP TABLE IF EXISTS public.t_service_info;
CREATE TABLE public.t_service_info
(
    service    character varying(64)                 NOT NULL,
    address    character varying(64)                 NOT NULL,
    webhook    character varying(1024)     DEFAULT NULL,
    mq_group   character varying(64)       DEFAULT NULL,
    mq_topic   character varying(64)       DEFAULT NULL,
    hook_type  integer                     DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (service)
);

COMMENT ON TABLE public.t_service_info IS '服务注册信息表';
COMMENT ON COLUMN public.t_service_info.service IS '业务服务名称';
COMMENT ON COLUMN public.t_service_info.address IS '需要扫描的Solana地址';
COMMENT ON COLUMN public.t_service_info.webhook IS 'Webhook回调URL';
COMMENT ON COLUMN public.t_service_info.mq_group IS 'Kafka消费者组（可选）';
COMMENT ON COLUMN public.t_service_info.mq_topic IS 'Kafka主题';
COMMENT ON COLUMN public.t_service_info.hook_type IS '通知类型: 0=Kafka, 1=Webhook';

-- 服务交易表（带死信队列功能）
DROP TABLE IF EXISTS public.t_service_tx;
CREATE TABLE public.t_service_tx
(
    record_id      public.ulid                 DEFAULT public.gen_ulid() NOT NULL,
    service        character varying(64)                                 NOT NULL,
    tx_id          character varying(128)                                NOT NULL,
    state          integer                     DEFAULT 0                 NOT NULL,
    retry_count    integer                     DEFAULT 0                 NOT NULL,
    next_retry_time timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    max_retries    integer                     DEFAULT 5                 NOT NULL,
    created_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (record_id)
);

COMMENT ON TABLE public.t_service_tx IS '服务交易表（带死信队列功能）';
COMMENT ON COLUMN public.t_service_tx.record_id IS '记录ID（ULID）';
COMMENT ON COLUMN public.t_service_tx.service IS '业务服务名称';
COMMENT ON COLUMN public.t_service_tx.tx_id IS '交易ID';
COMMENT ON COLUMN public.t_service_tx.state IS '交易状态: 0=初始化, <0=失败, >0=成功';
COMMENT ON COLUMN public.t_service_tx.retry_count IS '重试次数';
COMMENT ON COLUMN public.t_service_tx.next_retry_time IS '下次重试时间';
COMMENT ON COLUMN public.t_service_tx.max_retries IS '最大重试次数';

-- 创建索引
CREATE INDEX idx_service_tx_service ON public.t_service_tx (service);
CREATE INDEX idx_service_tx_tx_id ON public.t_service_tx (tx_id);
CREATE INDEX idx_service_tx_state ON public.t_service_tx (state);
CREATE INDEX idx_service_tx_next_retry_time ON public.t_service_tx (next_retry_time);
CREATE INDEX idx_service_tx_created_at ON public.t_service_tx (created_at);

-- 复合索引，用于优化查询性能
CREATE INDEX idx_service_tx_state_next_retry ON public.t_service_tx (state, next_retry_time);
CREATE INDEX idx_service_tx_service_state ON public.t_service_tx (service, state);

-- 添加外键约束（可选）
-- ALTER TABLE public.t_service_tx 
-- ADD CONSTRAINT fk_service_tx_service_info 
-- FOREIGN KEY (service) REFERENCES public.t_service_info(service);