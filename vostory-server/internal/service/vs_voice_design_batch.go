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

// ProcessItem 由后台 worker 调用：LLM 撰写音色描述 → Voice Design 生成音频 →
// 需要确认的保存为草稿，其余直接保存为声音配置并绑定角色
func (s *vsVoiceDesignService) ProcessItem(ctx context.Context, itemID uint64) error {
	claimed, err := s.itemRepo.Claim(ctx, itemID)
	if err != nil || !claimed {
		return err
	}
	item, err := s.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	// 后台任务没有请求上下文，沿用发起人作为操作人
	ctx = context.WithValue(ctx, "login_name", item.CreatedBy)
	ctx = context.WithValue(ctx, "dept_id", item.DeptID)

	fail := func(err error) error {
		return s.itemRepo.UpdateFields(ctx, itemID, map[string]interface{}{
			"status":        repository.VoiceDesignItemFailed,
			"error_message": truncate(err.Error(), 300),
		})
	}

	character, project, err := s.loadCharacter(ctx, item.CharacterID)
	if err != nil {
		return fail(err)
	}
	if item.NeedConfirm != "1" && character.VoiceProfileID != nil {
		return s.itemRepo.UpdateFields(ctx, itemID, map[string]interface{}{
			"status":        repository.VoiceDesignItemSkipped,
			"error_message": "角色已绑定声音",
		})
	}

	suggestion, err := s.Suggest(ctx, character.CharacterID)
	if err != nil {
		return fail(err)
	}
	data, seed, err := s.synthesizeDesign(ctx, project, suggestion.Instruction, suggestion.SampleText, nil)
	if err != nil {
		return fail(err)
	}
	draftPath := fmt.Sprintf("%sdraft_%d_%d.wav", voiceDesignDraftPrefix, itemID, time.Now().UnixMilli())
	if err := s.fileStorage.Put(ctx, draftPath, bytes.NewReader(data), int64(len(data)), "audio/wav"); err != nil {
		return fail(fmt.Errorf("保存音频失败: %w", err))
	}

	fields := map[string]interface{}{
		"instruction":   suggestion.Instruction,
		"sample_text":   suggestion.SampleText,
		"seed":          seed,
		"error_message": "",
	}
	if item.NeedConfirm == "1" {
		fields["status"] = repository.VoiceDesignItemDraft
		fields["preview_path"] = draftPath
		return s.itemRepo.UpdateFields(ctx, itemID, fields)
	}

	applied, err := s.saveDesignedVoice(ctx, character, project, draftPath, "", suggestion.Instruction, suggestion.SampleText, seed)
	if err != nil {
		_ = s.fileStorage.Delete(ctx, draftPath)
		return fail(err)
	}
	fields["status"] = repository.VoiceDesignItemApplied
	fields["voice_profile_id"] = applied.VoiceProfileID
	return s.itemRepo.UpdateFields(ctx, itemID, fields)
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
