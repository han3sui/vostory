package v1

import "time"

// VsLLMProviderListQuery 厂家列表查询参数
type VsLLMProviderListQuery struct {
	*BasePageQuery
	Name   string `json:"name"`   // 厂家名称
	Status string `json:"status"` // 状态（0正常 1停用）
}

// VsLLMProviderCreateRequest 创建厂家请求
type VsLLMProviderCreateRequest struct {
	Name           string `json:"name" binding:"required"`          // 厂家名称
	APIBaseURL     string `json:"api_base_url" binding:"required"`  // API地址
	APIKey         string `json:"api_key"`                          // API密钥
	MaxConcurrency int    `json:"max_concurrency"`                  // 最大并发数
	SortOrder      int    `json:"sort_order"`                       // 排序
	Status         string `json:"status" binding:"required"`        // 状态（0正常 1停用）
	Remark         string `json:"remark"`                           // 备注
}

// VsLLMProviderUpdateRequest 更新厂家请求
type VsLLMProviderUpdateRequest struct {
	VsLLMProviderCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsLLMProviderDeleteRequest 批量删除厂家请求
type VsLLMProviderDeleteRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// VsLLMProviderDetailResponse 厂家详情响应
type VsLLMProviderDetailResponse struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	APIBaseURL     string    `json:"api_base_url"`
	APIKey         string    `json:"api_key"` // 脱敏返回
	MaxConcurrency int       `json:"max_concurrency"`
	SortOrder      int       `json:"sort_order"`
	Status         string    `json:"status"`
	Remark         string    `json:"remark"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// VsLLMProviderTestRequest 连通性测试请求
type VsLLMProviderTestRequest struct {
	ID         uint64 `json:"id"`                                  // 已保存厂家 ID（api_key 脱敏时据此回查）
	APIBaseURL string `json:"api_base_url" binding:"required"`     // API地址
	APIKey     string `json:"api_key"`                             // API密钥
	Model      string `json:"model" binding:"required"`            // 用于测试的模型标识
}

// VsLLMProviderTestResponse 连通性测试响应
type VsLLMProviderTestResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Duration int64  `json:"duration"` // 耗时（毫秒）
}
