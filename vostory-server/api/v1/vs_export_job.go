package v1

import "time"

// ExportChapterAudioRequest 导出章节音频请求
type ExportChapterAudioRequest struct {
	Format string `json:"format" binding:"required,oneof=wav mp3"` // 导出格式
}

// ExportJobResponse 导出任务响应
type ExportJobResponse struct {
	ExportJobID uint64     `json:"export_job_id"`
	Status      string     `json:"status"`
	Format      string     `json:"format"`
	OutputURL   string     `json:"output_url,omitempty"`
	FileSize    int64      `json:"file_size"`
	Duration    int        `json:"duration"`
	Error       string     `json:"error,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
