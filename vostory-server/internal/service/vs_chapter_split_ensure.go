package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"iot-alert-center/internal/model"
)

type splitSpeaker struct {
	name        string
	gender      string
	description string
	level       string
	sortOrder   int
}

// ensureCharacters 收集 LLM 角色名，先按规则归一到已有角色或路人池，再创建无法归并的新角色。
func (s *vsChapterSplitService) ensureCharacters(
	ctx context.Context,
	result *llmSplitResult,
	projectID uint64,
	existing []*model.VsCharacter,
	charMap map[string]uint64,
	loginName string,
	deptID uint,
) (int, error) {
	speakers := collectSplitSpeakers(result, charMap)
	idx := newNameIndex(existing)
	aliasUpdates := map[uint64]model.StringList{}
	genericNames := map[string]string{}
	needed := map[string]*splitSpeaker{}

	bindExisting := func(raw string, c *model.VsCharacter, aliases []string) {
		id := c.CharacterID
		if c.MergedIntoID != nil {
			id = *c.MergedIntoID
		}
		charMap[strings.ToLower(raw)] = id
		if len(aliases) == 0 {
			return
		}
		merged := appendUniqueAliases(c.Aliases, aliases, c.Name)
		if len(merged) == len(c.Aliases) {
			return
		}
		c.Aliases = merged
		idx.add(c)
		aliasUpdates[c.CharacterID] = merged
	}

	keys := make([]string, 0, len(speakers))
	for key := range speakers {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := speakers[keys[i]], speakers[keys[j]]
		aParen, bParen := speakerNameHasParen(a.name), speakerNameHasParen(b.name)
		if aParen != bParen {
			return !aParen
		}
		return a.name < b.name
	})

	deferred := map[string]*splitSpeaker{}
	for _, key := range keys {
		info := speakers[key]
		if resolveCharacterID(charMap, info.name) != nil {
			continue
		}
		if info.name == narratorName {
			needed[key] = info
			continue
		}
		res := resolveSpeaker(idx, info.name, info.gender, 0)
		if res.Existing != nil {
			bindExisting(info.name, res.Existing, res.AddAliases)
			continue
		}
		if res.PoolName != "" {
			genericNames[key] = res.PoolName
			queuePoolSpeaker(needed, charMap, res.PoolName)
			continue
		}
		// 括号/不具名变体等本名入库后再归一，避免先建成独立角色
		if speakerNameHasParen(info.name) || stripUnnamedPrefix(info.name) != info.name {
			deferred[key] = info
			continue
		}
		needed[key] = info
	}

	created, err := s.createMissingCharacters(ctx, projectID, needed, charMap, idx, loginName, deptID)
	if err != nil {
		return 0, err
	}

	leftover := map[string]*splitSpeaker{}
	for key, info := range deferred {
		if resolveCharacterID(charMap, info.name) != nil {
			continue
		}
		res := resolveSpeaker(idx, info.name, info.gender, 0)
		if res.Existing != nil {
			bindExisting(info.name, res.Existing, res.AddAliases)
			continue
		}
		if res.PoolName != "" {
			genericNames[key] = res.PoolName
			queuePoolSpeaker(leftover, charMap, res.PoolName)
			continue
		}
		leftover[key] = info
	}
	n, err := s.createMissingCharacters(ctx, projectID, leftover, charMap, idx, loginName, deptID)
	if err != nil {
		return created, err
	}
	created += n

	for key, poolName := range genericNames {
		if id, ok := charMap[strings.ToLower(poolName)]; ok {
			charMap[key] = id
		}
	}
	if err := s.persistResolvedAliases(ctx, aliasUpdates, loginName); err != nil {
		return created, err
	}
	return created, nil
}

func collectSplitSpeakers(result *llmSplitResult, charMap map[string]uint64) map[string]*splitSpeaker {
	speakers := make(map[string]*splitSpeaker)
	if resolveCharacterID(charMap, narratorName) == nil {
		speakers[strings.ToLower(narratorName)] = &splitSpeaker{
			name:        narratorName,
			gender:      "unknown",
			description: "系统内置旁白角色，用于旁白与描述类型片段的语音合成",
			level:       "supporting",
			sortOrder:   9999,
		}
	}
	for _, sc := range result.Scenes {
		for _, seg := range sc.Segments {
			segType := normalizeSegmentType(seg.Type)
			if segType == "narration" || segType == "description" {
				continue
			}
			charName := strings.TrimSpace(seg.Character)
			if charName == "" {
				continue
			}
			key := strings.ToLower(charName)
			if _, exists := speakers[key]; exists {
				continue
			}
			speakers[key] = &splitSpeaker{
				name:        charName,
				gender:      normalizeGender(seg.CharacterGender),
				description: strings.TrimSpace(seg.CharacterDescription),
				level:       "minor",
			}
		}
	}
	return speakers
}

func queuePoolSpeaker(needed map[string]*splitSpeaker, charMap map[string]uint64, poolName string) {
	poolKey := strings.ToLower(poolName)
	if _, exists := charMap[poolKey]; exists {
		return
	}
	if _, queued := needed[poolKey]; queued {
		return
	}
	role, _ := findPoolRole(poolName)
	needed[poolKey] = &splitSpeaker{
		name:        role.Name,
		gender:      role.Gender,
		description: role.Description,
		level:       "minor",
		sortOrder:   poolSortOrder(role.Name),
	}
}

func speakerNameHasParen(name string) bool {
	_, _, ok := splitParenName(name)
	return ok
}

func (s *vsChapterSplitService) createMissingCharacters(
	ctx context.Context,
	projectID uint64,
	needed map[string]*splitSpeaker,
	charMap map[string]uint64,
	idx *nameIndex,
	loginName string,
	deptID uint,
) (int, error) {
	if len(needed) == 0 {
		return 0, nil
	}

	neededNames := make([]string, 0, len(needed))
	for _, info := range needed {
		neededNames = append(neededNames, info.name)
	}
	var existingChars []*model.VsCharacter
	if err := s.db.WithContext(ctx).Unscoped().
		Where("project_id = ? AND name IN ?", projectID, neededNames).
		Find(&existingChars).Error; err == nil {
		for _, ec := range existingChars {
			key := strings.ToLower(ec.Name)
			if ec.MergedIntoID != nil {
				charMap[key] = *ec.MergedIntoID
			} else {
				charMap[key] = ec.CharacterID
			}
			idx.add(ec)
			delete(needed, key)
		}
	}
	if len(needed) == 0 {
		return 0, nil
	}

	toCreate := make([]*model.VsCharacter, 0, len(needed))
	for _, info := range needed {
		toCreate = append(toCreate, &model.VsCharacter{
			ProjectID:   projectID,
			Name:        info.name,
			Level:       info.level,
			Gender:      info.gender,
			Description: info.description,
			Status:      "0",
			SortOrder:   info.sortOrder,
			BaseModel: model.BaseModel{
				CreatedBy: loginName,
				DeptID:    deptID,
			},
		})
	}
	if err := s.db.WithContext(ctx).CreateInBatches(toCreate, 100).Error; err != nil {
		return 0, fmt.Errorf("批量创建角色失败: %w", err)
	}
	for _, c := range toCreate {
		charMap[strings.ToLower(c.Name)] = c.CharacterID
		idx.add(c)
	}
	return len(toCreate), nil
}

func (s *vsChapterSplitService) persistResolvedAliases(ctx context.Context, updates map[uint64]model.StringList, loginName string) error {
	for id, aliases := range updates {
		if err := s.db.WithContext(ctx).Model(&model.VsCharacter{}).
			Where("character_id = ?", id).
			Updates(map[string]interface{}{
				"aliases":    aliases,
				"updated_by": loginName,
			}).Error; err != nil {
			return fmt.Errorf("回写角色别名失败: %w", err)
		}
	}
	return nil
}
