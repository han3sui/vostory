package model

import (
	"database/sql/driver"
)

// ModelList LLM 可用模型列表
type ModelList []string

func (l ModelList) Value() (driver.Value, error) {
	return JSONValueWithDefault(l, len(l) == 0, "[]")
}

func (l *ModelList) Scan(value interface{}) error {
	return JSONScan(value, l)
}

// ProviderParams 提供商自定义参数
type ProviderParams map[string]interface{}

func (m ProviderParams) Value() (driver.Value, error) {
	return JSONValueWithDefault(m, len(m) == 0, "{}")
}

func (m *ProviderParams) Scan(value interface{}) error {
	return JSONScan(value, m)
}

// VsLLMProvider LLM 服务商配置
type VsLLMProvider struct {
	BaseModel
	ProviderID   uint64         `json:"provider_id" gorm:"primaryKey;comment:提供商ID"`
	Name         string         `json:"name" gorm:"size:100;not null;comment:显示名称"`
	ProviderType string         `json:"provider_type" gorm:"size:30;not null;comment:提供商类型（openai/deepseek/anthropic/gemini/ollama/azure/aliyun/custom）"`
	APIBaseURL   string         `json:"api_base_url" gorm:"size:500;not null;comment:API地址"`
	APIKey       string         `json:"api_key" gorm:"size:500;comment:API密钥"`
	ModelList    ModelList      `json:"model_list" gorm:"type:text;comment:可用模型列表"`
	DefaultModel string        `json:"default_model" gorm:"size:100;comment:默认模型"`
	CustomParams ProviderParams `json:"custom_params" gorm:"type:text;comment:自定义参数"`
	SortOrder    int            `json:"sort_order" gorm:"default:0;comment:排序"`
	Status       string         `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
}

func (VsLLMProvider) TableName() string {
	return "vs_llm_provider"
}
