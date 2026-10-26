package v1

import "time"

// VsLLMModelListQuery 模型列表查询参数
type VsLLMModelListQuery struct {
	*BasePageQuery
	ProviderID uint64 `json:"provider_id"` // 所属厂家ID
	ModelName  string `json:"model_name"`  // 模型标识
	Status     string `json:"status"`      // 状态（0正常 1停用）
}

// VsLLMModelCreateRequest 创建模型请求
type VsLLMModelCreateRequest struct {
	ProviderID uint64 `json:"provider_id" binding:"required"` // 所属厂家ID
	ModelName  string `json:"model_name" binding:"required"`  // 模型标识
	SortOrder  int    `json:"sort_order"`                    // 排序
	Status     string `json:"status" binding:"required"`     // 状态（0正常 1停用）
	Remark     string `json:"remark"`                        // 备注
}

// VsLLMModelUpdateRequest 更新模型请求
type VsLLMModelUpdateRequest struct {
	VsLLMModelCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsLLMModelDeleteRequest 批量删除模型请求
type VsLLMModelDeleteRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// VsLLMModelDetailResponse 模型详情响应
type VsLLMModelDetailResponse struct {
	ID           uint64    `json:"id"`
	ProviderID   uint64    `json:"provider_id"`
	ProviderName string    `json:"provider_name"` // 冗余：厂家名称
	ModelName    string    `json:"model_name"`
	SortOrder    int       `json:"sort_order"`
	Status       string    `json:"status"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
