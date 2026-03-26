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

	"gorm.io/gorm"
)

type VsChapterSplitService interface {
	SplitChapter(ctx context.Context, chapterID uint64) (*v1.ChapterSplitResponse, error)
}

func NewVsChapterSplitService(
	service *Service,
	db *gorm.DB,
	chapterRepo repository.VsChapterRepository,
	projectRepo repository.VsProjectRepository,
	promptRepo repository.VsPromptTemplateRepository,
	providerRepo repository.VsLLMProviderRepository,
	sceneRepo repository.VsSceneRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	characterRepo repository.VsCharacterRepository,
	llmLogRepo repository.VsLLMLogRepository,
) VsChapterSplitService {
	return &vsChapterSplitService{
		Service:       service,
		db:            db,
		chapterRepo:   chapterRepo,
		projectRepo:   projectRepo,
		promptRepo:    promptRepo,
		providerRepo:  providerRepo,
		sceneRepo:     sceneRepo,
		segmentRepo:   segmentRepo,
		characterRepo: characterRepo,
		llmLogRepo:    llmLogRepo,
		llmClient:     llm.NewClient(),
	}
}

type vsChapterSplitService struct {
	*Service
	db            *gorm.DB
	chapterRepo   repository.VsChapterRepository
	projectRepo   repository.VsProjectRepository
	promptRepo    repository.VsPromptTemplateRepository
	providerRepo  repository.VsLLMProviderRepository
	sceneRepo     repository.VsSceneRepository
	segmentRepo   repository.VsScriptSegmentRepository
	characterRepo repository.VsCharacterRepository
	llmLogRepo    repository.VsLLMLogRepository
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
	Type            string `json:"type"`
	Content         string `json:"content"`
	Character       string `json:"character"`
	Emotion         string `json:"emotion"`
	EmotionStrength string `json:"emotion_strength"`
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

	var totalScenes int
	var totalSegments int
	var newCharacters int

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("chapter_id = ?", chapterID).Delete(&model.VsScriptSegment{}).Error; err != nil {
			return fmt.Errorf("清理旧片段失败: %w", err)
		}
		if err := tx.Unscoped().Where("chapter_id = ?", chapterID).Delete(&model.VsScene{}).Error; err != nil {
			return fmt.Errorf("清理旧场景失败: %w", err)
		}

		narratorID := resolveCharacterID(charMap, "旁白")
		if narratorID == nil {
			narrator := &model.VsCharacter{
				ProjectID:   project.ProjectID,
				Name:        "旁白",
				Level:       "supporting",
				Gender:      "unknown",
				Description: "系统内置旁白角色，用于旁白与描述类型片段的语音合成",
				Status:      "0",
				SortOrder:   9999,
				BaseModel: model.BaseModel{
					CreatedBy: loginName,
					DeptID:    deptID,
				},
			}
			if err := tx.Create(narrator).Error; err == nil {
				charMap[strings.ToLower("旁白")] = narrator.CharacterID
				narratorID = &narrator.CharacterID
				newCharacters++
			}
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
					if charID == nil {
						newChar := &model.VsCharacter{
							ProjectID: project.ProjectID,
							Name:      charName,
							Level:     "minor",
							Gender:    "unknown",
							Status:    "0",
							BaseModel: model.BaseModel{
								CreatedBy: loginName,
								DeptID:    deptID,
							},
						}
						if err := tx.Create(newChar).Error; err == nil {
							charMap[strings.ToLower(charName)] = newChar.CharacterID
							charID = &newChar.CharacterID
							newCharacters++
						}
					}
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
