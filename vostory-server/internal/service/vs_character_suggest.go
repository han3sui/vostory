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
	if !request.DryRun {
		// 确认合并必须显式传入勾选的分组，避免漏传参数时把扫描结果全部合并
		if len(request.Groups) == 0 {
			return nil, fmt.Errorf("请选择要合并的分组")
		}
		return s.applySuggestGroups(ctx, projectID, request.Groups)
	}

	characters, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	suggested := suggestMergeGroups(characters)
	resp := &v1.VsCharacterMergeSuggestResponse{DryRun: true, Groups: suggested}
	for _, g := range suggested {
		resp.MergedCount += len(g.SourceIDs)
	}
	return resp, nil
}

// suggestTarget 单个源角色的合并去向：characterID 非 0 为已有角色，否则为路人池
type suggestTarget struct {
	characterID uint64
	poolName    string
	rule        string
	reason      string
}

func suggestMergeGroups(characters []*model.VsCharacter) []v1.VsCharacterMergeSuggestGroup {
	idx := newNameIndex(characters)
	byID := make(map[uint64]*model.VsCharacter, len(characters))
	assigned := make(map[uint64]suggestTarget, len(characters))
	var sources []*model.VsCharacter

	for _, c := range characters {
		byID[c.CharacterID] = c
	}
	for _, src := range characters {
		if !canBeMergeSource(src) {
			continue
		}
		if _, done := assigned[src.CharacterID]; done {
			continue
		}
		res := resolveSpeaker(idx, src.Name, src.Gender, src.CharacterID)
		if res.Existing == nil && res.PoolName == "" {
			continue
		}
		if res.Existing != nil && (res.Existing.CharacterID == src.CharacterID || !canReceiveMerge(res.Existing)) {
			continue
		}
		t := suggestTarget{poolName: res.PoolName, rule: res.Rule, reason: res.Reason}
		if res.Existing != nil {
			t.characterID = res.Existing.CharacterID
		}
		assigned[src.CharacterID] = t
		sources = append(sources, src)
	}

	grouped := make(map[string]*v1.VsCharacterMergeSuggestGroup)
	order := make([]string, 0)
	for _, src := range sources {
		t, ok := resolveSuggestChain(assigned, assigned[src.CharacterID])
		if !ok || t.characterID == src.CharacterID {
			continue
		}

		var key, targetKind, targetName string
		if t.characterID != 0 {
			target := byID[t.characterID]
			if target == nil {
				continue
			}
			key = "c:" + strconv.FormatUint(t.characterID, 10)
			targetKind = "character"
			targetName = target.Name
		} else {
			key = "p:" + t.poolName
			targetKind = "pool"
			targetName = t.poolName
		}

		g, ok := grouped[key]
		if !ok {
			g = &v1.VsCharacterMergeSuggestGroup{
				Rule:       t.rule,
				Reason:     t.reason,
				TargetKind: targetKind,
				TargetID:   t.characterID,
				TargetName: targetName,
			}
			grouped[key] = g
			order = append(order, key)
		}
		g.SourceIDs = append(g.SourceIDs, src.CharacterID)
		g.SourceNames = append(g.SourceNames, src.Name)
	}

	result := make([]v1.VsCharacterMergeSuggestGroup, 0, len(order))
	for _, key := range order {
		result = append(result, *grouped[key])
	}
	return result
}

// resolveSuggestChain 目标本身也被建议合并时（如 不具名长老→长老→路人池），沿链找到最终去向，
// 保留源角色自己的规则与原因。出现环时返回 false，该源角色不给出建议。
func resolveSuggestChain(assigned map[uint64]suggestTarget, t suggestTarget) (suggestTarget, bool) {
	visited := map[uint64]bool{}
	for t.characterID != 0 {
		next, ok := assigned[t.characterID]
		if !ok {
			return t, true
		}
		if visited[t.characterID] {
			return t, false
		}
		visited[t.characterID] = true
		t = suggestTarget{characterID: next.characterID, poolName: next.poolName, rule: t.rule, reason: t.reason}
	}
	return t, true
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
	return findFinalMergeTargetTx(tx, projectID, g.TargetID)
}

// findFinalMergeTargetTx 目标已被合并（例如同一批次中先作为源角色合并掉）时，顺着 merged_into_id 找到最终角色
func findFinalMergeTargetTx(tx *gorm.DB, projectID, targetID uint64) (*model.VsCharacter, error) {
	visited := map[uint64]bool{}
	id := targetID
	for {
		if visited[id] {
			return nil, fmt.Errorf("角色合并关系出现循环")
		}
		visited[id] = true
		var target model.VsCharacter
		if err := tx.Where("project_id = ? AND character_id = ?", projectID, id).First(&target).Error; err != nil {
			return nil, fmt.Errorf("目标角色不存在")
		}
		if target.MergedIntoID == nil {
			return &target, nil
		}
		id = *target.MergedIntoID
	}
}
