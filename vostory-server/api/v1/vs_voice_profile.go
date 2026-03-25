package v1

import "time"

// VsVoiceProfileListQuery 声音配置列表查询参数
type VsVoiceProfileListQuery struct {
	*BasePageQuery
	ProjectID uint64 `json:"project_id"` // 所属项目
	Name      string `json:"name"`       // 配置名称
	Status    string `json:"status"`     // 状态
}

// VsVoiceProfileCreateRequest 创建声音配置请求
type VsVoiceProfileCreateRequest struct {
	ProjectID         uint64                 `json:"project_id" binding:"required"`
	Name              string                 `json:"name" binding:"required"`
	VoiceAssetID      *uint64                `json:"voice_asset_id"`
	ReferenceAudioURL string                 `json:"reference_audio_url"`
	ReferenceText     string                 `json:"reference_text"`
	TTSProviderID     *uint64                `json:"tts_provider_id"`
	TTSParams         map[string]interface{} `json:"tts_params"`
}

// VsVoiceProfileUpdateRequest 更新声音配置请求
type VsVoiceProfileUpdateRequest struct {
	VsVoiceProfileCreateRequest
	ID     uint64 `json:"id" binding:"required"`
	Status string `json:"status"`
}

// VsVoiceProfileDetailResponse 声音配置详情响应
type VsVoiceProfileDetailResponse struct {
	ID                uint64                 `json:"id"`
	ProjectID         uint64                 `json:"project_id"`
	Name              string                 `json:"name"`
	VoiceAssetID      *uint64                `json:"voice_asset_id"`
	ReferenceAudioURL string                 `json:"reference_audio_url"`
	ReferenceText     string                 `json:"reference_text"`
	TTSProviderID     *uint64                `json:"tts_provider_id"`
	TTSProviderName   string                 `json:"tts_provider_name"`
	TTSParams         map[string]interface{} `json:"tts_params"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// VsVoiceProfileOptionResponse 声音配置选项（下拉选择用）
type VsVoiceProfileOptionResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
