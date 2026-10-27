package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/llmconfig"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/llm"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type VsChapterSplitService interface {
	SplitChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, error)
	BatchSplitChapters(ctx context.Context, projectID uint64, chapterIDs []uint64, loginName string) (*v1.BatchSplitResponse, error)
	// RunQueuedChapter 由 LLM worker 调用，只处理仍处于 queued 的章节；claimed 为 false 表示已取消或已被处理
	RunQueuedChapter(ctx context.Context, chapterID uint64) (result *v1.ChapterSplitResponse, claimed bool, err error)
	// CancelSplitQueue 取消项目中进行中的批量切割任务，返回取消的排队章节数
	CancelSplitQueue(ctx context.Context, projectID uint64) (int64, error)
	// RecoverableChapters 服务重启恢复：把任务中卡在 running 的章节重置为 queued，返回仍需处理的章节
	RecoverableChapters(ctx context.Context, task *model.VsGenerationTask) ([]uint64, error)
}

// 章节切割状态
const (
	splitStatusQueued    = "queued"
	splitStatusRunning   = "running"
	splitStatusDone      = "done"
	splitStatusFailed    = "failed"
	splitStatusCancelled = "cancelled"
)

var activeSplitStatuses = []string{splitStatusQueued, splitStatusRunning}

// splitIdleCondition 章节不在切割队列中（兼容新增字段前的 NULL）
const splitIdleCondition = "(split_status IS NULL OR split_status NOT IN ?)"

func NewVsChapterSplitService(
	service *Service,
	db *gorm.DB,
	rdb *redis.Client,
	chapterRepo repository.VsChapterRepository,
	projectRepo repository.VsProjectRepository,
	promptRepo repository.VsPromptTemplateRepository,
	llmResolver llmconfig.Resolver,
	sceneRepo repository.VsSceneRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	characterRepo repository.VsCharacterRepository,
	llmLogRepo repository.VsLLMLogRepository,
	taskRepo repository.VsGenerationTaskRepository,
) VsChapterSplitService {
	return &vsChapterSplitService{
		Service:       service,
		db:            db,
		rdb:           rdb,
		chapterRepo:   chapterRepo,
		projectRepo:   projectRepo,
		promptRepo:    promptRepo,
		llmResolver:   llmResolver,
		sceneRepo:     sceneRepo,
		segmentRepo:   segmentRepo,
		characterRepo: characterRepo,
		llmLogRepo:    llmLogRepo,
		taskRepo:      taskRepo,
		llmClient:     llm.NewClient(),
	}
}

type vsChapterSplitService struct {
	*Service
	db            *gorm.DB
	rdb           *redis.Client
	chapterRepo   repository.VsChapterRepository
	projectRepo   repository.VsProjectRepository
	promptRepo    repository.VsPromptTemplateRepository
	llmResolver   llmconfig.Resolver
	sceneRepo     repository.VsSceneRepository
	segmentRepo   repository.VsScriptSegmentRepository
	characterRepo repository.VsCharacterRepository
	llmLogRepo    repository.VsLLMLogRepository
	taskRepo      repository.VsGenerationTaskRepository
	llmClient     *llm.Client
}

type llmSplitResult struct {
	Scenes []llmScene `json:"scenes"`
}

type llmScene struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Segments    []llmSegment `json:"segments"`
}

type llmSegment struct {
	Type                 string `json:"type"`
	Content              string `json:"content"`
	Character            string `json:"character"`
	CharacterGender      string `json:"character_gender"`
	CharacterDescription string `json:"character_description"`
	Instruction          string `json:"instruction"`
}

// SplitChapter 同步切割单个章节；章节已在切割队列中时拒绝，避免同一章被并发切割、互相覆盖
func (s *vsChapterSplitService) SplitChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, error) {
	if _, err := s.chapterRepo.FindByID(ctx, chapterID); err != nil {
		return nil, fmt.Errorf("章节不存在")
	}
	res := s.db.WithContext(ctx).Model(&model.VsChapter{}).
		Where("chapter_id = ? AND "+splitIdleCondition, chapterID, activeSplitStatuses).
		Update("split_status", splitStatusRunning)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("该章节正在切割队列中，请等待完成或取消队列后再试")
	}
	result, err := s.doSplitChapter(ctx, chapterID)
	s.finishSplit(ctx, chapterID, err)
	return result, err
}

func (s *vsChapterSplitService) RunQueuedChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, bool, error) {
	res := s.db.WithContext(ctx).Model(&model.VsChapter{}).
		Where("chapter_id = ? AND split_status = ?", chapterID, splitStatusQueued).
		Update("split_status", splitStatusRunning)
	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, false, nil
	}
	result, err := s.doSplitChapter(ctx, chapterID)
	s.finishSplit(ctx, chapterID, err)
	return result, true, err
}

func (s *vsChapterSplitService) finishSplit(ctx context.Context, chapterID uint64, splitErr error) {
	status := splitStatusDone
	if splitErr != nil {
		status = splitStatusFailed
	}
	if err := s.db.WithContext(ctx).Model(&model.VsChapter{}).
		Where("chapter_id = ?", chapterID).
		Update("split_status", status).Error; err != nil {
		s.logger.Warn(fmt.Sprintf("更新章节 %d 切割状态失败: %v", chapterID, err))
	}
}

func (s *vsChapterSplitService) doSplitChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, error) {
	chapter, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("章节不存在")
	}
	if chapter.Content == "" {
		return nil, fmt.Errorf("章节内容为空")
	}

	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	resolved, err := s.llmResolver.ResolveByProject(ctx, project)
	if err != nil {
		return nil, err
	}

	promptContent, templateID, err := s.resolvePrompt(ctx, project)
	if err != nil {
		return nil, err
	}

	characters, _ := s.characterRepo.FindByProjectID(ctx, project.ProjectID)
	prompt := injectKnownCharacters(promptContent, buildKnownCharactersBlock(characters, chapter.Content))
	prompt = strings.ReplaceAll(prompt, "{{content}}", chapter.Content)

	start := time.Now()
	chatResp, err := s.llmClient.ChatCompletionStream(ctx, &llm.ChatRequest{
		BaseURL: resolved.BaseURL,
		APIKey:  resolved.APIKey,
		Model:   resolved.ModelName,
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
	})

	costTime := time.Since(start).Milliseconds()

	projectID := project.ProjectID
	providerID := resolved.ProviderID
	logEntry := &model.VsLLMLog{
		ProjectID:    &projectID,
		ProviderID:   providerID,
		TemplateID:   templateID,
		ModelName:    resolved.ModelName,
		InputTokens:  0,
		OutputTokens: 0,
		CostTime:     costTime,
		Status:       0,
		BaseModel: model.BaseModel{
			CreatedBy: s.getLoginName(ctx),
		},
	}

	if err != nil {
		logEntry.Status = 1
		logEntry.ErrorMessage = err.Error()
		logEntry.InputSummary = truncate(prompt, 500)
		_ = s.llmLogRepo.Create(ctx, logEntry)
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	logEntry.InputTokens = chatResp.InputTokens
	logEntry.OutputTokens = chatResp.OutputTokens
	logEntry.InputSummary = truncate(prompt, 500)
	logEntry.OutputSummary = truncate(chatResp.Content, 500)
	_ = s.llmLogRepo.Create(ctx, logEntry)

	result, err := s.parseLLMResponse(chatResp.Content)
	if err != nil {
		return nil, fmt.Errorf("解析 LLM 返回结果失败: %w", err)
	}

	charMap := buildCharacterMap(characters)

	loginName := s.getLoginName(ctx)
	deptID := s.getDeptID(ctx)

	// 在事务之前，先收集 LLM 结果中所有角色名，预创建不存在的角色
	newCharacters, err := s.ensureCharacters(ctx, result, project.ProjectID, characters, charMap, loginName, deptID)
	if err != nil {
		return nil, fmt.Errorf("预创建角色失败: %w", err)
	}

	narratorID := resolveCharacterID(charMap, narratorName)

	var totalScenes int
	var totalSegments int

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("chapter_id = ?", chapterID).Delete(&model.VsScriptSegment{}).Error; err != nil {
			return fmt.Errorf("清理旧片段失败: %w", err)
		}
		if err := tx.Unscoped().Where("chapter_id = ?", chapterID).Delete(&model.VsScene{}).Error; err != nil {
			return fmt.Errorf("清理旧场景失败: %w", err)
		}

		segmentNum := 0
		for i, sc := range result.Scenes {
			scene := &model.VsScene{
				ChapterID:   chapterID,
				SceneNum:    i + 1,
				Title:       sc.Title,
				Description: sc.Description,
				Status:      "parsed",
				BaseModel: model.BaseModel{
					CreatedBy: loginName,
					DeptID:    deptID,
				},
			}
			if err := tx.Create(scene).Error; err != nil {
				return fmt.Errorf("创建场景失败: %w", err)
			}
			totalScenes++

			segments := make([]*model.VsScriptSegment, 0, len(sc.Segments))
			for _, seg := range sc.Segments {
				segmentNum++

				segType := normalizeSegmentType(seg.Type)
				var charID *uint64
				charName := strings.TrimSpace(seg.Character)

				if segType == "narration" || segType == "description" {
					charID = narratorID
				} else if charName != "" {
					charID = resolveCharacterID(charMap, charName)
				}

				segment := &model.VsScriptSegment{
					ProjectID:       project.ProjectID,
					SceneID:         scene.SceneID,
					ChapterID:       chapterID,
					SegmentNum:      segmentNum,
					SegmentType:     segType,
					Content:         seg.Content,
					OriginalContent: seg.Content,
					CharacterID:     charID,
					Instruction:     normalizeInstruction(seg.Instruction),
					Status:          "raw",
					Version:         1,
					BaseModel: model.BaseModel{
						CreatedBy: loginName,
						DeptID:    deptID,
					},
				}
				segments = append(segments, segment)
			}

			if len(segments) > 0 {
				if err := tx.CreateInBatches(segments, 100).Error; err != nil {
					return fmt.Errorf("创建片段失败: %w", err)
				}
				totalSegments += len(segments)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &v1.ChapterSplitResponse{
		SceneCount:    totalScenes,
		SegmentCount:  totalSegments,
		NewCharacters: newCharacters,
		InputTokens:   chatResp.InputTokens,
		OutputTokens:  chatResp.OutputTokens,
	}, nil
}

func (s *vsChapterSplitService) BatchSplitChapters(ctx context.Context, projectID uint64, chapterIDs []uint64, loginName string) (*v1.BatchSplitResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	if _, err := s.llmResolver.ResolveByProject(ctx, project); err != nil {
		return nil, err
	}
	if _, _, err := s.resolvePrompt(ctx, project); err != nil {
		return nil, err
	}

	chapterIDs = uniqueIDs(chapterIDs)
	if len(chapterIDs) == 0 {
		return nil, fmt.Errorf("请选择要切割的章节")
	}
	for _, chID := range chapterIDs {
		ch, err := s.chapterRepo.FindByID(ctx, chID)
		if err != nil {
			return nil, fmt.Errorf("章节 %d 不存在", chID)
		}
		if ch.ProjectID != projectID {
			return nil, fmt.Errorf("章节 %d 不属于当前项目", chID)
		}
		if ch.Content == "" {
			return nil, fmt.Errorf("章节「%s」内容为空", ch.Title)
		}
	}

	task := &model.VsGenerationTask{
		ProjectID:    projectID,
		TaskType:     "chapter_split",
		Status:       "running",
		TotalBatches: len(chapterIDs),
		SegmentIDs:   model.Uint64Array(chapterIDs),
		BaseModel: model.BaseModel{
			CreatedBy: loginName,
		},
	}
	// 在同一事务里把章节标为 queued 并创建任务：已在队列中的章节不会被重复入队
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.VsChapter{}).
			Where("chapter_id IN ? AND "+splitIdleCondition, chapterIDs, activeSplitStatuses).
			Update("split_status", splitStatusQueued)
		if res.Error != nil {
			return res.Error
		}
		if int(res.RowsAffected) != len(chapterIDs) {
			var busy []model.VsChapter
			tx.Select("title").Where("chapter_id IN ? AND split_status IN ?", chapterIDs, activeSplitStatuses).Find(&busy)
			titles := make([]string, 0, len(busy))
			for _, ch := range busy {
				titles = append(titles, ch.Title)
			}
			return fmt.Errorf("以下章节已在切割队列中，请等待完成或取消后再试：%s", strings.Join(titles, "、"))
		}
		if err := tx.Create(task).Error; err != nil {
			return fmt.Errorf("创建批量切割任务失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.taskRepo.SetStarted(ctx, task.TaskID); err != nil {
		_ = s.taskRepo.SetFailed(ctx, task.TaskID, fmt.Sprintf("更新任务开始状态失败: %v", err))
		return nil, fmt.Errorf("更新任务开始状态失败: %w", err)
	}

	pipe := s.rdb.TxPipeline()
	for _, chID := range chapterIDs {
		msg := fmt.Sprintf("%d:%d", task.TaskID, chID)
		pipe.LPush(ctx, llmQueueKey, msg)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		_ = s.taskRepo.SetFailed(ctx, task.TaskID, fmt.Sprintf("任务入队失败: %v", err))
		s.db.WithContext(ctx).Model(&model.VsChapter{}).
			Where("chapter_id IN ? AND split_status = ?", chapterIDs, splitStatusQueued).
			Update("split_status", splitStatusFailed)
		return nil, fmt.Errorf("批量入队失败: %w", err)
	}

	return &v1.BatchSplitResponse{
		TaskID: task.TaskID,
		Total:  len(chapterIDs),
	}, nil
}

func (s *vsChapterSplitService) resolvePrompt(ctx context.Context, project *model.VsProject) (string, *uint64, error) {
	if project.PromptTemplateIDs != nil {
		if tid, ok := project.PromptTemplateIDs["dialogue_parse"]; ok && tid > 0 {
			tmpl, err := s.promptRepo.FindByID(ctx, tid)
			if err == nil && tmpl.Status == "0" {
				return tmpl.Content, &tmpl.TemplateID, nil
			}
		}
	}

	templates, err := s.promptRepo.FindByType(ctx, "dialogue_parse")
	if err != nil || len(templates) == 0 {
		return "", nil, fmt.Errorf("未找到 dialogue_parse 类型的提示词模板")
	}
	for _, t := range templates {
		if t.IsSystem == "1" && t.Status == "0" {
			return t.Content, &t.TemplateID, nil
		}
	}
	return templates[0].Content, &templates[0].TemplateID, nil
}

func (s *vsChapterSplitService) parseLLMResponse(content string) (*llmSplitResult, error) {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```") {
		lines := strings.SplitN(content, "\n", 2)
		if len(lines) > 1 {
			content = lines[1]
		}
		if idx := strings.LastIndex(content, "```"); idx >= 0 {
			content = content[:idx]
		}
		content = strings.TrimSpace(content)
	}

	var result llmSplitResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w\n原始内容: %s", err, truncate(content, 300))
	}
	if len(result.Scenes) == 0 {
		return nil, fmt.Errorf("LLM 返回的场景列表为空")
	}
	return &result, nil
}

func (s *vsChapterSplitService) getLoginName(ctx context.Context) string {
	if v := ctx.Value("login_name"); v != nil {
		return v.(string)
	}
	return "system"
}

func (s *vsChapterSplitService) getDeptID(ctx context.Context) uint {
	if v := ctx.Value("dept_id"); v != nil {
		return v.(uint)
	}
	return 0
}

func buildCharacterMap(characters []*model.VsCharacter) map[string]uint64 {
	m := make(map[string]uint64)
	for _, c := range characters {
		m[strings.ToLower(c.Name)] = c.CharacterID
		for _, alias := range c.Aliases {
			m[strings.ToLower(alias)] = c.CharacterID
		}
	}
	return m
}

func resolveCharacterID(charMap map[string]uint64, name string) *uint64 {
	if name == "" {
		return nil
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if id, ok := charMap[name]; ok {
		return &id
	}
	return nil
}

func normalizeSegmentType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "dialogue", "narration", "monologue", "description":
		return t
	default:
		return "narration"
	}
}

// normalizeInstruction 演绎指令只保留一句短描述，超长截断，避免把整段分析塞给 TTS
func normalizeInstruction(s string) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) > maxInstructionRunes {
		runes = runes[:maxInstructionRunes]
	}
	return string(runes)
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

func (s *vsChapterSplitService) CancelSplitQueue(ctx context.Context, projectID uint64) (int64, error) {
	tasks, err := s.taskRepo.FindActiveByProjectID(ctx, projectID)
	if err != nil {
		return 0, err
	}
	var cancelled int64
	for _, task := range tasks {
		if task.TaskType != "chapter_split" {
			continue
		}
		// 先结束任务：worker 取到残留消息时会因任务已取消而跳过
		if err := s.taskRepo.SetCancelled(ctx, task.TaskID, "cancelled by user"); err != nil {
			return cancelled, err
		}
		if len(task.SegmentIDs) > 0 {
			// 正在切割的章节会继续完成，只取消还在排队的
			res := s.db.WithContext(ctx).Model(&model.VsChapter{}).
				Where("chapter_id IN ? AND split_status = ?", []uint64(task.SegmentIDs), splitStatusQueued).
				Update("split_status", splitStatusCancelled)
			if res.Error != nil {
				return cancelled, res.Error
			}
			cancelled += res.RowsAffected
		}
		if err := removeTaskQueueMessages(ctx, s.rdb, llmQueueKey, task.TaskID); err != nil {
			s.logger.Warn(fmt.Sprintf("清理切割任务 %d 的队列消息失败: %v", task.TaskID, err))
		}
	}
	return cancelled, nil
}

func (s *vsChapterSplitService) RecoverableChapters(ctx context.Context, task *model.VsGenerationTask) ([]uint64, error) {
	ids := []uint64(task.SegmentIDs)
	if len(ids) == 0 {
		return nil, nil
	}
	var chapters []model.VsChapter
	if err := s.db.WithContext(ctx).Select("chapter_id, split_status").
		Where("chapter_id IN ?", ids).Find(&chapters).Error; err != nil {
		return nil, err
	}
	statusByID := make(map[uint64]string, len(chapters))
	tracked := false
	for _, ch := range chapters {
		statusByID[ch.ChapterID] = ch.SplitStatus
		if ch.SplitStatus != "" {
			tracked = true
		}
	}

	var remaining []uint64
	if !tracked {
		// 新增切割状态字段之前创建的任务：章节没有状态可查，退回按已处理数量跳过
		processed := task.CompletedBatches + task.FailedBatches
		for i, id := range ids {
			if _, exists := statusByID[id]; exists && i >= processed {
				remaining = append(remaining, id)
			}
		}
	} else {
		for _, id := range ids {
			if st := statusByID[id]; st == splitStatusQueued || st == splitStatusRunning {
				remaining = append(remaining, id)
			}
		}
	}
	if len(remaining) == 0 {
		return nil, nil
	}
	// 上次退出时正在切割的章节重新排队
	if err := s.db.WithContext(ctx).Model(&model.VsChapter{}).
		Where("chapter_id IN ?", remaining).
		Update("split_status", splitStatusQueued).Error; err != nil {
		return nil, err
	}
	return remaining, nil
}
