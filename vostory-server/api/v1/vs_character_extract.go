package v1

// CharacterExtractFromTextRequest 从文本智能录入角色请求
type CharacterExtractFromTextRequest struct {
	ProjectID uint64 `json:"project_id" binding:"required"`
	Text      string `json:"text" binding:"required"`
}

// CharacterExtractResponse 角色提取响应
type CharacterExtractResponse struct {
	ExtractedCount int `json:"extracted_count"`
	NewCount       int `json:"new_count"`
	UpdatedCount   int `json:"updated_count"`
	SkippedCount   int `json:"skipped_count"`
	InputTokens    int `json:"input_tokens"`
	OutputTokens   int `json:"output_tokens"`
}
