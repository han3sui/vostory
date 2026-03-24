package model

// SysRoleMenu 角色和菜单关联表
type SysRoleMenu struct {
	RoleID uint `json:"role_id" gorm:"primaryKey;comment:角色ID"`
	MenuID uint `json:"menu_id" gorm:"primaryKey;comment:菜单ID"`
}

func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}
