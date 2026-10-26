package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/audio"
	"iot-alert-center/internal/filestorage"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsExportJobService interface {
	ExportChapterAudio(ctx context.Context, chapterID uint64, format string) (*v1.ExportJobResponse, error)
	GetExportJob(ctx context.Context, exportJobID uint64) (*v1.ExportJobResponse, error)
	OpenExportFile(ctx context.Context, exportJobID uint64) (r io.ReadCloser, contentType string, fileName string, err error)
}

func NewVsExportJobService(
	service *Service,
	exportJobRepo repository.VsExportJobRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	audioClipRepo repository.VsAudioClipRepository,
	chapterRepo repository.VsChapterRepository,
	fileStorage filestorage.FileStorage,
) VsExportJobService {
	return &vsExportJobService{
		Service:       service,
		exportJobRepo: exportJobRepo,
		segmentRepo:   segmentRepo,
		audioClipRepo: audioClipRepo,
		chapterRepo:   chapterRepo,
		fileStorage:   fileStorage,
	}
}

type vsExportJobService struct {
	*Service
	exportJobRepo repository.VsExportJobRepository
	segmentRepo   repository.VsScriptSegmentRepository
	audioClipRepo repository.VsAudioClipRepository
	chapterRepo   repository.VsChapterRepository
	fileStorage   filestorage.FileStorage
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
	var cleanups []func()
	defer func() {
		for _, fn := range cleanups {
			fn()
		}
	}()
	for _, seg := range segments {
		clip, ok := audioMap[seg.SegmentID]
		if !ok || clip.AudioURL == "" {
			continue
		}
		path, cleanup, err := s.fileStorage.DownloadToTemp(ctx, filestorage.NormalizeKey(clip.AudioURL))
		if err != nil {
			continue
		}
		cleanups = append(cleanups, cleanup)
		audioPaths = append(audioPaths, path)
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
	tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("vostory_export_%d%s", job.ExportJobID, ext))
	if err := audio.MergeAudioFiles(audioPaths, tmpOut, format); err != nil {
		now := time.Now()
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		job.CompletedAt = &now
		_ = s.exportJobRepo.Update(ctx, job)
		return nil, fmt.Errorf("音频合并失败: %w", err)
	}
	defer os.Remove(tmpOut)

	f, err := os.Open(tmpOut)
	if err != nil {
		return nil, fmt.Errorf("导出文件不可读: %w", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("导出文件不可读: %w", err)
	}

	contentType := "audio/wav"
	if format == "mp3" {
		contentType = "audio/mpeg"
	}
	key := fmt.Sprintf("export/%d/chapter_%d_%d%s", chapter.ProjectID, chapterID, job.ExportJobID, ext)
	if err := s.fileStorage.Put(ctx, key, f, fi.Size(), contentType); err != nil {
		now := time.Now()
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		job.CompletedAt = &now
		_ = s.exportJobRepo.Update(ctx, job)
		return nil, fmt.Errorf("保存导出文件失败: %w", err)
	}

	now := time.Now()
	job.Status = "completed"
	job.OutputURL = key
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

func (s *vsExportJobService) OpenExportFile(ctx context.Context, exportJobID uint64) (io.ReadCloser, string, string, error) {
	job, err := s.exportJobRepo.FindByID(ctx, exportJobID)
	if err != nil {
		return nil, "", "", fmt.Errorf("导出任务不存在: %w", err)
	}
	if job.Status != "completed" {
		return nil, "", "", fmt.Errorf("导出任务尚未完成，当前状态: %s", job.Status)
	}
	r, err := s.fileStorage.Open(ctx, filestorage.NormalizeKey(job.OutputURL))
	if err != nil {
		return nil, "", "", fmt.Errorf("导出文件不存在")
	}

	contentType := "audio/wav"
	ext := ".wav"
	if job.Format == "mp3" {
		contentType = "audio/mpeg"
		ext = ".mp3"
	}

	fileName := fmt.Sprintf("chapter_%d%s", *job.ChapterID, ext)
	return r, contentType, fileName, nil
}
