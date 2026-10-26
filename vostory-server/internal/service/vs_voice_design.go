package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/audio"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/tts"
	"iot-alert-center/pkg/llm"
)

const (
	voiceDesignTemplateType = "voice_design"
	voiceDesignMaxLines     = 8
	voiceDesignLineRunes    = 60
	voiceDesignPreviewTTL   = 24 * time.Hour
)

var (
	referenceAudioDir     = filepath.Join("storage", "reference-audio")
	voiceDesignPreviewDir = filepath.Join("storage", "reference-audio", "voice-design")
)

// VsVoiceDesignService 基于 Breeze-TTS-2 Voice Design 为角色生成固定的基础音色：
// LLM 根据角色信息撰写音色描述与参考台词 → TTS 按描述生成参考音频 → 保存为声音配置并绑定角色。
// 生成的参考音频与台词即成为该角色后续所有片段 Voice Direction 合成的音色锚点。
type VsVoiceDesignService interface {
	Suggest(ctx context.Context, characterID uint64) (*v1.VoiceDesignSuggestResponse, error)
	Preview(ctx context.Context, characterID uint64, request *v1.VoiceDesignPreviewRequest) (*v1.VoiceDesignPreviewResponse, error)
	Apply(ctx context.Context, characterID uint64, request *v1.VoiceDesignApplyRequest) (*v1.VoiceDesignApplyResponse, error)
	Auto(ctx context.Context, characterID uint64) (*v1.VoiceDesignAutoResponse, error)
}

func NewVsVoiceDesignService(
	service *Service,
	characterRepo repository.VsCharacterRepository,
	projectRepo repository.VsProjectRepository,
	voiceProfileRepo repository.VsVoiceProfileRepository,
	segmentRepo repository.VsScriptSegmentRepository,
	promptRepo repository.VsPromptTemplateRepository,
	llmProviderRepo repository.VsLLMProviderRepository,
	llmLogRepo repository.VsLLMLogRepository,
	ttsProviderRepo repository.VsTTSProviderRepository,
) VsVoiceDesignService {
	return &vsVoiceDesignService{
		Service:          service,
		characterRepo:    characterRepo,
		projectRepo:      projectRepo,
		voiceProfileRepo: voiceProfileRepo,
		segmentRepo:      segmentRepo,
		promptRepo:       promptRepo,
		llmProviderRepo:  llmProviderRepo,
		llmLogRepo:       llmLogRepo,
		ttsProviderRepo:  ttsProviderRepo,
		llmClient:        llm.NewClient(),
	}
}

type vsVoiceDesignService struct {
	*Service
	characterRepo    repository.VsCharacterRepository
	projectRepo      repository.VsProjectRepository
	voiceProfileRepo repository.VsVoiceProfileRepository
	segmentRepo      repository.VsScriptSegmentRepository
	promptRepo       repository.VsPromptTemplateRepository
	llmProviderRepo  repository.VsLLMProviderRepository
	llmLogRepo       repository.VsLLMLogRepository
	ttsProviderRepo  repository.VsTTSProviderRepository
	llmClient        *llm.Client
}

type voiceDesignCharacterInput struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases,omitempty"`
	Gender      string   `json:"gender"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
}

type llmVoiceDesignResult struct {
	Instruction string `json:"instruction"`
	SampleText  string `json:"sample_text"`
}

func (s *vsVoiceDesignService) Suggest(ctx context.Context, characterID uint64) (*v1.VoiceDesignSuggestResponse, error) {
	character, project, err := s.loadCharacter(ctx, characterID)
	if err != nil {
		return nil, err
	}
	provider, err := s.resolveLLMProvider(ctx, project)
	if err != nil {
		return nil, err
	}
	promptContent, templateID := s.resolvePrompt(ctx, project)

	charJSON, _ := json.Marshal(voiceDesignCharacterInput{
		Name:        character.Name,
		Aliases:     character.Aliases,
		Gender:      character.Gender,
		Level:       character.Level,
		Description: character.Description,
	})
	prompt := strings.ReplaceAll(promptContent, "{{character}}", string(charJSON))
	prompt = strings.ReplaceAll(prompt, "{{lines}}", s.sampleLines(ctx, characterID))

	start := time.Now()
	chatResp, err := s.llmClient.ChatCompletionStream(ctx, &llm.ChatRequest{
		BaseURL:        provider.APIBaseURL,
		APIKey:         provider.APIKey,
		Model:          provider.DefaultModel,
		Messages:       []llm.Message{{Role: "user", Content: prompt}},
		ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
		CustomParams:   provider.CustomParams,
	})

	projectID := project.ProjectID
	logEntry := &model.VsLLMLog{
		ProjectID:    &projectID,
		ProviderID:   provider.ProviderID,
		TemplateID:   templateID,
		ModelName:    provider.DefaultModel,
		CostTime:     time.Since(start).Milliseconds(),
		InputSummary: truncate(prompt, 500),
		BaseModel:    model.BaseModel{CreatedBy: getLoginName(ctx)},
	}
	if err != nil {
		logEntry.Status = 1
		logEntry.ErrorMessage = err.Error()
		_ = s.llmLogRepo.Create(ctx, logEntry)
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}
	logEntry.InputTokens = chatResp.InputTokens
	logEntry.OutputTokens = chatResp.OutputTokens
	logEntry.OutputSummary = truncate(chatResp.Content, 500)
	_ = s.llmLogRepo.Create(ctx, logEntry)

	var result llmVoiceDesignResult
	if err := json.Unmarshal([]byte(stripCodeFence(chatResp.Content)), &result); err != nil {
		return nil, fmt.Errorf("解析 LLM 返回结果失败: %w", err)
	}
	result.Instruction = strings.TrimSpace(result.Instruction)
	result.SampleText = strings.TrimSpace(result.SampleText)
	if result.Instruction == "" || result.SampleText == "" {
		return nil, fmt.Errorf("LLM 未返回音色描述或参考台词")
	}

	return &v1.VoiceDesignSuggestResponse{
		Instruction:  result.Instruction,
		SampleText:   result.SampleText,
		InputTokens:  chatResp.InputTokens,
		OutputTokens: chatResp.OutputTokens,
	}, nil
}

func (s *vsVoiceDesignService) Preview(ctx context.Context, characterID uint64, request *v1.VoiceDesignPreviewRequest) (*v1.VoiceDesignPreviewResponse, error) {
	_, project, err := s.loadCharacter(ctx, characterID)
	if err != nil {
		return nil, err
	}
	provider, err := s.resolveTTSProvider(ctx, project)
	if err != nil {
		return nil, err
	}

	seed := rand.Int63n(1 << 31)
	if request.Seed != nil {
		seed = *request.Seed
	}

	data, err := tts.NewClient(provider.APIBaseURL, provider.APIKey).Speech(tts.SpeechRequest{
		Text:        strings.TrimSpace(request.SampleText),
		Instruction: strings.TrimSpace(request.Instruction),
		CfgScale:    tts.DefaultCfgScale,
		Seed:        &seed,
	})
	if err != nil {
		return nil, fmt.Errorf("音色生成失败: %w", err)
	}
	if normalized, normErr := audio.NormalizeLoudness(data, -16.0); normErr == nil {
		data = normalized
	} else {
		s.logger.Warn(fmt.Sprintf("音色试听音量归一化失败，使用原始音频: %v", normErr))
	}

	if err := os.MkdirAll(voiceDesignPreviewDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	cleanupVoiceDesignPreviews()

	previewPath := filepath.Join(voiceDesignPreviewDir,
		fmt.Sprintf("preview_%d_%d.wav", characterID, time.Now().UnixMilli()))
	if err := os.WriteFile(previewPath, data, 0644); err != nil {
		return nil, fmt.Errorf("保存试听音频失败: %w", err)
	}

	return &v1.VoiceDesignPreviewResponse{
		PreviewPath: previewPath,
		Seed:        seed,
		AudioBase64: base64.StdEncoding.EncodeToString(data),
	}, nil
}

func (s *vsVoiceDesignService) Apply(ctx context.Context, characterID uint64, request *v1.VoiceDesignApplyRequest) (*v1.VoiceDesignApplyResponse, error) {
	character, project, err := s.loadCharacter(ctx, characterID)
	if err != nil {
		return nil, err
	}

	previewPath, err := validatePreviewPath(request.PreviewPath)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = character.Name
	}

	if err := os.MkdirAll(referenceAudioDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	finalPath := filepath.Join(referenceAudioDir,
		fmt.Sprintf("%d_%s_design.wav", time.Now().UnixMilli(), sanitizeFileName(name)))
	if err := os.Rename(previewPath, finalPath); err != nil {
		return nil, fmt.Errorf("保存参考音频失败: %w", err)
	}

	gender := character.Gender
	if gender == "" {
		gender = "unknown"
	}
	instruction := strings.TrimSpace(request.Instruction)
	profile := &model.VsVoiceProfile{
		ProjectID:         project.ProjectID,
		Name:              name,
		Gender:            gender,
		Description:       instruction,
		ReferenceAudioURL: finalPath,
		ReferenceText:     strings.TrimSpace(request.SampleText),
		TTSParams: model.TTSParamsMap{
			"seed":         request.Seed,
			"voice_design": instruction,
		},
		Status: "0",
		BaseModel: model.BaseModel{
			CreatedBy: getLoginName(ctx),
			DeptID:    getDeptID(ctx),
		},
	}
	if err := s.voiceProfileRepo.Create(ctx, profile); err != nil {
		_ = os.Remove(finalPath)
		return nil, fmt.Errorf("创建声音配置失败: %w", err)
	}

	if err := s.characterRepo.UpdateFields(ctx, character.CharacterID, map[string]interface{}{
		"voice_profile_id": profile.VoiceProfileID,
		"updated_by":       getLoginName(ctx),
	}); err != nil {
		return nil, fmt.Errorf("绑定角色声音失败: %w", err)
	}

	return &v1.VoiceDesignApplyResponse{
		VoiceProfileID:   profile.VoiceProfileID,
		VoiceProfileName: profile.Name,
	}, nil
}

func (s *vsVoiceDesignService) Auto(ctx context.Context, characterID uint64) (*v1.VoiceDesignAutoResponse, error) {
	suggestion, err := s.Suggest(ctx, characterID)
	if err != nil {
		return nil, err
	}
	preview, err := s.Preview(ctx, characterID, &v1.VoiceDesignPreviewRequest{
		Instruction: suggestion.Instruction,
		SampleText:  suggestion.SampleText,
	})
	if err != nil {
		return nil, err
	}
	applied, err := s.Apply(ctx, characterID, &v1.VoiceDesignApplyRequest{
		PreviewPath: preview.PreviewPath,
		Instruction: suggestion.Instruction,
		SampleText:  suggestion.SampleText,
		Seed:        preview.Seed,
	})
	if err != nil {
		return nil, err
	}
	return &v1.VoiceDesignAutoResponse{
		CharacterID:      characterID,
		VoiceProfileID:   applied.VoiceProfileID,
		VoiceProfileName: applied.VoiceProfileName,
		Instruction:      suggestion.Instruction,
		SampleText:       suggestion.SampleText,
		Seed:             preview.Seed,
	}, nil
}

func (s *vsVoiceDesignService) loadCharacter(ctx context.Context, characterID uint64) (*model.VsCharacter, *model.VsProject, error) {
	character, err := s.characterRepo.FindByID(ctx, characterID)
	if err != nil {
		return nil, nil, fmt.Errorf("角色不存在")
	}
	if character.MergedIntoID != nil {
		return nil, nil, fmt.Errorf("角色「%s」已被合并，请为合并后的角色设计音色", character.Name)
	}
	project, err := s.projectRepo.FindByID(ctx, character.ProjectID)
	if err != nil {
		return nil, nil, fmt.Errorf("项目不存在")
	}
	return character, project, nil
}

// sampleLines 取角色的若干条台词作为 LLM 把握说话风格的参考
func (s *vsVoiceDesignService) sampleLines(ctx context.Context, characterID uint64) string {
	segments, err := s.segmentRepo.FindByCharacterIDs(ctx, []uint64{characterID})
	if err != nil {
		return "（暂无）"
	}
	var lines []string
	for _, seg := range segments {
		if seg.SegmentType != "dialogue" && seg.SegmentType != "monologue" {
			continue
		}
		content := strings.TrimSpace(seg.Content)
		if content == "" {
			continue
		}
		lines = append(lines, "- "+truncate(content, voiceDesignLineRunes))
		if len(lines) >= voiceDesignMaxLines {
			break
		}
	}
	if len(lines) == 0 {
		return "（暂无）"
	}
	return strings.Join(lines, "\n")
}

func (s *vsVoiceDesignService) resolveLLMProvider(ctx context.Context, project *model.VsProject) (*model.VsLLMProvider, error) {
	if project.LLMProviderID == nil || *project.LLMProviderID == 0 {
		return nil, fmt.Errorf("项目未配置 LLM 提供商，请先在项目设置中绑定")
	}
	provider, err := s.llmProviderRepo.FindByID(ctx, *project.LLMProviderID)
	if err != nil {
		return nil, fmt.Errorf("LLM 提供商不存在")
	}
	if provider.Status != "0" {
		return nil, fmt.Errorf("LLM 提供商已停用")
	}
	return provider, nil
}

func (s *vsVoiceDesignService) resolveTTSProvider(ctx context.Context, project *model.VsProject) (*model.VsTTSProvider, error) {
	if project.TTSProviderID == nil {
		return nil, fmt.Errorf("项目「%s」未设置 TTS 提供商，请在项目设置中选择", project.Name)
	}
	provider, err := s.ttsProviderRepo.FindByID(ctx, *project.TTSProviderID)
	if err != nil {
		return nil, fmt.Errorf("项目绑定的 TTS 提供商不存在，请重新设置")
	}
	if provider.Status != "0" {
		return nil, fmt.Errorf("项目绑定的 TTS 提供商「%s」已停用，请更换", provider.Name)
	}
	return provider, nil
}

// resolvePrompt 优先使用项目指定模板，其次系统模板；库中尚无该类型模板时回退到内置默认内容
func (s *vsVoiceDesignService) resolvePrompt(ctx context.Context, project *model.VsProject) (string, *uint64) {
	if project.PromptTemplateIDs != nil {
		if tid, ok := project.PromptTemplateIDs[voiceDesignTemplateType]; ok && tid > 0 {
			if tmpl, err := s.promptRepo.FindByID(ctx, tid); err == nil && tmpl.Status == "0" {
				return tmpl.Content, &tmpl.TemplateID
			}
		}
	}
	if templates, err := s.promptRepo.FindByType(ctx, voiceDesignTemplateType); err == nil && len(templates) > 0 {
		for _, t := range templates {
			if t.IsSystem == "1" && t.Status == "0" {
				return t.Content, &t.TemplateID
			}
		}
		return templates[0].Content, &templates[0].TemplateID
	}
	for _, seed := range model.DefaultPromptTemplateSeeds {
		if seed.TemplateType == voiceDesignTemplateType {
			return seed.Content, nil
		}
	}
	return "", nil
}

// validatePreviewPath 只接受本服务生成在试听目录下的文件，防止通过路径参数操作任意文件
func validatePreviewPath(p string) (string, error) {
	cleaned := filepath.Clean(p)
	dir, file := filepath.Split(cleaned)
	if filepath.Clean(dir) != filepath.Clean(voiceDesignPreviewDir) || !strings.HasPrefix(file, "preview_") {
		return "", fmt.Errorf("无效的试听文件")
	}
	if _, err := os.Stat(cleaned); err != nil {
		return "", fmt.Errorf("试听文件已过期，请重新生成")
	}
	return cleaned, nil
}

// cleanupVoiceDesignPreviews 清理过期未保存的试听文件
func cleanupVoiceDesignPreviews() {
	entries, err := os.ReadDir(voiceDesignPreviewDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "preview_") {
			continue
		}
		if info, err := e.Info(); err == nil && time.Since(info.ModTime()) > voiceDesignPreviewTTL {
			_ = os.Remove(filepath.Join(voiceDesignPreviewDir, e.Name()))
		}
	}
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", " ", "_")
	return replacer.Replace(name)
}

func stripCodeFence(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		if parts := strings.SplitN(content, "\n", 2); len(parts) > 1 {
			content = parts[1]
		}
		if idx := strings.LastIndex(content, "```"); idx >= 0 {
			content = content[:idx]
		}
	}
	return strings.TrimSpace(content)
}
