package model

// VsLLMModel LLM 模型：从属于某厂家，项目绑定到具体模型
type VsLLMModel struct {
	BaseModel
	ModelID    uint64 `json:"model_id" gorm:"primaryKey;autoIncrement;comment:模型ID"`
	ProviderID uint64 `json:"provider_id" gorm:"index;not null;comment:所属厂家ID"`
	ModelName  string `json:"model_name" gorm:"size:100;not null;comment:模型标识（如 gpt-4o/deepseek-chat）"`
	SortOrder  int    `json:"sort_order" gorm:"default:0;comment:排序"`
	Status     string `json:"status" gorm:"size:1;not null;default:'0';comment:状态（0正常 1停用）"`
	Remark     string `json:"remark" gorm:"size:500;comment:备注"`

	Provider *VsLLMProvider `json:"provider" gorm:"foreignKey:ProviderID;references:ProviderID"`
}

func (VsLLMModel) TableName() string {
	return "vs_llm_model"
}
