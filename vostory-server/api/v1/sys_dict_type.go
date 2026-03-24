package v1

import "time"

// SysDictTypeListQuery 字典类型列表查询参数
type SysDictTypeListQuery struct {
	*BasePageQuery
	DictName string `json:"dict_name"` // 字典名称
	DictType string `json:"dict_type"` // 字典类型
	Status   string `json:"status"`    // 状态（0正常 1停用）
}

// SysDictTypeCreateRequest 创建字典类型请求
type SysDictTypeCreateRequest struct {
	DictName string `json:"dict_name" binding:"required"` // 字典名称
	DictType string `json:"dict_type" binding:"required"` // 字典类型
	Status   string `json:"status" binding:"required"`    // 状态（0正常 1停用）
	Remark   string `json:"remark"`                       // 备注
}

// SysDictTypeUpdateRequest 更新字典类型请求
type SysDictTypeUpdateRequest struct {
	SysDictTypeCreateRequest
	ID uint `json:"id" binding:"required"`
}

// SysDictTypeDetailResponse 字典类型详情响应
type SysDictTypeDetailResponse struct {
	ID        uint      `json:"id"`
	DictName  string    `json:"dict_name"`
	DictType  string    `json:"dict_type"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
