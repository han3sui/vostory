package model

// VsVoiceAsset 全局声音资产
type VsVoiceAsset struct {
	BaseModel
	VoiceAssetID      uint64     `json:"voice_asset_id" gorm:"primaryKey;comment:声音资产ID"`
	WorkspaceID       uint64     `json:"workspace_id" gorm:"not null;index;comment:所属工作空间"`
	Name              string     `json:"name" gorm:"size:100;not null;comment:音色名称"`
	Gender            string     `json:"gender" gorm:"size:10;comment:性别（male/female/unknown）"`
	Description       string     `json:"description" gorm:"size:500;comment:描述"`
	ReferenceAudioURL string     `json:"reference_audio_url" gorm:"size:500;comment:默认参考音频路径"`
	ReferenceText     string     `json:"reference_text" gorm:"size:1000;comment:参考音频对应文本"`
	TTSProviderID     *uint64    `json:"tts_provider_id" gorm:"comment:关联的TTS提供商"`
	Tags              StringList `json:"tags" gorm:"type:text;comment:标签"`
	Status            string     `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`

	Workspace   *VsWorkspace   `json:"workspace" gorm:"foreignKey:WorkspaceID;references:WorkspaceID"`
	TTSProvider *VsTTSProvider `json:"tts_provider" gorm:"foreignKey:TTSProviderID;references:ProviderID"`
}

func (VsVoiceAsset) TableName() string {
	return "vs_voice_asset"
}
