package v1

// VoiceMatchResponse 声音自动匹配响应
type VoiceMatchResponse struct {
	MatchedCount int                  `json:"matched_count"`
	SkippedCount int                  `json:"skipped_count"`
	FailedCount  int                  `json:"failed_count"`
	Details      []VoiceMatchDetail   `json:"details,omitempty"`
	InputTokens  int                  `json:"input_tokens"`
	OutputTokens int                  `json:"output_tokens"`
}

// VoiceMatchDetail 单个角色的匹配详情
type VoiceMatchDetail struct {
	CharacterID    uint64 `json:"character_id"`
	CharacterName  string `json:"character_name"`
	VoiceProfileID uint64 `json:"voice_profile_id"`
	VoiceName      string `json:"voice_name"`
	Reason         string `json:"reason"`
}
