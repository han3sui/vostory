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

// TTSPreparedSpeech 已备好参考音频的合成任务。Run 只打 GPU，Persist 落库，用完必须 Close。
type TTSPreparedSpeech struct {
	segment      *model.VsScriptSegment
	provider     *model.VsTTSProvider
	voiceProfile *model.VsVoiceProfile
	request      tts.SpeechRequest
	cleanup      func()
	closed       bool
}

// Close 删除参考音频临时文件。
func (j *TTSPreparedSpeech) Close() {
	if j == nil || j.closed {
		return
	}
	j.closed = true
	if j.cleanup != nil {
		j.cleanup()
	}
}

type VsTTSSynthesizeService interface {
	SynthesizeSegment(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error)
	PrepareSegmentSpeech(ctx context.Context, segmentID uint64) (*TTSPreparedSpeech, error)
	RunPreparedSpeech(ctx context.Context, job *TTSPreparedSpeech) ([]byte, error)
	PersistPreparedSpeech(ctx context.Context, job *TTSPreparedSpeech, audioData []byte) (*v1.TTSSynthesizeResponse, error)
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
	pronunciationDictRepo repository.VsPronunciationDictRepository
	audioClipRepo         repository.VsAudioClipRepository
	ttsProviderRepo       repository.VsTTSProviderRepository
	projectRepo           repository.VsProjectRepository
	taskRepo              repository.VsGenerationTaskRepository
	rdb                   *redis.Client
	fileStorage           filestorage.FileStorage
}

func (s *vsTTSSynthesizeService) failSegment(ctx context.Context, segmentID uint64, msg string) error {
	_ = s.segmentRepo.UpdateStatusWithError(ctx, segmentID, "failed", msg)
	return fmt.Errorf("%s", msg)
}

func (s *vsTTSSynthesizeService) SynthesizeSegment(ctx context.Context, segmentID uint64) (*v1.TTSSynthesizeResponse, error) {
	job, err := s.PrepareSegmentSpeech(ctx, segmentID)
	if err != nil {
		return nil, err
	}
	defer job.Close()
	audioData, err := s.RunPreparedSpeech(ctx, job)
	if err != nil {
		return nil, err
	}
	return s.PersistPreparedSpeech(ctx, job, audioData)
}

// PrepareSegmentSpeech 查库并下载参考音频，不占用 GPU。
func (s *vsTTSSynthesizeService) PrepareSegmentSpeech(ctx context.Context, segmentID uint64) (*TTSPreparedSpeech, error) {
	segment, err := s.segmentRepo.FindByID(ctx, segmentID)
	if err != nil {
		return nil, fmt.Errorf("片段不存在: %w", err)
	}
	if segment.CharacterID == nil {
		return nil, s.failSegment(ctx, segmentID, "该片段未关联角色，无法合成。请先为片段指定说话人（旁白/描述类型请指定「旁白」角色）")
	}

	character, err := s.characterRepo.FindByID(ctx, *segment.CharacterID)
	if err != nil {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("角色不存在: %v", err))
	}
	if character.VoiceProfileID == nil {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("角色「%s」未绑定声音配置，请先在角色管理中绑定", character.Name))
	}
	voiceProfile, err := s.voiceProfileRepo.FindByID(ctx, *character.VoiceProfileID)
	if err != nil {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("声音配置不存在: %v", err))
	}

	provider, err := s.resolveTTSProvider(ctx, segment.ChapterID)
	if err != nil {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("TTS 提供商不可用: %v", err))
	}

	referenceAudioURL := voiceProfile.ReferenceAudioURL
	referenceText := voiceProfile.ReferenceText
	if referenceAudioURL == "" {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("声音配置「%s」缺少参考音频，可在角色管理中「设计音色」自动生成", voiceProfile.Name))
	}
	if strings.TrimSpace(referenceText) == "" {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("声音配置「%s」缺少参考文本（参考音频的逐字转写），Breeze-TTS-2 克隆音色需要", voiceProfile.Name))
	}

	_ = s.segmentRepo.UpdateStatus(ctx, segmentID, "processing")

	refPath, cleanup, err := s.fileStorage.DownloadToTemp(ctx, filestorage.NormalizeKey(referenceAudioURL))
	if err != nil {
		return nil, s.failSegment(ctx, segmentID, fmt.Sprintf("参考音频不存在: %v", err))
	}

	return &TTSPreparedSpeech{
		segment:      segment,
		provider:     provider,
		voiceProfile: voiceProfile,
		request: tts.SpeechRequest{
			Text:         s.applyPronunciationDict(ctx, segment),
			RefAudioPath: refPath,
			RefText:      referenceText,
			Instruction:  strings.TrimSpace(segment.Instruction),
			CfgScale:     ttsParamFloat(voiceProfile.TTSParams, "cfg_scale"),
			Seed:         voiceSeed(voiceProfile),
		},
		cleanup: cleanup,
	}, nil
}

// RunPreparedSpeech 调用 TTS HTTP，这是唯一占用 GPU 的步骤。
func (s *vsTTSSynthesizeService) RunPreparedSpeech(ctx context.Context, job *TTSPreparedSpeech) ([]byte, error) {
	if job == nil {
		return nil, fmt.Errorf("合成任务为空")
	}
	client := tts.NewClient(job.provider.APIBaseURL, job.provider.APIKey)
	audioData, err := client.Speech(job.request)
	if err != nil {
		return nil, s.failSegment(ctx, job.segment.SegmentID, fmt.Sprintf("TTS 合成失败: %v", err))
	}
	return audioData, nil
}

// PersistPreparedSpeech 归一化、写文件、建版本，不占用 GPU。
func (s *vsTTSSynthesizeService) PersistPreparedSpeech(ctx context.Context, job *TTSPreparedSpeech, audioData []byte) (*v1.TTSSynthesizeResponse, error) {
	if job == nil {
		return nil, fmt.Errorf("合成任务为空")
	}
	normalizedData, normErr := audio.NormalizeLoudness(audioData, -16.0)
	if normErr != nil {
		s.logger.Warn(fmt.Sprintf("音量归一化失败，使用原始音频: %v", normErr))
	} else {
		audioData = normalizedData
	}

	audioURL, fileSize, err := s.saveAudioFile(ctx, job.segment, audioData)
	if err != nil {
		return nil, s.failSegment(ctx, job.segment.SegmentID, fmt.Sprintf("保存音频文件失败: %v", err))
	}
	clip, err := s.createAudioClip(ctx, job.segment, job.provider, job.voiceProfile, audioURL, fileSize)
	if err != nil {
		return nil, s.failSegment(ctx, job.segment.SegmentID, fmt.Sprintf("创建音频记录失败: %v", err))
	}
	_ = s.segmentRepo.UpdateStatus(ctx, job.segment.SegmentID, "generated")
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
		SegmentID:      segment.SegmentID,
		AudioURL:       audioURL,
		FileSize:       fileSize,
		Format:         "wav",
		TTSProviderID:  &providerID,
		VoiceProfileID: &profileID,
		Version:        maxVer + 1,
		IsCurrent:      "1",
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

// regenerateByCharacterIDs 只查出可生成的片段就返回，逐章建任务、入队放到后台。
// 旁白这类角色可能有几万个片段，同步做完要几十分钟；按章节有多个任务，所以不返回 TaskID
func (s *vsTTSSynthesizeService) regenerateByCharacterIDs(ctx context.Context, characterIDs []uint64) (*v1.BatchGenerateResponse, error) {
	refs, err := s.segmentRepo.FindRegenerableByCharacterIDs(ctx, characterIDs)
	if err != nil {
		return nil, fmt.Errorf("查询片段失败: %w", err)
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf("没有可重新生成的片段（所有片段可能已锁定、正在生成或为空）")
	}
	total, err := s.segmentRepo.CountByCharacterIDs(ctx, characterIDs)
	if err != nil {
		return nil, fmt.Errorf("查询片段失败: %w", err)
	}

	// gin 的 ctx 在请求结束后会被复用，后台只带走登录信息
	go s.regenerateInBackground(refs, getLoginName(ctx), getDeptID(ctx))

	return &v1.BatchGenerateResponse{
		TotalCount:   len(refs),
		SkippedCount: int(total) - len(refs),
	}, nil
}

func (s *vsTTSSynthesizeService) regenerateInBackground(refs []repository.SegmentRef, loginName string, deptID uint) {
	ctx := context.WithValue(context.Background(), "login_name", loginName)
	ctx = context.WithValue(ctx, "dept_id", deptID)
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(fmt.Sprintf("按角色重新生成入队时发生 panic: %v", r))
		}
	}()

	// refs 已按章节排好序，按章节顺序入队，前面的章节先合成
	var chapterOrder []uint64
	chapterSegs := make(map[uint64][]uint64)
	chapterProject := make(map[uint64]uint64)
	for _, ref := range refs {
		if _, ok := chapterSegs[ref.ChapterID]; !ok {
			chapterOrder = append(chapterOrder, ref.ChapterID)
			chapterProject[ref.ChapterID] = ref.ProjectID
		}
		chapterSegs[ref.ChapterID] = append(chapterSegs[ref.ChapterID], ref.SegmentID)
	}

	for _, chapterID := range chapterOrder {
		if err := s.regenerateChapterSegments(ctx, chapterProject[chapterID], chapterID, chapterSegs[chapterID]); err != nil {
			s.logger.Warn(fmt.Sprintf("章节 %d 重新生成入队失败: %v", chapterID, err))
		}
	}
}

// regenerateChapterSegments 认领片段和建任务放在同一个事务里：已被别的请求认领的片段不会再认领，
// 也就不会重复入队；服务若在入队前重启，worker 能按任务把已认领的片段恢复出来
func (s *vsTTSSynthesizeService) regenerateChapterSegments(ctx context.Context, projectID, chapterID uint64, segmentIDs []uint64) error {
	var task *model.VsGenerationTask
	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		claimed, err := s.segmentRepo.ClaimForGenerate(ctx, segmentIDs)
		if err != nil || len(claimed) == 0 {
			return err
		}
		task = &model.VsGenerationTask{
			ProjectID:    projectID,
			ChapterID:    &chapterID,
			TaskType:     "tts_generate",
			Status:       "running",
			TotalBatches: len(claimed),
			SegmentIDs:   model.Uint64Array(claimed),
		}
		return s.taskRepo.Create(ctx, task)
	})
	if err != nil || task == nil {
		return err
	}

	if err := s.pushSegmentMessages(ctx, task.TaskID, task.SegmentIDs); err != nil {
		// 已认领的片段退回 cancelled，之后还能重新生成，不会一直卡在 queued
		_, _ = s.cancelTaskQueuedSegments(ctx, task)
		return err
	}
	return nil
}

// pushSegmentMessages 分批 LPUSH，消费顺序与逐条 LPUSH 一致
func (s *vsTTSSynthesizeService) pushSegmentMessages(ctx context.Context, taskID uint64, segmentIDs []uint64) error {
	const batchSize = 500
	for start := 0; start < len(segmentIDs); start += batchSize {
		end := min(start+batchSize, len(segmentIDs))
		msgs := make([]interface{}, 0, end-start)
		for _, segID := range segmentIDs[start:end] {
			msgs = append(msgs, fmt.Sprintf("%d:%d", taskID, segID))
		}
		if err := s.rdb.LPush(ctx, ttsQueueKey, msgs...).Err(); err != nil {
			return fmt.Errorf("任务 %d 入队失败: %w", taskID, err)
		}
	}
	return nil
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
