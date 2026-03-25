package model

import (
	"database/sql/driver"
)

// PromptTemplateIDMap 项目绑定的各类型 Prompt 模板ID映射
// key: template_type (character_extract/dialogue_parse/emotion_tag/scene_split/text_correct)
// value: template_id
type PromptTemplateIDMap map[string]uint64

func (m PromptTemplateIDMap) Value() (driver.Value, error) {
	return JSONValueWithDefault(m, len(m) == 0, "{}")
}

func (m *PromptTemplateIDMap) Scan(value interface{}) error {
	return JSONScan(value, m)
}

// VsProject 项目
type VsProject struct {
	BaseModel
	ProjectID         uint64            `json:"project_id" gorm:"primaryKey;comment:项目ID"`
	WorkspaceID       uint64            `json:"workspace_id" gorm:"not null;index;comment:所属工作空间"`
	Name              string            `json:"name" gorm:"size:200;not null;comment:项目名称"`
	Description       string            `json:"description" gorm:"size:1000;comment:项目描述"`
	CoverURL          string            `json:"cover_url" gorm:"size:500;comment:封面图"`
	SourceType        string            `json:"source_type" gorm:"size:20;comment:导入来源（txt/docx/epub）"`
	SourceFileURL     string            `json:"source_file_url" gorm:"size:500;comment:原始文件存储路径"`
	Status            string            `json:"status" gorm:"size:20;default:'draft';comment:项目状态（draft/parsing/parsed/generating/completed）"`
	LLMProviderID     *uint64           `json:"llm_provider_id" gorm:"comment:项目绑定的LLM提供商"`
	TTSProviderID     *uint64           `json:"tts_provider_id" gorm:"comment:项目绑定的TTS提供商"`
	PromptTemplateIDs PromptTemplateIDMap `json:"prompt_template_ids" gorm:"type:text;comment:项目绑定的各类型Prompt模板ID映射"`
	TotalChapters     int               `json:"total_chapters" gorm:"default:0;comment:总章节数"`
	TotalCharacters   int               `json:"total_characters" gorm:"default:0;comment:总角色数"`
	Remark            string            `json:"remark" gorm:"size:500;comment:备注"`

	Workspace   *VsWorkspace   `json:"workspace" gorm:"foreignKey:WorkspaceID;references:WorkspaceID"`
	LLMProvider *VsLLMProvider `json:"llm_provider" gorm:"foreignKey:LLMProviderID;references:ProviderID"`
	TTSProvider *VsTTSProvider `json:"tts_provider" gorm:"foreignKey:TTSProviderID;references:ProviderID"`
}

func (VsProject) TableName() string {
	return "vs_project"
}
