package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"iot-alert-center/internal/filestorage"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/llm"
	"iot-alert-center/pkg/log"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func nopLogger() *log.Logger { return &log.Logger{Logger: zap.NewNop()} }

func newQueueTestEnv(t *testing.T) (*gorm.DB, *redis.Client, *repository.Repository) {
	t.Helper()
	db := newSuggestTestDB(t)
	if err := db.AutoMigrate(&model.VsChapter{}, &model.VsScene{}, &model.VsGenerationTask{}, &model.VsWorkspace{},
		&model.VsProject{}, &model.VsLLMProvider{}, &model.VsLLMModel{}, &model.VsTTSProvider{},
		&model.VsPromptTemplate{}, &model.VsLLMLog{}); err != nil {
		t.Fatal(err)
	}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return db, rdb, repository.NewRepository(nopLogger(), db, rdb)
}

// fakeLLMServer 返回固定内容的 OpenAI 兼容流式接口，并统计调用次数
func fakeLLMServer(t *testing.T, content string, calls *int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		chunk, _ := json.Marshal(map[string]interface{}{
			"id": "c1", "object": "chat.completion.chunk",
			"choices": []map[string]interface{}{{"index": 0, "delta": map[string]string{"content": content}}},
		})
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", chunk)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newSplitTestService(t *testing.T, db *gorm.DB, rdb *redis.Client, repo *repository.Repository, llmURL string) *vsChapterSplitService {
	t.Helper()
	if err := db.Create(&model.VsProject{ProjectID: 1, WorkspaceID: 1, Name: "测试"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.VsPromptTemplate{Name: "切分", TemplateType: "dialogue_parse", Content: "{{content}}", IsSystem: "1", Status: "0"}).Error; err != nil {
		t.Fatal(err)
	}
	return &vsChapterSplitService{
		Service:       &Service{logger: nopLogger()},
		db:            db,
		rdb:           rdb,
		chapterRepo:   repository.NewVsChapterRepository(repo),
		projectRepo:   repository.NewVsProjectRepository(repo),
		promptRepo:    repository.NewVsPromptTemplateRepository(repo),
		llmResolver:   fakeResolver{baseURL: llmURL},
		sceneRepo:     repository.NewVsSceneRepository(repo),
		segmentRepo:   repository.NewVsScriptSegmentRepository(repo),
		characterRepo: repository.NewVsCharacterRepository(repo),
		llmLogRepo:    repository.NewVsLLMLogRepository(repo),
		taskRepo:      repository.NewVsGenerationTaskRepository(repo),
		llmClient:     llm.NewClient(),
	}
}

func createTestChapters(t *testing.T, db *gorm.DB, titles ...string) []*model.VsChapter {
	t.Helper()
	var chapters []*model.VsChapter
	for i, title := range titles {
		ch := &model.VsChapter{ProjectID: 1, Title: title, ChapterNum: i + 1, Content: title + "的正文", Status: "raw"}
		if err := db.Create(ch).Error; err != nil {
			t.Fatal(err)
		}
		chapters = append(chapters, ch)
	}
	return chapters
}

func chapterSplitStatus(t *testing.T, db *gorm.DB, id uint64) string {
	t.Helper()
	var ch model.VsChapter
	if err := db.First(&ch, id).Error; err != nil {
		t.Fatal(err)
	}
	return ch.SplitStatus
}

const fakeSplitResult = `{"scenes":[{"title":"场景","description":"","segments":[{"type":"narration","content":"天亮了。","character":"","emotion":"neutral","emotion_strength":"light"}]}]}`

func TestBatchSplitRejectsChaptersAlreadyQueued(t *testing.T) {
	db, rdb, repo := newQueueTestEnv(t)
	var calls int32
	svc := newSplitTestService(t, db, rdb, repo, fakeLLMServer(t, fakeSplitResult, &calls).URL)
	chs := createTestChapters(t, db, "第一章", "第二章")
	ctx := context.Background()

	if _, err := svc.BatchSplitChapters(ctx, 1, []uint64{chs[0].ChapterID, chs[0].ChapterID}, "test"); err != nil {
		t.Fatal(err)
	}
	if got := chapterSplitStatus(t, db, chs[0].ChapterID); got != splitStatusQueued {
		t.Fatalf("split_status=%s, want queued", got)
	}
	if n, _ := rdb.LLen(ctx, llmQueueKey).Result(); n != 1 {
		t.Fatalf("queue len=%d, want 1 (duplicate ids collapsed)", n)
	}

	_, err := svc.BatchSplitChapters(ctx, 1, []uint64{chs[0].ChapterID, chs[1].ChapterID}, "test")
	if err == nil || !strings.Contains(err.Error(), "第一章") {
		t.Fatalf("err=%v, want conflict naming 第一章", err)
	}
	if got := chapterSplitStatus(t, db, chs[1].ChapterID); got != "" {
		t.Fatalf("第二章 split_status=%s, want untouched", got)
	}
	var tasks int64
	db.Model(&model.VsGenerationTask{}).Count(&tasks)
	if tasks != 1 {
		t.Fatalf("tasks=%d, want 1", tasks)
	}
	if _, err := svc.SplitChapter(ctx, chs[0].ChapterID); err == nil {
		t.Fatal("sync split of a queued chapter should be rejected")
	}
	if calls != 0 {
		t.Fatalf("llm calls=%d, want 0", calls)
	}
}

func TestRunQueuedChapterAndCancel(t *testing.T) {
	db, rdb, repo := newQueueTestEnv(t)
	var calls int32
	svc := newSplitTestService(t, db, rdb, repo, fakeLLMServer(t, fakeSplitResult, &calls).URL)
	chs := createTestChapters(t, db, "第一章", "第二章", "第三章")
	ctx := context.Background()

	// 其他任务的消息不能被误删
	rdb.LPush(ctx, llmQueueKey, "999:1")
	resp, err := svc.BatchSplitChapters(ctx, 1, []uint64{chs[0].ChapterID, chs[1].ChapterID, chs[2].ChapterID}, "test")
	if err != nil {
		t.Fatal(err)
	}

	result, claimed, err := svc.RunQueuedChapter(ctx, chs[0].ChapterID)
	if err != nil || !claimed || result.SegmentCount != 1 {
		t.Fatalf("result=%+v claimed=%v err=%v", result, claimed, err)
	}
	if got := chapterSplitStatus(t, db, chs[0].ChapterID); got != splitStatusDone {
		t.Fatalf("split_status=%s, want done", got)
	}
	// 同一章节的重复消息不会再切一次
	if _, claimed, _ := svc.RunQueuedChapter(ctx, chs[0].ChapterID); claimed {
		t.Fatal("done chapter should not be claimed again")
	}

	// 模拟第二章正在切割时取消
	db.Model(&model.VsChapter{}).Where("chapter_id = ?", chs[1].ChapterID).Update("split_status", splitStatusRunning)
	n, err := svc.CancelSplitQueue(ctx, 1)
	if err != nil || n != 1 {
		t.Fatalf("cancelled=%d err=%v, want 1 queued chapter", n, err)
	}
	task, _ := svc.taskRepo.FindByID(ctx, resp.TaskID)
	if task.Status != "cancelled" {
		t.Fatalf("task status=%s, want cancelled", task.Status)
	}
	if got := chapterSplitStatus(t, db, chs[2].ChapterID); got != splitStatusCancelled {
		t.Fatalf("第三章 split_status=%s, want cancelled", got)
	}
	if got := chapterSplitStatus(t, db, chs[1].ChapterID); got != splitStatusRunning {
		t.Fatalf("running chapter should keep running, got %s", got)
	}
	queue, _ := rdb.LRange(ctx, llmQueueKey, 0, -1).Result()
	if len(queue) != 1 || queue[0] != "999:1" {
		t.Fatalf("queue=%v, want only the other task's message", queue)
	}
	if _, claimed, _ := svc.RunQueuedChapter(ctx, chs[2].ChapterID); claimed {
		t.Fatal("cancelled chapter should not be split")
	}
	if calls != 1 {
		t.Fatalf("llm calls=%d, want 1", calls)
	}

	// 取消后可以重新入队
	if _, err := svc.BatchSplitChapters(ctx, 1, []uint64{chs[2].ChapterID}, "test"); err != nil {
		t.Fatal(err)
	}
}

func TestRecoverableChapters(t *testing.T) {
	db, rdb, repo := newQueueTestEnv(t)
	svc := newSplitTestService(t, db, rdb, repo, "http://127.0.0.1:1")
	chs := createTestChapters(t, db, "A", "B", "C", "D")
	ctx := context.Background()
	ids := model.Uint64Array{chs[0].ChapterID, chs[1].ChapterID, chs[2].ChapterID, chs[3].ChapterID}

	// 并发切割时 B 先完成、A 在切割中途服务退出：按章节状态应补 A 和 C，不能跳过 A、重切 B
	for id, st := range map[uint64]string{chs[0].ChapterID: splitStatusRunning, chs[1].ChapterID: splitStatusDone,
		chs[2].ChapterID: splitStatusQueued, chs[3].ChapterID: splitStatusCancelled} {
		db.Model(&model.VsChapter{}).Where("chapter_id = ?", id).Update("split_status", st)
	}
	task := &model.VsGenerationTask{ProjectID: 1, TaskType: "chapter_split", Status: "running", TotalBatches: 4, CompletedBatches: 1, SegmentIDs: ids}
	got, err := svc.RecoverableChapters(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got) != fmt.Sprint([]uint64{chs[0].ChapterID, chs[2].ChapterID}) {
		t.Fatalf("remaining=%v, want A and C", got)
	}
	if st := chapterSplitStatus(t, db, chs[0].ChapterID); st != splitStatusQueued {
		t.Fatalf("A split_status=%s, want requeued", st)
	}

	// 字段上线前创建的旧任务：章节没有状态，退回按已处理数量跳过
	db.Model(&model.VsChapter{}).Where("1 = 1").Update("split_status", "")
	got, _ = svc.RecoverableChapters(ctx, task)
	if fmt.Sprint(got) != fmt.Sprint([]uint64{chs[1].ChapterID, chs[2].ChapterID, chs[3].ChapterID}) {
		t.Fatalf("legacy remaining=%v, want B, C, D", got)
	}
}

func TestSingleGenerateRejectsQueuedSegment(t *testing.T) {
	db, rdb, repo := newQueueTestEnv(t)
	svc := &vsTTSSynthesizeService{
		Service:     &Service{logger: nopLogger()},
		segmentRepo: repository.NewVsScriptSegmentRepository(repo),
		chapterRepo: repository.NewVsChapterRepository(repo),
		taskRepo:    repository.NewVsGenerationTaskRepository(repo),
		rdb:         rdb,
	}
	ctx := context.Background()
	charID := uint64(1)
	for _, st := range []string{"queued", "processing"} {
		seg := &model.VsScriptSegment{ProjectID: 1, SceneID: 1, ChapterID: 1, SegmentType: "dialogue", Content: "x", CharacterID: &charID, Status: st}
		if err := db.Create(seg).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := svc.SingleGenerate(ctx, seg.SegmentID); err == nil || !strings.HasPrefix(err.Error(), "CONFLICT:") {
			t.Fatalf("status=%s err=%v, want conflict", st, err)
		}
	}
	if n, _ := rdb.LLen(ctx, ttsQueueKey).Result(); n != 0 {
		t.Fatalf("queue len=%d, want 0", n)
	}
}

func TestCancelTTSQueue(t *testing.T) {
	db, rdb, repo := newQueueTestEnv(t)
	svc := &vsTTSSynthesizeService{
		Service:     &Service{logger: nopLogger()},
		segmentRepo: repository.NewVsScriptSegmentRepository(repo),
		taskRepo:    repository.NewVsGenerationTaskRepository(repo),
		rdb:         rdb,
	}
	ctx := context.Background()
	charID := uint64(1)
	var ids model.Uint64Array
	for _, st := range []string{"queued", "processing", "generated"} {
		seg := &model.VsScriptSegment{ProjectID: 1, SceneID: 1, ChapterID: 5, SegmentType: "dialogue", Content: "x", CharacterID: &charID, Status: st}
		db.Create(seg)
		ids = append(ids, seg.SegmentID)
	}
	chapterID := uint64(5)
	task := &model.VsGenerationTask{ProjectID: 1, ChapterID: &chapterID, TaskType: "tts_generate", Status: "running", TotalBatches: 3, SegmentIDs: ids}
	db.Create(task)
	rdb.LPush(ctx, ttsQueueKey, fmt.Sprintf("%d:%d", task.TaskID, ids[0]), "4242:1")

	n, err := svc.CancelChapterQueue(ctx, chapterID)
	if err != nil || n != 1 {
		t.Fatalf("cancelled=%d err=%v, want 1", n, err)
	}
	got, _ := svc.taskRepo.FindByID(ctx, task.TaskID)
	if got.Status != "cancelled" {
		t.Fatalf("task status=%s, want cancelled", got.Status)
	}
	var statuses []string
	db.Model(&model.VsScriptSegment{}).Order("segment_id").Pluck("status", &statuses)
	if strings.Join(statuses, ",") != "cancelled,cancelled,generated" {
		t.Fatalf("statuses=%v", statuses)
	}
	queue, _ := rdb.LRange(ctx, ttsQueueKey, 0, -1).Result()
	if len(queue) != 1 || queue[0] != "4242:1" {
		t.Fatalf("queue=%v, want only the other task's message", queue)
	}
}

func newImportTestService(t *testing.T) (*vsFileImportService, *gorm.DB) {
	t.Helper()
	db, rdb, repo := newQueueTestEnv(t)
	_ = rdb
	return &vsFileImportService{
		Service:     &Service{logger: nopLogger()},
		projectRepo: repository.NewVsProjectRepository(repo),
		chapterRepo: repository.NewVsChapterRepository(repo),
		fileStorage: filestorage.NewLocal("storage"),
	}, db
}

func TestRecoverParsingReparsesInterruptedProject(t *testing.T) {
	svc, db := newImportTestService(t)
	path := filepath.Join(t.TempDir(), "book.txt")
	if err := os.WriteFile(path, []byte("第一章 起\n天亮了。\n第二章 承\n天黑了。\n"), 0644); err != nil {
		t.Fatal(err)
	}
	db.Create(&model.VsProject{ProjectID: 1, WorkspaceID: 1, Name: "书", Status: projectStatusParsing, SourceType: "txt", SourceFileURL: path,
		BaseModel: model.BaseModel{CreatedBy: "tester"}})
	db.Create(&model.VsProject{ProjectID: 2, WorkspaceID: 1, Name: "docx", Status: projectStatusParsing, SourceType: "docx"})

	svc.RecoverParsing(context.Background())

	var p1, p2 model.VsProject
	db.First(&p1, 1)
	db.First(&p2, 2)
	if p1.Status != "parsed" || p1.TotalChapters != 2 {
		t.Fatalf("project 1 status=%s chapters=%d, want parsed with 2 chapters", p1.Status, p1.TotalChapters)
	}
	if p2.Status != projectStatusParsing {
		t.Fatalf("docx project status=%s, want untouched", p2.Status)
	}
}

func TestParseInBackgroundRecoversPanic(t *testing.T) {
	svc, db := newImportTestService(t)
	path := filepath.Join(t.TempDir(), "book.txt")
	_ = os.WriteFile(path, []byte("第一章\n内容\n"), 0644)
	db.Create(&model.VsProject{ProjectID: 1, WorkspaceID: 1, Name: "书", Status: projectStatusParsing, SourceType: "txt", SourceFileURL: path})

	svc.chapterRepo = nil // 让保存章节时 panic
	svc.parseInBackground(1, "tester", 0)

	var p model.VsProject
	db.First(&p, 1)
	if p.Status != projectStatusParseFailed {
		t.Fatalf("status=%s, want parse_failed after panic", p.Status)
	}
}

func TestUploadRejectedWhileParsing(t *testing.T) {
	svc, db := newImportTestService(t)
	db.Create(&model.VsProject{ProjectID: 1, WorkspaceID: 1, Name: "书", Status: projectStatusParsing, SourceType: "txt"})
	ctx := context.WithValue(context.Background(), "login_name", "tester")
	if _, _, err := svc.UploadFile(ctx, 1, &multipart.FileHeader{Filename: "b.txt"}); err == nil || !strings.Contains(err.Error(), "正在解析") {
		t.Fatalf("err=%v, want rejection while parsing", err)
	}
}

func TestSaveChaptersIsAtomic(t *testing.T) {
	svc, db := newImportTestService(t)
	project := &model.VsProject{ProjectID: 1, WorkspaceID: 1, Name: "书", Status: projectStatusParsing}
	db.Create(project)
	// 已有第 2 章，批量写入第 1~3 章会在中途冲突
	db.Create(&model.VsChapter{ProjectID: 1, Title: "旧", ChapterNum: 2})
	ctx := context.WithValue(context.Background(), "login_name", "tester")
	_, _, err := svc.saveChapters(ctx, project, []parsedChapter{{Title: "一"}, {Title: "二"}, {Title: "三"}})
	if err == nil {
		t.Fatal("expected unique conflict")
	}
	var count int64
	db.Model(&model.VsChapter{}).Count(&count)
	if count != 1 {
		t.Fatalf("chapters=%d, want 1 (no partial insert)", count)
	}
}
