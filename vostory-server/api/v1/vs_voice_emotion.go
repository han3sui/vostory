package v1

import "time"

// VsVoiceEmotionListQuery 情绪音频列表查询参数
type VsVoiceEmotionListQuery struct {
	*BasePageQuery
	VoiceProfileID uint64 `json:"voice_profile_id"`
	VoiceAssetID   uint64 `json:"voice_asset_id"`
	EmotionType    string `json:"emotion_type"`
	EmotionStrength string `json:"emotion_strength"`
}

// VsVoiceEmotionCreateRequest 创建情绪音频请求
type VsVoiceEmotionCreateRequest struct {
	VoiceProfileID    *uint64 `json:"voice_profile_id"`
	VoiceAssetID      *uint64 `json:"voice_asset_id"`
	EmotionType       string  `json:"emotion_type" binding:"required"`
	EmotionStrength   string  `json:"emotion_strength" binding:"required"`
	ReferenceAudioURL string  `json:"reference_audio_url" binding:"required"`
	ReferenceText     string  `json:"reference_text"`
}

// VsVoiceEmotionUpdateRequest 更新情绪音频请求
type VsVoiceEmotionUpdateRequest struct {
	VsVoiceEmotionCreateRequest
	ID uint64 `json:"id" binding:"required"`
}

// VsVoiceEmotionDetailResponse 情绪音频详情响应
type VsVoiceEmotionDetailResponse struct {
	ID                uint64    `json:"id"`
	VoiceProfileID    *uint64   `json:"voice_profile_id"`
	VoiceAssetID      *uint64   `json:"voice_asset_id"`
	EmotionType       string    `json:"emotion_type"`
	EmotionStrength   string    `json:"emotion_strength"`
	ReferenceAudioURL string    `json:"reference_audio_url"`
	ReferenceText     string    `json:"reference_text"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
