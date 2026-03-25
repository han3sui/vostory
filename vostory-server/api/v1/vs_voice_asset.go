package v1

import "time"

// VsVoiceAssetListQuery 声音资产列表查询参数
type VsVoiceAssetListQuery struct {
	*BasePageQuery
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Status string `json:"status"`
}

// VsVoiceAssetCreateRequest 创建声音资产请求
type VsVoiceAssetCreateRequest struct {
	Name              string   `json:"name" binding:"required"`
	Gender            string   `json:"gender"`
	Description       string   `json:"description"`
	ReferenceAudioURL string   `json:"reference_audio_url"`
	ReferenceText     string   `json:"reference_text"`
	TTSProviderID     *uint64  `json:"tts_provider_id"`
	Tags              []string `json:"tags"`
}

// VsVoiceAssetUpdateRequest 更新声音资产请求
type VsVoiceAssetUpdateRequest struct {
	VsVoiceAssetCreateRequest
	ID     uint64 `json:"id" binding:"required"`
	Status string `json:"status"`
}

// VsVoiceAssetDetailResponse 声音资产详情响应
type VsVoiceAssetDetailResponse struct {
	ID                uint64    `json:"id"`
	Name              string    `json:"name"`
	Gender            string    `json:"gender"`
	Description       string    `json:"description"`
	ReferenceAudioURL string    `json:"reference_audio_url"`
	ReferenceText     string    `json:"reference_text"`
	TTSProviderID     *uint64   `json:"tts_provider_id"`
	TTSProviderName   string    `json:"tts_provider_name"`
	Tags              []string  `json:"tags"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// VsVoiceAssetOptionResponse 声音资产选项（下拉选择用）
type VsVoiceAssetOptionResponse struct {
	ID                uint64   `json:"id"`
	Name              string   `json:"name"`
	Gender            string   `json:"gender"`
	Tags              []string `json:"tags"`
	ReferenceAudioURL string   `json:"reference_audio_url"`
}
