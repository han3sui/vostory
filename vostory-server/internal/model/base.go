package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// JSONValue 通用 JSON 序列化方法，兼容 MySQL 和 PostgreSQL
// 返回 string 类型以确保 PostgreSQL 兼容性
func JSONValue(v interface{}) (driver.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// JSONValueWithDefault 通用 JSON 序列化方法，空值返回默认值
// defaultVal 通常为 "{}" 或 "[]"
func JSONValueWithDefault(v interface{}, isEmpty bool, defaultVal string) (driver.Value, error) {
	if isEmpty {
		return defaultVal, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// JSONScan 通用 JSON 反序列化方法，兼容 MySQL ([]byte) 和 PostgreSQL (string)
func JSONScan(value interface{}, dest interface{}) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("type assertion to []byte or string failed")
	}
	return json.Unmarshal(b, dest)
}

// BaseModel 基础模型定义
// @Description 所有数据模型的基础字段
type BaseModel struct {
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;comment:删除时间"`
	CreatedBy string         `json:"created_by" gorm:"size:64;comment:创建者"`
	UpdatedBy string         `json:"updated_by" gorm:"size:64;comment:更新者"`
	DeptID    uint           `json:"dept_id" gorm:"comment:部门ID"`
}

// DeleteTime 软删除时间类型
type DeleteTime sql.NullInt64

// Paginate 分页查询
func Paginate(PageNo int, PageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch {
		case PageSize > 1000:
			PageSize = 1000
		case PageSize <= 0:
			PageSize = 10
		}
		switch {
		case PageNo <= 0:
			PageNo = 1
		}
		offset := (PageNo - 1) * PageSize
		return db.Offset(offset).Limit(PageSize)
	}
}

// Scan implements the Scanner interface.
func (n *DeleteTime) Scan(value interface{}) error {
	return (*sql.NullInt64)(n).Scan(value)
}

// Value implements the driver Valuer interface.
func (n DeleteTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Int64, nil
}

func (DeleteTime) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{
		clause.Where{Exprs: []clause.Expression{
			clause.Eq{
				Column: clause.Column{Table: clause.CurrentTable, Name: f.DBName},
				Value:  nil,
			},
		}},
	}
}

// WithDataScope 返回一个应用数据权限过滤的scope函数
func WithDataScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return WithDataScopeTable(ctx, "")
}

// WithDataScopeTable 返回一个应用数据权限过滤的scope函数，支持指定表名
// tableName 为空时不添加表名前缀，非空时会添加 "tableName." 前缀
func WithDataScopeTable(ctx context.Context, tableName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 从上下文获取权限信息
		dataScope := ctx.Value("data_scope")
		dataScopeDepts := ctx.Value("data_scope_depts")
		loginName := ctx.Value("login_name")

		deptColumn := "dept_id"
		createdByColumn := "created_by"
		if tableName != "" {
			deptColumn = tableName + ".dept_id"
			createdByColumn = tableName + ".created_by"
		}

		//1:全部数据权限
		if dataScope == "1" {
			return db
		}

		//2:自定数据权限
		if dataScope == "2" {
			return db.Where(deptColumn+" IN (?)", dataScopeDepts)
		}

		//3:本部门数据权限
		if dataScope == "3" {
			return db.Where(deptColumn+" IN (?)", dataScopeDepts)
		}

		//4:本部门及以下数据权限
		if dataScope == "4" {
			return db.Where(deptColumn+" IN (?)", dataScopeDepts)
		}

		//5:仅本人数据权限
		if dataScope == "5" {
			return db.Where(createdByColumn+" = ?", loginName)
		}

		return db
	}
}
