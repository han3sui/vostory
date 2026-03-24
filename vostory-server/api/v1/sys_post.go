package v1

import "time"

// SysPostListQuery 岗位列表查询参数
type SysPostListQuery struct {
	*BasePageQuery
	PostCode string `json:"post_code"` // 岗位编码
	PostName string `json:"post_name"` // 岗位名称
	Status   string `json:"status"`    // 状态（0正常 1停用）
}

// SysPostCreateRequest 创建岗位请求
type SysPostCreateRequest struct {
	PostCode string `json:"post_code" binding:"required"` // 岗位编码
	PostName string `json:"post_name" binding:"required"` // 岗位名称
	PostSort int    `json:"post_sort""`                   // 显示顺序
	Status   string `json:"status" binding:"required"`    // 状态（0正常 1停用）
	Remark   string `json:"remark"`                       // 备注
}

// SysPostUpdateRequest 更新岗位请求
type SysPostUpdateRequest struct {
	SysPostCreateRequest
	ID uint `json:"id" binding:"required"`
}

// SysPostDetailResponse 岗位详情响应
type SysPostDetailResponse struct {
	ID        uint      `json:"id"`
	PostCode  string    `json:"post_code"`
	PostName  string    `json:"post_name"`
	PostSort  int       `json:"post_sort"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
