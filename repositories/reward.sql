-- 开启显式事务，保证所有操作原子性（失败则回滚，成功则提交）
BEGIN;

-- 创建 schema（不存在则创建）
CREATE SCHEMA IF NOT EXISTS exchange;

-- 新建 exchange.t_trade_order 表（订单信息表）- 关联用户表外键
CREATE TABLE IF NOT EXISTS exchange.t_trade_order
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT             NOT NULL,
    order_no   VARCHAR(64) UNIQUE NOT NULL,
    amount     DECIMAL(10, 2)     NOT NULL,
    status     SMALLINT                 DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- 外键关联account schema的用户表（PostgreSQL跨Schema外键支持）
    CONSTRAINT fk_order_user_id FOREIGN KEY (user_id) REFERENCES account.t_user_info (id)
);

-- 为 exchange.t_trade_order 添加 表注释 + 字段注释
COMMENT ON TABLE exchange.t_trade_order IS '订单信息表';
COMMENT ON COLUMN exchange.t_trade_order.id IS '自增主键';
COMMENT ON COLUMN exchange.t_trade_order.user_id IS '关联account.t_user_info的id';
COMMENT ON COLUMN exchange.t_trade_order.order_no IS '订单编号（唯一）';
COMMENT ON COLUMN exchange.t_trade_order.amount IS '订单金额（保留2位小数）';
COMMENT ON COLUMN exchange.t_trade_order.status IS '订单状态：0-待支付 1-已支付 2-已取消 3-已完成';
COMMENT ON COLUMN exchange.t_trade_order.created_at IS '创建时间';
COMMENT ON COLUMN exchange.t_trade_order.updated_at IS '更新时间';
COMMENT ON CONSTRAINT fk_order_user_id ON "exchange".t_trade_order IS '外键关联用户表id';

-- 插入 "exchange".t_trade_order 测试数据
INSERT INTO "exchange".t_trade_order (user_id, order_no, amount, status)
VALUES
    (1, 'ORDER20240501001', 99.99, 0),  -- 张三-待支付
    (1, 'ORDER20240501002', 199.99, 1), -- 张三-已支付
    (2, 'ORDER20240501003', 299.99, 2), -- 李四-已取消
    (3, 'ORDER20240501004', 399.99, 3); -- 王五（禁用）-已完成

-- 提交事务：所有操作执行成功后，持久化到数据库
COMMIT;
