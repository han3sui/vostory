package model

// VsPronunciationDict 发音词典（项目级）
type VsPronunciationDict struct {
	BaseModel
	DictID    uint64 `json:"dict_id" gorm:"primaryKey;autoIncrement;comment:词典ID"`
	ProjectID uint64 `json:"project_id" gorm:"not null;index;uniqueIndex:uk_project_word;comment:所属项目"`
	Word      string `json:"word" gorm:"size:100;not null;uniqueIndex:uk_project_word;comment:原始词"`
	Phoneme   string `json:"phoneme" gorm:"size:200;not null;comment:发音标注"`
	Remark    string `json:"remark" gorm:"size:500;comment:备注"`

	Project *VsProject `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
}

func (VsPronunciationDict) TableName() string {
	return "vs_pronunciation_dict"
}
