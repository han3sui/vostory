package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/filestorage"
	"iot-alert-center/internal/llmconfig"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestSampleEvenly(t *testing.T) {
	var contents []string
	for i := 0; i < 100; i++ {
		contents = append(contents, fmt.Sprintf("第%03d句台词内容", i))
	}
	picked := sampleEvenly(contents, 10, 6)
	if len(picked) != 10 {
		t.Fatalf("picked=%d, want 10", len(picked))
	}
	if picked[0] != contents[5] || picked[9] != contents[95] {
		t.Fatalf("picked first=%s last=%s, want spread across the book", picked[0], picked[9])
	}

	// 足够多时跳过过短的台词；不够时才用短台词补
	mixed := []string{"嗯", "是", "这是一句足够长的台词", "另一句足够长的台词"}
	if got := sampleEvenly(mixed, 2, 6); strings.Join(got, "|") != "这是一句足够长的台词|另一句足够长的台词" {
		t.Fatalf("got=%v, want long lines only", got)
	}
	if got := sampleEvenly(mixed, 3, 6); len(got) != 3 {
		t.Fatalf("got=%v, want 3 lines including short ones", got)
	}
}

func TestAppendTraitNotes(t *testing.T) {
	notes := appendTraitNotes("", "少年天才", "少年天才", "  斗气大陆强者 ", "斗气大陆强者", "")
	if notes != "斗气大陆强者" {
		t.Fatalf("notes=%q, want description and duplicates skipped", notes)
	}

	var adds []string
	for i := 0; i < maxTraitNotes+3; i++ {
		adds = append(adds, fmt.Sprintf("特征%d", i))
	}
	lines := splitTraitNotes(appendTraitNotes(notes, "", adds...))
	if len(lines) != maxTraitNotes || lines[len(lines)-1] != fmt.Sprintf("特征%d", maxTraitNotes+2) {
		t.Fatalf("lines=%v, want latest %d kept", lines, maxTraitNotes)
	}
}

func TestEnsureCharactersAccumulatesTraitNotes(t *testing.T) {
	db := newSuggestTestDB(t)
	split := &vsChapterSplitService{db: db}
	xiao := &model.VsCharacter{ProjectID: 1, Name: "萧炎", Gender: "male", Level: "main", Description: "少年天才", Status: "0"}
	if err := db.Create(xiao).Error; err != nil {
		t.Fatal(err)
	}

	runSplit := func(desc string) {
		t.Helper()
		charMap := buildCharacterMap([]*model.VsCharacter{xiao})
		result := &llmSplitResult{Scenes: []llmScene{{Segments: []llmSegment{
			{Type: "dialogue", Character: "萧炎", CharacterGender: "male", CharacterDescription: desc},
			{Type: "dialogue", Character: "老者", CharacterGender: "male", CharacterDescription: "白发苍苍的老人"},
		}}}}
		if _, err := split.ensureCharacters(context.Background(), result, 1, []*model.VsCharacter{xiao}, charMap, "test", 0); err != nil {
			t.Fatal(err)
		}
	}
	runSplit("少年天才")
	runSplit("加玛帝国的炼药师")
	runSplit("加玛帝国的炼药师")

	got := loadTestCharacter(t, db, xiao.CharacterID)
	if got.TraitNotes != "加玛帝国的炼药师" {
		t.Fatalf("trait_notes=%q, want only the new trait once", got.TraitNotes)
	}
	var pool model.VsCharacter
	if err := db.Where("name = ?", "路人-老年男").First(&pool).Error; err != nil {
		t.Fatal(err)
	}
	if pool.TraitNotes != "" {
		t.Fatalf("pool trait_notes=%q, want empty", pool.TraitNotes)
	}
}

func TestMergeCombinesTraitNotes(t *testing.T) {
	db := newSuggestTestDB(t)
	target := &model.VsCharacter{ProjectID: 1, Name: "萧炎", Level: "main", Description: "少年天才", TraitNotes: "炼药师", Status: "0"}
	source := &model.VsCharacter{ProjectID: 1, Name: "萧少爷", Level: "minor", Description: "萧家少爷", TraitNotes: "炼药师\n佩戴黑色戒指", Status: "0"}
	for _, c := range []*model.VsCharacter{target, source} {
		if err := db.Create(c).Error; err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := mergeCharactersTx(db, target, []*model.VsCharacter{source}, "test"); err != nil {
		t.Fatal(err)
	}
	got := loadTestCharacter(t, db, target.CharacterID)
	if got.TraitNotes != "炼药师\n萧家少爷\n佩戴黑色戒指" {
		t.Fatalf("trait_notes=%q", got.TraitNotes)
	}
}

type fakeResolver struct{ baseURL string }

func (r fakeResolver) ResolveByProject(context.Context, *model.VsProject) (*llmconfig.Resolved, error) {
	return &llmconfig.Resolved{BaseURL: r.baseURL, APIKey: "test", ModelName: "fake"}, nil
}

// fakeVoiceServers 模拟 OpenAI 兼容的流式 LLM 与 Breeze-TTS-2 的 /v1/audio/speech
type fakeVoiceServers struct {
	llm, tts *httptest.Server
	mu       sync.Mutex
	prompts  []string
	ttsCalls int
}

func newFakeVoiceServers(t *testing.T) *fakeVoiceServers {
	f := &fakeVoiceServers{}
	f.llm = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		f.mu.Lock()
		if len(req.Messages) > 0 {
			f.prompts = append(f.prompts, req.Messages[0].Content)
		}
		f.mu.Unlock()

		content, _ := json.Marshal(map[string]string{"instruction": "青年男声，清亮有力", "sample_text": "我命由我不由天。"})
		chunk, _ := json.Marshal(map[string]interface{}{
			"id": "c1", "object": "chat.completion.chunk",
			"choices": []map[string]interface{}{{"index": 0, "delta": map[string]string{"content": string(content)}}},
		})
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", chunk)
	}))
	f.tts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		f.mu.Lock()
		f.ttsCalls++
		f.mu.Unlock()
		_, _ = w.Write(make([]byte, 4800)) // 0.1 秒静音 PCM
	}))
	t.Cleanup(func() {
		f.llm.Close()
		f.tts.Close()
	})
	return f
}

func newBatchTestService(t *testing.T, fakes *fakeVoiceServers) (*vsVoiceDesignService, *gorm.DB) {
	t.Helper()
	// 草稿与参考音频写在相对路径 storage/ 下，切到临时目录避免污染仓库
	t.Chdir(t.TempDir())
	db := newSuggestTestDB(t)
	if err := db.AutoMigrate(&model.VsWorkspace{}, &model.VsProject{}, &model.VsLLMProvider{}, &model.VsLLMModel{},
		&model.VsTTSProvider{}, &model.VsVoiceProfile{}, &model.VsVoiceDesignItem{}, &model.VsPromptTemplate{}, &model.VsLLMLog{}); err != nil {
		t.Fatal(err)
	}
	provider := &model.VsTTSProvider{Name: "fake", ProviderType: "local", APIBaseURL: fakes.tts.URL, Status: "0"}
	if err := db.Create(provider).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.VsProject{ProjectID: 1, WorkspaceID: 1, Name: "测试", TTSProviderID: &provider.ProviderID}).Error; err != nil {
		t.Fatal(err)
	}

	repo := repository.NewRepository(&log.Logger{Logger: zap.NewNop()}, db, nil)
	svc := &vsVoiceDesignService{
		Service:          &Service{logger: &log.Logger{Logger: zap.NewNop()}},
		characterRepo:    repository.NewVsCharacterRepository(repo),
		projectRepo:      repository.NewVsProjectRepository(repo),
		voiceProfileRepo: repository.NewVsVoiceProfileRepository(repo),
		segmentRepo:      repository.NewVsScriptSegmentRepository(repo),
		promptRepo:       repository.NewVsPromptTemplateRepository(repo),
		llmResolver:      fakeResolver{baseURL: fakes.llm.URL},
		llmLogRepo:       repository.NewVsLLMLogRepository(repo),
		ttsProviderRepo:  repository.NewVsTTSProviderRepository(repo),
		itemRepo:         repository.NewVsVoiceDesignItemRepository(repo),
		fileStorage:      filestorage.NewLocal("storage"),
	}
	return svc, db
}

func processAllQueued(t *testing.T, svc *vsVoiceDesignService) {
	t.Helper()
	ctx := context.Background()
	for {
		item, err := svc.itemRepo.NextQueued(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			return
		}
		if err := svc.ProcessItem(ctx, item.ItemID); err != nil {
			t.Fatal(err)
		}
	}
}

func batchItemByName(t *testing.T, resp *v1.VoiceDesignBatchResponse, name string) v1.VoiceDesignBatchItem {
	t.Helper()
	for _, item := range resp.Items {
		if item.CharacterName == name {
			return item
		}
	}
	t.Fatalf("no batch item for %s: %+v", name, resp.Items)
	return v1.VoiceDesignBatchItem{}
}

func TestVoiceDesignBatchFlow(t *testing.T) {
	fakes := newFakeVoiceServers(t)
	svc, db := newBatchTestService(t, fakes)
	ctx := context.Background()

	xiao := &model.VsCharacter{ProjectID: 1, Name: "萧炎", Gender: "male", Level: "main", Description: "少年天才", TraitNotes: "加玛帝国的炼药师", Status: "0"}
	lin := &model.VsCharacter{ProjectID: 1, Name: "林动", Gender: "male", Level: "minor", Status: "0"}
	bound := &model.VsCharacter{ProjectID: 1, Name: "已绑定", Level: "minor", VoiceProfileID: u64(99), Status: "0"}
	merged := &model.VsCharacter{ProjectID: 1, Name: "被合并", Level: "minor", MergedIntoID: u64(1), Status: "1"}
	for _, c := range []*model.VsCharacter{xiao, lin, bound, merged} {
		if err := db.Create(c).Error; err != nil {
			t.Fatal(err)
		}
	}

	resp, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{
		CharacterIDs:     []uint64{xiao.CharacterID, lin.CharacterID, bound.CharacterID, merged.CharacterID},
		ConfirmImportant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 2 || len(resp.Skipped) != 2 || !resp.Running {
		t.Fatalf("start resp=%+v, want 2 items, 2 skipped, running", resp)
	}
	if !batchItemByName(t, resp, "萧炎").NeedConfirm || batchItemByName(t, resp, "林动").NeedConfirm {
		t.Fatal("want only the main character to need confirmation")
	}
	if _, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{CharacterIDs: []uint64{lin.CharacterID}}); err == nil {
		t.Fatal("expected error when a batch is already running")
	}

	processAllQueued(t, svc)

	resp, err = svc.GetBatch(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Running {
		t.Fatal("batch should be finished")
	}
	draft := batchItemByName(t, resp, "萧炎")
	if draft.Status != repository.VoiceDesignItemDraft || draft.Instruction == "" {
		t.Fatalf("萧炎 item=%+v, want draft", draft)
	}
	if got := loadTestCharacter(t, db, xiao.CharacterID); got.VoiceProfileID != nil {
		t.Fatal("draft should not bind a voice before confirmation")
	}
	linItem := batchItemByName(t, resp, "林动")
	if linItem.Status != repository.VoiceDesignItemApplied || linItem.VoiceProfileID == nil {
		t.Fatalf("林动 item=%+v, want applied", linItem)
	}
	if got := loadTestCharacter(t, db, lin.CharacterID); got.VoiceProfileID == nil || *got.VoiceProfileID != *linItem.VoiceProfileID {
		t.Fatalf("林动 voice_profile_id=%v, want %d", got.VoiceProfileID, *linItem.VoiceProfileID)
	}

	// 音色设计提示词带上了各章节特征
	joined := strings.Join(fakes.prompts, "\n")
	if !strings.Contains(joined, "加玛帝国的炼药师") {
		t.Fatal("voice design prompt should include trait notes")
	}

	audioPath, err := svc.ItemAudioPath(ctx, draft.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := svc.ApplyItem(ctx, draft.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	if filestorage.Exists(ctx, svc.fileStorage, audioPath) {
		t.Fatal("draft audio should be moved into reference audio dir after apply")
	}
	profile := model.VsVoiceProfile{}
	if err := db.First(&profile, applied.VoiceProfileID).Error; err != nil {
		t.Fatal(err)
	}
	if !filestorage.Exists(ctx, svc.fileStorage, profile.ReferenceAudioURL) || profile.ReferenceText != "我命由我不由天。" {
		t.Fatalf("profile=%+v, want saved reference audio and text", profile)
	}
	if got := loadTestCharacter(t, db, xiao.CharacterID); got.VoiceProfileID == nil || *got.VoiceProfileID != applied.VoiceProfileID {
		t.Fatal("萧炎 should be bound after confirming the draft")
	}
	if err := svc.RegenerateItem(ctx, draft.ItemID); err == nil {
		t.Fatal("applied item should not be regenerated")
	}
}

func TestVoiceDesignDraftRegenerateAndDiscard(t *testing.T) {
	fakes := newFakeVoiceServers(t)
	svc, db := newBatchTestService(t, fakes)
	ctx := context.Background()

	yao := &model.VsCharacter{ProjectID: 1, Name: "药老", Gender: "male", Level: "supporting", Status: "0"}
	if err := db.Create(yao).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{CharacterIDs: []uint64{yao.CharacterID}, ConfirmImportant: true}); err != nil {
		t.Fatal(err)
	}
	processAllQueued(t, svc)
	resp, _ := svc.GetBatch(ctx, 1)
	item := batchItemByName(t, resp, "药老")
	firstAudio, err := svc.ItemAudioPath(ctx, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}

	// 草稿在新批次里仍可见，且同一角色不会重复排队
	if _, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{CharacterIDs: []uint64{yao.CharacterID}, ConfirmImportant: true}); err == nil {
		t.Fatal("expected no items when the character already has a pending draft")
	}

	if err := svc.RegenerateItem(ctx, item.ItemID); err != nil {
		t.Fatal(err)
	}
	if filestorage.Exists(ctx, svc.fileStorage, firstAudio) {
		t.Fatal("old draft audio should be removed on regenerate")
	}
	processAllQueued(t, svc)
	secondAudio, err := svc.ItemAudioPath(ctx, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.DiscardItem(ctx, item.ItemID); err != nil {
		t.Fatal(err)
	}
	if filestorage.Exists(ctx, svc.fileStorage, secondAudio) {
		t.Fatal("draft audio should be removed on discard")
	}
	resp, _ = svc.GetBatch(ctx, 1)
	if got := batchItemByName(t, resp, "药老"); got.Status != repository.VoiceDesignItemDiscarded {
		t.Fatalf("status=%s, want discarded", got.Status)
	}
	if fakes.ttsCalls != 2 {
		t.Fatalf("tts calls=%d, want 2", fakes.ttsCalls)
	}
}

func TestVoiceDesignProcessSkipsCharacterBoundMeanwhile(t *testing.T) {
	fakes := newFakeVoiceServers(t)
	svc, db := newBatchTestService(t, fakes)
	ctx := context.Background()

	lin := &model.VsCharacter{ProjectID: 1, Name: "林动", Level: "minor", Status: "0"}
	if err := db.Create(lin).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{CharacterIDs: []uint64{lin.CharacterID}}); err != nil {
		t.Fatal(err)
	}
	// 排队期间用户手动绑定了声音
	db.Model(&model.VsCharacter{}).Where("character_id = ?", lin.CharacterID).Update("voice_profile_id", 7)
	processAllQueued(t, svc)

	resp, _ := svc.GetBatch(ctx, 1)
	if got := batchItemByName(t, resp, "林动"); got.Status != repository.VoiceDesignItemSkipped {
		t.Fatalf("status=%s, want skipped", got.Status)
	}
	if len(fakes.prompts) != 0 || fakes.ttsCalls != 0 {
		t.Fatal("skipped item should not call LLM or TTS")
	}
}

func TestVoiceDesignCancelBatch(t *testing.T) {
	fakes := newFakeVoiceServers(t)
	svc, db := newBatchTestService(t, fakes)
	ctx := context.Background()

	var ids []uint64
	for _, name := range []string{"甲", "乙"} {
		c := &model.VsCharacter{ProjectID: 1, Name: name, Level: "minor", Status: "0"}
		if err := db.Create(c).Error; err != nil {
			t.Fatal(err)
		}
		ids = append(ids, c.CharacterID)
	}
	if _, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{CharacterIDs: ids}); err != nil {
		t.Fatal(err)
	}
	n, err := svc.CancelBatch(ctx, 1)
	if err != nil || n != 2 {
		t.Fatalf("cancelled=%d err=%v, want 2", n, err)
	}
	resp, _ := svc.GetBatch(ctx, 1)
	if resp.Running {
		t.Fatal("batch should not be running after cancel")
	}
	// 取消后可以重新开始
	if _, err := svc.StartBatch(ctx, 1, &v1.VoiceDesignBatchRequest{CharacterIDs: ids}); err != nil {
		t.Fatal(err)
	}
}

func u64(v uint64) *uint64 { return &v }
