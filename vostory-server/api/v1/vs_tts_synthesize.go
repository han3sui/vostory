package v1

import (
	"time"
)

// TTSSynthesizeResponse TTS 合成响应
type TTSSynthesizeResponse struct {
	ClipID   uint64 `json:"clip_id"`
	AudioURL string `json:"audio_url"`
	Duration int    `json:"duration"`
	Version  int    `json:"version"`
}

// BatchGenerateRequest 批量生成请求
type BatchGenerateRequest struct {
	ChapterID uint64 `json:"chapter_id" binding:"required"`
}

// BatchGenerateChaptersRequest 多章节批量生成请求
type BatchGenerateChaptersRequest struct {
	ProjectID  uint64   `json:"project_id" binding:"required"`
	ChapterIDs []uint64 `json:"chapter_ids" binding:"required,min=1"`
}

// BatchGenerateChaptersResponse 多章节批量生成响应（立即返回，入队在后台进行）
type BatchGenerateChaptersResponse struct {
	ChapterCount    int `json:"chapter_count"`    // 实际提交的章节数
	TotalCount      int `json:"total_count"`      // 提交的片段数
	SkippedCount    int `json:"skipped_count"`    // 不可生成的片段数（已锁定、排队/合成中、无内容或未指定说话人）
	SkippedChapters int `json:"skipped_chapters"` // 已有生成任务在运行而跳过的章节数
}

// BatchGenerateResponse 批量生成响应（立即返回）
type BatchGenerateResponse struct {
	TaskID       uint64 `json:"task_id"`
	TotalCount   int    `json:"total_count"`
	SkippedCount int    `json:"skipped_count"`
}

// TaskProgressResponse 任务进度响应
type TaskProgressResponse struct {
	TaskID         uint64     `json:"task_id"`
	Status         string     `json:"status"`
	Progress       int        `json:"progress"`
	TotalCount     int        `json:"total_count"`
	CompletedCount int        `json:"completed_count"`
	FailedCount    int        `json:"failed_count"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// ProjectTaskProgressResponse 项目级活跃任务进度
type ProjectTaskProgressResponse struct {
	TaskID         uint64   `json:"task_id"`
	ChapterID      *uint64  `json:"chapter_id"`
	ChapterTitle   string   `json:"chapter_title"`
	TaskType       string   `json:"task_type"`
	Status         string   `json:"status"`
	Progress       int      `json:"progress"`
	TotalCount     int      `json:"total_count"`
	CompletedCount int      `json:"completed_count"`
	FailedCount    int      `json:"failed_count"`
	SegmentIDs     []uint64 `json:"segment_ids,omitempty"`
}

// CancelQueueResponse 取消队列响应
type CancelQueueResponse struct {
	CancelledCount int64 `json:"cancelled_count"`
}

// BatchLockResponse 批量锁定/解锁响应
type BatchLockResponse struct {
	AffectedCount int64 `json:"affected_count"`
}
