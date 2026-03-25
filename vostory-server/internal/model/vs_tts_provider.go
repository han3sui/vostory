package model

import (
	"database/sql/driver"
)

// FeatureList TTS 支持的能力列表
type FeatureList []string

func (l FeatureList) Value() (driver.Value, error) {
	return JSONValueWithDefault(l, len(l) == 0, "[]")
}

func (l *FeatureList) Scan(value interface{}) error {
	return JSONScan(value, l)
}

// VsTTSProvider TTS 服务商配置
type VsTTSProvider struct {
	BaseModel
	ProviderID        uint64         `json:"provider_id" gorm:"primaryKey;comment:提供商ID"`
	Name              string         `json:"name" gorm:"size:100;not null;comment:显示名称"`
	ProviderType      string         `json:"provider_type" gorm:"size:30;not null;comment:提供商类型（local/online/custom）"`
	APIBaseURL        string         `json:"api_base_url" gorm:"size:500;not null;comment:API地址"`
	APIKey            string         `json:"api_key" gorm:"size:500;comment:API密钥"`
	SupportedFeatures FeatureList    `json:"supported_features" gorm:"type:text;comment:支持的能力（emotion/clone/multi_speaker等）"`
	CustomParams      ProviderParams `json:"custom_params" gorm:"type:text;comment:自定义参数"`
	SortOrder         int            `json:"sort_order" gorm:"default:0;comment:排序"`
	Status            string         `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
}

func (VsTTSProvider) TableName() string {
	return "vs_tts_provider"
}
