package model

// VsPromptTemplate Prompt 模板
type VsPromptTemplate struct {
	BaseModel
	TemplateID   uint64 `json:"template_id" gorm:"primaryKey;comment:模板ID"`
	Name         string `json:"name" gorm:"size:100;not null;comment:模板名称"`
	TemplateType string `json:"template_type" gorm:"size:30;not null;comment:模板类型（character_extract/dialogue_parse/emotion_tag/scene_split/text_correct）"`
	Content      string `json:"content" gorm:"type:text;not null;comment:Prompt内容"`
	Description  string `json:"description" gorm:"size:500;comment:模板描述"`
	IsSystem     string `json:"is_system" gorm:"size:1;default:'0';comment:是否系统内置（1是 0否）"`
	Version      int    `json:"version" gorm:"default:1;comment:版本号"`
	SortOrder    int    `json:"sort_order" gorm:"default:0;comment:排序"`
	Status       string `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
}

func (VsPromptTemplate) TableName() string {
	return "vs_prompt_template"
}
