package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/tts"
)

// IndexTTS2 8-dim emotion vector order:
// [happy, angry, sad, afraid, disgusted, melancholic, surprised, calm]
var emotionVectorMap = map[string][8]float64{
	"neutral":  {0, 0, 0, 0, 0, 0, 0, 1.0},
	"happy":    {1.0, 0, 0, 0, 0, 0, 0, 0},
	"sad":      {0, 0, 1.0, 0, 0, 0, 0, 0},
	"angry":    {0, 1.0, 0, 0, 0, 0, 0, 0},
	"fear":     {0, 0, 0, 1.0, 0, 0, 0, 0},
	"surprise": {0, 0, 0, 0, 0, 0, 1.0, 0},
	"disgust":  {0, 0, 0, 0, 1.0, 0, 0, 0},
}

var strengthAlphaMap = map[string]float64{
	"light":  0.3,
	"medium": 0.6,
	"strong": 0.9,
}

type VsTTSSynthesizeService interface {
	SynthesizeSegment(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error)
	GetSegmentAudio(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error)
}

func NewVsTTSSynthesizeService(
	service *Service,
	segmentRepo repository.VsScriptSegmentRepository,
	characterRepo repository.VsCharacterRepository,
	voiceProfileRepo repository.VsVoiceProfileRepository,
	voiceEmotionRepo repository.VsVoiceEmotionRepository,
	pronunciationDictRepo repository.VsPronunciationDictRepository,
	audioClipRepo repository.VsAudioClipRepository,
	ttsProviderRepo repository.VsTTSProviderRepository,
	projectRepo repository.VsProjectRepository,
) VsTTSSynthesizeService {
	return &vsTTSSynthesizeService{
		Service:               service,
		segmentRepo:           segmentRepo,
		characterRepo:         characterRepo,
		voiceProfileRepo:      voiceProfileRepo,
		voiceEmotionRepo:      voiceEmotionRepo,
		pronunciationDictRepo: pronunciationDictRepo,
		audioClipRepo:         audioClipRepo,
		ttsProviderRepo:       ttsProviderRepo,
		projectRepo:           projectRepo,
	}
}

type vsTTSSynthesizeService struct {
	*Service
	segmentRepo           repository.VsScriptSegmentRepository
	characterRepo         repository.VsCharacterRepository
	voiceProfileRepo      repository.VsVoiceProfileRepository
	voiceEmotionRepo      repository.VsVoiceEmotionRepository
	pronunciationDictRepo repository.VsPronunciationDictRepository
	audioClipRepo         repository.VsAudioClipRepository
	ttsProviderRepo       repository.VsTTSProviderRepository
	projectRepo           repository.VsProjectRepository
}

func (s *vsTTSSynthesizeService) SynthesizeSegment(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error) {
	segment, err := s.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		return nil, fmt.Errorf("片段不存在: %w", err)
	}

	if segment.CharacterID == nil {
		return nil, fmt.Errorf("该片段未关联角色，无法合成。请先为片段指定说话人（旁白/描述类型请指定「旁白」角色）")
	}

	character, err := s.characterRepo.FindByID(ctx, *segment.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}
	if character.VoiceProfileID == nil {
		return nil, fmt.Errorf("角色「%s」未绑定声音配置，请先在角色管理中绑定", character.Name)
	}
	voiceProfile, err := s.voiceProfileRepo.FindByID(ctx, *character.VoiceProfileID)
	if err != nil {
		return nil, fmt.Errorf("声音配置不存在: %w", err)
	}

	provider, err := s.resolveTTSProvider(ctx, voiceProfile, segment.ChapterID)
	if err != nil {
		return nil, err
	}

	referenceAudioURL := voiceProfile.ReferenceAudioURL
	emotion, _ := s.voiceEmotionRepo.FindByMatch(ctx, voiceProfile.VoiceProfileID, segment.EmotionType, segment.EmotionStrength)
	if emotion != nil {
		referenceAudioURL = emotion.ReferenceAudioURL
	}

	if referenceAudioURL == "" {
		return nil, fmt.Errorf("声音配置 %s 缺少参考音频", voiceProfile.Name)
	}

	_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	text := s.applyPronunciationDict(ctx, segment)

	emoVector, _ := buildEmotionVector(segment.EmotionType, segment.EmotionStrength)

	client := tts.NewClient(provider.APIBaseURL)

	remoteKey := filepath.Base(referenceAudioURL)
	if err := client.EnsureAudioUploaded(referenceAudioURL, remoteKey); err != nil {
		_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "failed")
		return nil, fmt.Errorf("上传参考音频失败: %w", err)
	}

	audioData, err := client.Synthesize(text, remoteKey, emoVector, "")
	if err != nil {
		_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "failed")
		return nil, fmt.Errorf("TTS 合成失败: %w", err)
	}

	audioURL, fileSize, err := s.saveAudioFile(segment, audioData)
	if err != nil {
		_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "failed")
		return nil, fmt.Errorf("保存音频文件失败: %w", err)
	}

	clip, err := s.createAudioClip(ctx, segment, provider, voiceProfile, audioURL, fileSize)
	if err != nil {
		_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "failed")
		return nil, fmt.Errorf("创建音频记录失败: %w", err)
	}

	_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "generated")

	return &v1.TTSSynthesizeResponse{
		ClipID:   clip.ClipID,
		AudioURL: clip.AudioURL,
		Duration: clip.Duration,
		Version:  clip.Version,
	}, nil
}

func (s *vsTTSSynthesizeService) GetSegmentAudio(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error) {
	clip, err := s.audioClipRepo.FindCurrentBySegmentID(ctx, segmentID)
	if err != nil {
		return nil, fmt.Errorf("该片段暂无音频")
	}
	return &v1.TTSSynthesizeResponse{
		ClipID:   clip.ClipID,
		AudioURL: clip.AudioURL,
		Duration: clip.Duration,
		Version:  clip.Version,
	}, nil
}

func (s *vsTTSSynthesizeService) resolveTTSProvider(ctx context.Context, profile *model.VsVoiceProfile, chapterID uint64) (*model.VsTTSProvider, error) {
	if profile.TTSProviderID != nil {
		provider, err := s.ttsProviderRepo.FindByID(ctx, *profile.TTSProviderID)
		if err == nil && provider.Status == "0" {
			return provider, nil
		}
	}

	chapter, err := s.segmentRepo.FindByID(ctx, chapterID)
	_ = chapter

	providers, err := s.ttsProviderRepo.FindAllEnabled(ctx)
	if err != nil || len(providers) == 0 {
		return nil, fmt.Errorf("没有可用的 TTS 提供商，请先在 AI 配置中添加")
	}
	return providers[0], nil
}

func (s *vsTTSSynthesizeService) applyPronunciationDict(ctx context.Context, segment *model.VsScriptSegment) string {
	text := segment.Content

	projectDicts, _ := s.pronunciationDictRepo.FindByProjectID(ctx, 0)
	_ = projectDicts

	dicts, err := s.pronunciationDictRepo.FindByProjectID(ctx, segment.ChapterID)
	if err != nil {
		return text
	}

	for _, d := range dicts {
		text = strings.ReplaceAll(text, d.Word, d.Phoneme)
	}
	return text
}

func buildEmotionVector(emotionType, emotionStrength string) ([]float64, float64) {
	baseVector, ok := emotionVectorMap[emotionType]
	if !ok {
		baseVector = emotionVectorMap["neutral"]
	}

	alpha, ok := strengthAlphaMap[emotionStrength]
	if !ok {
		alpha = 0.6
	}

	result := make([]float64, 8)
	for i, v := range baseVector {
		result[i] = v * alpha
	}
	return result, alpha
}

func (s *vsTTSSynthesizeService) saveAudioFile(segment *model.VsScriptSegment, audioData []byte) (string, int64, error) {
	maxVer, _ := s.audioClipRepo.GetMaxVersion(context.Background(), segment.SegmentID)
	version := maxVer + 1

	dir := filepath.Join("storage", "audio",
		fmt.Sprintf("%d", segment.ChapterID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, err
	}

	filename := fmt.Sprintf("%d_v%d.wav", segment.SegmentID, version)
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, audioData, 0644); err != nil {
		return "", 0, err
	}

	return fullPath, int64(len(audioData)), nil
}

func (s *vsTTSSynthesizeService) createAudioClip(
	ctx context.Context,
	segment *model.VsScriptSegment,
	provider *model.VsTTSProvider,
	profile *model.VsVoiceProfile,
	audioURL string,
	fileSize int64,
) (*model.VsAudioClip, error) {
	if err := s.audioClipRepo.SetAllNonCurrent(ctx, segment.SegmentID); err != nil {
		return nil, err
	}

	maxVer, _ := s.audioClipRepo.GetMaxVersion(ctx, segment.SegmentID)

	providerID := provider.ProviderID
	profileID := profile.VoiceProfileID
	clip := &model.VsAudioClip{
		SegmentID:       segment.SegmentID,
		AudioURL:        audioURL,
		FileSize:        fileSize,
		Format:          "wav",
		TTSProviderID:   &providerID,
		VoiceProfileID:  &profileID,
		EmotionType:     segment.EmotionType,
		EmotionStrength: segment.EmotionStrength,
		Version:         maxVer + 1,
		IsCurrent:       "1",
		BaseModel: model.BaseModel{
			CreatedBy: getLoginName(ctx),
			DeptID:    getDeptID(ctx),
		},
	}

	if err := s.audioClipRepo.Create(ctx, clip); err != nil {
		return nil, err
	}
	return clip, nil
}

func getLoginName(ctx context.Context) string {
	if v := ctx.Value("login_name"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getDeptID(ctx context.Context) uint {
	if v := ctx.Value("dept_id"); v != nil {
		if d, ok := v.(uint); ok {
			return d
		}
	}
	return 0
}
