package model

import "time"

// VsGenerationTask 生成任务
type VsGenerationTask struct {
	BaseModel
	TaskID           uint64     `json:"task_id" gorm:"primaryKey;comment:任务ID"`
	ProjectID        uint64     `json:"project_id" gorm:"not null;index;index:idx_project_type_status;comment:所属项目"`
	ChapterID        *uint64    `json:"chapter_id" gorm:"index;comment:关联章节"`
	TaskType         string     `json:"task_type" gorm:"size:30;not null;index:idx_project_type_status;comment:任务类型（text_parse/character_extract/emotion_tag/tts_generate/audio_merge）"`
	Status           string     `json:"status" gorm:"size:20;default:'pending';index:idx_project_type_status;comment:任务状态（pending/running/completed/failed/cancelled）"`
	Progress         int        `json:"progress" gorm:"default:0;comment:进度百分比（0-100）"`
	TotalBatches     int        `json:"total_batches" gorm:"default:0;comment:总批次数"`
	CompletedBatches int        `json:"completed_batches" gorm:"default:0;comment:已完成批次数"`
	LLMProviderID    *uint64    `json:"llm_provider_id" gorm:"comment:使用的LLM提供商"`
	TTSProviderID    *uint64    `json:"tts_provider_id" gorm:"comment:使用的TTS提供商"`
	PromptTemplateID *uint64    `json:"prompt_template_id" gorm:"comment:使用的Prompt模板"`
	ErrorMessage     string     `json:"error_message" gorm:"size:2000;comment:错误信息"`
	RetryCount       int        `json:"retry_count" gorm:"default:0;comment:已重试次数"`
	MaxRetries       int        `json:"max_retries" gorm:"default:3;comment:最大重试次数"`
	StartedAt        *time.Time `json:"started_at" gorm:"comment:开始执行时间"`
	CompletedAt      *time.Time `json:"completed_at" gorm:"comment:完成时间"`

	Project        *VsProject        `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
	Chapter        *VsChapter        `json:"chapter" gorm:"foreignKey:ChapterID;references:ChapterID"`
	LLMProvider    *VsLLMProvider    `json:"llm_provider" gorm:"foreignKey:LLMProviderID;references:ProviderID"`
	TTSProvider    *VsTTSProvider    `json:"tts_provider" gorm:"foreignKey:TTSProviderID;references:ProviderID"`
	PromptTemplate *VsPromptTemplate `json:"prompt_template" gorm:"foreignKey:PromptTemplateID;references:TemplateID"`
}

func (VsGenerationTask) TableName() string {
	return "vs_generation_task"
}
