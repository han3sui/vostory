package v1

import "time"

// VsLLMProviderListQuery LLM提供商列表查询参数
type VsLLMProviderListQuery struct {
	*BasePageQuery
	Name         string `json:"name"`          // 显示名称
	ProviderType string `json:"provider_type"` // 提供商类型
	Status       string `json:"status"`        // 状态（0正常 1停用）
}

// VsLLMProviderCreateRequest 创建LLM提供商请求
type VsLLMProviderCreateRequest struct {
	Name         string                 `json:"name" binding:"required"`          // 显示名称
	ProviderType string                 `json:"provider_type" binding:"required"` // 提供商类型（openai/deepseek/anthropic/gemini/ollama/azure/aliyun/custom）
	APIBaseURL   string                 `json:"api_base_url" binding:"required"`  // API地址
	APIKey       string                 `json:"api_key"`                          // API密钥
	ModelList    []string               `json:"model_list"`                       // 可用模型列表
	DefaultModel string                 `json:"default_model"`                    // 默认模型
	CustomParams   map[string]interface{} `json:"custom_params"`                    // 自定义参数
	MaxConcurrency int                    `json:"max_concurrency"`                  // 最大并发数
	SortOrder      int                    `json:"sort_order"`                       // 排序
	Status         string                 `json:"status" binding:"required"`        // 状态（0正常 1停用）
}

// VsLLMProviderUpdateRequest 更新LLM提供商请求
type VsLLMProviderUpdateRequest struct {
	VsLLMProviderCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsLLMProviderDetailResponse LLM提供商详情响应
type VsLLMProviderDetailResponse struct {
	ID             uint64                 `json:"id"`
	Name           string                 `json:"name"`
	ProviderType   string                 `json:"provider_type"`
	APIBaseURL     string                 `json:"api_base_url"`
	APIKey         string                 `json:"api_key"`
	ModelList      []string               `json:"model_list"`
	DefaultModel   string                 `json:"default_model"`
	CustomParams   map[string]interface{} `json:"custom_params"`
	MaxConcurrency int                    `json:"max_concurrency"`
	SortOrder      int                    `json:"sort_order"`
	Status         string                 `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// VsLLMProviderOptionResponse LLM提供商选项响应（下拉选择用）
type VsLLMProviderOptionResponse struct {
	ID           uint64   `json:"id"`
	Name         string   `json:"name"`
	ProviderType string   `json:"provider_type"`
	ModelList    []string `json:"model_list"`
	DefaultModel string   `json:"default_model"`
}

// VsLLMProviderTestRequest 连通性测试请求
type VsLLMProviderTestRequest struct {
	ProviderType string                 `json:"provider_type" binding:"required"`
	APIBaseURL   string                 `json:"api_base_url" binding:"required"`
	APIKey       string                 `json:"api_key"`
	Model        string                 `json:"model"`
	CustomParams map[string]interface{} `json:"custom_params"`
}

// VsLLMProviderTestResponse 连通性测试响应
type VsLLMProviderTestResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	Models   []string `json:"models,omitempty"`
	Duration int64    `json:"duration"` // 耗时（毫秒）
}
