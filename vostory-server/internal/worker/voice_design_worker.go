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

// VoiceDesignWorker 批量音色设计：LLM 与 GPU 错开，当前条推理时预取下一条描述。
// Breeze 仍单并发，同一时刻只有一条 Speech。
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
	var ready *service.VoiceDesignPrepared
	for {
		if ctx.Err() != nil {
			return
		}
		if ready == nil {
			ready = w.claimAndPrepare(ctx)
			if ready == nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(voiceDesignPollInterval):
				}
				continue
			}
		}

		// 推理当前条时，下一条先写 LLM 描述
		nextCh := make(chan *service.VoiceDesignPrepared, 1)
		go func() {
			nextCh <- w.claimAndPrepare(ctx)
		}()

		data, seed, synthErr := w.svc.RunPreparedDesign(ctx, ready)
		job := ready
		ready = nil

		// 落库不占 GPU，与下一条推理重叠
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			if synthErr != nil {
				if ctx.Err() != nil {
					return
				}
				w.logger.Error("voice design: synthesize failed",
					zap.Uint64("item_id", job.ItemID()), zap.Error(synthErr))
				_ = w.svc.FailPreparedDesign(job, synthErr)
				return
			}
			if err := w.svc.FinishPreparedDesign(job, data, seed); err != nil {
				w.logger.Error("voice design: persist failed",
					zap.Uint64("item_id", job.ItemID()), zap.Error(err))
			}
		}()

		ready = <-nextCh
	}
}

func (w *VoiceDesignWorker) claimAndPrepare(ctx context.Context) *service.VoiceDesignPrepared {
	for {
		if ctx.Err() != nil {
			return nil
		}
		item, err := w.itemRepo.NextQueued(ctx)
		if err != nil {
			w.logger.Error("voice design: query queued item failed", zap.Error(err))
			return nil
		}
		if item == nil {
			return nil
		}
		job, err := w.svc.PrepareItem(ctx, item.ItemID)
		if err != nil {
			w.logger.Error("voice design: prepare item failed",
				zap.Uint64("item_id", item.ItemID), zap.Error(err))
			continue
		}
		if job != nil {
			return job
		}
	}
}
