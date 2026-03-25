package v1

import "time"

// VsCharacterListQuery 角色列表查询参数
type VsCharacterListQuery struct {
	*BasePageQuery
	ProjectID uint64 `json:"project_id"` // 所属项目
	Name      string `json:"name"`       // 角色名称
	Gender    string `json:"gender"`     // 性别
	Level     string `json:"level"`      // 角色层级
	Status    string `json:"status"`     // 状态
}

// VsCharacterCreateRequest 创建角色请求
type VsCharacterCreateRequest struct {
	ProjectID      uint64   `json:"project_id" binding:"required"` // 所属项目
	Name           string   `json:"name" binding:"required"`       // 角色名称
	Aliases        []string `json:"aliases"`                       // 别名列表
	Gender         string   `json:"gender"`                        // 性别
	Description    string   `json:"description"`                   // 角色描述
	Level          string   `json:"level"`                         // 角色层级
	VoiceProfileID *uint64  `json:"voice_profile_id"`              // 绑定声音配置
	SortOrder      int      `json:"sort_order"`                    // 排序
	Status         string   `json:"status" binding:"required"`     // 状态
}

// VsCharacterUpdateRequest 更新角色请求
type VsCharacterUpdateRequest struct {
	VsCharacterCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsCharacterDetailResponse 角色详情响应
type VsCharacterDetailResponse struct {
	ID             uint64    `json:"id"`
	ProjectID      uint64    `json:"project_id"`
	Name           string    `json:"name"`
	Aliases        []string  `json:"aliases"`
	Gender         string    `json:"gender"`
	Description    string    `json:"description"`
	Level          string    `json:"level"`
	VoiceProfileID *uint64   `json:"voice_profile_id"`
	SortOrder      int       `json:"sort_order"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// VsCharacterOptionResponse 角色选项响应（下拉选择用）
type VsCharacterOptionResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
