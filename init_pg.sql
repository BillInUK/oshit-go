-- 创建通用触发器函数：实现updated_at字段自动更新（PostgreSQL替代ON UPDATE CURRENT_TIMESTAMP）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;