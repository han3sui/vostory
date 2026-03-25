package model

// VsLLMLog LLM 调用日志
type VsLLMLog struct {
	BaseModel
	LogID         uint64  `json:"log_id" gorm:"primaryKey;autoIncrement;comment:日志ID"`
	ProjectID     *uint64 `json:"project_id" gorm:"index;comment:关联项目"`
	TaskID        *uint64 `json:"task_id" gorm:"comment:关联任务"`
	ProviderID    uint64  `json:"provider_id" gorm:"comment:使用的LLM提供商"`
	TemplateID    *uint64 `json:"template_id" gorm:"comment:使用的Prompt模板"`
	ModelName     string  `json:"model_name" gorm:"size:100;comment:实际使用的模型"`
	InputTokens   int     `json:"input_tokens" gorm:"default:0;comment:输入token数"`
	OutputTokens  int     `json:"output_tokens" gorm:"default:0;comment:输出token数"`
	InputSummary  string  `json:"input_summary" gorm:"size:2000;comment:输入摘要"`
	OutputSummary string  `json:"output_summary" gorm:"size:2000;comment:输出摘要"`
	CostTime      int64   `json:"cost_time" gorm:"default:0;comment:耗时（毫秒）"`
	Status        int     `json:"status" gorm:"default:0;comment:状态（0成功 1失败）"`
	ErrorMessage  string  `json:"error_message" gorm:"size:2000;comment:错误信息"`

	Project        *VsProject        `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
	Task           *VsGenerationTask `json:"task" gorm:"foreignKey:TaskID;references:TaskID"`
	LLMProvider    *VsLLMProvider    `json:"llm_provider" gorm:"foreignKey:ProviderID;references:ProviderID"`
	PromptTemplate *VsPromptTemplate `json:"prompt_template" gorm:"foreignKey:TemplateID;references:TemplateID"`
}

func (VsLLMLog) TableName() string {
	return "vs_llm_log"
}
