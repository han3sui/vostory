package worker

import (
	"context"
	"sync"
	"time"

	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/service"
	"iot-alert-center/pkg/log"

	"go.uber.org/zap"
)

const voiceDesignPollInterval = 2 * time.Second

// VoiceDesignWorker 逐条处理批量音色设计任务项。
// Breeze-TTS-2 单并发，这里只用一个协程顺序处理；任务项状态存在数据库里，服务重启后可继续。
type VoiceDesignWorker struct {
	logger   *log.Logger
	itemRepo repository.VsVoiceDesignItemRepository
	svc      service.VsVoiceDesignService
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewVoiceDesignWorker(
	logger *log.Logger,
	itemRepo repository.VsVoiceDesignItemRepository,
	svc service.VsVoiceDesignService,
) *VoiceDesignWorker {
	return &VoiceDesignWorker{logger: logger, itemRepo: itemRepo, svc: svc}
}

func (w *VoiceDesignWorker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	// 上次退出时处理到一半的任务项重新排队
	if n, err := w.itemRepo.ResetRunning(workerCtx); err != nil {
		w.logger.Error("voice design: reset running items failed", zap.Error(err))
	} else if n > 0 {
		w.logger.Info("voice design: requeued interrupted items", zap.Int64("count", n))
	}

	w.wg.Add(1)
	go w.loop(workerCtx)
	w.logger.Info("VoiceDesignWorker started")
}

func (w *VoiceDesignWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	w.logger.Info("VoiceDesignWorker stopped")
}

func (w *VoiceDesignWorker) loop(ctx context.Context) {
	defer w.wg.Done()
	for {
		if ctx.Err() != nil {
			return
		}
		item, err := w.itemRepo.NextQueued(ctx)
		if err != nil {
			w.logger.Error("voice design: query queued item failed", zap.Error(err))
		}
		if err != nil || item == nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(voiceDesignPollInterval):
			}
			continue
		}
		if err := w.svc.ProcessItem(ctx, item.ItemID); err != nil {
			w.logger.Error("voice design: process item failed",
				zap.Uint64("item_id", item.ItemID), zap.Error(err))
		}
	}
}
