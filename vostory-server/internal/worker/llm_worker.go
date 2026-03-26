package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/service"
	"iot-alert-center/pkg/log"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const llmQueueKey = "vs:llm:queue"

type LLMWorker struct {
	rdb         *redis.Client
	logger      *log.Logger
	taskRepo    repository.VsGenerationTaskRepository
	chapterRepo repository.VsChapterRepository
	splitSvc    service.VsChapterSplitService
	cancel      context.CancelFunc
}

func NewLLMWorker(
	rdb *redis.Client,
	logger *log.Logger,
	taskRepo repository.VsGenerationTaskRepository,
	chapterRepo repository.VsChapterRepository,
	splitSvc service.VsChapterSplitService,
) *LLMWorker {
	return &LLMWorker{
		rdb:         rdb,
		logger:      logger,
		taskRepo:    taskRepo,
		chapterRepo: chapterRepo,
		splitSvc:    splitSvc,
	}
}

func (w *LLMWorker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.recoverTasks(workerCtx)

	go w.consumeLoop(workerCtx)
	w.logger.Info("LLMWorker started")
}

func (w *LLMWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.logger.Info("LLMWorker stopped")
}

func (w *LLMWorker) recoverTasks(ctx context.Context) {
	tasks, err := w.taskRepo.FindAllByStatusesAndType(ctx, []string{"running", "pending"}, "chapter_split")
	if err != nil {
		w.logger.Error("llm recover tasks: query failed", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	w.logger.Info("recovering interrupted LLM tasks", zap.Int("count", len(tasks)))

	for _, task := range tasks {
		if task.Status == "running" {
			if err := w.taskRepo.ResetStatus(ctx, task.TaskID, "pending"); err != nil {
				w.logger.Error("llm recover: reset status failed",
					zap.Uint64("task_id", task.TaskID), zap.Error(err))
				continue
			}
		}
		_ = w.taskRepo.SetStarted(ctx, task.TaskID)

		processed := task.CompletedBatches + task.FailedBatches
		for i, chID := range task.SegmentIDs {
			if i < processed {
				continue
			}
			msg := fmt.Sprintf("%d:%d", task.TaskID, chID)
			if err := w.rdb.LPush(ctx, llmQueueKey, msg).Err(); err != nil {
				w.logger.Error("llm recover: enqueue chapter failed",
					zap.Uint64("task_id", task.TaskID), zap.Uint64("chapter_id", chID), zap.Error(err))
			}
		}
	}
}

func (w *LLMWorker) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := w.rdb.BRPop(ctx, 5*time.Second, llmQueueKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("LLM BRPOP error", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		var taskID, chapterID uint64
		if _, err := fmt.Sscanf(result[1], "%d:%d", &taskID, &chapterID); err != nil {
			w.logger.Error("llm parse message failed", zap.String("raw", result[1]), zap.Error(err))
			continue
		}

		w.processChapter(ctx, taskID, chapterID)
	}
}

func (w *LLMWorker) processChapter(ctx context.Context, taskID, chapterID uint64) {
	chapter, err := w.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		w.logger.Warn("chapter not found, skipping", zap.Uint64("chapter_id", chapterID))
		w.incrementFailedAndPublish(ctx, taskID, chapterID, "章节不存在")
		return
	}

	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		w.logger.Error("task not found", zap.Uint64("task_id", taskID))
		return
	}
	if task.Status == "cancelled" {
		w.logger.Info("task cancelled, skipping", zap.Uint64("task_id", taskID))
		return
	}

	splitResult, splitErr := w.splitSvc.SplitChapter(ctx, chapterID)

	var status, errMsg string
	var sceneCount, segmentCount int

	if splitErr != nil {
		w.logger.Warn("chapter split failed",
			zap.Uint64("task_id", taskID), zap.Uint64("chapter_id", chapterID), zap.Error(splitErr))
		if _, err := w.taskRepo.IncrementFailed(ctx, taskID); err != nil {
			w.logger.Error("increment failed count error", zap.Uint64("task_id", taskID), zap.Error(err))
		}
		status = "failed"
		errMsg = splitErr.Error()
	} else {
		if _, err := w.taskRepo.IncrementCompleted(ctx, taskID); err != nil {
			w.logger.Error("increment completed failed", zap.Uint64("task_id", taskID), zap.Error(err))
		}
		status = "completed"
		if splitResult != nil {
			sceneCount = splitResult.SceneCount
			segmentCount = splitResult.SegmentCount
		}
	}

	task, err = w.taskRepo.FindByID(ctx, taskID)
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
				fmt.Sprintf("%d/%d chapters failed", task.FailedBatches, task.TotalBatches))
		} else {
			taskStatus = "completed"
			_ = w.taskRepo.SetCompleted(ctx, taskID)
		}
	}

	chapterTitle := chapter.Title

	w.publishEvent(ctx, LLMEvent{
		Type:         "chapter_split_done",
		TaskID:       taskID,
		ChapterID:    chapterID,
		ChapterTitle: chapterTitle,
		Status:       status,
		ErrorMsg:     errMsg,
		SceneCount:   sceneCount,
		SegmentCount: segmentCount,
		Progress:     progress,
		Completed:    task.CompletedBatches,
		Failed:       task.FailedBatches,
		Total:        task.TotalBatches,
		TaskDone:     taskDone,
		TaskStatus:   taskStatus,
	}, task.ProjectID)
}

func (w *LLMWorker) incrementFailedAndPublish(ctx context.Context, taskID, chapterID uint64, errMsg string) {
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
	}
	_ = w.taskRepo.UpdateProgress(ctx, taskID, task.CompletedBatches, progress)

	taskDone := task.TotalBatches <= 0 || processed >= task.TotalBatches
	taskStatus := task.Status
	if taskDone {
		taskStatus = "failed"
		_ = w.taskRepo.SetFailed(ctx, taskID, fmt.Sprintf("%d/%d chapters failed", task.FailedBatches, task.TotalBatches))
	}

	w.publishEvent(ctx, LLMEvent{
		Type:       "chapter_split_done",
		TaskID:     taskID,
		ChapterID:  chapterID,
		Status:     "failed",
		ErrorMsg:   errMsg,
		Progress:   progress,
		Completed:  task.CompletedBatches,
		Failed:     task.FailedBatches,
		Total:      task.TotalBatches,
		TaskDone:   taskDone,
		TaskStatus: taskStatus,
	}, task.ProjectID)
}

type LLMEvent struct {
	Type         string `json:"type"`
	TaskID       uint64 `json:"task_id"`
	ChapterID    uint64 `json:"chapter_id"`
	ChapterTitle string `json:"chapter_title,omitempty"`
	Status       string `json:"status"`
	ErrorMsg     string `json:"error_message,omitempty"`
	SceneCount   int    `json:"scene_count,omitempty"`
	SegmentCount int    `json:"segment_count,omitempty"`
	Progress     int    `json:"progress"`
	Completed    int    `json:"completed"`
	Failed       int    `json:"failed"`
	Total        int    `json:"total"`
	TaskDone     bool   `json:"task_done"`
	TaskStatus   string `json:"task_status"`
}

func (w *LLMWorker) publishEvent(ctx context.Context, evt LLMEvent, projectID uint64) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	channel := fmt.Sprintf("vs:tts:events:project:%d", projectID)
	w.rdb.Publish(ctx, channel, string(data))
}
