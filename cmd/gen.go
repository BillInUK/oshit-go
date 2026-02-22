package main

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const PostgreSQLDSN = "postgres://postgres:postgres@localhost:5432/my_db"

func connectDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
		NamingStrategy: schema.NamingStrategy{
			// 移除 TablePrefix，避免影响 schema
			NoLowerCase: true, // 保留，使字段名与数据库列名一致
		},
	})
	if err != nil {
		panic(fmt.Errorf("connect db fail: %w", err))
	}
	return db
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run gen.go <service>")
		fmt.Println("Service options: public, account, order, etc.")
		os.Exit(1)
	}

	service := os.Args[1]
	if service == "" {
		panic("service name cannot be empty")
	}

	db := connectDB(PostgreSQLDSN)
	db = db.Debug()
	db.Exec("set search_path to " + service) // 设置当前 schema

	g := gen.NewGenerator(gen.Config{
		OutPath:      fmt.Sprintf("../app/%s/dal/query", service),
		ModelPkgPath: fmt.Sprintf("../app/%s/dal/model", service),
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	// 自定义 JSON 标签（首字母小写）
	g.WithJSONTagNameStrategy(func(columnName string) string {
		return strings.ToLower(columnName[:1]) + columnName[1:]
	})

	// 自定义模型名称：去除表名的 "t_" 前缀，并将下划线命名转为驼峰
	g.WithModelNameStrategy(func(tableName string) string {
		// 去掉 "t_" 前缀
		name := strings.TrimPrefix(tableName, "t_")
		// 下划线转驼峰
		parts := strings.Split(name, "_")
		for i, p := range parts {
			parts[i] = strings.Title(p)
		}
		return strings.Join(parts, "")
	})

	// 自定义表名：添加 schema 前缀
	g.WithTableNameStrategy(func(tableName string) string {
		return service + "." + tableName
	})

	g.UseDB(db)

	tables := g.GenerateAllTable()
	g.ApplyBasic(tables...)
	g.Execute()
}
