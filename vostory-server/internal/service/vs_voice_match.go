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
)

type VsVoiceMatchService interface {
	MatchVoices(ctx context.Context, projectID uint64) (*v1.VoiceMatchResponse, error)
}

func NewVsVoiceMatchService(
	service *Service,
	projectRepo repository.VsProjectRepository,
	characterRepo repository.VsCharacterRepository,
	voiceProfileRepo repository.VsVoiceProfileRepository,
	promptRepo repository.VsPromptTemplateRepository,
	providerRepo repository.VsLLMProviderRepository,
	llmLogRepo repository.VsLLMLogRepository,
) VsVoiceMatchService {
	return &vsVoiceMatchService{
		Service:          service,
		projectRepo:      projectRepo,
		characterRepo:    characterRepo,
		voiceProfileRepo: voiceProfileRepo,
		promptRepo:       promptRepo,
		providerRepo:     providerRepo,
		llmLogRepo:       llmLogRepo,
		llmClient:        llm.NewClient(),
	}
}

type vsVoiceMatchService struct {
	*Service
	projectRepo      repository.VsProjectRepository
	characterRepo    repository.VsCharacterRepository
	voiceProfileRepo repository.VsVoiceProfileRepository
	promptRepo       repository.VsPromptTemplateRepository
	providerRepo     repository.VsLLMProviderRepository
	llmLogRepo       repository.VsLLMLogRepository
	llmClient        *llm.Client
}

type llmMatchResult struct {
	Matches []llmMatchItem `json:"matches"`
}

type llmMatchItem struct {
	CharacterID    uint64  `json:"character_id"`
	VoiceProfileID *uint64 `json:"voice_profile_id"`
	Reason         string  `json:"reason"`
}

type characterInput struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Description string `json:"description"`
}

type voiceInput struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Description string `json:"description"`
}

func (s *vsVoiceMatchService) MatchVoices(ctx context.Context, projectID uint64) (*v1.VoiceMatchResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
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

	allCharacters, err := s.characterRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	var unboundCharacters []*model.VsCharacter
	skippedCount := 0
	for _, c := range allCharacters {
		if c.VoiceProfileID != nil && *c.VoiceProfileID > 0 {
			skippedCount++
		} else {
			unboundCharacters = append(unboundCharacters, c)
		}
	}

	if len(unboundCharacters) == 0 {
		return &v1.VoiceMatchResponse{
			MatchedCount: 0,
			SkippedCount: skippedCount,
			FailedCount:  0,
		}, nil
	}

	voiceProfiles, err := s.voiceProfileRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("查询声音配置失败: %w", err)
	}
	if len(voiceProfiles) == 0 {
		return nil, fmt.Errorf("项目没有可用的声音配置，请先添加声音配置")
	}

	characters := make([]characterInput, len(unboundCharacters))
	characterMap := make(map[uint64]*model.VsCharacter, len(unboundCharacters))
	for i, c := range unboundCharacters {
		characters[i] = characterInput{
			ID:          c.CharacterID,
			Name:        c.Name,
			Gender:      c.Gender,
			Description: c.Description,
		}
		characterMap[c.CharacterID] = c
	}

	voices := make([]voiceInput, len(voiceProfiles))
	voiceMap := make(map[uint64]*model.VsVoiceProfile, len(voiceProfiles))
	for i, v := range voiceProfiles {
		voices[i] = voiceInput{
			ID:          v.VoiceProfileID,
			Name:        v.Name,
			Gender:      v.Gender,
			Description: v.Description,
		}
		voiceMap[v.VoiceProfileID] = v
	}

	charsJSON, _ := json.Marshal(characters)
	voicesJSON, _ := json.Marshal(voices)

	prompt := strings.ReplaceAll(promptContent, "{{characters}}", string(charsJSON))
	prompt = strings.ReplaceAll(prompt, "{{voices}}", string(voicesJSON))

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

	logEntry := &model.VsLLMLog{
		ProjectID:  &projectID,
		ProviderID: provider.ProviderID,
		TemplateID: templateID,
		ModelName:  provider.DefaultModel,
		CostTime:   costTime,
		Status:     0,
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

	matchResult, err := s.parseMatchResponse(chatResp.Content)
	if err != nil {
		return nil, fmt.Errorf("解析 LLM 返回结果失败: %w", err)
	}

	matchedCount := 0
	failedCount := 0
	var details []v1.VoiceMatchDetail

	for _, m := range matchResult.Matches {
		char, ok := characterMap[m.CharacterID]
		if !ok {
			continue
		}

		if m.VoiceProfileID == nil || *m.VoiceProfileID == 0 {
			failedCount++
			continue
		}

		voice, ok := voiceMap[*m.VoiceProfileID]
		if !ok {
			failedCount++
			continue
		}

		char.VoiceProfileID = m.VoiceProfileID
		char.UpdatedBy = s.getLoginName(ctx)
		if err := s.characterRepo.Update(ctx, char); err != nil {
			s.logger.Warn(fmt.Sprintf("更新角色 %s 声音配置失败: %v", char.Name, err))
			failedCount++
			continue
		}

		matchedCount++
		details = append(details, v1.VoiceMatchDetail{
			CharacterID:    char.CharacterID,
			CharacterName:  char.Name,
			VoiceProfileID: voice.VoiceProfileID,
			VoiceName:      voice.Name,
			Reason:         m.Reason,
		})
	}

	unmatchedInResult := make(map[uint64]bool)
	for _, m := range matchResult.Matches {
		unmatchedInResult[m.CharacterID] = true
	}
	for id := range characterMap {
		if !unmatchedInResult[id] {
			failedCount++
		}
	}

	return &v1.VoiceMatchResponse{
		MatchedCount: matchedCount,
		SkippedCount: skippedCount,
		FailedCount:  failedCount,
		Details:      details,
		InputTokens:  chatResp.InputTokens,
		OutputTokens: chatResp.OutputTokens,
	}, nil
}

func (s *vsVoiceMatchService) resolveProvider(ctx context.Context, project *model.VsProject) (*model.VsLLMProvider, error) {
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

func (s *vsVoiceMatchService) resolvePrompt(ctx context.Context, project *model.VsProject) (string, *uint64, error) {
	if project.PromptTemplateIDs != nil {
		if tid, ok := project.PromptTemplateIDs["voice_match"]; ok && tid > 0 {
			tmpl, err := s.promptRepo.FindByID(ctx, tid)
			if err == nil && tmpl.Status == "0" {
				return tmpl.Content, &tmpl.TemplateID, nil
			}
		}
	}

	templates, err := s.promptRepo.FindByType(ctx, "voice_match")
	if err != nil || len(templates) == 0 {
		return "", nil, fmt.Errorf("未找到 voice_match 类型的提示词模板")
	}
	for _, t := range templates {
		if t.IsSystem == "1" && t.Status == "0" {
			return t.Content, &t.TemplateID, nil
		}
	}
	return templates[0].Content, &templates[0].TemplateID, nil
}

func (s *vsVoiceMatchService) parseMatchResponse(content string) (*llmMatchResult, error) {
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

	var result llmMatchResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w\n原始内容: %s", err, truncate(content, 200))
	}
	return &result, nil
}

func (s *vsVoiceMatchService) getLoginName(ctx context.Context) string {
	if v := ctx.Value("login_name"); v != nil {
		return v.(string)
	}
	return "system"
}
