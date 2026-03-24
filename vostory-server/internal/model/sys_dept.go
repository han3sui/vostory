package model

// SysDept 部门表
type SysDept struct {
	BaseModel
	DeptID    uint   `json:"dept_id" gorm:"primaryKey;autoIncrement;comment:部门id"`
	ParentID  uint   `json:"parent_id" gorm:"default:0;comment:父部门id"`
	Ancestors string `json:"ancestors" gorm:"size:50;default:'';comment:祖级列表"`
	DeptName  string `json:"dept_name" gorm:"size:100;default:'';comment:部门名称"`
	OrderNum  int    `json:"order_num" gorm:"default:0;comment:显示顺序"`
	LeaderID  *uint  `json:"leader_id" gorm:"comment:负责人ID"`
	Leader    string `json:"leader" gorm:"size:20;comment:负责人姓名"`
	Phone     string `json:"phone" gorm:"size:11;comment:联系电话"`
	Email     string `json:"email" gorm:"size:50;comment:邮箱"`
	Status    string `json:"status" gorm:"size:1;default:'0';comment:部门状态（0正常 1停用）"`
	Remark    string `json:"remark" gorm:"size:500;comment:备注"`

	// 关联关系
	LeaderUser *SysUser `json:"leader_user" gorm:"foreignKey:LeaderID;references:UserID"`
}

func (SysDept) TableName() string {
	return "sys_dept"
}
