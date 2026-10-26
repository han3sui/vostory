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
	ID               uint64    `json:"id"`
	ProjectID        uint64    `json:"project_id"`
	Name             string    `json:"name"`
	Aliases          []string  `json:"aliases"`
	Gender           string    `json:"gender"`
	Description      string    `json:"description"`
	Level            string    `json:"level"`
	VoiceProfileID   *uint64   `json:"voice_profile_id"`
	VoiceProfileName string    `json:"voice_profile_name,omitempty"`
	SortOrder        int       `json:"sort_order"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// VsCharacterOptionResponse 角色选项响应（下拉选择用）
type VsCharacterOptionResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// VsCharacterMergeRequest 合并角色请求：将 source_ids 合并到 target_id
type VsCharacterMergeRequest struct {
	TargetID  uint64   `json:"target_id" binding:"required"`        // 合并目标角色
	SourceIDs []uint64 `json:"source_ids" binding:"required,min=1"` // 被合并的角色
}

// VsCharacterMergeResponse 合并角色响应
type VsCharacterMergeResponse struct {
	TargetID         uint64 `json:"target_id"`
	MergedCount      int    `json:"merged_count"`      // 被合并的角色数
	AffectedSegments int64  `json:"affected_segments"` // 迁移到目标角色的片段数
	NeedRegenerate   int64  `json:"need_regenerate"`   // 已生成音频但音色发生变化、建议重新生成的片段数
}

// VsCharacterMergeGenericRequest 泛称角色归入路人池请求
type VsCharacterMergeGenericRequest struct {
	DryRun bool `json:"dry_run"` // 仅预览，不实际合并
}

// VsCharacterMergeGenericGroup 单个路人池的归并明细
type VsCharacterMergeGenericGroup struct {
	PoolName    string   `json:"pool_name"`
	PoolID      uint64   `json:"pool_id,omitempty"`
	SourceNames []string `json:"source_names"`
}

// VsCharacterMergeGenericResponse 泛称角色归入路人池响应
type VsCharacterMergeGenericResponse struct {
	DryRun           bool                           `json:"dry_run"`
	MergedCount      int                            `json:"merged_count"`
	AffectedSegments int64                          `json:"affected_segments"`
	NeedRegenerate   int64                          `json:"need_regenerate"`
	Groups           []VsCharacterMergeGenericGroup `json:"groups"`
	SkippedNames     []string                       `json:"skipped_names,omitempty"` // 有片段正在生成而跳过的角色
}

// VsCharacterMergeSuggestRequest 规则扫描合并请求
type VsCharacterMergeSuggestRequest struct {
	DryRun bool                           `json:"dry_run"` // 仅预览
	Groups []VsCharacterMergeSuggestGroup `json:"groups"`  // 确认合并时传入勾选的分组；为空则按扫描结果全部合并
}

// VsCharacterMergeSuggestGroup 一条建议：多个源角色合并到同一目标或路人池
type VsCharacterMergeSuggestGroup struct {
	Rule        string   `json:"rule"`                   // paren_alias / unnamed_to_named / unnamed_pool / numbered_pool / generic_pool
	Reason      string   `json:"reason"`                 // 给用户看的原因
	TargetKind  string   `json:"target_kind"`            // character | pool
	TargetID    uint64   `json:"target_id,omitempty"`    // 已有目标角色
	TargetName  string   `json:"target_name"`            // 目标名称（路人池名或角色名）
	SourceIDs   []uint64 `json:"source_ids"`             // 被合并角色
	SourceNames []string `json:"source_names,omitempty"` // 预览用
}

// VsCharacterMergeSuggestResponse 规则扫描合并响应
type VsCharacterMergeSuggestResponse struct {
	DryRun           bool                           `json:"dry_run"`
	MergedCount      int                            `json:"merged_count"`
	AffectedSegments int64                          `json:"affected_segments"`
	NeedRegenerate   int64                          `json:"need_regenerate"`
	Groups           []VsCharacterMergeSuggestGroup `json:"groups"`
	SkippedNames     []string                       `json:"skipped_names,omitempty"`
}
