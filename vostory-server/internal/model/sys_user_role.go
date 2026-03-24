package model

// SysUserRole 用户和角色关联表
type SysUserRole struct {
	UserID uint `json:"user_id" gorm:"primaryKey;comment:用户ID"`
	RoleID uint `json:"role_id" gorm:"primaryKey;comment:角色ID"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}
