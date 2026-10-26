package model

import (
	"database/sql/driver"
)

// ProviderParams 提供商自定义参数（TTS 等模块复用）
type ProviderParams map[string]interface{}

func (m ProviderParams) Value() (driver.Value, error) {
	return JSONValueWithDefault(m, len(m) == 0, "{}")
}

func (m *ProviderParams) Scan(value interface{}) error {
	return JSONScan(value, m)
}

// VsLLMProvider LLM 厂家（接入点）：只存 API 基址与凭证，供其下模型复用
type VsLLMProvider struct {
	BaseModel
	ProviderID     uint64 `json:"provider_id" gorm:"primaryKey;autoIncrement;comment:厂家ID"`
	Name           string `json:"name" gorm:"size:100;not null;comment:厂家名称"`
	APIBaseURL     string `json:"api_base_url" gorm:"size:500;not null;comment:API地址"`
	APIKey         string `json:"api_key" gorm:"size:500;comment:API密钥"`
	MaxConcurrency int    `json:"max_concurrency" gorm:"default:1;comment:最大并发数"`
	SortOrder      int    `json:"sort_order" gorm:"default:0;comment:排序"`
	Status         string `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
	Remark         string `json:"remark" gorm:"size:500;comment:备注"`
}

func (VsLLMProvider) TableName() string {
	return "vs_llm_provider"
}
