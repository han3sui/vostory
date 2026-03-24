package model

// SysRoleDept 角色和部门关联表
type SysRoleDept struct {
	RoleID uint `json:"role_id" gorm:"primaryKey;comment:角色ID"`
	DeptID uint `json:"dept_id" gorm:"primaryKey;comment:部门ID"`
}

func (SysRoleDept) TableName() string {
	return "sys_role_dept"
}
