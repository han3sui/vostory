package model

// VsScene 场景
type VsScene struct {
	BaseModel
	SceneID     uint64 `json:"scene_id" gorm:"primaryKey;autoIncrement;comment:场景ID"`
	ChapterID   uint64 `json:"chapter_id" gorm:"not null;index;uniqueIndex:uk_chapter_scene_num;comment:所属章节"`
	SceneNum    int    `json:"scene_num" gorm:"not null;uniqueIndex:uk_chapter_scene_num;comment:场景序号"`
	Title       string `json:"title" gorm:"size:200;comment:场景标题"`
	Description string `json:"description" gorm:"size:1000;comment:场景描述"`
	Status      string `json:"status" gorm:"size:20;default:'raw';comment:场景状态（raw/parsed/edited）"`

	Chapter *VsChapter `json:"chapter" gorm:"foreignKey:ChapterID;references:ChapterID"`
}

func (VsScene) TableName() string {
	return "vs_scene"
}
