package service

import (
	"context"
	"fmt"
	"strings"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

// 正在排队/生成中的片段不允许迁移角色，避免 worker 用旧角色的音色写回结果
var activeSegmentStatuses = []string{"queued", "processing"}

// 已有音频的片段，合并后若音色变化需要重新生成
var generatedSegmentStatuses = []string{"generated", "locked"}

var characterLevelRank = map[string]int{"main": 3, "supporting": 2, "minor": 1}

func (s *vsCharacterService) Merge(ctx context.Context, request *v1.VsCharacterMergeRequest) (*v1.VsCharacterMergeResponse, error) {
	sourceIDs := uniqueIDs(request.SourceIDs)
	if len(sourceIDs) == 0 {
		return nil, fmt.Errorf("请选择被合并的角色")
	}
	for _, id := range sourceIDs {
		if id == request.TargetID {
			return nil, fmt.Errorf("目标角色不能出现在被合并列表中")
		}
	}

	loginName := getLoginName(ctx)
	resp := &v1.VsCharacterMergeResponse{TargetID: request.TargetID}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target model.VsCharacter
		if err := tx.Where("character_id = ?", request.TargetID).First(&target).Error; err != nil {
			return fmt.Errorf("目标角色不存在")
		}
		if target.MergedIntoID != nil {
			return fmt.Errorf("目标角色「%s」已被合并，请选择其合并后的角色", target.Name)
		}

		var sources []*model.VsCharacter
		if err := tx.Where("character_id IN ?", sourceIDs).Find(&sources).Error; err != nil {
			return err
		}
		if len(sources) != len(sourceIDs) {
			return fmt.Errorf("部分被合并角色不存在")
		}
		for _, src := range sources {
			if src.ProjectID != target.ProjectID {
				return fmt.Errorf("角色「%s」不属于同一项目，无法合并", src.Name)
			}
			if src.MergedIntoID != nil {
				return fmt.Errorf("角色「%s」已被合并过", src.Name)
			}
			if src.Name == narratorName {
				return fmt.Errorf("「旁白」是系统内置角色，不能被合并")
			}
		}

		active, err := countSegmentsByStatus(tx, sourceIDs, activeSegmentStatuses)
		if err != nil {
			return err
		}
		if active > 0 {
			return fmt.Errorf("被合并角色有 %d 个片段正在排队或生成中，请等待完成或取消后再合并", active)
		}

		affected, needRegen, err := mergeCharactersTx(tx, &target, sources, loginName)
		if err != nil {
			return err
		}
		resp.MergedCount = len(sources)
		resp.AffectedSegments = affected
		resp.NeedRegenerate = needRegen
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *vsCharacterService) MergeGenericToPool(ctx context.Context, projectID uint64, request *v1.VsCharacterMergeGenericRequest) (*v1.VsCharacterMergeGenericResponse, error) {
	characters, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	// 按路人池分组，保持 characterPool 的顺序输出
	grouped := make(map[string][]*model.VsCharacter)
	for _, c := range characters {
		if c.Name == narratorName || isPoolCharacterName(c.Name) || c.Level == "main" {
			continue
		}
		if poolName, ok := resolveGenericPoolName(c.Name, c.Gender); ok {
			grouped[poolName] = append(grouped[poolName], c)
		}
	}

	resp := &v1.VsCharacterMergeGenericResponse{DryRun: request.DryRun, Groups: []v1.VsCharacterMergeGenericGroup{}}
	if len(grouped) == 0 {
		return resp, nil
	}

	loginName := getLoginName(ctx)
	deptID := getDeptID(ctx)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, role := range characterPool {
			sources := grouped[role.Name]
			if len(sources) == 0 {
				continue
			}

			ids := make([]uint64, len(sources))
			for i, c := range sources {
				ids[i] = c.CharacterID
			}
			busy, err := activeSegmentCharacterIDs(tx, ids)
			if err != nil {
				return err
			}
			var mergeable []*model.VsCharacter
			for _, c := range sources {
				if busy[c.CharacterID] {
					resp.SkippedNames = append(resp.SkippedNames, c.Name)
					continue
				}
				mergeable = append(mergeable, c)
			}
			if len(mergeable) == 0 {
				continue
			}

			group := v1.VsCharacterMergeGenericGroup{PoolName: role.Name}
			for _, c := range mergeable {
				group.SourceNames = append(group.SourceNames, c.Name)
			}

			if request.DryRun {
				var existing model.VsCharacter
				if res := tx.Where("project_id = ? AND name = ?", projectID, role.Name).Limit(1).Find(&existing); res.Error == nil && res.RowsAffected > 0 {
					group.PoolID = existing.CharacterID
				}
				resp.Groups = append(resp.Groups, group)
				resp.MergedCount += len(mergeable)
				continue
			}

			pool, err := ensurePoolCharacterTx(tx, projectID, role.Name, loginName, deptID)
			if err != nil {
				return err
			}
			affected, needRegen, err := mergeCharactersTx(tx, pool, mergeable, loginName)
			if err != nil {
				return err
			}
			group.PoolID = pool.CharacterID
			resp.Groups = append(resp.Groups, group)
			resp.MergedCount += len(mergeable)
			resp.AffectedSegments += affected
			resp.NeedRegenerate += needRegen
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// mergeCharactersTx 在事务内将 sources 合并到 target：
// 迁移片段归属、拉平已有的合并链、标记被合并角色，并把名称/别名及缺失信息并入 target。
// 返回迁移的片段数，以及其中已生成音频且音色发生变化（建议重新生成）的片段数。
func mergeCharactersTx(tx *gorm.DB, target *model.VsCharacter, sources []*model.VsCharacter, loginName string) (int64, int64, error) {
	sourceIDs := make([]uint64, len(sources))
	for i, src := range sources {
		sourceIDs[i] = src.CharacterID
	}

	finalVoice := target.VoiceProfileID
	for _, src := range sources {
		if finalVoice == nil && src.VoiceProfileID != nil {
			v := *src.VoiceProfileID
			finalVoice = &v
		}
	}

	var voiceChangedIDs []uint64
	for _, src := range sources {
		if !sameVoiceProfile(src.VoiceProfileID, finalVoice) {
			voiceChangedIDs = append(voiceChangedIDs, src.CharacterID)
		}
	}
	var needRegen int64
	if len(voiceChangedIDs) > 0 {
		n, err := countSegmentsByStatus(tx, voiceChangedIDs, generatedSegmentStatuses)
		if err != nil {
			return 0, 0, err
		}
		needRegen = n
	}

	res := tx.Model(&model.VsScriptSegment{}).
		Where("character_id IN ?", sourceIDs).
		Updates(map[string]interface{}{"character_id": target.CharacterID, "updated_by": loginName})
	if res.Error != nil {
		return 0, 0, fmt.Errorf("迁移片段失败: %w", res.Error)
	}

	// 之前已合并到 sources 的角色改为直接指向 target，保证合并链只有一层
	if err := tx.Model(&model.VsCharacter{}).
		Where("merged_into_id IN ?", sourceIDs).
		Updates(map[string]interface{}{"merged_into_id": target.CharacterID, "updated_by": loginName}).Error; err != nil {
		return 0, 0, fmt.Errorf("更新合并链失败: %w", err)
	}

	if err := tx.Model(&model.VsCharacter{}).
		Where("character_id IN ?", sourceIDs).
		Updates(map[string]interface{}{"merged_into_id": target.CharacterID, "status": "1", "updated_by": loginName}).Error; err != nil {
		return 0, 0, fmt.Errorf("标记被合并角色失败: %w", err)
	}

	fields := map[string]interface{}{
		"aliases":    mergeCharacterAliases(target, sources),
		"updated_by": loginName,
	}
	if target.VoiceProfileID == nil && finalVoice != nil {
		fields["voice_profile_id"] = *finalVoice
	}
	// 路人池角色的性别、描述、层级是固定定义，不从被合并角色继承
	if !isPoolCharacterName(target.Name) {
		gender, description, level := target.Gender, target.Description, target.Level
		for _, src := range sources {
			if (gender == "" || gender == "unknown") && src.Gender != "" && src.Gender != "unknown" {
				gender = src.Gender
			}
			if strings.TrimSpace(description) == "" && strings.TrimSpace(src.Description) != "" {
				description = src.Description
			}
			if characterLevelRank[src.Level] > characterLevelRank[level] {
				level = src.Level
			}
		}
		fields["gender"] = gender
		fields["description"] = description
		fields["level"] = level
	}
	if err := tx.Model(&model.VsCharacter{}).
		Where("character_id = ?", target.CharacterID).
		Updates(fields).Error; err != nil {
		return 0, 0, fmt.Errorf("更新目标角色失败: %w", err)
	}

	return res.RowsAffected, needRegen, nil
}

// ensurePoolCharacterTx 查找或创建项目内的路人池角色。
// 与唯一索引一致，查询包含停用的记录；被停用的路人池角色会重新启用。
func ensurePoolCharacterTx(tx *gorm.DB, projectID uint64, poolName, loginName string, deptID uint) (*model.VsCharacter, error) {
	var pool model.VsCharacter
	res := tx.Unscoped().Where("project_id = ? AND name = ?", projectID, poolName).Limit(1).Find(&pool)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected > 0 {
		if pool.MergedIntoID != nil {
			var target model.VsCharacter
			if err := tx.Where("character_id = ?", *pool.MergedIntoID).First(&target).Error; err != nil {
				return nil, fmt.Errorf("路人池角色「%s」的合并目标不存在", poolName)
			}
			return &target, nil
		}
		if pool.DeletedAt.Valid || pool.Status != "0" {
			if err := tx.Unscoped().Model(&model.VsCharacter{}).
				Where("character_id = ?", pool.CharacterID).
				Updates(map[string]interface{}{"status": "0", "deleted_at": nil, "updated_by": loginName}).Error; err != nil {
				return nil, fmt.Errorf("启用路人池角色失败: %w", err)
			}
		}
		return &pool, nil
	}

	created := newPoolCharacter(projectID, poolName, loginName, deptID)
	if err := tx.Create(created).Error; err != nil {
		return nil, fmt.Errorf("创建路人池角色失败: %w", err)
	}
	return created, nil
}

func countSegmentsByStatus(tx *gorm.DB, characterIDs []uint64, statuses []string) (int64, error) {
	var count int64
	err := tx.Model(&model.VsScriptSegment{}).
		Where("character_id IN ? AND status IN ?", characterIDs, statuses).
		Count(&count).Error
	return count, err
}

func activeSegmentCharacterIDs(tx *gorm.DB, characterIDs []uint64) (map[uint64]bool, error) {
	var ids []uint64
	if err := tx.Model(&model.VsScriptSegment{}).
		Where("character_id IN ? AND status IN ?", characterIDs, activeSegmentStatuses).
		Distinct().Pluck("character_id", &ids).Error; err != nil {
		return nil, err
	}
	busy := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		busy[id] = true
	}
	return busy, nil
}

// mergeCharacterAliases 合并别名：保留 target 原有别名顺序，追加 sources 的名称与别名，大小写不敏感去重
func mergeCharacterAliases(target *model.VsCharacter, sources []*model.VsCharacter) model.StringList {
	seen := map[string]bool{strings.ToLower(target.Name): true}
	result := make(model.StringList, 0, len(target.Aliases)+len(sources))
	add := func(name string) {
		name = strings.TrimSpace(name)
		key := strings.ToLower(name)
		if name == "" || seen[key] {
			return
		}
		seen[key] = true
		result = append(result, name)
	}
	for _, a := range target.Aliases {
		add(a)
	}
	for _, src := range sources {
		add(src.Name)
		for _, a := range src.Aliases {
			add(a)
		}
	}
	return result
}

func sameVoiceProfile(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func uniqueIDs(ids []uint64) []uint64 {
	seen := make(map[uint64]bool, len(ids))
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
