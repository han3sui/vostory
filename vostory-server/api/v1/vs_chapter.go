package v1

import "time"

// VsChapterListQuery 章节列表查询参数
type VsChapterListQuery struct {
	*BasePageQuery
	ProjectID uint64 `json:"project_id"` // 所属项目
	Title     string `json:"title"`      // 章节标题
	Status    string `json:"status"`     // 章节状态
}

// VsChapterCreateRequest 创建章节请求
type VsChapterCreateRequest struct {
	ProjectID  uint64 `json:"project_id" binding:"required"` // 所属项目
	Title      string `json:"title"`                         // 章节标题
	ChapterNum int    `json:"chapter_num" binding:"required"` // 章节序号
	Content    string `json:"content"`                       // 章节原文
	Remark     string `json:"remark"`                        // 备注
}

// VsChapterUpdateRequest 更新章节请求
type VsChapterUpdateRequest struct {
	ID         uint64 `json:"id" binding:"required"`
	Title      string `json:"title"`      // 章节标题
	ChapterNum int    `json:"chapter_num"` // 章节序号
	Content    string `json:"content"`    // 章节原文
	Status     string `json:"status"`     // 章节状态
	Remark     string `json:"remark"`     // 备注
}

// VsChapterListResponse 章节列表响应（不含 content，避免大文本传输）
type VsChapterListResponse struct {
	ID           uint64    `json:"id"`
	ProjectID    uint64    `json:"project_id"`
	Title        string    `json:"title"`
	ChapterNum   int       `json:"chapter_num"`
	WordCount    int       `json:"word_count"`
	SegmentCount int       `json:"segment_count"`
	Status       string    `json:"status"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// VsChapterDetailResponse 章节详情响应
type VsChapterDetailResponse struct {
	ID         uint64    `json:"id"`
	ProjectID  uint64    `json:"project_id"`
	Title      string    `json:"title"`
	ChapterNum int       `json:"chapter_num"`
	Content    string    `json:"content"`
	WordCount  int       `json:"word_count"`
	Status     string    `json:"status"`
	Remark     string    `json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
