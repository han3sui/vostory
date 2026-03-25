package v1

import "time"

// VsPromptTemplateListQuery Prompt模板列表查询参数
type VsPromptTemplateListQuery struct {
	*BasePageQuery
	Name         string `json:"name"`          // 模板名称
	TemplateType string `json:"template_type"` // 模板类型
	IsSystem     string `json:"is_system"`     // 是否系统内置（1是 0否）
	Status       string `json:"status"`        // 状态（0正常 1停用）
}

// VsPromptTemplateCreateRequest 创建Prompt模板请求
type VsPromptTemplateCreateRequest struct {
	Name         string `json:"name" binding:"required"`          // 模板名称
	TemplateType string `json:"template_type" binding:"required"` // 模板类型（character_extract/dialogue_parse/emotion_tag/scene_split/text_correct）
	Content      string `json:"content" binding:"required"`       // Prompt内容
	Description  string `json:"description"`                      // 模板描述
	SortOrder    int    `json:"sort_order"`                       // 排序
	Status       string `json:"status" binding:"required"`        // 状态（0正常 1停用）
}

// VsPromptTemplateUpdateRequest 更新Prompt模板请求
type VsPromptTemplateUpdateRequest struct {
	VsPromptTemplateCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsPromptTemplateDetailResponse Prompt模板详情响应
type VsPromptTemplateDetailResponse struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	TemplateType string    `json:"template_type"`
	Content      string    `json:"content"`
	Description  string    `json:"description"`
	IsSystem     string    `json:"is_system"`
	Version      int       `json:"version"`
	SortOrder    int       `json:"sort_order"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// VsPromptTemplateOptionResponse Prompt模板选项响应（下拉选择用）
type VsPromptTemplateOptionResponse struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	TemplateType string `json:"template_type"`
}
