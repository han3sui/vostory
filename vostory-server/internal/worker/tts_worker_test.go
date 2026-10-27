package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/log"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestAcquireGPUReportsIdleOnlyForSingleSlot(t *testing.T) {
	w := &TTSWorker{gpuReleasedAt: make(map[chan struct{}]time.Time)}

	single := make(chan struct{}, 1)
	if idle := w.acquireGPU(single); idle != -1 {
		t.Fatalf("first acquire idle=%v, want -1 (unknown)", idle)
	}
	w.releaseGPU(single)
	time.Sleep(20 * time.Millisecond)
	if idle := w.acquireGPU(single); idle < 20*time.Millisecond {
		t.Fatalf("idle=%v, want >= 20ms since last release", idle)
	}
	w.releaseGPU(single)

	multi := make(chan struct{}, 2)
	w.acquireGPU(multi)
	w.releaseGPU(multi)
	if idle := w.acquireGPU(multi); idle != -1 {
		t.Fatalf("multi-slot idle=%v, want -1", idle)
	}
}

func newRecoverTestWorker(t *testing.T) (*TTSWorker, *gorm.DB, *redis.Client) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.VsChapter{}, &model.VsScriptSegment{}, &model.VsGenerationTask{}); err != nil {
		t.Fatal(err)
	}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	logger := &log.Logger{Logger: zap.NewNop()}
	repo := repository.NewRepository(logger, db, rdb)
	w := &TTSWorker{
		rdb:           rdb,
		logger:        logger,
		taskRepo:      repository.NewVsGenerationTaskRepository(repo),
		segmentRepo:   repository.NewVsScriptSegmentRepository(repo),
		gpuReleasedAt: make(map[chan struct{}]time.Time),
	}
	return w, db, rdb
}

func TestRecoverQueueAssignsEachSegmentOnce(t *testing.T) {
	w, db, rdb := newRecoverTestWorker(t)
	ctx := context.Background()
	charID := uint64(1)
	mk := func(num int, status string) uint64 {
		seg := &model.VsScriptSegment{ProjectID: 1, SceneID: 1, ChapterID: 1, SegmentNum: num, SegmentType: "dialogue", Content: "x", CharacterID: &charID, Status: status}
		if err := db.Create(seg).Error; err != nil {
			t.Fatal(err)
		}
		return seg.SegmentID
	}
	s1 := mk(1, "queued")
	s2 := mk(2, "processing")
	s3 := mk(3, "generated")
	mk(4, "queued") // 不属于任何未结束任务

	chapterID := uint64(1)
	mkTask := func(ids []uint64, total, completed int) *model.VsGenerationTask {
		task := &model.VsGenerationTask{ProjectID: 1, ChapterID: &chapterID, TaskType: "tts_generate", Status: "running",
			TotalBatches: total, CompletedBatches: completed, SegmentIDs: model.Uint64Array(ids)}
		if err := db.Create(task).Error; err != nil {
			t.Fatal(err)
		}
		return task
	}
	old := mkTask([]uint64{s1, s2, s3}, 3, 1) // 重复提交留下的旧任务
	newer := mkTask([]uint64{s1, s2}, 2, 0)
	drained := mkTask([]uint64{s3}, 1, 0) // 片段已生成完，没有可恢复的

	// 重启前残留的消息：同一片段在多个任务下各有一条
	rdb.LPush(ctx, redisQueueKey,
		fmt.Sprintf("%d:%d", old.TaskID, s1), fmt.Sprintf("%d:%d", old.TaskID, s1), fmt.Sprintf("%d:%d", newer.TaskID, s1), "9999:1")

	recovered := w.recoverQueue(ctx)

	queue, _ := rdb.LRange(ctx, redisQueueKey, 0, -1).Result()
	want := []string{fmt.Sprintf("%d:%d", newer.TaskID, s2), fmt.Sprintf("%d:%d", newer.TaskID, s1)}
	if fmt.Sprint(queue) != fmt.Sprint(want) {
		t.Fatalf("queue=%v, want %v (每个片段一条，归最新任务，BRPOP 先取 s1)", queue, want)
	}
	var statuses []string
	db.Model(&model.VsScriptSegment{}).Order("segment_num").Pluck("status", &statuses)
	if fmt.Sprint(statuses) != fmt.Sprint([]string{"queued", "queued", "generated", "cancelled"}) {
		t.Fatalf("statuses=%v", statuses)
	}

	w.reconcileRecoveredTasks(ctx, recovered)

	check := func(task *model.VsGenerationTask, status string, total int) {
		t.Helper()
		got, err := w.taskRepo.FindByID(ctx, task.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != status || got.TotalBatches != total {
			t.Fatalf("task %d status=%s total=%d, want %s/%d", task.TaskID, got.Status, got.TotalBatches, status, total)
		}
	}
	check(old, "completed", 1) // 片段让给了新任务，按已处理数收口
	check(newer, "running", 2) // 继续跑
	check(drained, "completed", 0)
}
