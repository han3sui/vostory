package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/llm"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type VsChapterSplitService interface {
	SplitChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, error)
	BatchSplitChapters(ctx context.Context, projectID uint64, chapterIDs []uint64, loginName string) (*v1.BatchSplitResponse, error)
}

func NewVsChapterSplitService(
	service *Service,
	db *gorm.DB,
	rdb *redis.Client,
	chapterRepo repository.VsChapterRepository,
	projectRepo repository.VsProjectRepository,
	promptRepo repository.VsPromptTemplateRepository,
	providerRepo repository.VsLLMProviderRepository,
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
		providerRepo:  providerRepo,
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
	providerRepo  repository.VsLLMProviderRepository
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
	Emotion              string `json:"emotion"`
	EmotionStrength      string `json:"emotion_strength"`
}

func (s *vsChapterSplitService) SplitChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, error) {
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

	provider, err := s.resolveProvider(ctx, project)
	if err != nil {
		return nil, err
	}

	promptContent, templateID, err := s.resolvePrompt(ctx, project)
	if err != nil {
		return nil, err
	}

	prompt := strings.ReplaceAll(promptContent, "{{content}}", chapter.Content)

	start := time.Now()
	chatResp, err := s.llmClient.ChatCompletion(ctx, &llm.ChatRequest{
		BaseURL: provider.APIBaseURL,
		APIKey:  provider.APIKey,
		Model:   provider.DefaultModel,
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
		CustomParams:   provider.CustomParams,
	})

	costTime := time.Since(start).Milliseconds()

	projectID := project.ProjectID
	providerID := provider.ProviderID
	logEntry := &model.VsLLMLog{
		ProjectID:    &projectID,
		ProviderID:   providerID,
		TemplateID:   templateID,
		ModelName:    provider.DefaultModel,
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

	characters, _ := s.characterRepo.FindByProjectID(ctx, project.ProjectID)
	charMap := buildCharacterMap(characters)

	loginName := s.getLoginName(ctx)
	deptID := s.getDeptID(ctx)

	// 在事务之前，先收集 LLM 结果中所有角色名，预创建不存在的角色
	newCharacters, err := s.ensureCharacters(ctx, result, project.ProjectID, charMap, loginName, deptID)
	if err != nil {
		return nil, fmt.Errorf("预创建角色失败: %w", err)
	}

	narratorID := resolveCharacterID(charMap, "旁白")

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
					SceneID:         scene.SceneID,
					ChapterID:       chapterID,
					SegmentNum:      segmentNum,
					SegmentType:     segType,
					Content:         seg.Content,
					OriginalContent: seg.Content,
					CharacterID:     charID,
					EmotionType:     normalizeEmotion(seg.Emotion),
					EmotionStrength: normalizeStrength(seg.EmotionStrength),
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

	if _, err := s.resolveProvider(ctx, project); err != nil {
		return nil, err
	}
	if _, _, err := s.resolvePrompt(ctx, project); err != nil {
		return nil, err
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
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("创建批量切割任务失败: %w", err)
	}

	if err := s.taskRepo.SetStarted(ctx, task.TaskID); err != nil {
		_ = s.taskRepo.SetFailed(ctx, task.TaskID, fmt.Sprintf("更新任务开始状态失败: %v", err))
		return nil, fmt.Errorf("更新任务开始状态失败: %w", err)
	}

	pipe := s.rdb.TxPipeline()
	for _, chID := range chapterIDs {
		msg := fmt.Sprintf("%d:%d", task.TaskID, chID)
		pipe.LPush(ctx, "vs:llm:queue", msg)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		_ = s.taskRepo.SetFailed(ctx, task.TaskID, fmt.Sprintf("任务入队失败: %v", err))
		return nil, fmt.Errorf("批量入队失败: %w", err)
	}

	return &v1.BatchSplitResponse{
		TaskID: task.TaskID,
		Total:  len(chapterIDs),
	}, nil
}

func (s *vsChapterSplitService) resolveProvider(ctx context.Context, project *model.VsProject) (*model.VsLLMProvider, error) {
	if project.LLMProviderID == nil || *project.LLMProviderID == 0 {
		return nil, fmt.Errorf("项目未配置 LLM 提供商，请先在项目设置中绑定")
	}
	provider, err := s.providerRepo.FindByID(ctx, *project.LLMProviderID)
	if err != nil {
		return nil, fmt.Errorf("LLM 提供商不存在")
	}
	if provider.Status != "0" {
		return nil, fmt.Errorf("LLM 提供商已停用")
	}
	return provider, nil
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

func normalizeEmotion(e string) string {
	e = strings.ToLower(strings.TrimSpace(e))
	switch e {
	case "neutral", "happy", "sad", "angry", "fear", "surprise", "disgust":
		return e
	default:
		return "neutral"
	}
}

func normalizeStrength(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "light", "medium", "strong":
		return s
	default:
		return "medium"
	}
}

// ensureCharacters 从 LLM 结果中收集所有角色名（含"旁白"），
// 与 charMap 对比后，仅创建不存在的角色，并将新角色 ID 写回 charMap。
// 返回新创建的角色数量。
func (s *vsChapterSplitService) ensureCharacters(
	ctx context.Context,
	result *llmSplitResult,
	projectID uint64,
	charMap map[string]uint64,
	loginName string,
	deptID uint,
) (int, error) {
	type charInfo struct {
		name        string
		gender      string
		description string
		level       string
		sortOrder   int
	}

	needed := make(map[string]*charInfo)

	if resolveCharacterID(charMap, "旁白") == nil {
		needed["旁白"] = &charInfo{
			name:        "旁白",
			gender:      "unknown",
			description: "系统内置旁白角色，用于旁白与描述类型片段的语音合成",
			level:       "supporting",
			sortOrder:   9999,
		}
	}

	for _, sc := range result.Scenes {
		for _, seg := range sc.Segments {
			segType := normalizeSegmentType(seg.Type)
			if segType == "narration" || segType == "description" {
				continue
			}
			charName := strings.TrimSpace(seg.Character)
			if charName == "" {
				continue
			}
			key := strings.ToLower(charName)
			if _, exists := charMap[key]; exists {
				continue
			}
			if _, queued := needed[key]; queued {
				continue
			}
			needed[key] = &charInfo{
				name:        charName,
				gender:      normalizeGender(seg.CharacterGender),
				description: strings.TrimSpace(seg.CharacterDescription),
				level:       "minor",
			}
		}
	}

	if len(needed) == 0 {
		return 0, nil
	}

	// 一次性查出库中已存在的角色（含被停用和软删除的，与唯一索引覆盖范围一致）
	neededNames := make([]string, 0, len(needed))
	for _, info := range needed {
		neededNames = append(neededNames, info.name)
	}
	var existingChars []*model.VsCharacter
	if err := s.db.WithContext(ctx).Unscoped().
		Where("project_id = ? AND name IN ?", projectID, neededNames).
		Find(&existingChars).Error; err == nil {
		for _, ec := range existingChars {
			key := strings.ToLower(ec.Name)
			charMap[key] = ec.CharacterID
			delete(needed, key)
		}
	}

	if len(needed) == 0 {
		return 0, nil
	}

	// 批量创建真正不存在的角色
	toCreate := make([]*model.VsCharacter, 0, len(needed))
	for _, info := range needed {
		toCreate = append(toCreate, &model.VsCharacter{
			ProjectID:   projectID,
			Name:        info.name,
			Level:       info.level,
			Gender:      info.gender,
			Description: info.description,
			Status:      "0",
			SortOrder:   info.sortOrder,
			BaseModel: model.BaseModel{
				CreatedBy: loginName,
				DeptID:    deptID,
			},
		})
	}
	if err := s.db.WithContext(ctx).CreateInBatches(toCreate, 100).Error; err != nil {
		return 0, fmt.Errorf("批量创建角色失败: %w", err)
	}
	for _, c := range toCreate {
		charMap[strings.ToLower(c.Name)] = c.CharacterID
	}
	return len(toCreate), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
