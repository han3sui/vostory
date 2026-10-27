package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
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

	mu             sync.RWMutex
	providerSems   map[uint64]chan struct{} // GPU 坑：只罩住 Speech HTTP
	providerPipes  map[uint64]chan struct{} // 流水线坑：准备 + 落库，比 GPU 多 1 条以便预取
	defaultSem     chan struct{}
	defaultPipeSem chan struct{}

	gpuMu         sync.Mutex
	gpuReleasedAt map[chan struct{}]time.Time // GPU 坑上一次释放的时间，用于统计 GPU 空闲

	// persistSem 上传 + 收尾的并发上限，与 GPU、流水线坑分开，所有服务商共用
	persistSem chan struct{}
}

// persistConcurrency 同时上传落库的片段数。推理结束就让出流水线坑，上传慢时片段在这里排队，不影响 GPU
const persistConcurrency = 4

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
		rdb:            rdb,
		logger:         logger,
		taskRepo:       taskRepo,
		segmentRepo:    segmentRepo,
		ttsSvc:         ttsSvc,
		providerRepo:   providerRepo,
		projectRepo:    projectRepo,
		providerSems:   make(map[uint64]chan struct{}),
		providerPipes:  make(map[uint64]chan struct{}),
		defaultSem:     make(chan struct{}, 1),
		defaultPipeSem: make(chan struct{}, 2),
		gpuReleasedAt:  make(map[chan struct{}]time.Time),
		persistSem:     make(chan struct{}, persistConcurrency),
	}
}

func (w *TTSWorker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.refreshProviderSems(workerCtx)
	recovered := w.recoverQueue(workerCtx)

	go w.consumeLoop(workerCtx)
	go w.concurrencyRefreshLoop(workerCtx)
	go w.reconcileRecoveredTasks(workerCtx, recovered)
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
		pipeC := maxC + 1
		existing, ok := w.providerSems[p.ProviderID]
		pipe, pipeOK := w.providerPipes[p.ProviderID]
		if ok && cap(existing) == maxC && pipeOK && cap(pipe) == pipeC {
			continue
		}
		w.logger.Info("tts: updating provider concurrency",
			zap.Uint64("provider_id", p.ProviderID),
			zap.String("provider_name", p.Name),
			zap.Int("concurrency", maxC),
			zap.Int("pipeline", pipeC))
		w.providerSems[p.ProviderID] = make(chan struct{}, maxC)
		w.providerPipes[p.ProviderID] = make(chan struct{}, pipeC)
	}

	for pid := range w.providerSems {
		if !active[pid] {
			delete(w.providerSems, pid)
			delete(w.providerPipes, pid)
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

func (w *TTSWorker) getProviderPipe(providerID uint64) chan struct{} {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if sem, ok := w.providerPipes[providerID]; ok {
		return sem
	}
	return w.defaultPipeSem
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

// recoveredTask 重启后某个未结束任务这次分到的片段数，用于后台修正任务计数
type recoveredTask struct {
	task     *model.VsGenerationTask
	assigned int
}

// recoverQueue 服务重启后按数据库重建 TTS 队列，只做几次批量查询，不挡 consumeLoop 启动。
// 同一片段可能出现在多个未结束的任务里（例如重复提交重新生成），只分给最新的任务；
// Redis 里的旧消息整体清掉重推，否则同一片段会被合成多次。任务计数修正交给 reconcileRecoveredTasks 在后台做
func (w *TTSWorker) recoverQueue(ctx context.Context) []recoveredTask {
	tasks, err := w.taskRepo.FindAllByStatusesAndType(ctx, []string{"running", "pending"}, "tts_generate")
	if err != nil {
		w.logger.Error("recover tasks: query failed", zap.Error(err))
		return nil
	}
	if len(tasks) == 0 {
		return nil
	}

	var chapterIDs []uint64
	seenChapter := make(map[uint64]struct{})
	for _, task := range tasks {
		if task.ChapterID == nil {
			continue
		}
		if _, ok := seenChapter[*task.ChapterID]; !ok {
			seenChapter[*task.ChapterID] = struct{}{}
			chapterIDs = append(chapterIDs, *task.ChapterID)
		}
	}
	refs, err := w.segmentRepo.FindInQueueByChapterIDs(ctx, chapterIDs)
	if err != nil {
		w.logger.Error("recover tasks: query segments failed", zap.Error(err))
		return nil
	}

	owner := assignRecoveredSegments(tasks, refs)

	// 中断时正在合成的片段退回排队；不属于任何未结束任务的片段标为 cancelled，否则会一直卡在排队、也没法重新生成
	var processingIDs, orphanQueued, orphanProcessing []uint64
	for _, ref := range refs {
		_, owned := owner[ref.SegmentID]
		switch {
		case owned && ref.Status == "processing":
			processingIDs = append(processingIDs, ref.SegmentID)
		case !owned && ref.Status == "queued":
			orphanQueued = append(orphanQueued, ref.SegmentID)
		case !owned && ref.Status == "processing":
			orphanProcessing = append(orphanProcessing, ref.SegmentID)
		}
	}
	for _, u := range []struct {
		ids      []uint64
		from, to string
	}{
		{processingIDs, "processing", "queued"},
		{orphanQueued, "queued", "cancelled"},
		{orphanProcessing, "processing", "cancelled"},
	} {
		if len(u.ids) == 0 {
			continue
		}
		if _, err := w.segmentRepo.BatchUpdateStatus(ctx, u.ids, u.from, u.to); err != nil {
			w.logger.Error("recover tasks: reset segment status failed",
				zap.String("from", u.from), zap.String("to", u.to), zap.Error(err))
			return nil
		}
	}

	// refs 已按章节、片段顺序排好，按这个顺序推，前面的章节先合成
	msgs := make([]interface{}, 0, len(owner))
	assigned := make(map[uint64]int, len(tasks))
	for _, ref := range refs {
		taskID, ok := owner[ref.SegmentID]
		if !ok {
			continue
		}
		msgs = append(msgs, fmt.Sprintf("%d:%d", taskID, ref.SegmentID))
		assigned[taskID]++
	}
	const batchSize = 1000
	if _, err := w.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, redisQueueKey)
		for start := 0; start < len(msgs); start += batchSize {
			pipe.LPush(ctx, redisQueueKey, msgs[start:min(start+batchSize, len(msgs))]...)
		}
		return nil
	}); err != nil {
		w.logger.Error("recover tasks: rebuild queue failed", zap.Error(err))
		return nil
	}

	recovered := make([]recoveredTask, len(tasks))
	for i, task := range tasks {
		recovered[i] = recoveredTask{task: task, assigned: assigned[task.TaskID]}
	}
	w.logger.Info("recovered tts queue",
		zap.Int("tasks", len(tasks)),
		zap.Int("tasks_with_segments", len(assigned)),
		zap.Int("segments", len(msgs)),
		zap.Int("reset_processing", len(processingIDs)),
		zap.Int("orphans_cancelled", len(orphanQueued)+len(orphanProcessing)))
	return recovered
}

// assignRecoveredSegments 把排队中的片段分给包含它的最新任务（任务按 ID 升序，后面的覆盖前面的）。
// 没记录 SegmentIDs 的旧任务视为包含整章
func assignRecoveredSegments(tasks []*model.VsGenerationTask, refs []repository.SegmentRef) map[uint64]uint64 {
	byChapter := make(map[uint64][]uint64)
	for _, ref := range refs {
		byChapter[ref.ChapterID] = append(byChapter[ref.ChapterID], ref.SegmentID)
	}
	owner := make(map[uint64]uint64, len(refs))
	for _, task := range tasks {
		if task.ChapterID == nil {
			continue
		}
		if len(task.SegmentIDs) == 0 {
			for _, segID := range byChapter[*task.ChapterID] {
				owner[segID] = task.TaskID
			}
			continue
		}
		inChapter := make(map[uint64]struct{}, len(byChapter[*task.ChapterID]))
		for _, segID := range byChapter[*task.ChapterID] {
			inChapter[segID] = struct{}{}
		}
		for _, segID := range task.SegmentIDs {
			if _, ok := inChapter[segID]; ok {
				owner[segID] = task.TaskID
			}
		}
	}
	return owner
}

// reconcileRecoveredTasks 修正任务总数：已处理数 + 这次分到的片段数。片段被分给更新的任务后，
// 旧任务不会再收到它们，不改总数就永远结束不了；没分到片段的任务直接收口
func (w *TTSWorker) reconcileRecoveredTasks(ctx context.Context, recovered []recoveredTask) {
	finalized := 0
	for _, r := range recovered {
		if ctx.Err() != nil {
			return
		}
		task := r.task
		total := task.CompletedBatches + task.FailedBatches + r.assigned
		if total != task.TotalBatches {
			if err := w.taskRepo.SetTotalBatches(ctx, task.TaskID, total); err != nil {
				w.logger.Error("recover: set total batches failed", zap.Uint64("task_id", task.TaskID), zap.Error(err))
				continue
			}
		}
		if r.assigned > 0 {
			if task.Status == "pending" {
				_ = w.taskRepo.SetStarted(ctx, task.TaskID)
			}
			if total == task.TotalBatches {
				continue
			}
		}
		// worker 可能在总数修正前就处理完了这个任务的片段，重新读一次判断是否该收口
		fresh, err := w.taskRepo.FindByID(ctx, task.TaskID)
		if err != nil || fresh.Status == "completed" || fresh.Status == "failed" || fresh.Status == "cancelled" {
			continue
		}
		if fresh.CompletedBatches+fresh.FailedBatches >= total {
			fresh.TotalBatches = total
			w.finalizeStaleTask(ctx, fresh)
			finalized++
		}
	}
	w.logger.Info("reconciled recovered tts tasks", zap.Int("tasks", len(recovered)), zap.Int("finalized", finalized))
}

func (w *TTSWorker) finalizeStaleTask(ctx context.Context, task *model.VsGenerationTask) {
	if task.FailedBatches > 0 {
		_ = w.taskRepo.SetFailed(ctx, task.TaskID,
			fmt.Sprintf("%d/%d segments failed", task.FailedBatches, task.TotalBatches))
		return
	}
	_ = w.taskRepo.SetCompleted(ctx, task.TaskID)
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
		// 流水线坑比 GPU 多 1：当前条在推理时，下一条可以先准备好；推理一结束就让出坑位，
		// 上传和收尾不再挡着后面片段的准备
		pipe := w.getProviderPipe(providerID)
		pipeStart := time.Now()
		pipe <- struct{}{}
		pipeWait := time.Since(pipeStart)
		var releaseOnce sync.Once
		releasePipe := func() { releaseOnce.Do(func() { <-pipe }) }
		w.wg.Add(1)
		go func(tID, sID, pID uint64) {
			defer w.wg.Done()
			defer releasePipe()
			w.processSegment(ctx, tID, sID, pID, pipeWait, releasePipe)
		}(taskID, segmentID, providerID)
	}
}

func (w *TTSWorker) isTaskCancelled(ctx context.Context, taskID uint64) bool {
	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return false
	}
	return task.Status == "cancelled"
}

// segmentTiming 单条片段各阶段耗时，处理结束输出一行日志，用来判断瓶颈在准备、GPU 还是上传
type segmentTiming struct {
	start    time.Time
	pipeWait time.Duration // 等流水线坑（前面的片段还没推理完）
	prepare  time.Duration // 查库 + 下载参考音频
	gpuWait  time.Duration // 等 GPU 坑
	gpuIdle  time.Duration // GPU 并发为 1 时，拿到坑前 GPU 已空闲的时长
	infer    time.Duration // TTS 推理 HTTP
	persistW time.Duration // 等上传并发名额
	persist  time.Duration // 音量归一化 + 上传存储 + 写音频版本
	finish   time.Duration // 更新任务进度 + 推送事件
	textLen  int
	result   string
	stage    string // 失败或跳过时所处的阶段
}

func (w *TTSWorker) logSegmentTiming(taskID, segmentID, providerID uint64, t *segmentTiming) {
	fields := []zap.Field{
		zap.Uint64("task_id", taskID),
		zap.Uint64("segment_id", segmentID),
		zap.Uint64("provider_id", providerID),
		zap.String("result", t.result),
		zap.Int("text_len", t.textLen),
		zap.Int64("pipe_wait_ms", t.pipeWait.Milliseconds()),
		zap.Int64("prepare_ms", t.prepare.Milliseconds()),
		zap.Int64("gpu_wait_ms", t.gpuWait.Milliseconds()),
		zap.Int64("infer_ms", t.infer.Milliseconds()),
		zap.Int64("persist_wait_ms", t.persistW.Milliseconds()),
		zap.Int64("persist_ms", t.persist.Milliseconds()),
		zap.Int64("finish_ms", t.finish.Milliseconds()),
		zap.Int64("total_ms", time.Since(t.start).Milliseconds()),
	}
	if t.gpuIdle >= 0 {
		fields = append(fields, zap.Int64("gpu_idle_ms", t.gpuIdle.Milliseconds()))
	}
	if t.stage != "" {
		fields = append(fields, zap.String("stage", t.stage))
	}
	w.logger.Info("tts segment timing", fields...)
}

// acquireGPU 拿 GPU 坑。并发为 1 时顺便算出 GPU 在这之前空闲了多久（上一条释放到这一条拿到），
// 并发大于 1 时无法区分是哪块卡空着，返回 -1
func (w *TTSWorker) acquireGPU(sem chan struct{}) (idle time.Duration) {
	sem <- struct{}{}
	if cap(sem) != 1 {
		return -1
	}
	w.gpuMu.Lock()
	defer w.gpuMu.Unlock()
	last, ok := w.gpuReleasedAt[sem]
	if !ok {
		return -1
	}
	return time.Since(last)
}

func (w *TTSWorker) releaseGPU(sem chan struct{}) {
	if cap(sem) == 1 {
		w.gpuMu.Lock()
		w.gpuReleasedAt[sem] = time.Now()
		w.gpuMu.Unlock()
	}
	<-sem
}

// processSegment 处理一条片段。releasePipe 在推理结束后调用，让出流水线坑位给下一条片段准备
func (w *TTSWorker) processSegment(ctx context.Context, taskID, segmentID, providerID uint64, pipeWait time.Duration, releasePipe func()) {
	t := &segmentTiming{start: time.Now(), pipeWait: pipeWait, gpuIdle: -1, result: "skipped"}
	defer w.logSegmentTiming(taskID, segmentID, providerID, t)

	finish := func(seg *model.VsScriptSegment, res *v1.TTSSynthesizeResponse, synthErr error) {
		if synthErr != nil {
			t.result = "failed"
		} else {
			t.result = "generated"
		}
		start := time.Now()
		w.finishSegment(ctx, taskID, segmentID, seg, res, synthErr)
		t.finish = time.Since(start)
	}

	stageStart := time.Now()
	if w.isTaskCancelled(ctx, taskID) {
		w.logger.Info("task cancelled, skipping segment",
			zap.Uint64("task_id", taskID), zap.Uint64("segment_id", segmentID))
		t.stage = "cancelled"
		return
	}

	seg, err := w.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		w.logger.Warn("segment not found, skipping", zap.Uint64("segment_id", segmentID))
		w.incrementFailedAndPublish(ctx, taskID, segmentID, "skipped", "segment not found")
		t.stage = "not_found"
		return
	}
	t.textLen = len([]rune(seg.Content))
	if seg.Status == "cancelled" {
		// 任务仍在进行但片段已被取消（例如取消逻辑调整前遗留的数据），按失败计入，保证任务能收口
		w.logger.Info("segment cancelled, skipping", zap.Uint64("segment_id", segmentID))
		w.incrementFailedAndPublish(ctx, taskID, segmentID, "cancelled", "")
		t.stage = "cancelled"
		return
	}

	_ = w.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	job, synthErr := w.ttsSvc.PrepareSegmentSpeech(ctx, segmentID)
	t.prepare = time.Since(stageStart)
	if synthErr != nil {
		t.stage = "prepare"
		finish(seg, nil, synthErr)
		return
	}
	defer job.Close()

	if w.isTaskCancelled(ctx, taskID) {
		w.logger.Info("task cancelled before speech, skipping",
			zap.Uint64("task_id", taskID), zap.Uint64("segment_id", segmentID))
		t.stage = "cancelled"
		return
	}

	// GPU 坑只罩住 HTTP；落库放到坑外，下一条可以马上开跑
	gpuSem := w.getProviderSem(providerID)
	stageStart = time.Now()
	t.gpuIdle = w.acquireGPU(gpuSem)
	t.gpuWait = time.Since(stageStart)
	stageStart = time.Now()
	audioData, synthErr := w.ttsSvc.RunPreparedSpeech(ctx, job)
	t.infer = time.Since(stageStart)
	w.releaseGPU(gpuSem)
	releasePipe()
	if synthErr != nil {
		t.stage = "infer"
		finish(seg, nil, synthErr)
		return
	}

	stageStart = time.Now()
	if w.persistSem != nil {
		w.persistSem <- struct{}{}
		defer func() { <-w.persistSem }()
	}
	t.persistW = time.Since(stageStart)
	stageStart = time.Now()
	synthResult, synthErr := w.ttsSvc.PersistPreparedSpeech(ctx, job, audioData)
	t.persist = time.Since(stageStart)
	if synthErr != nil {
		t.stage = "persist"
	}
	finish(seg, synthResult, synthErr)
}

func (w *TTSWorker) finishSegment(
	ctx context.Context,
	taskID, segmentID uint64,
	seg *model.VsScriptSegment,
	synthResult *v1.TTSSynthesizeResponse,
	synthErr error,
) {

	var segStatus, segErrMsg string
	var clipID *uint64
	var audioURL string
	if synthErr != nil {
		w.logger.Warn("segment synthesis failed",
			zap.Uint64("task_id", taskID), zap.Uint64("segment_id", segmentID), zap.Error(synthErr))
		segStatus = "failed"
		segErrMsg = synthErr.Error()
	} else {
		segStatus = "generated"
		if synthResult != nil {
			clipID = &synthResult.ClipID
			audioURL = synthResult.AudioURL
		}
	}

	// 合成期间任务被取消：结果已经落库，不再计入任务进度，但仍推送片段事件，
	// 否则前端会一直把这个片段显示为「生成中」
	if w.isTaskCancelled(ctx, taskID) {
		w.logger.Info("task cancelled during synthesis, publish segment result only",
			zap.Uint64("task_id", taskID), zap.Uint64("segment_id", segmentID))
		w.publishCancelledSegment(ctx, taskID, seg, segStatus, segErrMsg, clipID, audioURL)
		return
	}

	if synthErr != nil {
		if _, err := w.taskRepo.IncrementFailed(ctx, taskID); err != nil {
			w.logger.Error("increment failed count error", zap.Uint64("task_id", taskID), zap.Error(err))
		}
	} else if _, err := w.taskRepo.IncrementCompleted(ctx, taskID); err != nil {
		w.logger.Error("increment completed failed", zap.Uint64("task_id", taskID), zap.Error(err))
		return
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

// publishCancelledSegment 推送已取消任务中片段的最终结果，task_status 为 cancelled，前端据此移出队列
func (w *TTSWorker) publishCancelledSegment(ctx context.Context, taskID uint64, seg *model.VsScriptSegment, status, errMsg string, clipID *uint64, audioURL string) {
	task, err := w.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return
	}
	w.publishEvent(ctx, TTSEvent{
		Type:       "segment_done",
		TaskID:     taskID,
		ChapterID:  seg.ChapterID,
		SegmentID:  seg.SegmentID,
		Status:     status,
		ErrorMsg:   errMsg,
		ClipID:     clipID,
		AudioURL:   audioURL,
		Progress:   task.Progress,
		Completed:  task.CompletedBatches,
		Failed:     task.FailedBatches,
		Total:      task.TotalBatches,
		TaskStatus: task.Status,
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
