package model

import "time"

// SysOperLog 操作日志记录
type SysOperLog struct {
	BaseModel
	OperID        uint      `json:"oper_id" gorm:"primaryKey;autoIncrement;comment:日志主键"`
	Title         string    `json:"title" gorm:"size:50;default:'';comment:模块标题"`
	BusinessType  int       `json:"business_type" gorm:"default:0;comment:业务类型（0其它 1新增 2修改 3删除）"`
	Method        string    `json:"method" gorm:"size:200;default:'';comment:方法名称"`
	RequestMethod string    `json:"request_method" gorm:"size:10;default:'';comment:请求方式"`
	OperatorType  int       `json:"operator_type" gorm:"default:0;comment:操作类别（0其它 1后台用户 2手机端用户）"`
	OperName      string    `json:"oper_name" gorm:"size:50;default:'';comment:操作人员"`
	DeptName      string    `json:"dept_name" gorm:"size:50;default:'';comment:部门名称"`
	OperURL       string    `json:"oper_url" gorm:"size:255;default:'';comment:请求URL"`
	OperIP        string    `json:"oper_ip" gorm:"size:128;default:'';comment:主机地址"`
	OperLocation  string    `json:"oper_location" gorm:"size:255;default:'';comment:操作地点"`
	OperParam     string    `json:"oper_param" gorm:"size:2000;default:'';comment:请求参数"`
	JSONResult    string    `json:"json_result" gorm:"size:2000;default:'';comment:返回参数"`
	Status        int       `json:"status" gorm:"default:0;comment:操作状态（0正常 1异常）"`
	ErrorMsg      string    `json:"error_msg" gorm:"size:2000;default:'';comment:错误消息"`
	OperTime      time.Time `json:"oper_time" gorm:"comment:操作时间"`
	CostTime      int64     `json:"cost_time" gorm:"default:0;comment:消耗时间"`
}

func (SysOperLog) TableName() string {
	return "sys_oper_log"
}
