package service

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/filestorage"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

// VoiceDesignPrepared LLM 已写好描述，只等 GPU 生成。Finish / Fail 用 job 自带的 ctx。
type VoiceDesignPrepared struct {
	ctx        context.Context
	item       *model.VsVoiceDesignItem
	character  *model.VsCharacter
	project    *model.VsProject
	suggestion *v1.VoiceDesignSuggestResponse
}

func (j *VoiceDesignPrepared) ItemID() uint64 {
	if j == nil || j.item == nil {
		return 0
	}
	return j.item.ItemID
}

func (s *vsVoiceDesignService) StartBatch(ctx context.Context, projectID uint64, request *v1.VoiceDesignBatchRequest) (*v1.VoiceDesignBatchResponse, error) {
	active, err := s.itemRepo.CountActiveByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if active > 0 {
		return nil, fmt.Errorf("该项目已有音色设计任务在进行中，请等待完成或停止后再开始")
	}

	draftIDs, err := s.itemRepo.FindDraftCharacterIDs(ctx, projectID)
	if err != nil {
		return nil, err
	}
	hasDraft := make(map[uint64]bool, len(draftIDs))
	for _, id := range draftIDs {
		hasDraft[id] = true
	}

	ids := uniqueIDs(request.CharacterIDs)
	chars, err := s.characterRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	batchID := uint64(time.Now().UnixMilli())
	loginName := getLoginName(ctx)
	deptID := getDeptID(ctx)
	var items []*model.VsVoiceDesignItem
	var skipped []string
	for _, id := range ids {
		c, ok := chars[id]
		if !ok || c.ProjectID != projectID {
			skipped = append(skipped, fmt.Sprintf("角色 %d：不存在或不属于该项目", id))
			continue
		}
		switch {
		case c.MergedIntoID != nil:
			skipped = append(skipped, fmt.Sprintf("%s：已被合并", c.Name))
			continue
		case c.VoiceProfileID != nil:
			skipped = append(skipped, fmt.Sprintf("%s：已绑定声音", c.Name))
			continue
		case hasDraft[c.CharacterID]:
			skipped = append(skipped, fmt.Sprintf("%s：已有待确认的草稿", c.Name))
			continue
		}
		needConfirm := "0"
		if request.ConfirmImportant && isImportantForVoice(c) {
			needConfirm = "1"
		}
		items = append(items, &model.VsVoiceDesignItem{
			ProjectID:   projectID,
			BatchID:     batchID,
			CharacterID: c.CharacterID,
			NeedConfirm: needConfirm,
			Status:      repository.VoiceDesignItemQueued,
			BaseModel:   model.BaseModel{CreatedBy: loginName, DeptID: deptID},
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("没有需要生成音色的角色")
	}
	if err := s.itemRepo.CreateBatch(ctx, items); err != nil {
		return nil, fmt.Errorf("创建音色设计任务失败: %w", err)
	}

	resp, err := s.GetBatch(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resp.Skipped = skipped
	return resp, nil
}

// isImportantForVoice 主角、配角和旁白的音色影响整本书，批量生成时需要试听确认
func isImportantForVoice(c *model.VsCharacter) bool {
	if isPoolCharacterName(c.Name) {
		return false
	}
	return c.Name == narratorName || c.Level == "main" || c.Level == "supporting"
}

func (s *vsVoiceDesignService) GetBatch(ctx context.Context, projectID uint64) (*v1.VoiceDesignBatchResponse, error) {
	items, err := s.itemRepo.FindLatestByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resp := &v1.VoiceDesignBatchResponse{Items: []v1.VoiceDesignBatchItem{}}
	if len(items) == 0 {
		return resp, nil
	}

	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.CharacterID)
		if item.BatchID > resp.BatchID {
			resp.BatchID = item.BatchID
		}
		if item.Status == repository.VoiceDesignItemQueued || item.Status == repository.VoiceDesignItemRunning {
			resp.Running = true
		}
	}
	chars, err := s.characterRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		out := v1.VoiceDesignBatchItem{
			ItemID:         item.ItemID,
			BatchID:        item.BatchID,
			CharacterID:    item.CharacterID,
			NeedConfirm:    item.NeedConfirm == "1",
			Status:         item.Status,
			Instruction:    item.Instruction,
			SampleText:     item.SampleText,
			Seed:           item.Seed,
			VoiceProfileID: item.VoiceProfileID,
			ErrorMessage:   item.ErrorMessage,
		}
		if c, ok := chars[item.CharacterID]; ok {
			out.CharacterName = c.Name
			out.CharacterLevel = c.Level
		}
		resp.Items = append(resp.Items, out)
	}
	return resp, nil
}

func (s *vsVoiceDesignService) CancelBatch(ctx context.Context, projectID uint64) (int64, error) {
	return s.itemRepo.CancelQueuedByProject(ctx, projectID)
}

// ProcessItem 串行跑完一条，测试和单条重试用；批量走 worker 流水线。
func (s *vsVoiceDesignService) ProcessItem(ctx context.Context, itemID uint64) error {
	job, err := s.PrepareItem(ctx, itemID)
	if err != nil || job == nil {
		return err
	}
	data, seed, err := s.RunPreparedDesign(ctx, job)
	if err != nil {
		return s.FailPreparedDesign(job, err)
	}
	return s.FinishPreparedDesign(job, data, seed)
}

// PrepareItem 抢占任务并调 LLM，不占 GPU。未抢到或跳过时返回 nil, nil。
func (s *vsVoiceDesignService) PrepareItem(ctx context.Context, itemID uint64) (*VoiceDesignPrepared, error) {
	claimed, err := s.itemRepo.Claim(ctx, itemID)
	if err != nil || !claimed {
		return nil, err
	}
	item, err := s.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	// 与 worker 生命周期解耦，落库时即使请求 ctx 取消也要写完
	jobCtx := context.WithValue(context.WithoutCancel(ctx), "login_name", item.CreatedBy)
	jobCtx = context.WithValue(jobCtx, "dept_id", item.DeptID)

	character, project, err := s.loadCharacter(jobCtx, item.CharacterID)
	if err != nil {
		return nil, s.failItem(jobCtx, itemID, err)
	}
	if item.NeedConfirm != "1" && character.VoiceProfileID != nil {
		_ = s.itemRepo.UpdateFields(jobCtx, itemID, map[string]interface{}{
			"status":        repository.VoiceDesignItemSkipped,
			"error_message": "角色已绑定声音",
		})
		return nil, nil
	}

	suggestion, err := s.Suggest(jobCtx, character.CharacterID)
	if err != nil {
		return nil, s.failItem(jobCtx, itemID, err)
	}
	return &VoiceDesignPrepared{
		ctx:        jobCtx,
		item:       item,
		character:  character,
		project:    project,
		suggestion: suggestion,
	}, nil
}

// RunPreparedDesign 调用 Voice Design，这是唯一占用 GPU 的步骤。
func (s *vsVoiceDesignService) RunPreparedDesign(ctx context.Context, job *VoiceDesignPrepared) ([]byte, int64, error) {
	if job == nil || job.suggestion == nil {
		return nil, 0, fmt.Errorf("音色设计任务为空")
	}
	return s.synthesizeDesign(ctx, job.project, job.suggestion.Instruction, job.suggestion.SampleText, nil)
}

// FinishPreparedDesign 写对象存储并绑定角色，不占 GPU。
func (s *vsVoiceDesignService) FinishPreparedDesign(job *VoiceDesignPrepared, data []byte, seed int64) error {
	if job == nil {
		return fmt.Errorf("音色设计任务为空")
	}
	draftPath := fmt.Sprintf("%sdraft_%d_%d.wav", voiceDesignDraftPrefix, job.item.ItemID, time.Now().UnixMilli())
	if err := s.fileStorage.Put(job.ctx, draftPath, bytes.NewReader(data), int64(len(data)), "audio/wav"); err != nil {
		return s.FailPreparedDesign(job, fmt.Errorf("保存音频失败: %w", err))
	}

	fields := map[string]interface{}{
		"instruction":   job.suggestion.Instruction,
		"sample_text":   job.suggestion.SampleText,
		"seed":          seed,
		"error_message": "",
	}
	if job.item.NeedConfirm == "1" {
		fields["status"] = repository.VoiceDesignItemDraft
		fields["preview_path"] = draftPath
		return s.itemRepo.UpdateFields(job.ctx, job.item.ItemID, fields)
	}

	applied, err := s.saveDesignedVoice(job.ctx, job.character, job.project, draftPath, "", job.suggestion.Instruction, job.suggestion.SampleText, seed)
	if err != nil {
		_ = s.fileStorage.Delete(job.ctx, draftPath)
		return s.FailPreparedDesign(job, err)
	}
	fields["status"] = repository.VoiceDesignItemApplied
	fields["voice_profile_id"] = applied.VoiceProfileID
	return s.itemRepo.UpdateFields(job.ctx, job.item.ItemID, fields)
}

func (s *vsVoiceDesignService) FailPreparedDesign(job *VoiceDesignPrepared, cause error) error {
	if job == nil {
		return cause
	}
	return s.failItem(job.ctx, job.item.ItemID, cause)
}

func (s *vsVoiceDesignService) failItem(ctx context.Context, itemID uint64, cause error) error {
	_ = s.itemRepo.UpdateFields(ctx, itemID, map[string]interface{}{
		"status":        repository.VoiceDesignItemFailed,
		"error_message": truncate(cause.Error(), 300),
	})
	return cause
}

func (s *vsVoiceDesignService) loadDraft(ctx context.Context, itemID uint64) (*model.VsVoiceDesignItem, error) {
	item, err := s.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("任务项不存在")
	}
	if item.Status != repository.VoiceDesignItemDraft {
		return nil, fmt.Errorf("该任务项不是待确认的草稿")
	}
	return item, nil
}

// ApplyItem 确认草稿：保存为声音配置并绑定角色（会替换角色当前绑定的声音）
func (s *vsVoiceDesignService) ApplyItem(ctx context.Context, itemID uint64) (*v1.VoiceDesignApplyResponse, error) {
	item, err := s.loadDraft(ctx, itemID)
	if err != nil {
		return nil, err
	}
	character, project, err := s.loadCharacter(ctx, item.CharacterID)
	if err != nil {
		return nil, err
	}
	draftKey, err := s.validateDraftKey(ctx, item.PreviewPath)
	if err != nil {
		return nil, err
	}
	applied, err := s.saveDesignedVoice(ctx, character, project, draftKey, "", item.Instruction, item.SampleText, item.Seed)
	if err != nil {
		return nil, err
	}
	if err := s.itemRepo.UpdateFields(ctx, itemID, map[string]interface{}{
		"status":           repository.VoiceDesignItemApplied,
		"preview_path":     "",
		"voice_profile_id": applied.VoiceProfileID,
	}); err != nil {
		return nil, err
	}
	return applied, nil
}

// RegenerateItem 丢弃当前草稿（或失败结果）并重新排队，由 worker 重新撰写描述并生成
func (s *vsVoiceDesignService) RegenerateItem(ctx context.Context, itemID uint64) error {
	item, err := s.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("任务项不存在")
	}
	if item.Status != repository.VoiceDesignItemDraft && item.Status != repository.VoiceDesignItemFailed {
		return fmt.Errorf("只有草稿或失败的任务项可以重新生成")
	}
	s.removeObject(ctx, item.PreviewPath)
	return s.itemRepo.UpdateFields(ctx, itemID, map[string]interface{}{
		"status":        repository.VoiceDesignItemQueued,
		"preview_path":  "",
		"error_message": "",
		"updated_at":    time.Now(), // 让 NextQueued 把它排到队首
	})
}

func (s *vsVoiceDesignService) DiscardItem(ctx context.Context, itemID uint64) error {
	item, err := s.loadDraft(ctx, itemID)
	if err != nil {
		return err
	}
	s.removeObject(ctx, item.PreviewPath)
	return s.itemRepo.UpdateFields(ctx, itemID, map[string]interface{}{
		"status":       repository.VoiceDesignItemDiscarded,
		"preview_path": "",
	})
}

func (s *vsVoiceDesignService) ItemAudioPath(ctx context.Context, itemID uint64) (string, error) {
	item, err := s.loadDraft(ctx, itemID)
	if err != nil {
		return "", err
	}
	return s.validateDraftKey(ctx, item.PreviewPath)
}

// validateDraftKey 只允许访问草稿目录下的对象
func (s *vsVoiceDesignService) validateDraftKey(ctx context.Context, p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("草稿音频不存在，请重新生成")
	}
	key := filestorage.NormalizeKey(p)
	base := path.Base(key)
	if !strings.HasPrefix(key, voiceDesignDraftPrefix) || !strings.HasPrefix(base, "draft_") {
		return "", fmt.Errorf("无效的草稿音频")
	}
	if !filestorage.Exists(ctx, s.fileStorage, key) {
		return "", fmt.Errorf("草稿音频不存在，请重新生成")
	}
	return key, nil
}

func (s *vsVoiceDesignService) removeObject(ctx context.Context, p string) {
	if key, err := s.validateDraftKey(ctx, p); err == nil {
		_ = s.fileStorage.Delete(ctx, key)
	}
}
