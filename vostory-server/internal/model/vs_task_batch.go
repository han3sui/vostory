package model

import "time"

// VsTaskBatch 任务批次
type VsTaskBatch struct {
	BaseModel
	BatchID        uint64     `json:"batch_id" gorm:"primaryKey;comment:批次ID"`
	TaskID         uint64     `json:"task_id" gorm:"not null;index;uniqueIndex:uk_task_batch_index;comment:所属任务"`
	BatchIndex     int        `json:"batch_index" gorm:"not null;uniqueIndex:uk_task_batch_index;comment:批次序号"`
	InputText      string     `json:"input_text" gorm:"type:text;comment:输入文本"`
	InputWordCount int        `json:"input_word_count" gorm:"default:0;comment:输入字数"`
	OutputResult   string     `json:"output_result" gorm:"type:text;comment:输出结果（JSON）"`
	Status         string     `json:"status" gorm:"size:20;default:'pending';comment:批次状态（pending/running/completed/failed）"`
	ErrorMessage   string     `json:"error_message" gorm:"size:2000;comment:错误信息"`
	RetryCount     int        `json:"retry_count" gorm:"default:0;comment:已重试次数"`
	ModelName      string     `json:"model_name" gorm:"size:100;comment:实际使用的模型名"`
	CostTime       int64      `json:"cost_time" gorm:"default:0;comment:耗时（毫秒）"`
	StartedAt      *time.Time `json:"started_at" gorm:"comment:开始时间"`
	CompletedAt    *time.Time `json:"completed_at" gorm:"comment:完成时间"`

	Task *VsGenerationTask `json:"task" gorm:"foreignKey:TaskID;references:TaskID"`
}

func (VsTaskBatch) TableName() string {
	return "vs_task_batch"
}
