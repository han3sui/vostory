package v1

// ChapterSplitResponse 章节切割响应
type ChapterSplitResponse struct {
	SceneCount    int `json:"scene_count"`
	SegmentCount  int `json:"segment_count"`
	NewCharacters int `json:"new_characters"`
	InputTokens   int `json:"input_tokens"`
	OutputTokens  int `json:"output_tokens"`
}
