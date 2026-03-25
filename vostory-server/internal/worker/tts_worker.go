package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/service"
	"iot-alert-center/pkg/log"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const redisQueueKey = "vs:tts:queue"

type TTSWorker struct {
	rdb         *redis.Client
	logger      *log.Logger
	taskRepo    repository.VsGenerationTaskRepository
	segmentRepo repository.VsScriptSegmentRepository
	ttsSvc      service.VsTTSSynthesizeService
	cancel      context.CancelFunc
}

func NewTTSWorker(
	rdb *redis.Client,
	logger *log.Logger,
	taskRepo repository.VsGenerationTaskRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	ttsSvc service.VsTTSSynthesizeService,
) *TTSWorker {
	return &TTSWorker{
		rdb:         rdb,
		logger:      logger,
		taskRepo:    taskRepo,
		segmentRepo: segmentRepo,
		ttsSvc:      ttsSvc,
	}
}

func (w *TTSWorker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.recoverTasks(workerCtx)

	go w.consumeLoop(workerCtx)
	w.logger.Info("TTSWorker started")
}

func (w *TTSWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.logger.Info("TTSWorker stopped")
}

func (w *TTSWorker) Enqueue(ctx context.Context, taskID uint64) error {
	return w.rdb.LPush(ctx, redisQueueKey, taskID).Err()
}

func (w *TTSWorker) recoverTasks(ctx context.Context) {
	tasks, err := w.taskRepo.FindAllByStatuses(ctx, []string{"running", "pending"})
	if err != nil {
		w.logger.Error("recover tasks: query failed", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	w.logger.Info("recovering interrupted tasks", zap.Int("count", len(tasks)))

	for _, task := range tasks {
		if task.Status == "running" {
			if err := w.taskRepo.ResetStatus(ctx, task.TaskID, "pending"); err != nil {
				w.logger.Error("recover: reset status failed",
					zap.Uint64("task_id", task.TaskID), zap.Error(err))
				continue
			}
		}
		if err := w.rdb.LPush(ctx, redisQueueKey, task.TaskID).Err(); err != nil {
			w.logger.Error("recover: enqueue failed",
				zap.Uint64("task_id", task.TaskID), zap.Error(err))
		}
	}
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

		var taskID uint64
		if _, err := fmt.Sscanf(result[1], "%d", &taskID); err != nil {
			w.logger.Error("parse task_id failed", zap.String("raw", result[1]), zap.Error(err))
			continue
		}

		w.executeTask(ctx, taskID)
	}
}

func (w *TTSWorker) executeTask(ctx context.Context, taskID uint64) {
	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		w.logger.Error("task not found", zap.Uint64("task_id", taskID), zap.Error(err))
		return
	}

	if task.Status != "pending" {
		w.logger.Warn("skip non-pending task",
			zap.Uint64("task_id", taskID), zap.String("status", task.Status))
		return
	}

	_ = w.taskRepo.SetStarted(ctx, taskID)

	var eligible []*model.VsScriptSegment

	if len(task.SegmentIDs) > 0 {
		for _, segID := range task.SegmentIDs {
			seg, err := w.segmentRepo.FindByID(ctx, segID)
			if err != nil {
				w.logger.Warn("segment not found", zap.Uint64("segment_id", segID), zap.Error(err))
				continue
			}
			eligible = append(eligible, seg)
		}
	} else {
		if task.ChapterID == nil {
			_ = w.taskRepo.UpdateStatus(ctx, taskID, "failed", "任务缺少章节ID")
			return
		}
		segments, err := w.segmentRepo.FindByChapterID(ctx, *task.ChapterID)
		if err != nil {
			_ = w.taskRepo.UpdateStatus(ctx, taskID, "failed", "获取片段失败: "+err.Error())
			return
		}
		for _, seg := range segments {
			if seg.CharacterID == nil || seg.Status == "processing" {
				continue
			}
			eligible = append(eligible, seg)
		}
	}

	if len(eligible) == 0 {
		_ = w.taskRepo.UpdateStatus(ctx, taskID, "failed", "没有可生成的片段")
		return
	}

	total := len(eligible)
	var completed int32
	var failedCount int32

	for _, seg := range eligible {
		if ctx.Err() != nil {
			_ = w.taskRepo.UpdateStatus(ctx, taskID, "pending", "服务关闭，任务中断")
			return
		}

		_, err := w.ttsSvc.SynthesizeSegment(ctx, seg.SegmentID)
		if err != nil {
			atomic.AddInt32(&failedCount, 1)
			w.logger.Warn("batch segment failed",
				zap.Uint64("segment_id", seg.SegmentID), zap.Error(err))
		}

		done := int(atomic.AddInt32(&completed, 1))
		progress := done * 100 / total
		_ = w.taskRepo.UpdateProgress(ctx, taskID, done, progress)
	}

	if failedCount > 0 {
		errMsg := fmt.Sprintf("%d 个片段合成失败", failedCount)
		if failedCount == int32(total) {
			_ = w.taskRepo.UpdateStatus(ctx, taskID, "failed", errMsg)
		} else {
			_ = w.taskRepo.SetCompleted(ctx, taskID)
			_ = w.taskRepo.UpdateStatus(ctx, taskID, "completed", errMsg)
		}
	} else {
		_ = w.taskRepo.SetCompleted(ctx, taskID)
	}
}
