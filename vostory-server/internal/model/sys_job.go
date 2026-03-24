package model

// SysJob 定时任务调度表
type SysJob struct {
	BaseModel
	JobID          uint   `json:"job_id" gorm:"primaryKey;autoIncrement;comment:任务ID"`
	JobName        string `json:"job_name" gorm:"size:64;default:'';comment:任务名称"`
	JobGroup       string `json:"job_group" gorm:"size:64;default:'DEFAULT';comment:任务组名"`
	InvokeTarget   string `json:"invoke_target" gorm:"size:500;not null;comment:调用目标字符串"`
	CronExpression string `json:"cron_expression" gorm:"size:255;default:'';comment:cron执行表达式"`
	MisfirePolicy  string `json:"misfire_policy" gorm:"size:20;default:'3';comment:计划执行错误策略（1立即执行 2执行一次 3放弃执行）"`
	Concurrent     string `json:"concurrent" gorm:"size:1;default:'1';comment:是否并发执行（0允许 1禁止）"`
	Status         string `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1暂停）"`
	Remark         string `json:"remark" gorm:"size:500;default:'';comment:备注信息"`
}

func (SysJob) TableName() string {
	return "sys_job"
}
