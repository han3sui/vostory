package v1

import "time"

// VsFileImportResponse 文件上传响应
type VsFileImportResponse struct {
	ProjectID     uint64 `json:"project_id"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	SourceType    string `json:"source_type"`
	SourceFileURL string `json:"source_file_url"`
}

// VsLLMLogListQuery LLM调用日志列表查询参数
type VsLLMLogListQuery struct {
	*BasePageQuery
	ProjectID  uint64 `json:"project_id"`  // 关联项目
	ProviderID uint64 `json:"provider_id"` // LLM提供商
	ModelName  string `json:"model_name"`  // 模型名称
	Status     int    `json:"status"`      // 状态（0成功 1失败, -1全部）
}

// VsLLMLogDetailResponse LLM调用日志详情响应
type VsLLMLogDetailResponse struct {
	ID                 uint64    `json:"id"`
	ProjectID          *uint64   `json:"project_id"`
	ProjectName        string    `json:"project_name"`
	ProviderID         uint64    `json:"provider_id"`
	ProviderName       string    `json:"provider_name"`
	TemplateID         *uint64   `json:"template_id"`
	TemplateName       string    `json:"template_name"`
	ModelName          string    `json:"model_name"`
	InputTokens        int       `json:"input_tokens"`
	OutputTokens       int       `json:"output_tokens"`
	InputSummary       string    `json:"input_summary"`
	OutputSummary      string    `json:"output_summary"`
	CostTime           int64     `json:"cost_time"`
	Status             int       `json:"status"`
	ErrorMessage       string    `json:"error_message"`
	CreatedAt          time.Time `json:"created_at"`
}
