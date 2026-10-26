package service

import (
	"context"
	"fmt"
	"strconv"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

func (s *vsCharacterService) SuggestMerge(ctx context.Context, projectID uint64, request *v1.VsCharacterMergeSuggestRequest) (*v1.VsCharacterMergeSuggestResponse, error) {
	if request == nil {
		request = &v1.VsCharacterMergeSuggestRequest{DryRun: true}
	}
	characters, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	suggested := suggestMergeGroups(characters)
	resp := &v1.VsCharacterMergeSuggestResponse{DryRun: request.DryRun, Groups: suggested}
	for _, g := range suggested {
		resp.MergedCount += len(g.SourceIDs)
	}
	if request.DryRun {
		return resp, nil
	}

	toApply := request.Groups
	if len(toApply) == 0 {
		toApply = suggested
	}
	return s.applySuggestGroups(ctx, projectID, toApply)
}

func suggestMergeGroups(characters []*model.VsCharacter) []v1.VsCharacterMergeSuggestGroup {
	idx := newNameIndex(characters)
	used := make(map[uint64]bool, len(characters))
	grouped := make(map[string]*v1.VsCharacterMergeSuggestGroup)
	order := make([]string, 0)

	for _, src := range characters {
		if !canBeMergeSource(src) || used[src.CharacterID] {
			continue
		}
		res := resolveSpeaker(idx, src.Name, src.Gender, src.CharacterID)
		if res.Existing == nil && res.PoolName == "" {
			continue
		}
		if res.Existing != nil && (res.Existing.CharacterID == src.CharacterID || !canReceiveMerge(res.Existing)) {
			continue
		}

		var key, targetKind, targetName string
		var targetID uint64
		if res.Existing != nil {
			key = "c:" + strconv.FormatUint(res.Existing.CharacterID, 10)
			targetKind = "character"
			targetName = res.Existing.Name
			targetID = res.Existing.CharacterID
		} else {
			key = "p:" + res.PoolName
			targetKind = "pool"
			targetName = res.PoolName
		}

		g, ok := grouped[key]
		if !ok {
			g = &v1.VsCharacterMergeSuggestGroup{
				Rule:       res.Rule,
				Reason:     res.Reason,
				TargetKind: targetKind,
				TargetID:   targetID,
				TargetName: targetName,
			}
			grouped[key] = g
			order = append(order, key)
		}
		g.SourceIDs = append(g.SourceIDs, src.CharacterID)
		g.SourceNames = append(g.SourceNames, src.Name)
		used[src.CharacterID] = true
	}

	result := make([]v1.VsCharacterMergeSuggestGroup, 0, len(order))
	for _, key := range order {
		result = append(result, *grouped[key])
	}
	return result
}

func canReceiveMerge(c *model.VsCharacter) bool {
	return c != nil && c.MergedIntoID == nil && c.Name != narratorName
}

func (s *vsCharacterService) applySuggestGroups(ctx context.Context, projectID uint64, groups []v1.VsCharacterMergeSuggestGroup) (*v1.VsCharacterMergeSuggestResponse, error) {
	loginName := getLoginName(ctx)
	deptID := getDeptID(ctx)
	resp := &v1.VsCharacterMergeSuggestResponse{DryRun: false, Groups: []v1.VsCharacterMergeSuggestGroup{}}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seen := map[uint64]bool{}
		for _, g := range groups {
			sourceIDs := uniqueIDs(g.SourceIDs)
			if len(sourceIDs) == 0 {
				continue
			}
			for _, id := range sourceIDs {
				if seen[id] {
					return fmt.Errorf("角色被重复选入多个合并组")
				}
				seen[id] = true
			}

			var sources []*model.VsCharacter
			if err := tx.Where("project_id = ? AND character_id IN ?", projectID, sourceIDs).Find(&sources).Error; err != nil {
				return err
			}
			if len(sources) != len(sourceIDs) {
				return fmt.Errorf("部分被合并角色不存在或不属于该项目")
			}

			target, err := resolveSuggestTarget(tx, projectID, g, loginName, deptID)
			if err != nil {
				return err
			}
			if seen[target.CharacterID] && targetKindOf(g) == "character" {
				return fmt.Errorf("目标角色「%s」不能同时作为被合并角色", target.Name)
			}

			var mergeable []*model.VsCharacter
			busy, err := activeSegmentCharacterIDs(tx, sourceIDs)
			if err != nil {
				return err
			}
			applied := v1.VsCharacterMergeSuggestGroup{
				Rule:       g.Rule,
				Reason:     g.Reason,
				TargetKind: targetKindOf(g),
				TargetID:   target.CharacterID,
				TargetName: target.Name,
			}
			for _, src := range sources {
				if src.CharacterID == target.CharacterID || !canBeMergeSource(src) {
					continue
				}
				if busy[src.CharacterID] {
					resp.SkippedNames = append(resp.SkippedNames, src.Name)
					continue
				}
				mergeable = append(mergeable, src)
				applied.SourceIDs = append(applied.SourceIDs, src.CharacterID)
				applied.SourceNames = append(applied.SourceNames, src.Name)
			}
			if len(mergeable) == 0 {
				continue
			}

			affected, needRegen, err := mergeCharactersTx(tx, target, mergeable, loginName)
			if err != nil {
				return err
			}
			resp.Groups = append(resp.Groups, applied)
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

func targetKindOf(g v1.VsCharacterMergeSuggestGroup) string {
	if g.TargetKind == "pool" || (g.TargetID == 0 && isPoolCharacterName(g.TargetName)) {
		return "pool"
	}
	return "character"
}

func resolveSuggestTarget(tx *gorm.DB, projectID uint64, g v1.VsCharacterMergeSuggestGroup, loginName string, deptID uint) (*model.VsCharacter, error) {
	if targetKindOf(g) == "pool" {
		if !isPoolCharacterName(g.TargetName) {
			return nil, fmt.Errorf("无效的路人池名称「%s」", g.TargetName)
		}
		return ensurePoolCharacterTx(tx, projectID, g.TargetName, loginName, deptID)
	}
	if g.TargetID == 0 {
		return nil, fmt.Errorf("缺少合并目标角色")
	}
	var target model.VsCharacter
	if err := tx.Where("project_id = ? AND character_id = ?", projectID, g.TargetID).First(&target).Error; err != nil {
		return nil, fmt.Errorf("目标角色不存在")
	}
	if target.MergedIntoID != nil {
		return nil, fmt.Errorf("目标角色「%s」已被合并", target.Name)
	}
	return &target, nil
}
