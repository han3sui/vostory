package model

import "time"

// SysUserOnline 在线用户记录
type SysUserOnline struct {
	BaseModel
	SessionID      string     `json:"sessionId" gorm:"primaryKey;size:50;default:'';comment:用户会话id"`
	LoginName      string     `json:"login_name" gorm:"size:50;default:'';comment:登录账号"`
	DeptName       string     `json:"dept_name" gorm:"size:50;default:'';comment:部门名称"`
	IPAddr         string     `json:"ipaddr" gorm:"size:128;default:'';comment:登录IP地址"`
	LoginLocation  string     `json:"login_location" gorm:"size:255;default:'';comment:登录地点"`
	Browser        string     `json:"browser" gorm:"size:50;default:'';comment:浏览器类型"`
	OS             string     `json:"os" gorm:"size:50;default:'';comment:操作系统"`
	Status         string     `json:"status" gorm:"size:10;default:'';comment:在线状态on_line在线off_line离线"`
	StartTimestamp *time.Time `json:"start_timestamp" gorm:"comment:session创建时间"`
	LastAccessTime *time.Time `json:"last_access_time" gorm:"comment:session最后访问时间"`
	ExpireTime     int        `json:"expire_time" gorm:"default:0;comment:超时时间，单位为分钟"`
}

func (SysUserOnline) TableName() string {
	return "sys_user_online"
}
