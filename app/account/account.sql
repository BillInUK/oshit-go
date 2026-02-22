-- 开启显式事务，保证所有操作原子性（失败则回滚，成功则提交）
BEGIN;

-- 1. 创建Schema（不存在则创建）
CREATE SCHEMA IF NOT EXISTS account;

-- 用户信息表
CREATE TABLE IF NOT EXISTS account.t_user_info
(
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL,
    phone      VARCHAR(20) UNIQUE,
    email      VARCHAR(100) UNIQUE,
    password   VARCHAR(100) NOT NULL,
    status     SMALLINT                 DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE account.t_user_info IS '用户信息表';
COMMENT ON COLUMN account.t_user_info.id IS '自增主键';
COMMENT ON COLUMN account.t_user_info.username IS '用户名';
COMMENT ON COLUMN account.t_user_info.phone IS '手机号（唯一）';
COMMENT ON COLUMN account.t_user_info.email IS '邮箱（唯一）';
COMMENT ON COLUMN account.t_user_info.password IS '密码（演示用，实际需加密）';
COMMENT ON COLUMN account.t_user_info.status IS '状态：1-正常 0-禁用';
COMMENT ON COLUMN account.t_user_info.created_at IS '创建时间';
COMMENT ON COLUMN account.t_user_info.updated_at IS '更新时间';

-- 插入account.t_user_info测试数据
INSERT INTO account.t_user_info (username, phone, email, password, status)
VALUES
    ('zhangsan', '13800138000', 'zhangsan@example.com', 'demo123456', 1),
    ('lisi', '13800138001', 'lisi@example.com', 'demo123456', 1),
    ('wangwu', '13800138002', 'wangwu@example.com', 'demo123456', 0);

-- 提交事务：所有操作执行成功后，持久化到数据库
COMMIT;
