package model

import "time"

// SysUser 用户信息表
type SysUser struct {
	BaseModel
	UserID        uint       `json:"user_id" gorm:"primaryKey;autoIncrement;comment:用户ID"`
	UserDeptID    *uint      `json:"dept_id" gorm:"column:dept_id;comment:部门ID"`
	SuperiorID    *uint      `json:"superior_id" gorm:"comment:直属上级ID"`
	LoginName     string     `json:"login_name" gorm:"size:30;not null;comment:登录账号"`
	UserName      string     `json:"user_name" gorm:"size:100;default:'';comment:用户昵称"`
	UserType      string     `json:"user_type" gorm:"size:2;default:'00';comment:用户类型（00系统用户 01注册用户 99超级管理员）"`
	Email         string     `json:"email" gorm:"size:50;default:'';comment:用户邮箱"`
	Phonenumber   string     `json:"phonenumber" gorm:"size:11;default:'';comment:手机号码"`
	Sex           string     `json:"sex" gorm:"size:1;default:'0';comment:用户性别（0男 1女 2未知）"`
	Avatar        string     `json:"avatar" gorm:"size:100;default:'';comment:头像路径"`
	Password      string     `json:"password" gorm:"size:500;default:'';comment:密码"`
	Status        string     `json:"status" gorm:"size:1;default:'0';comment:账号状态（0正常 1停用）"`
	LoginIP       string     `json:"login_ip" gorm:"size:128;default:'';comment:最后登录IP"`
	LoginDate     *time.Time `json:"login_date" gorm:"comment:最后登录时间"`
	PwdUpdateDate *time.Time `json:"pwd_update_date" gorm:"comment:密码最后更新时间"`
	Remark        string     `json:"remark" gorm:"size:500;comment:备注"`

	// 关联关系
	Dept     *SysDept  `json:"dept" gorm:"foreignKey:UserDeptID;references:DeptID"`
	Superior *SysUser  `json:"superior" gorm:"foreignKey:SuperiorID;references:UserID"`
	Roles    []SysRole `json:"roles" gorm:"many2many:sys_user_role;foreignKey:UserID;joinForeignKey:UserID;References:RoleID;joinReferences:RoleID"`
	Posts    []SysPost `json:"posts" gorm:"many2many:sys_user_post;foreignKey:UserID;joinForeignKey:UserID;References:PostID;joinReferences:PostID"`
}

func (SysUser) TableName() string {
	return "sys_user"
}
