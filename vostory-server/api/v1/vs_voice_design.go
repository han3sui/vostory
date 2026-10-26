package v1

// VoiceDesignSuggestResponse LLM 根据角色信息生成的音色描述与参考台词
type VoiceDesignSuggestResponse struct {
	Instruction  string `json:"instruction"`
	SampleText   string `json:"sample_text"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// VoiceDesignPreviewRequest 音色设计试听请求
type VoiceDesignPreviewRequest struct {
	Instruction string `json:"instruction" binding:"required"` // 音色描述
	SampleText  string `json:"sample_text" binding:"required"` // 参考台词（同时作为参考音频的转写文本）
	Seed        *int64 `json:"seed"`                           // 随机种子，为空时随机生成
}

// VoiceDesignPreviewResponse 音色设计试听结果
type VoiceDesignPreviewResponse struct {
	PreviewPath string `json:"preview_path"` // 试听文件路径，保存时回传
	Seed        int64  `json:"seed"`
	AudioBase64 string `json:"audio_base64"` // WAV 音频
}

// VoiceDesignApplyRequest 保存音色设计结果为声音配置并绑定角色
type VoiceDesignApplyRequest struct {
	PreviewPath string `json:"preview_path" binding:"required"`
	Instruction string `json:"instruction" binding:"required"`
	SampleText  string `json:"sample_text" binding:"required"`
	Seed        int64  `json:"seed"`
	Name        string `json:"name"` // 声音配置名称，默认使用角色名
}

// VoiceDesignApplyResponse 保存结果
type VoiceDesignApplyResponse struct {
	VoiceProfileID   uint64 `json:"voice_profile_id"`
	VoiceProfileName string `json:"voice_profile_name"`
}

// VoiceDesignAutoResponse 一键自动生成音色结果
type VoiceDesignAutoResponse struct {
	CharacterID      uint64 `json:"character_id"`
	VoiceProfileID   uint64 `json:"voice_profile_id"`
	VoiceProfileName string `json:"voice_profile_name"`
	Instruction      string `json:"instruction"`
	SampleText       string `json:"sample_text"`
	Seed             int64  `json:"seed"`
}
