package v1

import "time"

// VsPronunciationDictListQuery 发音词典列表查询参数
type VsPronunciationDictListQuery struct {
	*BasePageQuery
	WorkspaceID uint64 `json:"workspace_id"` // 所属工作空间
	ProjectID   int64  `json:"project_id"`   // 所属项目（0=全局，>0=项目级，-1=全部）
	Word        string `json:"word"`         // 原始词
}

// VsPronunciationDictCreateRequest 创建发音词典请求
type VsPronunciationDictCreateRequest struct {
	WorkspaceID uint64  `json:"workspace_id" binding:"required"` // 所属工作空间
	ProjectID   *uint64 `json:"project_id"`                      // 所属项目（空=全局）
	Word        string  `json:"word" binding:"required"`         // 原始词
	Phoneme     string  `json:"phoneme" binding:"required"`      // 发音标注
	Remark      string  `json:"remark"`                          // 备注
}

// VsPronunciationDictUpdateRequest 更新发音词典请求
type VsPronunciationDictUpdateRequest struct {
	ID      uint64 `json:"id" binding:"required"`
	Word    string `json:"word" binding:"required"`
	Phoneme string `json:"phoneme" binding:"required"`
	Remark  string `json:"remark"`
}

// VsPronunciationDictDetailResponse 发音词典详情响应
type VsPronunciationDictDetailResponse struct {
	ID            uint64    `json:"id"`
	WorkspaceID   uint64    `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	ProjectID     *uint64   `json:"project_id"`
	ProjectName   string    `json:"project_name"`
	Word          string    `json:"word"`
	Phoneme       string    `json:"phoneme"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
