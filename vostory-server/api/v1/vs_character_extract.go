package v1

// CharacterExtractResponse 角色提取响应
type CharacterExtractResponse struct {
	ExtractedCount int `json:"extracted_count"`
	NewCount       int `json:"new_count"`
	SkippedCount   int `json:"skipped_count"`
	InputTokens    int `json:"input_tokens"`
	OutputTokens   int `json:"output_tokens"`
}
