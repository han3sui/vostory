package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/service"
	"iot-alert-center/pkg/log"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const redisQueueKey = "vs:tts:queue"

type TTSWorker struct {
	rdb          *redis.Client
	logger       *log.Logger
	taskRepo     repository.VsGenerationTaskRepository
	segmentRepo  repository.VsScriptSegmentRepository
	ttsSvc       service.VsTTSSynthesizeService
	providerRepo repository.VsTTSProviderRepository
	projectRepo  repository.VsProjectRepository
	cancel       context.CancelFunc
	wg           sync.WaitGroup

	mu           sync.RWMutex
	providerSems map[uint64]chan struct{}
	defaultSem   chan struct{}
}

func NewTTSWorker(
	rdb *redis.Client,
	logger *log.Logger,
	taskRepo repository.VsGenerationTaskRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	ttsSvc service.VsTTSSynthesizeService,
	providerRepo repository.VsTTSProviderRepository,
	projectRepo repository.VsProjectRepository,
) *TTSWorker {
	return &TTSWorker{
		rdb:          rdb,
		logger:       logger,
		taskRepo:     taskRepo,
		segmentRepo:  segmentRepo,
		ttsSvc:       ttsSvc,
		providerRepo: providerRepo,
		projectRepo:  projectRepo,
		providerSems: make(map[uint64]chan struct{}),
		defaultSem:   make(chan struct{}, 1),
	}
}

func (w *TTSWorker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.refreshProviderSems(workerCtx)
	w.recoverTasks(workerCtx)

	go w.consumeLoop(workerCtx)
	go w.concurrencyRefreshLoop(workerCtx)
	w.logger.Info("TTSWorker started")
}

func (w *TTSWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	w.logger.Info("TTSWorker stopped")
}

func (w *TTSWorker) refreshProviderSems(ctx context.Context) {
	providers, err := w.providerRepo.FindAllEnabled(ctx)
	if err != nil {
		w.logger.Warn("tts: failed to load providers for concurrency", zap.Error(err))
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	active := make(map[uint64]bool, len(providers))
	for _, p := range providers {
		active[p.ProviderID] = true
		maxC := p.MaxConcurrency
		if maxC < 1 {
			maxC = 1
		}
		existing, ok := w.providerSems[p.ProviderID]
		if ok && cap(existing) == maxC {
			continue
		}
		w.logger.Info("tts: updating provider concurrency",
			zap.Uint64("provider_id", p.ProviderID),
			zap.String("provider_name", p.Name),
			zap.Int("concurrency", maxC))
		w.providerSems[p.ProviderID] = make(chan struct{}, maxC)
	}

	for pid := range w.providerSems {
		if !active[pid] {
			delete(w.providerSems, pid)
		}
	}
}

func (w *TTSWorker) getProviderSem(providerID uint64) chan struct{} {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if sem, ok := w.providerSems[providerID]; ok {
		return sem
	}
	return w.defaultSem
}

func (w *TTSWorker) concurrencyRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.refreshProviderSems(ctx)
		}
	}
}

// EnqueueSegment pushes a "taskID:segmentID" message into the queue.
func (w *TTSWorker) EnqueueSegment(ctx context.Context, taskID, segmentID uint64) error {
	msg := fmt.Sprintf("%d:%d", taskID, segmentID)
	return w.rdb.LPush(ctx, redisQueueKey, msg).Err()
}

func (w *TTSWorker) recoverTasks(ctx context.Context) {
	tasks, err := w.taskRepo.FindAllByStatusesAndType(ctx, []string{"running", "pending"}, "tts_generate")
	if err != nil {
		w.logger.Error("recover tasks: query failed", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	w.logger.Info("recovering interrupted tasks", zap.Int("count", len(tasks)))

	for _, task := range tasks {
		if task.ChapterID == nil {
			w.logger.Warn("recover: task chapter_id is nil, skipping", zap.Uint64("task_id", task.TaskID))
			continue
		}

		if task.Status == "running" {
			if err := w.taskRepo.ResetStatus(ctx, task.TaskID, "pending"); err != nil {
				w.logger.Error("recover: reset status failed",
					zap.Uint64("task_id", task.TaskID), zap.Error(err))
				continue
			}
		}

		_ = w.taskRepo.SetStarted(ctx, task.TaskID)

		segments, err := w.segmentRepo.FindByChapterIDAndStatus(ctx, *task.ChapterID, "queued")
		if err != nil || len(segments) == 0 {
			allSegs, _ := w.segmentRepo.FindByChapterID(ctx, *task.ChapterID)
			for _, seg := range allSegs {
				if seg.CharacterID == nil {
					continue
				}
			if seg.Status == "queued" || seg.Status == "processing" {
				_ = w.segmentRepo.UpdateStatus(ctx, seg.SegmentID, "queued")
				msg := fmt.Sprintf("%d:%d", task.TaskID, seg.SegmentID)
				_ = w.rdb.LRem(ctx, redisQueueKey, 0, msg).Err()
				if err := w.rdb.LPush(ctx, redisQueueKey, msg).Err(); err != nil {
					w.logger.Error("recover: enqueue segment failed",
						zap.Uint64("task_id", task.TaskID), zap.Uint64("segment_id", seg.SegmentID), zap.Error(err))
				}
			}
			}
		} else {
		for _, seg := range segments {
			msg := fmt.Sprintf("%d:%d", task.TaskID, seg.SegmentID)
			_ = w.rdb.LRem(ctx, redisQueueKey, 0, msg).Err()
			if err := w.rdb.LPush(ctx, redisQueueKey, msg).Err(); err != nil {
				w.logger.Error("recover: enqueue segment failed",
					zap.Uint64("task_id", task.TaskID), zap.Uint64("segment_id", seg.SegmentID), zap.Error(err))
			}
		}
		}
	}
}

func (w *TTSWorker) resolveProviderID(ctx context.Context, taskID uint64) uint64 {
	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return 0
	}
	project, err := w.projectRepo.FindByID(ctx, task.ProjectID)
	if err != nil {
		return 0
	}
	if project.TTSProviderID != nil {
		return *project.TTSProviderID
	}
	return 0
}

func (w *TTSWorker) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := w.rdb.BRPop(ctx, 5*time.Second, redisQueueKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("BRPOP error", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		var taskID, segmentID uint64
		if _, err := fmt.Sscanf(result[1], "%d:%d", &taskID, &segmentID); err != nil {
			w.logger.Error("parse message failed", zap.String("raw", result[1]), zap.Error(err))
			continue
		}

		providerID := w.resolveProviderID(ctx, taskID)
		sem := w.getProviderSem(providerID)
		sem <- struct{}{}
		w.wg.Add(1)
		go func(tID, sID uint64) {
			defer w.wg.Done()
			defer func() { <-sem }()
			w.processSegment(ctx, tID, sID)
		}(taskID, segmentID)
	}
}

func (w *TTSWorker) processSegment(ctx context.Context, taskID, segmentID uint64) {
	seg, err := w.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		w.logger.Warn("segment not found, skipping", zap.Uint64("segment_id", segmentID))
		w.incrementFailedAndPublish(ctx, taskID, segmentID, "skipped", "segment not found")
		return
	}
	if seg.Status == "cancelled" {
		w.logger.Info("segment cancelled, skipping", zap.Uint64("segment_id", segmentID))
		w.incrementFailedAndPublish(ctx, taskID, segmentID, "cancelled", "")
		return
	}

	_ = w.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	synthResult, synthErr := w.ttsSvc.SynthesizeSegment(ctx, segmentID)

	var segStatus, segErrMsg string
	var clipID *uint64
	var audioURL string

	if synthErr != nil {
		w.logger.Warn("segment synthesis failed",
			zap.Uint64("task_id", taskID), zap.Uint64("segment_id", segmentID), zap.Error(synthErr))

		if _, err := w.taskRepo.IncrementFailed(ctx, taskID); err != nil {
			w.logger.Error("increment failed count error", zap.Uint64("task_id", taskID), zap.Error(err))
		}
		segStatus = "failed"
		segErrMsg = synthErr.Error()
	} else {
		if _, err := w.taskRepo.IncrementCompleted(ctx, taskID); err != nil {
			w.logger.Error("increment completed failed", zap.Uint64("task_id", taskID), zap.Error(err))
			return
		}
		segStatus = "generated"
		if synthResult != nil {
			clipID = &synthResult.ClipID
			audioURL = synthResult.AudioURL
		}
	}

	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return
	}

	processed := task.CompletedBatches + task.FailedBatches
	progress := 0
	if task.TotalBatches <= 0 {
		progress = 100
	} else {
		progress = processed * 100 / task.TotalBatches
		if progress > 100 {
			progress = 100
		}
	}
	_ = w.taskRepo.UpdateProgress(ctx, taskID, task.CompletedBatches, progress)

	taskDone := false
	taskStatus := task.Status
	if task.TotalBatches <= 0 || processed >= task.TotalBatches {
		taskDone = true
		if task.FailedBatches > 0 {
			taskStatus = "failed"
			_ = w.taskRepo.SetFailed(ctx, taskID,
				fmt.Sprintf("%d/%d segments failed", task.FailedBatches, task.TotalBatches))
		} else {
			taskStatus = "completed"
			_ = w.taskRepo.SetCompleted(ctx, taskID)
		}
	}

	var chapterID uint64
	var chapterTitle string
	if task.ChapterID != nil {
		chapterID = *task.ChapterID
	}
	if task.Chapter != nil {
		chapterTitle = task.Chapter.Title
	}

	w.publishEvent(ctx, TTSEvent{
		Type:         "segment_done",
		TaskID:       taskID,
		ChapterID:    chapterID,
		ChapterTitle: chapterTitle,
		SegmentID:    segmentID,
		Status:       segStatus,
		ErrorMsg:     segErrMsg,
		ClipID:       clipID,
		AudioURL:     audioURL,
		Progress:     progress,
		Completed:    task.CompletedBatches,
		Failed:       task.FailedBatches,
		Total:        task.TotalBatches,
		TaskDone:     taskDone,
		TaskStatus:   taskStatus,
	}, task.ProjectID)
}

func (w *TTSWorker) incrementFailedAndPublish(ctx context.Context, taskID, segmentID uint64, segStatus, errMsg string) {
	if _, err := w.taskRepo.IncrementFailed(ctx, taskID); err != nil {
		w.logger.Error("increment failed count error", zap.Uint64("task_id", taskID), zap.Error(err))
	}

	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return
	}

	processed := task.CompletedBatches + task.FailedBatches
	progress := 0
	if task.TotalBatches > 0 {
		progress = processed * 100 / task.TotalBatches
		if progress > 100 {
			progress = 100
		}
	}
	_ = w.taskRepo.UpdateProgress(ctx, taskID, task.CompletedBatches, progress)

	taskDone := task.TotalBatches <= 0 || processed >= task.TotalBatches
	taskStatus := task.Status
	if taskDone {
		taskStatus = "failed"
		_ = w.taskRepo.SetFailed(ctx, taskID, fmt.Sprintf("%d/%d segments failed", task.FailedBatches, task.TotalBatches))
	}

	var chapterID uint64
	var chapterTitle string
	if task.ChapterID != nil {
		chapterID = *task.ChapterID
	}
	if task.Chapter != nil {
		chapterTitle = task.Chapter.Title
	}

	w.publishEvent(ctx, TTSEvent{
		Type:         "segment_done",
		TaskID:       taskID,
		ChapterID:    chapterID,
		ChapterTitle: chapterTitle,
		SegmentID:    segmentID,
		Status:       segStatus,
		ErrorMsg:     errMsg,
		Progress:     progress,
		Completed:    task.CompletedBatches,
		Failed:       task.FailedBatches,
		Total:        task.TotalBatches,
		TaskDone:     taskDone,
		TaskStatus:   taskStatus,
	}, task.ProjectID)
}

// TTSEvent is the SSE event payload published via Redis Pub/Sub.
type TTSEvent struct {
	Type         string  `json:"type"`
	TaskID       uint64  `json:"task_id"`
	ChapterID    uint64  `json:"chapter_id"`
	ChapterTitle string  `json:"chapter_title,omitempty"`
	SegmentID    uint64  `json:"segment_id"`
	Status       string  `json:"status"`
	ErrorMsg     string  `json:"error_message,omitempty"`
	ClipID       *uint64 `json:"clip_id,omitempty"`
	AudioURL     string  `json:"audio_url,omitempty"`
	Progress     int     `json:"progress"`
	Completed    int     `json:"completed"`
	Failed       int     `json:"failed"`
	Total        int     `json:"total"`
	TaskDone     bool    `json:"task_done"`
	TaskStatus   string  `json:"task_status"`
}

func (w *TTSWorker) publishEvent(ctx context.Context, evt TTSEvent, projectID uint64) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	channel := fmt.Sprintf("vs:tts:events:project:%d", projectID)
	w.rdb.Publish(ctx, channel, string(data))
}
