package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/llm"
)

const maxExtractChars = 30000

type VsCharacterExtractService interface {
	ExtractCharacters(ctx context.Context, projectID uint64) (*v1.CharacterExtractResponse, error)
	ExtractFromText(ctx context.Context, req *v1.CharacterExtractFromTextRequest) (*v1.CharacterExtractResponse, error)
}

func NewVsCharacterExtractService(
	service *Service,
	projectRepo repository.VsProjectRepository,
	chapterRepo repository.VsChapterRepository,
	characterRepo repository.VsCharacterRepository,
	promptRepo repository.VsPromptTemplateRepository,
	providerRepo repository.VsLLMProviderRepository,
	llmLogRepo repository.VsLLMLogRepository,
) VsCharacterExtractService {
	return &vsCharacterExtractService{
		Service:       service,
		projectRepo:   projectRepo,
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
		promptRepo:    promptRepo,
		providerRepo:  providerRepo,
		llmLogRepo:    llmLogRepo,
		llmClient:     llm.NewClient(),
	}
}

type vsCharacterExtractService struct {
	*Service
	projectRepo   repository.VsProjectRepository
	chapterRepo   repository.VsChapterRepository
	characterRepo repository.VsCharacterRepository
	promptRepo    repository.VsPromptTemplateRepository
	providerRepo  repository.VsLLMProviderRepository
	llmLogRepo    repository.VsLLMLogRepository
	llmClient     *llm.Client
}

type llmExtractResult struct {
	Characters []llmCharacter `json:"characters"`
}

type llmCharacter struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases"`
	Gender      string   `json:"gender"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
}

func (s *vsCharacterExtractService) ExtractCharacters(ctx context.Context, projectID uint64) (*v1.CharacterExtractResponse, error) {
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

	sampleText, err := s.buildSampleText(ctx, projectID)
	if err != nil {
		return nil, err
	}

	prompt := strings.ReplaceAll(promptContent, "{{content}}", sampleText)

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

	result, err := s.parseLLMResponse(chatResp.Content)
	if err != nil {
		return nil, fmt.Errorf("解析 LLM 返回结果失败: %w", err)
	}

	resp, err := s.saveCharacters(ctx, projectID, result)
	if err != nil {
		return nil, err
	}
	resp.InputTokens = chatResp.InputTokens
	resp.OutputTokens = chatResp.OutputTokens
	return resp, nil
}

func (s *vsCharacterExtractService) ExtractFromText(ctx context.Context, req *v1.CharacterExtractFromTextRequest) (*v1.CharacterExtractResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
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

	prompt := strings.ReplaceAll(promptContent, "{{content}}", req.Text)

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
		ProjectID:  &req.ProjectID,
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

	result, err := s.parseLLMResponse(chatResp.Content)
	if err != nil {
		return nil, fmt.Errorf("解析 LLM 返回结果失败: %w", err)
	}

	return s.saveCharacters(ctx, req.ProjectID, result)
}

func (s *vsCharacterExtractService) saveCharacters(ctx context.Context, projectID uint64, result *llmExtractResult) (*v1.CharacterExtractResponse, error) {
	existing, _ := s.characterRepo.FindByProjectID(ctx, projectID)
	existingMap := make(map[string]*model.VsCharacter)
	for _, c := range existing {
		existingMap[strings.ToLower(c.Name)] = c
		for _, alias := range c.Aliases {
			existingMap[strings.ToLower(alias)] = c
		}
	}

	loginName := s.getLoginName(ctx)
	deptID := s.getDeptID(ctx)
	newCount := 0
	updatedCount := 0
	skippedCount := 0

	for i, ch := range result.Characters {
		if ch.Name == "" {
			continue
		}

		if existChar, ok := existingMap[strings.ToLower(ch.Name)]; ok {
			fields := map[string]interface{}{}
			if existChar.Gender == "unknown" && normalizeGender(ch.Gender) != "unknown" {
				fields["gender"] = normalizeGender(ch.Gender)
			}
			if existChar.Description == "" && strings.TrimSpace(ch.Description) != "" {
				fields["description"] = strings.TrimSpace(ch.Description)
			}
			if len(existChar.Aliases) == 0 && len(ch.Aliases) > 0 {
				fields["aliases"] = model.StringList(ch.Aliases)
			}
			if existChar.Level == "minor" && normalizeLevel(ch.Level) != "minor" {
				fields["level"] = normalizeLevel(ch.Level)
			}
			if len(fields) > 0 {
				fields["updated_by"] = loginName
				if err := s.characterRepo.UpdateFields(ctx, existChar.CharacterID, fields); err != nil {
					s.logger.Warn(fmt.Sprintf("更新角色 %s 失败: %v", ch.Name, err))
				} else {
					updatedCount++
				}
			} else {
				skippedCount++
			}
			continue
		}

		character := &model.VsCharacter{
			ProjectID:   projectID,
			Name:        ch.Name,
			Aliases:     model.StringList(ch.Aliases),
			Gender:      normalizeGender(ch.Gender),
			Description: ch.Description,
			Level:       normalizeLevel(ch.Level),
			SortOrder:   i,
			Status:      "0",
			BaseModel: model.BaseModel{
				CreatedBy: loginName,
				DeptID:    deptID,
			},
		}

		if err := s.characterRepo.Create(ctx, character); err != nil {
			s.logger.Warn(fmt.Sprintf("创建角色 %s 失败: %v", ch.Name, err))
			skippedCount++
			continue
		}

		existingMap[strings.ToLower(ch.Name)] = character
		for _, alias := range ch.Aliases {
			existingMap[strings.ToLower(alias)] = character
		}
		newCount++
	}

	return &v1.CharacterExtractResponse{
		ExtractedCount: len(result.Characters),
		NewCount:       newCount,
		UpdatedCount:   updatedCount,
		SkippedCount:   skippedCount,
	}, nil
}

func (s *vsCharacterExtractService) buildSampleText(ctx context.Context, projectID uint64) (string, error) {
	chapters, err := s.chapterRepo.FindByProjectID(ctx, projectID)
	if err != nil || len(chapters) == 0 {
		return "", fmt.Errorf("项目没有章节数据")
	}

	// FindByProjectID uses Select without content, need full content
	var sb strings.Builder
	totalChars := 0

	for _, ch := range chapters {
		full, err := s.chapterRepo.FindByID(ctx, ch.ChapterID)
		if err != nil || full.Content == "" {
			continue
		}

		chTitle := full.Title
		if chTitle == "" {
			chTitle = fmt.Sprintf("第%d章", full.ChapterNum)
		}

		chContent := full.Content
		chLen := utf8.RuneCountInString(chContent)

		if totalChars+chLen > maxExtractChars {
			remaining := maxExtractChars - totalChars
			if remaining > 500 {
				runes := []rune(chContent)
				if remaining < len(runes) {
					chContent = string(runes[:remaining])
				}
				sb.WriteString(fmt.Sprintf("\n\n【%s】\n%s", chTitle, chContent))
			}
			break
		}

		sb.WriteString(fmt.Sprintf("\n\n【%s】\n%s", chTitle, chContent))
		totalChars += chLen
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("章节内容为空")
	}
	return sb.String(), nil
}

func (s *vsCharacterExtractService) resolveProvider(ctx context.Context, project *model.VsProject) (*model.VsLLMProvider, error) {
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

func (s *vsCharacterExtractService) resolvePrompt(ctx context.Context, project *model.VsProject) (string, *uint64, error) {
	if project.PromptTemplateIDs != nil {
		if tid, ok := project.PromptTemplateIDs["character_extract"]; ok && tid > 0 {
			tmpl, err := s.promptRepo.FindByID(ctx, tid)
			if err == nil && tmpl.Status == "0" {
				return tmpl.Content, &tmpl.TemplateID, nil
			}
		}
	}

	templates, err := s.promptRepo.FindByType(ctx, "character_extract")
	if err != nil || len(templates) == 0 {
		return "", nil, fmt.Errorf("未找到 character_extract 类型的提示词模板")
	}
	for _, t := range templates {
		if t.IsSystem == "1" && t.Status == "0" {
			return t.Content, &t.TemplateID, nil
		}
	}
	return templates[0].Content, &templates[0].TemplateID, nil
}

func (s *vsCharacterExtractService) parseLLMResponse(content string) (*llmExtractResult, error) {
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

	var result llmExtractResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w\n原始内容: %s", err, truncate(content, 200))
	}
	if len(result.Characters) == 0 {
		return nil, fmt.Errorf("LLM 返回的角色列表为空")
	}
	return &result, nil
}

func (s *vsCharacterExtractService) getLoginName(ctx context.Context) string {
	if v := ctx.Value("login_name"); v != nil {
		return v.(string)
	}
	return "system"
}

func (s *vsCharacterExtractService) getDeptID(ctx context.Context) uint {
	if v := ctx.Value("dept_id"); v != nil {
		return v.(uint)
	}
	return 0
}

func normalizeGender(g string) string {
	g = strings.ToLower(strings.TrimSpace(g))
	switch g {
	case "male", "female":
		return g
	default:
		return "unknown"
	}
}

func normalizeLevel(l string) string {
	l = strings.ToLower(strings.TrimSpace(l))
	switch l {
	case "main", "supporting", "minor":
		return l
	default:
		return "supporting"
	}
}
