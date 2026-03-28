package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/audio"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsExportJobService interface {
	ExportChapterAudio(ctx context.Context, chapterID uint64, format string) (*v1.ExportJobResponse, error)
	GetExportJob(ctx context.Context, exportJobID uint64) (*v1.ExportJobResponse, error)
	GetExportFilePath(ctx context.Context, exportJobID uint64) (filePath string, contentType string, fileName string, err error)
}

func NewVsExportJobService(
	service *Service,
	exportJobRepo repository.VsExportJobRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	audioClipRepo repository.VsAudioClipRepository,
	chapterRepo repository.VsChapterRepository,
) VsExportJobService {
	return &vsExportJobService{
		Service:       service,
		exportJobRepo: exportJobRepo,
		segmentRepo:   segmentRepo,
		audioClipRepo: audioClipRepo,
		chapterRepo:   chapterRepo,
	}
}

type vsExportJobService struct {
	*Service
	exportJobRepo repository.VsExportJobRepository
	segmentRepo   repository.VsScriptSegmentRepository
	audioClipRepo repository.VsAudioClipRepository
	chapterRepo   repository.VsChapterRepository
}

func (s *vsExportJobService) ExportChapterAudio(ctx context.Context, chapterID uint64, format string) (*v1.ExportJobResponse, error) {
	chapter, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("章节不存在: %w", err)
	}

	segments, err := s.segmentRepo.FindByChapterID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("获取片段失败: %w", err)
	}

	segmentIDs := make([]uint64, 0, len(segments))
	for _, seg := range segments {
		segmentIDs = append(segmentIDs, seg.SegmentID)
	}

	audioMap, err := s.audioClipRepo.FindCurrentBySegmentIDs(ctx, segmentIDs)
	if err != nil {
		return nil, fmt.Errorf("获取音频失败: %w", err)
	}

	var audioPaths []string
	for _, seg := range segments {
		clip, ok := audioMap[seg.SegmentID]
		if !ok || clip.AudioURL == "" {
			continue
		}
		if _, err := os.Stat(clip.AudioURL); err != nil {
			continue
		}
		audioPaths = append(audioPaths, clip.AudioURL)
	}

	if len(audioPaths) == 0 {
		return nil, fmt.Errorf("该章节没有可导出的音频片段")
	}

	job := &model.VsExportJob{
		ProjectID:  chapter.ProjectID,
		ChapterID:  &chapterID,
		ExportType: "chapter",
		Format:     format,
		Status:     "processing",
	}
	if err := s.exportJobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("创建导出任务失败: %w", err)
	}

	ext := "." + format
	outputDir := filepath.Join("storage", "export", fmt.Sprintf("%d", chapter.ProjectID))
	outputPath := filepath.Join(outputDir, fmt.Sprintf("chapter_%d_%d%s", chapterID, job.ExportJobID, ext))

	if err := audio.MergeAudioFiles(audioPaths, outputPath, format); err != nil {
		now := time.Now()
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		job.CompletedAt = &now
		_ = s.exportJobRepo.Update(ctx, job)
		return nil, fmt.Errorf("音频合并失败: %w", err)
	}

	fi, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("导出文件不可读: %w", err)
	}
	now := time.Now()
	job.Status = "completed"
	job.OutputURL = outputPath
	job.FileSize = fi.Size()
	job.CompletedAt = &now
	if err := s.exportJobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("更新导出任务失败: %w", err)
	}

	return &v1.ExportJobResponse{
		ExportJobID: job.ExportJobID,
		Status:      job.Status,
		Format:      job.Format,
		FileSize:    job.FileSize,
		CompletedAt: job.CompletedAt,
	}, nil
}

func (s *vsExportJobService) GetExportJob(ctx context.Context, exportJobID uint64) (*v1.ExportJobResponse, error) {
	job, err := s.exportJobRepo.FindByID(ctx, exportJobID)
	if err != nil {
		return nil, fmt.Errorf("导出任务不存在: %w", err)
	}
	return &v1.ExportJobResponse{
		ExportJobID: job.ExportJobID,
		Status:      job.Status,
		Format:      job.Format,
		FileSize:    job.FileSize,
		Duration:    job.Duration,
		Error:       job.ErrorMessage,
		CompletedAt: job.CompletedAt,
	}, nil
}

func (s *vsExportJobService) GetExportFilePath(ctx context.Context, exportJobID uint64) (string, string, string, error) {
	job, err := s.exportJobRepo.FindByID(ctx, exportJobID)
	if err != nil {
		return "", "", "", fmt.Errorf("导出任务不存在: %w", err)
	}
	if job.Status != "completed" {
		return "", "", "", fmt.Errorf("导出任务尚未完成，当前状态: %s", job.Status)
	}
	if _, err := os.Stat(job.OutputURL); err != nil {
		return "", "", "", fmt.Errorf("导出文件不存在")
	}

	contentType := "audio/wav"
	ext := ".wav"
	if job.Format == "mp3" {
		contentType = "audio/mpeg"
		ext = ".mp3"
	}

	fileName := fmt.Sprintf("chapter_%d%s", *job.ChapterID, ext)
	return job.OutputURL, contentType, fileName, nil
}
