package model

type SysApi struct {
	BaseModel
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:API ID"`
	Name        string `json:"name" gorm:"size:50;not null;comment:API名称"`
	Path        string `json:"path" gorm:"size:200;not null;comment:API路径"`
	Method      string `json:"method" gorm:"size:10;not null;comment:API方法"`
	Description string `json:"description" gorm:"size:200;not null;comment:API描述"`
	Perms       string `json:"perms" gorm:"size:100;comment:权限标识"`
	Tag         string `json:"tag" gorm:"size:100;comment:API标签"`
}

func (SysApi) TableName() string {
	return "sys_api"
}
