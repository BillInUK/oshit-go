package main

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const PostgreSQLDSN = "postgres://postgres:postgres@localhost:5432/oshit_db"

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

// 运行方法: go run gen.go ${service}，例如: go run gen.go base
func main() {
	db := connectDB(PostgreSQLDSN)
	db = db.Debug()

	g := gen.NewGenerator(gen.Config{
		OutPath:      "../common/pkg/dal/query",
		ModelPkgPath: "../common/pkg/dal/model",
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	// 自定义 JSON 标签：将 snake_case 列名转为 camelCase
	g.WithJSONTagNameStrategy(func(columnName string) string {
		parts := strings.Split(columnName, "_")
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > 0 {
				parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
		return strings.Join(parts, "")
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
		return tableName
	})

	g.UseDB(db)

	tables := g.GenerateAllTable()
	g.ApplyBasic(tables...)
	g.Execute()
}

// 运行方法: go run gen.go ${service}，例如: go run gen.go base
//func main() {
//	if len(os.Args) < 2 {
//		fmt.Println("Usage: go run gen.go <service>")
//		fmt.Println("Service options: public, account, order, etc.")
//		os.Exit(1)
//	}
//
//	service := os.Args[1]
//	if service == "" {
//		panic("service name cannot be empty")
//	}
//
//	db := connectDB(PostgreSQLDSN)
//	db = db.Debug()
//
//	// 如果是base模块，则只生成public里面的表
//	dbSchema := "public"
//	dalPath := "../app/dal"
//	if service != "base" {
//		dbSchema = service
//		dalPath = fmt.Sprintf("../app/%s/dal", service)
//	}
//
//	// 设置当前 schema
//	db.Exec("set search_path to " + dbSchema)
//
//	g := gen.NewGenerator(gen.Config{
//		OutPath:      fmt.Sprintf("%s/query", dalPath),
//		ModelPkgPath: fmt.Sprintf("%s/model", dalPath),
//		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
//	})
//
//	// 自定义 JSON 标签（首字母小写）
//	g.WithJSONTagNameStrategy(func(columnName string) string {
//		return strings.ToLower(columnName[:1]) + columnName[1:]
//	})
//
//	// 自定义模型名称：去除表名的 "t_" 前缀，并将下划线命名转为驼峰
//	g.WithModelNameStrategy(func(tableName string) string {
//		// 去掉 "t_" 前缀
//		name := strings.TrimPrefix(tableName, "t_")
//		// 下划线转驼峰
//		parts := strings.Split(name, "_")
//		for i, p := range parts {
//			parts[i] = strings.Title(p)
//		}
//		return strings.Join(parts, "")
//	})
//
//	// 自定义表名：添加 schema 前缀
//	g.WithTableNameStrategy(func(tableName string) string {
//		return dbSchema + "." + tableName
//	})
//
//	g.UseDB(db)
//
//	tables := g.GenerateAllTable()
//	g.ApplyBasic(tables...)
//	g.Execute()
//}
