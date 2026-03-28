package v1

import "time"

// VsPronunciationDictListQuery 发音词典列表查询参数
type VsPronunciationDictListQuery struct {
	*BasePageQuery
	ProjectID uint64 `json:"project_id"` // 所属项目
	Word      string `json:"word"`       // 原始词
}

// VsPronunciationDictCreateRequest 创建发音词典请求
type VsPronunciationDictCreateRequest struct {
	ProjectID uint64 `json:"project_id" binding:"required"` // 所属项目
	Word      string `json:"word" binding:"required"`       // 原始词
	Phoneme   string `json:"phoneme" binding:"required"`    // 发音标注
	Remark    string `json:"remark"`                        // 备注
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
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	Word      string    `json:"word"`
	Phoneme   string    `json:"phoneme"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
