# 数据库迁移指南

## 概述

此迁移将数据库从旧的命名约定（双引号驼峰命名）转换为新的命名约定（小写下划线命名），并统一表名格式（将 `t_sol_` 开头的表名改为 `t_` 开头）。

## 生成的文件

1. **`/repositories/oshit_db.sql`** - 新的数据库结构文件
   - 所有字段名从驼峰命名转换为小写下划线命名（如 `RecordId` → `record_id`）
   - 所有 `t_sol_` 开头的表名改为 `t_` 开头（如 `t_sol_airdrop_record` → `t_airdrop_record`）
   - 移除了字段名周围的双引号

2. **`/migrate/meme_to_oshit.sql`** - 数据迁移脚本
   - 将数据从旧表复制到新表
   - 处理了字段名和表名的映射关系

3. **`/convert_db_schema.py`** - 转换脚本（可重复使用）
   - 用于生成上述文件的Python脚本
   - 可以处理其他类似的迁移需求

## 迁移步骤

### 步骤1：备份现有数据库
```bash
# 备份当前数据库
pg_dump -U username -d meme_db -f meme_db_backup.sql
```

### 步骤2：创建新的数据库结构
```bash
# 创建新的数据库
createdb -U username oshit_db

# 应用新的数据库结构
psql -U username -d oshit_db -f repositories/oshit_db.sql
```

### 步骤3：迁移数据
```bash
# 在新的数据库中运行迁移脚本
# 注意：需要确保旧数据库（meme_db）和新数据库（oshit_db）都在运行
# 或者将旧数据导入到新数据库后再运行迁移

# 方法1：如果旧数据库还在运行
psql -U username -d oshit_db -f migrate/meme_to_oshit.sql

# 方法2：如果已经将旧数据导入到新数据库
# 首先导入旧数据库的备份
psql -U username -d oshit_db -f meme_db_backup.sql
# 然后运行迁移脚本
psql -U username -d oshit_db -f migrate/meme_to_oshit.sql
```

### 步骤4：验证数据完整性
```bash
# 检查表数量
psql -U username -d oshit_db -c "\dt"

# 检查一些关键表的数据量
psql -U username -d oshit_db -c "SELECT COUNT(*) FROM t_airdrop_record;"
psql -U username -d oshit_db -c "SELECT COUNT(*) FROM t_fund_flow;"
```

### 步骤5：清理（可选）
```bash
# 删除旧表（如果不再需要）
# 注意：迁移脚本不会自动删除旧表，需要手动清理
```

## 主要变化

### 表名变化示例
- `t_sol_airdrop_record` → `t_airdrop_record`
- `t_sol_fund_flow` → `t_fund_flow`
- `t_sol_game_buy_property_record` → `t_game_buy_property_record`

### 字段名变化示例
- `"RecordId"` → `record_id`
- `"Brand"` → `brand`
- `"TokenSymbol"` → `token_symbol`
- `"CreateTime"` → `create_time`
- `"UpdateTime"` → `update_time`

### 不变的表
- 不以 `t_sol_` 开头的表名保持不变（如 `audit_comment_template`, `user_basic` 等）
- 已经是小写下划线命名的字段保持不变

## 注意事项

1. **备份重要**：在执行任何迁移操作前，请务必备份现有数据库。

2. **测试环境**：建议先在测试环境中执行完整的迁移流程。

3. **应用程序更新**：迁移完成后，需要更新应用程序代码中的数据库查询：
   - 更新GORM模型中的字段标签
   - 更新SQL查询中的表名和字段名
   - 更新数据库连接配置（如果需要切换到新数据库）

4. **依赖关系**：迁移脚本假设新旧表在同一个数据库中。如果需要在不同数据库间迁移，需要调整连接参数。

5. **性能考虑**：对于大型数据库，迁移可能需要较长时间。建议在低峰期执行。

## 故障排除

### 常见问题

1. **权限错误**：确保数据库用户有足够的权限创建表和执行插入操作。

2. **重复表名**：如果新旧表名相同（非t_sol开头的表），迁移脚本会跳过这些表。

3. **字段类型不匹配**：如果原始数据中有类型不匹配的问题，迁移可能会失败。

### 恢复步骤

如果迁移失败，可以按以下步骤恢复：
1. 删除新数据库：`dropdb oshit_db`
2. 从备份恢复：`psql -U username -d meme_db -f meme_db_backup.sql`
3. 分析错误原因并修复转换脚本
4. 重新执行迁移流程

## 支持

如有问题，请参考生成的SQL文件或运行转换脚本时查看详细的转换日志。