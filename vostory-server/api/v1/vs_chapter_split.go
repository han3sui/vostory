package v1

// ChapterSplitResponse 章节切割响应
type ChapterSplitResponse struct {
	SceneCount    int `json:"scene_count"`
	SegmentCount  int `json:"segment_count"`
	NewCharacters int `json:"new_characters"`
	InputTokens   int `json:"input_tokens"`
	OutputTokens  int `json:"output_tokens"`
}

// BatchSplitRequest 批量切割请求
type BatchSplitRequest struct {
	ProjectID  uint64   `json:"project_id" binding:"required"`
	ChapterIDs []uint64 `json:"chapter_ids" binding:"required,min=1"`
}

// BatchSplitResponse 批量切割响应
type BatchSplitResponse struct {
	TaskID uint64 `json:"task_id"`
	Total  int    `json:"total"`
}
