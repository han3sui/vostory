package utils

import (
	"database/sql"
	"fmt"
	"iot-alert-center/pkg/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// createMySQLDatabase 创建MySQL数据库
func CreateMySQLDatabase(dsn string, logger *log.Logger) error {
	// 解析DSN获取数据库名
	dbName, err := extractMySQLDbName(dsn)
	if err != nil {
		return fmt.Errorf("解析MySQL DSN失败: %v", err)
	}

	// 构建无数据库名的连接字符串
	connectDSN := removeMySQLDbName(dsn)

	// 连接到MySQL服务器
	db, err := sql.Open("mysql", connectDSN)
	if err != nil {
		return fmt.Errorf("连接MySQL服务器失败: %v", err)
	}
	defer db.Close()

	// 检查数据库是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查MySQL数据库是否存在失败: %v", err)
	}

	if count == 0 {
		// 创建数据库
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
		if err != nil {
			return fmt.Errorf("创建MySQL数据库失败: %v", err)
		}
		logger.Info(fmt.Sprintf("MySQL数据库 '%s' 创建成功", dbName))
	} else {
		logger.Info(fmt.Sprintf("MySQL数据库 '%s' 已存在", dbName))
	}

	return nil
}

// createPostgreSQLDatabase 创建PostgreSQL数据库
func CreatePostgreSQLDatabase(dsn string, logger *log.Logger) error {
	// 解析DSN获取数据库名
	dbName, err := extractPostgreSQLDbName(dsn)
	if err != nil {
		return fmt.Errorf("解析PostgreSQL DSN失败: %v", err)
	}

	// 构建连接到postgres数据库的DSN
	connectDSN := replacePostgreSQLDbName(dsn, "postgres")

	// 连接到PostgreSQL服务器
	db, err := sql.Open("postgres", connectDSN)
	if err != nil {
		return fmt.Errorf("连接PostgreSQL服务器失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL服务器连接测试失败: %v", err)
	}

	// 检查数据库是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pg_database WHERE datname = $1", dbName).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查PostgreSQL数据库是否存在失败: %v", err)
	}

	if count == 0 {
		// 创建数据库
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
		if err != nil {
			return fmt.Errorf("创建PostgreSQL数据库失败: %v", err)
		}
		logger.Info(fmt.Sprintf("PostgreSQL数据库 '%s' 创建成功", dbName))
	} else {
		logger.Info(fmt.Sprintf("PostgreSQL数据库 '%s' 已存在", dbName))
	}

	return nil
}

// createSQLiteDatabase 创建SQLite数据库文件
func CreateSQLiteDatabase(dsn string, logger *log.Logger) error {
	// 移除查询参数，获取文件路径
	dbPath := strings.Split(dsn, "?")[0]

	// 创建目录
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "/" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("创建SQLite目录失败: %v", err)
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// 创建空的SQLite数据库文件
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("创建SQLite数据库文件失败: %v", err)
		}
		file.Close()
		logger.Info(fmt.Sprintf("SQLite数据库文件 '%s' 创建成功", dbPath))
	} else {
		logger.Info(fmt.Sprintf("SQLite数据库文件 '%s' 已存在", dbPath))
	}

	return nil
}

// extractMySQLDbName 从MySQL DSN中提取数据库名
func extractMySQLDbName(dsn string) (string, error) {
	// MySQL DSN格式: user:password@tcp(host:port)/dbname?params
	re := regexp.MustCompile(`/([^?]+)`)
	matches := re.FindStringSubmatch(dsn)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法从DSN中提取数据库名: %s", dsn)
	}
	return matches[1], nil
}

// removeMySQLDbName 从MySQL DSN中移除数据库名
func removeMySQLDbName(dsn string) string {
	// 移除数据库名部分，保留参数
	re := regexp.MustCompile(`(/[^?]+)(\?.*)?$`)
	if re.MatchString(dsn) {
		// 如果有参数，保留参数；否则只保留到/
		matches := re.FindStringSubmatch(dsn)
		if len(matches) > 2 && matches[2] != "" {
			return re.ReplaceAllString(dsn, "/"+matches[2])
		} else {
			return re.ReplaceAllString(dsn, "/")
		}
	}
	return dsn
}

// extractPostgreSQLDbName 从PostgreSQL DSN中提取数据库名
func extractPostgreSQLDbName(dsn string) (string, error) {
	// PostgreSQL DSN格式: host=localhost user=user password=pass dbname=db port=5432
	// 支持多种格式: 关键字=值 格式和 postgres://user:pass@host:port/dbname 格式

	// 首先检查是否是 postgres:// URL 格式
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// 从URL中提取数据库名
		parts := strings.Split(dsn, "/")
		if len(parts) < 4 {
			return "", fmt.Errorf("无效的PostgreSQL URL格式: %s", dsn)
		}
		dbName := parts[len(parts)-1]
		// 移除查询参数
		if idx := strings.Index(dbName, "?"); idx != -1 {
			dbName = dbName[:idx]
		}
		if dbName == "" {
			return "", fmt.Errorf("PostgreSQL URL中未找到数据库名: %s", dsn)
		}
		return dbName, nil
	}

	// 处理关键字=值格式
	re := regexp.MustCompile(`dbname=([^\s]+)`)
	matches := re.FindStringSubmatch(dsn)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法从DSN中提取数据库名: %s", dsn)
	}
	return matches[1], nil
}

// replacePostgreSQLDbName 替换PostgreSQL DSN中的数据库名
func replacePostgreSQLDbName(dsn, newDbName string) string {
	// 首先检查是否是 postgres:// URL 格式
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// 替换URL中的数据库名
		parts := strings.Split(dsn, "/")
		if len(parts) >= 4 {
			// 保留查询参数
			lastPart := parts[len(parts)-1]
			if idx := strings.Index(lastPart, "?"); idx != -1 {
				parts[len(parts)-1] = newDbName + lastPart[idx:]
			} else {
				parts[len(parts)-1] = newDbName
			}
			return strings.Join(parts, "/")
		}
		return dsn
	}

	// 处理关键字=值格式
	re := regexp.MustCompile(`dbname=[^\s]+`)
	return re.ReplaceAllString(dsn, "dbname="+newDbName)
}
