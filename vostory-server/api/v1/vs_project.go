package v1

import "time"

// VsProjectListQuery 项目列表查询参数
type VsProjectListQuery struct {
	*BasePageQuery
	WorkspaceID uint64 `json:"workspace_id"` // 所属工作空间
	Name        string `json:"name"`         // 项目名称
	Status      string `json:"status"`       // 项目状态
	SourceType  string `json:"source_type"`  // 导入来源
}

// VsProjectCreateRequest 创建项目请求
type VsProjectCreateRequest struct {
	WorkspaceID       uint64            `json:"workspace_id" binding:"required"` // 所属工作空间
	Name              string            `json:"name" binding:"required"`         // 项目名称
	Description       string            `json:"description"`                     // 项目描述
	CoverURL          string            `json:"cover_url"`                       // 封面图
	LLMProviderID     *uint64           `json:"llm_provider_id"`                 // LLM提供商ID
	TTSProviderID     *uint64           `json:"tts_provider_id"`                 // TTS提供商ID
	PromptTemplateIDs map[string]uint64 `json:"prompt_template_ids"`             // Prompt模板ID映射
	Remark            string            `json:"remark"`                          // 备注
}

// VsProjectUpdateRequest 更新项目请求
type VsProjectUpdateRequest struct {
	ID                uint64            `json:"id" binding:"required"`
	Name              string            `json:"name" binding:"required"` // 项目名称
	Description       string            `json:"description"`             // 项目描述
	CoverURL          string            `json:"cover_url"`               // 封面图
	LLMProviderID     *uint64           `json:"llm_provider_id"`         // LLM提供商ID
	TTSProviderID     *uint64           `json:"tts_provider_id"`         // TTS提供商ID
	PromptTemplateIDs map[string]uint64 `json:"prompt_template_ids"`     // Prompt模板ID映射
	Remark            string            `json:"remark"`                  // 备注
}

// VsProjectDetailResponse 项目详情响应
type VsProjectDetailResponse struct {
	ID                uint64            `json:"id"`
	WorkspaceID       uint64            `json:"workspace_id"`
	WorkspaceName     string            `json:"workspace_name"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	CoverURL          string            `json:"cover_url"`
	SourceType        string            `json:"source_type"`
	SourceFileURL     string            `json:"source_file_url"`
	Status            string            `json:"status"`
	LLMProviderID     *uint64           `json:"llm_provider_id"`
	LLMProviderName   string            `json:"llm_provider_name"`
	TTSProviderID     *uint64           `json:"tts_provider_id"`
	TTSProviderName   string            `json:"tts_provider_name"`
	PromptTemplateIDs map[string]uint64 `json:"prompt_template_ids"`
	TotalChapters     int               `json:"total_chapters"`
	TotalCharacters   int               `json:"total_characters"`
	Remark            string            `json:"remark"`
	CreatedBy         string            `json:"created_by"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// VsProjectOptionResponse 项目选项响应（下拉选择用）
type VsProjectOptionResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
