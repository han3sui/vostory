package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/audio"
	"iot-alert-center/internal/filestorage"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/tts"

	"github.com/redis/go-redis/v9"
)

// emotionInstructions 情绪类型+强度 → Breeze-TTS-2 演绎指令（片段未填写 instruction 时兜底）。
// neutral / auto 不下发指令，按参考音色原样克隆，由模型根据文本自行把握语气。
var emotionInstructions = map[string]map[string]string{
	"happy": {
		"light":  "语气轻快，略带笑意",
		"medium": "语气愉快，带着笑意",
		"strong": "非常开心，语调上扬，充满兴奋",
	},
	"sad": {
		"light":  "语气略显低落",
		"medium": "语气悲伤，语速稍慢",
		"strong": "极度悲伤，带着哭腔，声音哽咽",
	},
	"angry": {
		"light":  "语气不满，略带怒意",
		"medium": "语气愤怒，语速加快",
		"strong": "暴怒，大声呵斥",
	},
	"fear": {
		"light":  "语气有些不安",
		"medium": "语气紧张害怕，声音发颤",
		"strong": "极度恐惧，声音颤抖急促",
	},
	"surprise": {
		"light":  "语气略带惊讶",
		"medium": "语气惊讶",
		"strong": "非常震惊，脱口而出",
	},
	"disgust": {
		"light":  "语气略带嫌弃",
		"medium": "语气厌恶",
		"strong": "极度厌恶，语带鄙夷",
	},
}

type VsTTSSynthesizeService interface {
	SynthesizeSegment(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error)
	GetSegmentAudio(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error)
	SingleGenerate(ctx context.Context, segmentID uint64) (*v1.BatchGenerateResponse, error)
	BatchGenerate(ctx context.Context, chapterID uint64) (*v1.BatchGenerateResponse, error)
	GetTaskProgress(ctx context.Context, taskID uint64) (*v1.TaskProgressResponse, error)
	GetActiveTaskByChapter(ctx context.Context, chapterID uint64) (*v1.TaskProgressResponse, error)
	GetActiveTasksByProject(ctx context.Context, projectID uint64) ([]*v1.ProjectTaskProgressResponse, error)
	OpenAudioClip(ctx context.Context, clipID uint64) (r io.ReadCloser, contentType string, err error)
	LockSegment(ctx context.Context, segmentID uint64) error
	UnlockSegment(ctx context.Context, segmentID uint64) error
	BatchLockByChapter(ctx context.Context, chapterID uint64) (int64, error)
	BatchUnlockByChapter(ctx context.Context, chapterID uint64) (int64, error)
	CancelChapterQueue(ctx context.Context, chapterID uint64) (int64, error)
	CancelProjectQueue(ctx context.Context, projectID uint64) (int64, error)
	RegenerateByVoiceProfile(ctx context.Context, voiceProfileID uint64) (*v1.BatchGenerateResponse, error)
	RegenerateByCharacter(ctx context.Context, characterID uint64) (*v1.BatchGenerateResponse, error)
}

func NewVsTTSSynthesizeService(
	service *Service,
	segmentRepo repository.VsScriptSegmentRepository,
	chapterRepo repository.VsChapterRepository,
	characterRepo repository.VsCharacterRepository,
	voiceProfileRepo repository.VsVoiceProfileRepository,
	voiceEmotionRepo repository.VsVoiceEmotionRepository,
	pronunciationDictRepo repository.VsPronunciationDictRepository,
	audioClipRepo repository.VsAudioClipRepository,
	ttsProviderRepo repository.VsTTSProviderRepository,
	projectRepo repository.VsProjectRepository,
	taskRepo repository.VsGenerationTaskRepository,
	rdb *redis.Client,
	fileStorage filestorage.FileStorage,
) VsTTSSynthesizeService {
	return &vsTTSSynthesizeService{
		Service:               service,
		segmentRepo:           segmentRepo,
		chapterRepo:           chapterRepo,
		characterRepo:         characterRepo,
		voiceProfileRepo:      voiceProfileRepo,
		voiceEmotionRepo:      voiceEmotionRepo,
		pronunciationDictRepo: pronunciationDictRepo,
		audioClipRepo:         audioClipRepo,
		ttsProviderRepo:       ttsProviderRepo,
		projectRepo:           projectRepo,
		taskRepo:              taskRepo,
		rdb:                   rdb,
		fileStorage:           fileStorage,
	}
}

type vsTTSSynthesizeService struct {
	*Service
	segmentRepo           repository.VsScriptSegmentRepository
	chapterRepo           repository.VsChapterRepository
	characterRepo         repository.VsCharacterRepository
	voiceProfileRepo      repository.VsVoiceProfileRepository
	voiceEmotionRepo      repository.VsVoiceEmotionRepository
	pronunciationDictRepo repository.VsPronunciationDictRepository
	audioClipRepo         repository.VsAudioClipRepository
	ttsProviderRepo       repository.VsTTSProviderRepository
	projectRepo           repository.VsProjectRepository
	taskRepo              repository.VsGenerationTaskRepository
	rdb                   *redis.Client
	fileStorage           filestorage.FileStorage
}

func (s *vsTTSSynthesizeService) SynthesizeSegment(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error) {
	segment, err := s.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		return nil, fmt.Errorf("片段不存在: %w", err)
	}

	failAndReturn := func(msg string) (*v1.TTSSynthesizeResponse, error) {
		_ = s.segmentRepo.UpdateStatusWithError(ctx, segmentID, "failed", msg)
		return nil, fmt.Errorf("%s", msg)
	}

	if segment.CharacterID == nil {
		return failAndReturn("该片段未关联角色，无法合成。请先为片段指定说话人（旁白/描述类型请指定「旁白」角色）")
	}

	character, err := s.characterRepo.FindByID(ctx, *segment.CharacterID)
	if err != nil {
		return failAndReturn(fmt.Sprintf("角色不存在: %v", err))
	}
	if character.VoiceProfileID == nil {
		return failAndReturn(fmt.Sprintf("角色「%s」未绑定声音配置，请先在角色管理中绑定", character.Name))
	}
	voiceProfile, err := s.voiceProfileRepo.FindByID(ctx, *character.VoiceProfileID)
	if err != nil {
		return failAndReturn(fmt.Sprintf("声音配置不存在: %v", err))
	}

	provider, err := s.resolveTTSProvider(ctx, segment.ChapterID)
	if err != nil {
		return failAndReturn(fmt.Sprintf("TTS 提供商不可用: %v", err))
	}

	// 情绪参考音频（如有）优先，否则使用声音配置的基础音色
	referenceAudioURL := voiceProfile.ReferenceAudioURL
	referenceText := voiceProfile.ReferenceText
	emotion, _ := s.voiceEmotionRepo.FindByMatch(ctx, voiceProfile.VoiceProfileID, segment.EmotionType, segment.EmotionStrength)
	if emotion != nil && emotion.ReferenceAudioURL != "" && emotion.ReferenceText != "" {
		referenceAudioURL = emotion.ReferenceAudioURL
		referenceText = emotion.ReferenceText
	}

	if referenceAudioURL == "" {
		return failAndReturn(fmt.Sprintf("声音配置「%s」缺少参考音频，可在角色管理中「设计音色」自动生成", voiceProfile.Name))
	}
	if strings.TrimSpace(referenceText) == "" {
		return failAndReturn(fmt.Sprintf("声音配置「%s」缺少参考文本（参考音频的逐字转写），Breeze-TTS-2 克隆音色需要", voiceProfile.Name))
	}

	_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	text := s.applyPronunciationDict(ctx, segment)

	refPath, cleanup, err := s.fileStorage.DownloadToTemp(ctx, filestorage.NormalizeKey(referenceAudioURL))
	if err != nil {
		return failAndReturn(fmt.Sprintf("参考音频不存在: %v", err))
	}
	defer cleanup()

	client := tts.NewClient(provider.APIBaseURL, provider.APIKey)

	audioData, err := client.Speech(tts.SpeechRequest{
		Text:         text,
		RefAudioPath: refPath,
		RefText:      referenceText,
		Instruction:  segmentInstruction(segment),
		CfgScale:     ttsParamFloat(voiceProfile.TTSParams, "cfg_scale"),
		Seed:         voiceSeed(voiceProfile),
	})
	if err != nil {
		return failAndReturn(fmt.Sprintf("TTS 合成失败: %v", err))
	}

	normalizedData, normErr := audio.NormalizeLoudness(audioData, -16.0)
	if normErr != nil {
		s.logger.Warn(fmt.Sprintf("音量归一化失败，使用原始音频: %v", normErr))
	} else {
		audioData = normalizedData
	}

	audioURL, fileSize, err := s.saveAudioFile(ctx, segment, audioData)
	if err != nil {
		return failAndReturn(fmt.Sprintf("保存音频文件失败: %v", err))
	}

	clip, err := s.createAudioClip(ctx, segment, provider, voiceProfile, audioURL, fileSize)
	if err != nil {
		return failAndReturn(fmt.Sprintf("创建音频记录失败: %v", err))
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

func (s *vsTTSSynthesizeService) resolveTTSProvider(ctx context.Context, chapterID uint64) (*model.VsTTSProvider, error) {
	chapter, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("章节不存在: %w", err)
	}
	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在: %w", err)
	}
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

func (s *vsTTSSynthesizeService) applyPronunciationDict(ctx context.Context, segment *model.VsScriptSegment) string {
	text := segment.Content

	dicts, err := s.pronunciationDictRepo.FindByProjectID(ctx, segment.ProjectID)
	if err != nil || len(dicts) == 0 {
		return text
	}

	for _, d := range dicts {
		text = strings.ReplaceAll(text, d.Word, d.Phoneme)
	}
	return text
}

// segmentInstruction 片段的演绎指令：优先使用片段上的 instruction（由切分时 LLM 按场景生成或人工编辑），
// 为空时按情绪类型+强度兜底。
func segmentInstruction(segment *model.VsScriptSegment) string {
	if inst := strings.TrimSpace(segment.Instruction); inst != "" {
		return inst
	}
	if byStrength, ok := emotionInstructions[segment.EmotionType]; ok {
		if inst, ok := byStrength[segment.EmotionStrength]; ok {
			return inst
		}
		return byStrength["medium"]
	}
	return ""
}

// voiceSeed 同一声音配置固定使用同一个随机种子，保证不同片段之间音色稳定。
// 优先取 tts_params.seed（音色设计时记录），否则以声音配置 ID 作为种子。
func voiceSeed(profile *model.VsVoiceProfile) *int64 {
	if v, ok := ttsParamInt(profile.TTSParams, "seed"); ok {
		return &v
	}
	seed := int64(profile.VoiceProfileID)
	return &seed
}

func ttsParamFloat(params model.TTSParamsMap, key string) float64 {
	switch v := params[key].(type) {
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

func ttsParamInt(params model.TTSParamsMap, key string) (int64, bool) {
	switch v := params[key].(type) {
	case float64:
		return int64(v), true
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}

func (s *vsTTSSynthesizeService) saveAudioFile(ctx context.Context, segment *model.VsScriptSegment, audioData []byte) (string, int64, error) {
	maxVer, _ := s.audioClipRepo.GetMaxVersion(ctx, segment.SegmentID)
	version := maxVer + 1
	key := fmt.Sprintf("audio/%d/%d_v%d.wav", segment.ChapterID, segment.SegmentID, version)
	if err := s.fileStorage.Put(ctx, key, bytes.NewReader(audioData), int64(len(audioData)), "audio/wav"); err != nil {
		return "", 0, err
	}
	return key, int64(len(audioData)), nil
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

func (s *vsTTSSynthesizeService) enqueueSegments(ctx context.Context, task *model.VsGenerationTask, segmentIDs []uint64) error {
	for _, segID := range segmentIDs {
		_ = s.segmentRepo.UpdateStatus(ctx, segID, "queued")

		msg := fmt.Sprintf("%d:%d", task.TaskID, segID)
		if err := s.rdb.LPush(ctx, ttsQueueKey, msg).Err(); err != nil {
			return fmt.Errorf("片段 %d 入队失败: %w", segID, err)
		}
	}
	return nil
}

func (s *vsTTSSynthesizeService) SingleGenerate(ctx context.Context, segmentID uint64) (*v1.BatchGenerateResponse, error) {
	segment, err := s.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		return nil, fmt.Errorf("片段不存在: %w", err)
	}
	if segment.CharacterID == nil {
		return nil, fmt.Errorf("该片段未关联角色，无法合成")
	}
	// 已在队列中的片段再入队会被合成两次，并发时还可能写出同名音频文件、两条当前版本
	if segment.Status == "queued" || segment.Status == "processing" {
		return nil, fmt.Errorf("CONFLICT:该片段正在排队或合成中，请等待完成")
	}

	chapterID := segment.ChapterID
	projectID, err := s.getProjectIDByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	task := &model.VsGenerationTask{
		ProjectID:    projectID,
		ChapterID:    &chapterID,
		TaskType:     "tts_generate",
		Status:       "running",
		TotalBatches: 1,
		SegmentIDs:   model.Uint64Array{segmentID},
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("创建生成任务失败: %w", err)
	}

	if err := s.enqueueSegments(ctx, task, []uint64{segmentID}); err != nil {
		_ = s.taskRepo.UpdateStatus(ctx, task.TaskID, "failed", err.Error())
		return nil, fmt.Errorf("任务入队失败: %w", err)
	}

	return &v1.BatchGenerateResponse{
		TaskID:       task.TaskID,
		TotalCount:   1,
		SkippedCount: 0,
	}, nil
}

func (s *vsTTSSynthesizeService) BatchGenerate(ctx context.Context, chapterID uint64) (*v1.BatchGenerateResponse, error) {
	activeTask, _ := s.taskRepo.FindActiveByChapterID(ctx, chapterID)
	if activeTask != nil {
		return nil, fmt.Errorf("CONFLICT:该章节已有正在运行的生成任务（ID: %d），请等待完成", activeTask.TaskID)
	}

	segments, err := s.segmentRepo.FindByChapterID(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("获取章节片段失败: %w", err)
	}

	var eligible []*model.VsScriptSegment
	for _, seg := range segments {
		if seg.CharacterID == nil || seg.Status == "queued" || seg.Status == "processing" || seg.Status == "locked" {
			continue
		}
		eligible = append(eligible, seg)
	}

	if len(eligible) == 0 {
		return nil, fmt.Errorf("没有可生成的片段（请确保片段已关联角色）")
	}

	segIDs := make([]uint64, len(eligible))
	for i, seg := range eligible {
		segIDs[i] = seg.SegmentID
	}

	projectID, err := s.getProjectIDByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	task := &model.VsGenerationTask{
		ProjectID:    projectID,
		ChapterID:    &chapterID,
		TaskType:     "tts_generate",
		Status:       "running",
		TotalBatches: len(eligible),
		SegmentIDs:   model.Uint64Array(segIDs),
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("创建生成任务失败: %w", err)
	}

	if err := s.enqueueSegments(ctx, task, segIDs); err != nil {
		_ = s.taskRepo.UpdateStatus(ctx, task.TaskID, "failed", err.Error())
		return nil, fmt.Errorf("任务入队失败: %w", err)
	}

	return &v1.BatchGenerateResponse{
		TaskID:       task.TaskID,
		TotalCount:   len(eligible),
		SkippedCount: len(segments) - len(eligible),
	}, nil
}

func (s *vsTTSSynthesizeService) GetActiveTaskByChapter(ctx context.Context, chapterID uint64) (*v1.TaskProgressResponse, error) {
	task, err := s.taskRepo.FindActiveByChapterID(ctx, chapterID)
	if err != nil {
		return nil, nil
	}
	return &v1.TaskProgressResponse{
		TaskID:         task.TaskID,
		Status:         task.Status,
		Progress:       task.Progress,
		TotalCount:     task.TotalBatches,
		CompletedCount: task.CompletedBatches,
		FailedCount:    task.FailedBatches,
		ErrorMessage:   task.ErrorMessage,
		StartedAt:      task.StartedAt,
		CompletedAt:    task.CompletedAt,
	}, nil
}

func (s *vsTTSSynthesizeService) GetTaskProgress(ctx context.Context, taskID uint64) (*v1.TaskProgressResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}

	return &v1.TaskProgressResponse{
		TaskID:         task.TaskID,
		Status:         task.Status,
		Progress:       task.Progress,
		TotalCount:     task.TotalBatches,
		CompletedCount: task.CompletedBatches,
		FailedCount:    task.FailedBatches,
		ErrorMessage:   task.ErrorMessage,
		StartedAt:      task.StartedAt,
		CompletedAt:    task.CompletedAt,
	}, nil
}

func (s *vsTTSSynthesizeService) OpenAudioClip(ctx context.Context, clipID uint64) (io.ReadCloser, string, error) {
	clip, err := s.audioClipRepo.FindByID(ctx, clipID)
	if err != nil {
		return nil, "", fmt.Errorf("音频片段不存在: %w", err)
	}

	r, err := s.fileStorage.Open(ctx, filestorage.NormalizeKey(clip.AudioURL))
	if err != nil {
		return nil, "", fmt.Errorf("音频文件不存在: %w", err)
	}

	contentType := "audio/wav"
	if clip.Format == "mp3" {
		contentType = "audio/mpeg"
	} else if clip.Format == "flac" {
		contentType = "audio/flac"
	}

	return r, contentType, nil
}

func (s *vsTTSSynthesizeService) GetActiveTasksByProject(ctx context.Context, projectID uint64) ([]*v1.ProjectTaskProgressResponse, error) {
	tasks, err := s.taskRepo.FindActiveByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var result []*v1.ProjectTaskProgressResponse
	for _, t := range tasks {
		item := &v1.ProjectTaskProgressResponse{
			TaskID:         t.TaskID,
			ChapterID:      t.ChapterID,
			TaskType:       t.TaskType,
			Status:         t.Status,
			Progress:       t.Progress,
			TotalCount:     t.TotalBatches,
			CompletedCount: t.CompletedBatches,
			FailedCount:    t.FailedBatches,
			SegmentIDs:     []uint64(t.SegmentIDs),
		}
		if t.Chapter != nil {
			item.ChapterTitle = t.Chapter.Title
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *vsTTSSynthesizeService) LockSegment(ctx context.Context, segmentID uint64) error {
	seg, err := s.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		return fmt.Errorf("片段不存在")
	}
	if seg.Status != "generated" {
		return fmt.Errorf("只有已生成状态的片段才能锁定，当前状态: %s", seg.Status)
	}
	return s.segmentRepo.UpdateStatus(ctx, segmentID, "locked")
}

func (s *vsTTSSynthesizeService) UnlockSegment(ctx context.Context, segmentID uint64) error {
	seg, err := s.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		return fmt.Errorf("片段不存在")
	}
	if seg.Status != "locked" {
		return fmt.Errorf("只有已锁定状态的片段才能解锁，当前状态: %s", seg.Status)
	}
	return s.segmentRepo.UpdateStatus(ctx, segmentID, "generated")
}

func (s *vsTTSSynthesizeService) BatchLockByChapter(ctx context.Context, chapterID uint64) (int64, error) {
	return s.segmentRepo.BatchUpdateStatusByChapter(ctx, chapterID, "generated", "locked")
}

func (s *vsTTSSynthesizeService) BatchUnlockByChapter(ctx context.Context, chapterID uint64) (int64, error) {
	return s.segmentRepo.BatchUpdateStatusByChapter(ctx, chapterID, "locked", "generated")
}

func (s *vsTTSSynthesizeService) CancelChapterQueue(ctx context.Context, chapterID uint64) (int64, error) {
	tasks, err := s.taskRepo.FindActiveListByChapterID(ctx, chapterID)
	if err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	var totalAffected int64
	for _, task := range tasks {
		affected, err := s.cancelTaskQueuedSegments(ctx, task)
		if err != nil {
			continue
		}
		totalAffected += affected
	}
	return totalAffected, nil
}

func (s *vsTTSSynthesizeService) getProjectIDByChapter(ctx context.Context, chapterID uint64) (uint64, error) {
	chapter, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return 0, fmt.Errorf("章节不存在: %w", err)
	}
	return chapter.ProjectID, nil
}

func (s *vsTTSSynthesizeService) CancelProjectQueue(ctx context.Context, projectID uint64) (int64, error) {
	tasks, err := s.taskRepo.FindActiveByProjectID(ctx, projectID)
	if err != nil {
		return 0, err
	}
	var totalAffected int64
	for _, task := range tasks {
		if task.TaskType != "tts_generate" {
			continue
		}
		affected, err := s.cancelTaskQueuedSegments(ctx, task)
		if err != nil {
			continue
		}
		totalAffected += affected
	}
	return totalAffected, nil
}

// cancelTaskQueuedSegments 结束任务、取消排队片段并清掉 Redis 中的残留消息。
// 先把任务标为 cancelled：worker 处理每条消息前都会检查任务状态，这样取消过程中不会再有片段被计入进度、
// 也不会推送 running 状态把任务重新显示到队列里。
func (s *vsTTSSynthesizeService) cancelTaskQueuedSegments(ctx context.Context, task *model.VsGenerationTask) (int64, error) {
	if task.TaskType != "tts_generate" {
		return 0, nil
	}
	if err := s.taskRepo.SetCancelled(ctx, task.TaskID, "cancelled by user"); err != nil {
		return 0, err
	}

	// processing 也一并标为 cancelled：正在合成的片段完成后会被 SynthesizeSegment 改写为最终状态；
	// 若 worker 已经退出（如服务重启），片段不会永远卡在 processing，之后还能重新生成
	var affected int64
	if len(task.SegmentIDs) > 0 {
		n, err := s.segmentRepo.BatchUpdateStatus(ctx, []uint64(task.SegmentIDs), "queued", "cancelled")
		if err != nil {
			return 0, err
		}
		affected = n
		_, _ = s.segmentRepo.BatchUpdateStatus(ctx, []uint64(task.SegmentIDs), "processing", "cancelled")
	} else if task.ChapterID != nil {
		n, err := s.segmentRepo.BatchUpdateStatusByChapter(ctx, *task.ChapterID, "queued", "cancelled")
		if err != nil {
			return 0, err
		}
		affected = n
		_, _ = s.segmentRepo.BatchUpdateStatusByChapter(ctx, *task.ChapterID, "processing", "cancelled")
	}
	if err := s.removeQueueMessagesByTask(ctx, task.TaskID); err != nil {
		// 清理失败不影响取消结果，残留消息会被 worker 按任务状态跳过
		s.logger.Warn(fmt.Sprintf("清理任务 %d 的队列消息失败: %v", task.TaskID, err))
	}
	return affected, nil
}

func (s *vsTTSSynthesizeService) removeQueueMessagesByTask(ctx context.Context, taskID uint64) error {
	return removeTaskQueueMessages(ctx, s.rdb, ttsQueueKey, taskID)
}

func (s *vsTTSSynthesizeService) RegenerateByVoiceProfile(ctx context.Context, voiceProfileID uint64) (*v1.BatchGenerateResponse, error) {
	profile, err := s.voiceProfileRepo.FindByID(ctx, voiceProfileID)
	if err != nil {
		return nil, fmt.Errorf("声音配置不存在: %w", err)
	}

	characters, err := s.characterRepo.FindByVoiceProfileID(ctx, profile.VoiceProfileID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	if len(characters) == 0 {
		return nil, fmt.Errorf("没有角色绑定了该声音配置")
	}

	characterIDs := make([]uint64, len(characters))
	for i, c := range characters {
		characterIDs[i] = c.CharacterID
	}

	return s.regenerateByCharacterIDs(ctx, characterIDs)
}

func (s *vsTTSSynthesizeService) RegenerateByCharacter(ctx context.Context, characterID uint64) (*v1.BatchGenerateResponse, error) {
	_, err := s.characterRepo.FindByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}
	return s.regenerateByCharacterIDs(ctx, []uint64{characterID})
}

func (s *vsTTSSynthesizeService) regenerateByCharacterIDs(ctx context.Context, characterIDs []uint64) (*v1.BatchGenerateResponse, error) {
	segments, err := s.segmentRepo.FindByCharacterIDs(ctx, characterIDs)
	if err != nil {
		return nil, fmt.Errorf("查询片段失败: %w", err)
	}

	var eligible []*model.VsScriptSegment
	for _, seg := range segments {
		if seg.Status == "queued" || seg.Status == "processing" || seg.Status == "locked" {
			continue
		}
		if seg.Content == "" {
			continue
		}
		eligible = append(eligible, seg)
	}

	if len(eligible) == 0 {
		return nil, fmt.Errorf("没有可重新生成的片段（所有片段可能已锁定或为空）")
	}

	chapterSegMap := make(map[uint64][]uint64)
	for _, seg := range eligible {
		chapterSegMap[seg.ChapterID] = append(chapterSegMap[seg.ChapterID], seg.SegmentID)
	}

	var totalCount int
	var firstTaskID uint64
	for chapterID, segIDs := range chapterSegMap {
		projectID, err := s.getProjectIDByChapter(ctx, chapterID)
		if err != nil {
			continue
		}
		task := &model.VsGenerationTask{
			ProjectID:    projectID,
			ChapterID:    &chapterID,
			TaskType:     "tts_generate",
			Status:       "running",
			TotalBatches: len(segIDs),
			SegmentIDs:   model.Uint64Array(segIDs),
		}
		if err := s.taskRepo.Create(ctx, task); err != nil {
			continue
		}
		if firstTaskID == 0 {
			firstTaskID = task.TaskID
		}
		if err := s.enqueueSegments(ctx, task, segIDs); err != nil {
			_ = s.taskRepo.UpdateStatus(ctx, task.TaskID, "failed", err.Error())
			continue
		}
		totalCount += len(segIDs)
	}

	if totalCount == 0 {
		return nil, fmt.Errorf("任务创建失败")
	}

	return &v1.BatchGenerateResponse{
		TaskID:       firstTaskID,
		TotalCount:   totalCount,
		SkippedCount: len(segments) - len(eligible),
	}, nil
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
