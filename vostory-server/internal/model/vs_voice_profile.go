package model

import (
	"database/sql/driver"
)

// TTSParamsMap TTS 额外参数
type TTSParamsMap map[string]interface{}

func (m TTSParamsMap) Value() (driver.Value, error) {
	return JSONValueWithDefault(m, len(m) == 0, "{}")
}

func (m *TTSParamsMap) Scan(value interface{}) error {
	return JSONScan(value, m)
}

// VsVoiceProfile 声音配置（项目级）
type VsVoiceProfile struct {
	BaseModel
	VoiceProfileID    uint64       `json:"voice_profile_id" gorm:"primaryKey;autoIncrement;comment:声音配置ID"`
	ProjectID         uint64       `json:"project_id" gorm:"not null;index;comment:所属项目"`
	VoiceAssetID      *uint64      `json:"voice_asset_id" gorm:"comment:引用的全局声音资产"`
	Name              string       `json:"name" gorm:"size:100;not null;comment:配置名称"`
	Gender            string       `json:"gender" gorm:"size:10;comment:性别（male/female/unknown）"`
	Description       string       `json:"description" gorm:"size:500;comment:声音描述"`
	ReferenceAudioURL string       `json:"reference_audio_url" gorm:"size:500;comment:项目级参考音频（覆盖全局）"`
	ReferenceText     string       `json:"reference_text" gorm:"size:1000;comment:参考音频对应文本"`
	TTSProviderID     *uint64      `json:"tts_provider_id" gorm:"comment:TTS提供商（覆盖项目默认）"`
	TTSParams         TTSParamsMap `json:"tts_params" gorm:"type:text;comment:TTS额外参数"`
	Status            string       `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`

	Project     *VsProject     `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
	VoiceAsset  *VsVoiceAsset  `json:"voice_asset" gorm:"foreignKey:VoiceAssetID;references:VoiceAssetID"`
	TTSProvider *VsTTSProvider `json:"tts_provider" gorm:"foreignKey:TTSProviderID;references:ProviderID"`
}

func (VsVoiceProfile) TableName() string {
	return "vs_voice_profile"
}
