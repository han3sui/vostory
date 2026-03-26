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

	"github.com/redis/go-redis/v9"
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
	SingleGenerate(ctx context.Context, segmentID uint64) (*v1.BatchGenerateResponse, error)
	BatchGenerate(ctx context.Context, chapterID uint64) (*v1.BatchGenerateResponse, error)
	GetTaskProgress(ctx context.Context, taskID uint64) (*v1.TaskProgressResponse, error)
	GetActiveTaskByChapter(ctx context.Context, chapterID uint64) (*v1.TaskProgressResponse, error)
	GetActiveTasksByProject(ctx context.Context, projectID uint64) ([]*v1.ProjectTaskProgressResponse, error)
	GetAudioClipFile(ctx context.Context, clipID uint64) (filePath string, contentType string, err error)
	LockSegment(ctx context.Context, segmentID uint64) error
	UnlockSegment(ctx context.Context, segmentID uint64) error
	BatchLockByChapter(ctx context.Context, chapterID uint64) (int64, error)
	BatchUnlockByChapter(ctx context.Context, chapterID uint64) (int64, error)
	CancelChapterQueue(ctx context.Context, chapterID uint64) (int64, error)
	CancelProjectQueue(ctx context.Context, projectID uint64) (int64, error)
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
	taskRepo repository.VsGenerationTaskRepository,
	rdb *redis.Client,
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
		taskRepo:              taskRepo,
		rdb:                   rdb,
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
	taskRepo              repository.VsGenerationTaskRepository
	rdb                   *redis.Client
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

	provider, err := s.resolveTTSProvider(ctx, voiceProfile, segment.ChapterID)
	if err != nil {
		return failAndReturn(fmt.Sprintf("TTS 提供商不可用: %v", err))
	}

	referenceAudioURL := voiceProfile.ReferenceAudioURL
	emotion, _ := s.voiceEmotionRepo.FindByMatch(ctx, voiceProfile.VoiceProfileID, segment.EmotionType, segment.EmotionStrength)
	if emotion != nil {
		referenceAudioURL = emotion.ReferenceAudioURL
	}

	if referenceAudioURL == "" {
		return failAndReturn(fmt.Sprintf("声音配置「%s」缺少参考音频", voiceProfile.Name))
	}

	_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	text := s.applyPronunciationDict(ctx, segment)

	emoVector, _ := buildEmotionVector(segment.EmotionType, segment.EmotionStrength)

	client := tts.NewClient(provider.APIBaseURL)

	remoteKey := filepath.Base(referenceAudioURL)
	if err := client.EnsureAudioUploaded(referenceAudioURL, remoteKey); err != nil {
		return failAndReturn(fmt.Sprintf("上传参考音频失败: %v", err))
	}

	audioData, err := client.Synthesize(text, remoteKey, emoVector, "")
	if err != nil {
		return failAndReturn(fmt.Sprintf("TTS 合成失败: %v", err))
	}

	audioURL, fileSize, err := s.saveAudioFile(segment, audioData)
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

func (s *vsTTSSynthesizeService) enqueueSegments(ctx context.Context, task *model.VsGenerationTask, segmentIDs []uint64) error {
	for _, segID := range segmentIDs {
		_ = s.segmentRepo.UpdateStatus(ctx, segID, "queued")

		msg := fmt.Sprintf("%d:%d", task.TaskID, segID)
		if err := s.rdb.LPush(ctx, "vs:tts:queue", msg).Err(); err != nil {
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

	chapterID := segment.ChapterID
	task := &model.VsGenerationTask{
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

	task := &model.VsGenerationTask{
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

func (s *vsTTSSynthesizeService) GetAudioClipFile(ctx context.Context, clipID uint64) (string, string, error) {
	clip, err := s.audioClipRepo.FindByID(ctx, clipID)
	if err != nil {
		return "", "", fmt.Errorf("音频片段不存在: %w", err)
	}

	if _, err := os.Stat(clip.AudioURL); err != nil {
		return "", "", fmt.Errorf("音频文件不存在: %w", err)
	}

	contentType := "audio/wav"
	if clip.Format == "mp3" {
		contentType = "audio/mpeg"
	} else if clip.Format == "flac" {
		contentType = "audio/flac"
	}

	return clip.AudioURL, contentType, nil
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
			Status:         t.Status,
			Progress:       t.Progress,
			TotalCount:     t.TotalBatches,
			CompletedCount: t.CompletedBatches,
			FailedCount:    t.FailedBatches,
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

func (s *vsTTSSynthesizeService) CancelProjectQueue(ctx context.Context, projectID uint64) (int64, error) {
	tasks, err := s.taskRepo.FindActiveByProjectID(ctx, projectID)
	if err != nil {
		return 0, err
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

// cancelTaskQueuedSegments 按 task.SegmentIDs 精确取消 queued 片段，并重算任务终态。
func (s *vsTTSSynthesizeService) cancelTaskQueuedSegments(ctx context.Context, task *model.VsGenerationTask) (int64, error) {
	if len(task.SegmentIDs) == 0 {
		return 0, nil
	}
	affected, err := s.segmentRepo.BatchUpdateStatus(ctx, []uint64(task.SegmentIDs), "queued", "cancelled")
	if err != nil {
		return 0, err
	}
	if affected > 0 {
		_ = s.taskRepo.ReduceTotalBatches(ctx, task.TaskID, int(affected))
	}

	// 重新读取 task，判断是否已全部处理完毕
	updated, err := s.taskRepo.FindByID(ctx, task.TaskID)
	if err != nil {
		return affected, nil
	}
	processed := updated.CompletedBatches + updated.FailedBatches
	if processed >= updated.TotalBatches {
		if updated.FailedBatches > 0 {
			_ = s.taskRepo.SetFailed(ctx, task.TaskID,
				fmt.Sprintf("%d/%d segments failed", updated.FailedBatches, updated.TotalBatches))
		} else {
			_ = s.taskRepo.SetCompleted(ctx, task.TaskID)
		}
	}
	return affected, nil
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
