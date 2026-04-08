package v1

import "time"

// VsTTSProviderListQuery TTS提供商列表查询参数
type VsTTSProviderListQuery struct {
	*BasePageQuery
	Name         string `json:"name"`          // 显示名称
	ProviderType string `json:"provider_type"` // 提供商类型
	Status       string `json:"status"`        // 状态（0正常 1停用）
}

// VsTTSProviderCreateRequest 创建TTS提供商请求
type VsTTSProviderCreateRequest struct {
	Name              string                 `json:"name" binding:"required"`          // 显示名称
	ProviderType      string                 `json:"provider_type" binding:"required"` // 提供商类型（local/online/custom）
	APIBaseURL        string                 `json:"api_base_url" binding:"required"`  // API地址
	APIKey            string                 `json:"api_key"`                          // API密钥
	SupportedFeatures []string               `json:"supported_features"`               // 支持的能力（emotion/clone/multi_speaker等）
	CustomParams      map[string]interface{} `json:"custom_params"`                    // 自定义参数
	MaxConcurrency    int                    `json:"max_concurrency"`                  // 最大并发数
	SortOrder         int                    `json:"sort_order"`                       // 排序
	Status            string                 `json:"status" binding:"required"`        // 状态（0正常 1停用）
}

// VsTTSProviderUpdateRequest 更新TTS提供商请求
type VsTTSProviderUpdateRequest struct {
	VsTTSProviderCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsTTSProviderDetailResponse TTS提供商详情响应
type VsTTSProviderDetailResponse struct {
	ID                uint64                 `json:"id"`
	Name              string                 `json:"name"`
	ProviderType      string                 `json:"provider_type"`
	APIBaseURL        string                 `json:"api_base_url"`
	APIKey            string                 `json:"api_key"`
	SupportedFeatures []string               `json:"supported_features"`
	CustomParams      map[string]interface{} `json:"custom_params"`
	MaxConcurrency    int                    `json:"max_concurrency"`
	SortOrder         int                    `json:"sort_order"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// VsTTSProviderOptionResponse TTS提供商选项响应（下拉选择用）
type VsTTSProviderOptionResponse struct {
	ID                uint64   `json:"id"`
	Name              string   `json:"name"`
	ProviderType      string   `json:"provider_type"`
	SupportedFeatures []string `json:"supported_features"`
}

// VsTTSProviderTestRequest 连通性测试请求
type VsTTSProviderTestRequest struct {
	ProviderType string                 `json:"provider_type" binding:"required"`
	APIBaseURL   string                 `json:"api_base_url" binding:"required"`
	APIKey       string                 `json:"api_key"`
	CustomParams map[string]interface{} `json:"custom_params"`
}

// VsTTSProviderTestResponse 连通性测试响应
type VsTTSProviderTestResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Duration int64  `json:"duration"` // 耗时（毫秒）
}

// VsTTSProviderStatusCPU CPU状态
type VsTTSProviderStatusCPU struct {
	Name          string  `json:"name"`
	Percent       float64 `json:"percent"`
	CountPhysical *int    `json:"count_physical"`
	CountLogical  int     `json:"count_logical"`
}

// VsTTSProviderStatusMemory 内存状态
type VsTTSProviderStatusMemory struct {
	TotalMB int     `json:"total_mb"`
	UsedMB  int     `json:"used_mb"`
	Percent float64 `json:"percent"`
}

// VsTTSProviderStatusGPU GPU状态
type VsTTSProviderStatusGPU struct {
	Name             string  `json:"name"`
	GPUUtilization   *int    `json:"gpu_utilization"`
	MemoryTotalMB    int     `json:"memory_total_mb"`
	MemoryAllocatedMB int   `json:"memory_allocated_mb"`
	MemoryReservedMB int     `json:"memory_reserved_mb"`
}

// VsTTSProviderStatusResponse 系统状态响应
type VsTTSProviderStatusResponse struct {
	CPU    *VsTTSProviderStatusCPU    `json:"cpu"`
	Memory *VsTTSProviderStatusMemory `json:"memory"`
	GPU    *VsTTSProviderStatusGPU    `json:"gpu"`
}
