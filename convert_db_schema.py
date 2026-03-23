#!/usr/bin/env python3
import re
import sys

def camel_to_snake(name):
    """将驼峰命名转换为下划线命名"""
    # 处理特殊情况：首字母大写的情况
    if name.startswith('"'):
        name = name[1:-1]  # 去掉双引号
    
    # 将驼峰转换为下划线
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    result = re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()
    
    # 处理特殊情况：数字
    result = re.sub(r'([a-z])(\d)', r'\1_\2', result)
    result = re.sub(r'(\d)([a-z])', r'\1_\2', result)
    
    return result

def convert_table_name(table_name):
    """转换表名"""
    # 去掉双引号
    if table_name.startswith('"') and table_name.endswith('"'):
        table_name = table_name[1:-1]
    
    # 将 t_sol 开头的表名改为 t_ 开头
    if table_name.startswith('t_sol_'):
        table_name = 't_' + table_name[6:]
    
    # 转换为小写下划线
    return camel_to_snake(table_name)

def convert_column_name(column_name):
    """转换列名"""
    # 去掉双引号
    if column_name.startswith('"') and column_name.endswith('"'):
        column_name = column_name[1:-1]
    
    # 转换为小写下划线
    return camel_to_snake(column_name)

def convert_sql_line(line):
    """转换SQL语句中的一行"""
    # 跳过注释和空行
    if line.strip().startswith('--') or not line.strip():
        return line
    
    # 处理CREATE TABLE语句
    create_table_match = re.match(r'^\s*CREATE\s+TABLE\s+(public\.)?("[^"]+"|\w+)', line, re.IGNORECASE)
    if create_table_match:
        table_name = create_table_match.group(2)
        new_table_name = convert_table_name(table_name)
        line = line.replace(table_name, new_table_name)
    
    # 处理列定义
    # 查找双引号包围的列名
    column_pattern = r'"([^"]+)"'
    columns = re.findall(column_pattern, line)
    for column in columns:
        new_column = convert_column_name(column)
        line = line.replace(f'"{column}"', new_column)
    
    # 处理约束名
    constraint_pattern = r'CONSTRAINT\s+"([^"]+)"'
    constraints = re.findall(constraint_pattern, line, re.IGNORECASE)
    for constraint in constraints:
        new_constraint = convert_column_name(constraint)
        line = line.replace(f'"{constraint}"', new_constraint)
    
    # 处理索引名
    index_pattern = r'INDEX\s+"([^"]+)"'
    indexes = re.findall(index_pattern, line, re.IGNORECASE)
    for index in indexes:
        new_index = convert_column_name(index)
        line = line.replace(f'"{index}"', new_index)
    
    # 处理外键名
    fk_pattern = r'FOREIGN\s+KEY\s+"([^"]+)"'
    fks = re.findall(fk_pattern, line, re.IGNORECASE)
    for fk in fks:
        new_fk = convert_column_name(fk)
        line = line.replace(f'"{fk}"', new_fk)
    
    # 处理主键名
    pk_pattern = r'PRIMARY\s+KEY\s+"([^"]+)"'
    pks = re.findall(pk_pattern, line, re.IGNORECASE)
    for pk in pks:
        new_pk = convert_column_name(pk)
        line = line.replace(f'"{pk}"', new_pk)
    
    # 处理唯一约束名
    unique_pattern = r'UNIQUE\s+"([^"]+)"'
    uniques = re.findall(unique_pattern, line, re.IGNORECASE)
    for unique in uniques:
        new_unique = convert_column_name(unique)
        line = line.replace(f'"{unique}"', new_unique)
    
    # 处理表注释中的表名
    comment_table_pattern = r'COMMENT\s+ON\s+TABLE\s+(public\.)?("[^"]+"|\w+)'
    comment_table_match = re.search(comment_table_pattern, line, re.IGNORECASE)
    if comment_table_match:
        table_name = comment_table_match.group(2)
        if table_name.startswith('"'):
            new_table_name = convert_table_name(table_name)
            line = line.replace(table_name, new_table_name)
    
    # 处理列注释中的列名
    comment_column_pattern = r'COMMENT\s+ON\s+COLUMN\s+(public\.)?("[^"]+"|\w+)\.("[^"]+"|\w+)'
    comment_column_match = re.search(comment_column_pattern, line, re.IGNORECASE)
    if comment_column_match:
        table_name = comment_column_match.group(2)
        column_name = comment_column_match.group(3)
        
        if table_name.startswith('"'):
            new_table_name = convert_table_name(table_name)
            line = line.replace(table_name, new_table_name)
        
        if column_name.startswith('"'):
            new_column_name = convert_column_name(column_name)
            line = line.replace(column_name, new_column_name)
    
    # 处理ALTER TABLE语句中的表名
    alter_table_pattern = r'ALTER\s+TABLE\s+(public\.)?("[^"]+"|\w+)'
    alter_table_match = re.search(alter_table_pattern, line, re.IGNORECASE)
    if alter_table_match:
        table_name = alter_table_match.group(2)
        if table_name.startswith('"'):
            new_table_name = convert_table_name(table_name)
            line = line.replace(table_name, new_table_name)
    
    # 处理ALTER TABLE OWNER语句
    owner_pattern = r'ALTER\s+TABLE\s+(public\.)?("[^"]+"|\w+)\s+OWNER\s+TO'
    owner_match = re.search(owner_pattern, line, re.IGNORECASE)
    if owner_match:
        table_name = owner_match.group(2)
        if table_name.startswith('"'):
            new_table_name = convert_table_name(table_name)
            line = line.replace(table_name, new_table_name)
        elif table_name.startswith('t_sol_'):
            # 处理没有双引号的t_sol表名
            new_table_name = convert_table_name(table_name)
            line = line.replace(table_name, new_table_name)
    
    # 处理SEQUENCE OWNED BY语句
    sequence_pattern = r'ALTER\s+SEQUENCE\s+("[^"]+"|\w+)\s+OWNED\s+BY\s+(public\.)?("[^"]+"|\w+)\.("[^"]+"|\w+)'
    sequence_match = re.search(sequence_pattern, line, re.IGNORECASE)
    if sequence_match:
        table_name = sequence_match.group(3)
        column_name = sequence_match.group(4)
        
        if table_name.startswith('"'):
            new_table_name = convert_table_name(table_name)
            line = line.replace(table_name, new_table_name)
        
        if column_name.startswith('"'):
            new_column_name = convert_column_name(column_name)
            line = line.replace(column_name, new_column_name)
    
    return line

def main():
    input_file = '/Users/mac/go/oshit-go/repositories/meme_db_schema.sql'
    output_file = '/Users/mac/go/oshit-go/repositories/oshit_db.sql'
    
    print(f"正在转换 {input_file} -> {output_file}")
    
    with open(input_file, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    converted_lines = []
    for i, line in enumerate(lines):
        try:
            converted_line = convert_sql_line(line)
            converted_lines.append(converted_line)
        except Exception as e:
            print(f"第 {i+1} 行转换出错: {e}")
            print(f"原始内容: {line}")
            converted_lines.append(line)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        f.writelines(converted_lines)
    
    print(f"转换完成！输出文件: {output_file}")
    
    # 生成迁移脚本
    generate_migration_script()

def generate_migration_script():
    """生成数据迁移脚本"""
    migration_file = '/Users/mac/go/oshit-go/migrate/meme_to_oshit.sql'
    
    print(f"正在生成迁移脚本: {migration_file}")
    
    # 读取原始文件以获取表名
    with open('/Users/mac/go/oshit-go/repositories/meme_db_schema.sql', 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 提取所有表名
    table_pattern = r'CREATE\s+TABLE\s+(public\.)?("[^"]+"|\w+)'
    tables = re.findall(table_pattern, content, re.IGNORECASE)
    
    migration_sql = []
    migration_sql.append("-- 数据迁移脚本：从 meme_db 迁移到 oshit_db")
    migration_sql.append("-- 注意：请在执行前备份数据！")
    migration_sql.append("-- 执行步骤：")
    migration_sql.append("-- 1. 创建新的数据库结构（使用 oshit_db.sql）")
    migration_sql.append("-- 2. 运行本迁移脚本将数据从旧表复制到新表")
    migration_sql.append("-- 3. 验证数据完整性")
    migration_sql.append("-- 4. 删除旧表（可选）")
    migration_sql.append("")
    
    for table_match in tables:
        old_table_name = table_match[1]
        
        # 转换表名
        if old_table_name.startswith('"'):
            old_table_clean = old_table_name[1:-1]
        else:
            old_table_clean = old_table_name
        
        new_table_name = convert_table_name(old_table_name)
        
        # 如果表名没有变化，跳过（避免自引用）
        if old_table_clean == new_table_name:
            continue
        
        # 获取列信息
        # 查找表定义
        table_def_pattern = rf'CREATE\s+TABLE\s+(public\.)?{re.escape(old_table_name)}\s*\((.*?)\);'
        table_def_match = re.search(table_def_pattern, content, re.IGNORECASE | re.DOTALL)
        
        if table_def_match:
            table_body = table_def_match.group(2)
            
            # 提取列名（去掉双引号）
            column_pattern = r'"([^"]+)"\s+[^,]+(?:,|$)'
            columns = re.findall(column_pattern, table_body)
            
            if columns:
                # 构建INSERT语句
                old_columns = [f'"{col}"' for col in columns]
                new_columns = [convert_column_name(col) for col in columns]
                
                insert_sql = f"INSERT INTO {new_table_name} ({', '.join(new_columns)})\n"
                insert_sql += f"SELECT {', '.join(old_columns)}\n"
                insert_sql += f"FROM {old_table_clean};\n"
                
                migration_sql.append(insert_sql)
                migration_sql.append("")
    
    with open(migration_file, 'w', encoding='utf-8') as f:
        f.write('\n'.join(migration_sql))
    
    print(f"迁移脚本生成完成！输出文件: {migration_file}")

if __name__ == '__main__':
    main()