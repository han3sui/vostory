package model

import "time"

// VsExportJob 导出任务
type VsExportJob struct {
	BaseModel
	ExportJobID  uint64     `json:"export_job_id" gorm:"primaryKey;comment:导出任务ID"`
	ProjectID    uint64     `json:"project_id" gorm:"not null;index;comment:所属项目"`
	ChapterID    *uint64    `json:"chapter_id" gorm:"comment:导出的章节（NULL表示整项目）"`
	ExportType   string     `json:"export_type" gorm:"size:20;not null;comment:导出类型（chapter/project）"`
	Format       string     `json:"format" gorm:"size:10;not null;comment:导出格式（wav/mp3）"`
	Status       string     `json:"status" gorm:"size:20;default:'pending';comment:导出状态（pending/processing/completed/failed）"`
	OutputURL    string     `json:"output_url" gorm:"size:500;comment:导出文件路径"`
	FileSize     int64      `json:"file_size" gorm:"default:0;comment:文件大小（字节）"`
	Duration     int        `json:"duration" gorm:"default:0;comment:总时长（毫秒）"`
	ErrorMessage string     `json:"error_message" gorm:"size:2000;comment:错误信息"`
	CompletedAt  *time.Time `json:"completed_at" gorm:"comment:完成时间"`

	Project *VsProject `json:"project" gorm:"foreignKey:ProjectID;references:ProjectID"`
	Chapter *VsChapter `json:"chapter" gorm:"foreignKey:ChapterID;references:ChapterID"`
}

func (VsExportJob) TableName() string {
	return "vs_export_job"
}
