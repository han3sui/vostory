package v1

import "time"

// VsWorkspaceListQuery 工作空间列表查询参数
type VsWorkspaceListQuery struct {
	*BasePageQuery
	Name   string `json:"name"`   // 空间名称
	Status string `json:"status"` // 状态（0正常 1停用）
}

// VsWorkspaceCreateRequest 创建工作空间请求
type VsWorkspaceCreateRequest struct {
	Name        string `json:"name" binding:"required"`   // 空间名称
	Description string `json:"description"`               // 描述
	Status      string `json:"status" binding:"required"` // 状态（0正常 1停用）
}

// VsWorkspaceUpdateRequest 更新工作空间请求
type VsWorkspaceUpdateRequest struct {
	VsWorkspaceCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsWorkspaceDetailResponse 工作空间详情响应
type VsWorkspaceDetailResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VsWorkspaceOptionResponse 工作空间选项响应（下拉选择用）
type VsWorkspaceOptionResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
