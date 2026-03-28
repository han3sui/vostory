package model

// VsVoiceEmotion 多情绪参考音频（可关联 voice_profile 或 voice_asset，二选一）
type VsVoiceEmotion struct {
	BaseModel
	VoiceEmotionID    uint64  `json:"voice_emotion_id" gorm:"primaryKey;autoIncrement;comment:情绪音频ID"`
	VoiceProfileID    *uint64 `json:"voice_profile_id" gorm:"index;comment:所属声音配置（项目级）"`
	VoiceAssetID      *uint64 `json:"voice_asset_id" gorm:"index;comment:所属音色资产（全局级）"`
	EmotionType       string  `json:"emotion_type" gorm:"size:50;not null;comment:情绪类型"`
	EmotionStrength   string  `json:"emotion_strength" gorm:"size:20;default:'medium';comment:情绪强度（light/medium/strong）"`
	ReferenceAudioURL string  `json:"reference_audio_url" gorm:"size:500;not null;comment:该情绪的参考音频"`
	ReferenceText     string  `json:"reference_text" gorm:"size:1000;comment:参考音频对应文本"`

	VoiceProfile *VsVoiceProfile `json:"voice_profile" gorm:"foreignKey:VoiceProfileID;references:VoiceProfileID"`
	VoiceAsset   *VsVoiceAsset   `json:"voice_asset" gorm:"foreignKey:VoiceAssetID;references:VoiceAssetID"`
}

func (VsVoiceEmotion) TableName() string {
	return "vs_voice_emotion"
}
