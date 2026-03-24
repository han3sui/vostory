package model

// SysRole 角色信息表
type SysRole struct {
	BaseModel
	RoleID    uint   `json:"role_id" gorm:"primaryKey;autoIncrement;comment:角色ID"`
	RoleName  string `json:"role_name" gorm:"size:30;not null;comment:角色名称"`
	RoleKey   string `json:"role_key" gorm:"size:100;not null;comment:角色权限字符串"`
	RoleSort  int    `json:"role_sort" gorm:"not null;comment:显示顺序"`
	DataScope string `json:"data_scope" gorm:"size:1;default:'1';comment:数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限 5：仅本人数据权限）"`
	Status    string `json:"status" gorm:"size:1;not null;comment:角色状态（0正常 1停用）"`
	Remark    string `json:"remark" gorm:"size:500;comment:备注"`
}

func (SysRole) TableName() string {
	return "sys_role"
}
