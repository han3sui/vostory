package model

import "time"

// SysLogininfor 系统访问记录
type SysLogininfor struct {
	InfoID        uint      `json:"info_id" gorm:"primaryKey;autoIncrement;comment:访问ID"`
	LoginName     string    `json:"login_name" gorm:"size:50;default:'';comment:登录账号"`
	IPAddr        string    `json:"ipaddr" gorm:"size:128;default:'';comment:登录IP地址"`
	LoginLocation string    `json:"login_location" gorm:"size:255;default:'';comment:登录地点"`
	Browser       string    `json:"browser" gorm:"size:50;default:'';comment:浏览器类型"`
	OS            string    `json:"os" gorm:"size:50;default:'';comment:操作系统"`
	Status        string    `json:"status" gorm:"size:1;default:'0';comment:登录状态（0成功 1失败）"`
	Msg           string    `json:"msg" gorm:"size:255;default:'';comment:提示消息"`
	LoginTime     time.Time `json:"login_time" gorm:"comment:访问时间"`
}

func (SysLogininfor) TableName() string {
	return "sys_logininfor"
}
