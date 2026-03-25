package model

import (
	"database/sql/driver"
)

// StringList JSON 字符串数组类型
type StringList []string

func (l StringList) Value() (driver.Value, error) {
	return JSONValueWithDefault(l, len(l) == 0, "[]")
}

func (l *StringList) Scan(value interface{}) error {
	return JSONScan(value, l)
}

// VsCharacter 角色
type VsCharacter struct {
	BaseModel
	CharacterID    uint64     `json:"character_id" gorm:"primaryKey;autoIncrement;comment:角色ID"`
	ProjectID      uint64     `json:"project_id" gorm:"not null;index;uniqueIndex:uk_project_character_name;comment:所属项目"`
	Name           string     `json:"name" gorm:"size:100;not null;uniqueIndex:uk_project_character_name;comment:角色名称"`
	Aliases        StringList `json:"aliases" gorm:"type:text;comment:别名列表"`
	Gender         string     `json:"gender" gorm:"size:10;comment:性别（male/female/unknown）"`
	Description    string     `json:"description" gorm:"size:1000;comment:角色描述"`
	Level          string     `json:"level" gorm:"size:20;default:'main';comment:角色层级（main/supporting/minor）"`
	VoiceProfileID *uint64    `json:"voice_profile_id" gorm:"comment:绑定的声音配置"`
	SortOrder      int        `json:"sort_order" gorm:"default:0;comment:排序"`
	Status         string     `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`

	Project      *VsProject      `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
	VoiceProfile *VsVoiceProfile `json:"voice_profile" gorm:"foreignKey:VoiceProfileID;references:VoiceProfileID"`
}

func (VsCharacter) TableName() string {
	return "vs_character"
}
