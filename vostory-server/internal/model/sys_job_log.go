package model

// SysJobLog 定时任务调度日志表
type SysJobLog struct {
	BaseModel
	JobLogID      uint   `json:"job_log_id" gorm:"primaryKey;autoIncrement;comment:任务日志ID"`
	JobName       string `json:"job_name" gorm:"size:64;not null;comment:任务名称"`
	JobGroup      string `json:"job_group" gorm:"size:64;not null;comment:任务组名"`
	InvokeTarget  string `json:"invoke_target" gorm:"size:500;not null;comment:调用目标字符串"`
	JobMessage    string `json:"job_message" gorm:"size:500;comment:日志信息"`
	Status        string `json:"status" gorm:"size:1;default:'0';comment:执行状态（0正常 1失败）"`
	ExceptionInfo string `json:"exception_info" gorm:"size:2000;default:'';comment:异常信息"`
}

func (SysJobLog) TableName() string {
	return "sys_job_log"
}
