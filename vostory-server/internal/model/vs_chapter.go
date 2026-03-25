package model

// VsChapter 章节
type VsChapter struct {
	BaseModel
	ChapterID  uint64 `json:"chapter_id" gorm:"primaryKey;autoIncrement;comment:章节ID"`
	ProjectID  uint64 `json:"project_id" gorm:"not null;index;uniqueIndex:uk_project_chapter_num;comment:所属项目"`
	Title      string `json:"title" gorm:"size:200;comment:章节标题"`
	ChapterNum int    `json:"chapter_num" gorm:"not null;uniqueIndex:uk_project_chapter_num;comment:章节序号"`
	Content    string `json:"content" gorm:"type:text;comment:章节原文"`
	WordCount  int    `json:"word_count" gorm:"default:0;comment:字数"`
	Status     string `json:"status" gorm:"size:20;default:'raw';comment:章节状态（raw/parsed/edited/generated/exported）"`
	Remark     string `json:"remark" gorm:"size:500;comment:备注"`

	Project *VsProject `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
}

func (VsChapter) TableName() string {
	return "vs_chapter"
}
