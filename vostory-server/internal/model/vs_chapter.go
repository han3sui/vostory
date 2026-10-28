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
	// SplitStatus 按章节记录切割进度，用于防重复入队、取消和服务重启后的恢复
	SplitStatus string `json:"split_status" gorm:"size:20;default:'';index;comment:切割状态（空/queued/running/done/failed/cancelled）"`

	SegmentCount    int `json:"segment_count" gorm:"->"`
	AudioCount      int `json:"audio_count" gorm:"->"`       // 已有当前音频的片段数
	FailedCount     int `json:"failed_count" gorm:"->"`      // 配音失败的片段数
	QueuedCount     int `json:"queued_count" gorm:"->"`      // 排队中的片段数
	ProcessingCount int `json:"processing_count" gorm:"->"`  // 正在生成的片段数

	Project *VsProject `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
}

func (VsChapter) TableName() string {
	return "vs_chapter"
}
