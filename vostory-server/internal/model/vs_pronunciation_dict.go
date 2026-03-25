package model

// VsPronunciationDict 发音词典
type VsPronunciationDict struct {
	BaseModel
	DictID      uint64  `json:"dict_id" gorm:"primaryKey;autoIncrement;comment:词典ID"`
	ProjectID   *uint64 `json:"project_id" gorm:"index;uniqueIndex:uk_workspace_project_word;comment:所属项目（NULL表示全局词典）"`
	WorkspaceID uint64  `json:"workspace_id" gorm:"not null;index;uniqueIndex:uk_workspace_project_word;comment:所属工作空间"`
	Word        string  `json:"word" gorm:"size:100;not null;uniqueIndex:uk_workspace_project_word;comment:原始词"`
	Phoneme     string  `json:"phoneme" gorm:"size:200;not null;comment:发音标注"`
	Remark      string  `json:"remark" gorm:"size:500;comment:备注"`

	Project   *VsProject   `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
	Workspace *VsWorkspace `json:"workspace" gorm:"foreignKey:WorkspaceID;references:WorkspaceID"`
}

func (VsPronunciationDict) TableName() string {
	return "vs_pronunciation_dict"
}
