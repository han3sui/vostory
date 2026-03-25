package worker

import (
	"context"
	"fmt"
	"time"

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

// EnqueueSegment pushes a "taskID:segmentID" message into the queue.
func (w *TTSWorker) EnqueueSegment(ctx context.Context, taskID, segmentID uint64) error {
	msg := fmt.Sprintf("%d:%d", taskID, segmentID)
	return w.rdb.LPush(ctx, redisQueueKey, msg).Err()
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
					if err := w.rdb.LPush(ctx, redisQueueKey, fmt.Sprintf("%d:%d", task.TaskID, seg.SegmentID)).Err(); err != nil {
						w.logger.Error("recover: enqueue segment failed",
							zap.Uint64("task_id", task.TaskID), zap.Uint64("segment_id", seg.SegmentID), zap.Error(err))
					}
				}
			}
		} else {
			for _, seg := range segments {
				if err := w.rdb.LPush(ctx, redisQueueKey, fmt.Sprintf("%d:%d", task.TaskID, seg.SegmentID)).Err(); err != nil {
					w.logger.Error("recover: enqueue segment failed",
						zap.Uint64("task_id", task.TaskID), zap.Uint64("segment_id", seg.SegmentID), zap.Error(err))
				}
			}
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

		var taskID, segmentID uint64
		if _, err := fmt.Sscanf(result[1], "%d:%d", &taskID, &segmentID); err != nil {
			w.logger.Error("parse message failed", zap.String("raw", result[1]), zap.Error(err))
			continue
		}

		w.processSegment(ctx, taskID, segmentID)
	}
}

func (w *TTSWorker) processSegment(ctx context.Context, taskID, segmentID uint64) {
	_ = w.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	_, err := w.ttsSvc.SynthesizeSegment(ctx, segmentID)

	if err != nil {
		w.logger.Warn("segment synthesis failed",
			zap.Uint64("task_id", taskID), zap.Uint64("segment_id", segmentID), zap.Error(err))
	}

	completed, err := w.taskRepo.IncrementCompleted(ctx, taskID)
	if err != nil {
		w.logger.Error("increment completed failed", zap.Uint64("task_id", taskID), zap.Error(err))
		return
	}

	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return
	}

	progress := int(completed) * 100 / task.TotalBatches
	_ = w.taskRepo.UpdateProgress(ctx, taskID, int(completed), progress)

	if int(completed) >= task.TotalBatches {
		_ = w.taskRepo.SetCompleted(ctx, taskID)
	}
}
