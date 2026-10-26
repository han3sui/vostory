package service

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"iot-alert-center/internal/model"
)

const unnamedPrefix = "不具名"

var (
	parenPairs = [][2]string{{"（", "）"}, {"(", ")"}}
	// 甲乙丙及少量拉丁/数字后缀，用于「学子甲」「不具名长老A」
	numberedSuffixes = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	codeLikeNameRe   = regexp.MustCompile(`^[一二三四五六七八九十零〇0-9]+号$`)
)

// speakerResolve 将切分/扫描得到的角色名归一到已有角色或路人池
type speakerResolve struct {
	Existing   *model.VsCharacter
	PoolName   string
	Rule       string
	Reason     string
	AddAliases []string
}

type nameIndex struct {
	byKey map[string]*model.VsCharacter
}

func newNameIndex(characters []*model.VsCharacter) *nameIndex {
	idx := &nameIndex{byKey: make(map[string]*model.VsCharacter, len(characters)*2)}
	for _, c := range characters {
		idx.add(c)
	}
	return idx
}

func (idx *nameIndex) add(c *model.VsCharacter) {
	if c == nil {
		return
	}
	idx.put(c.Name, c)
	for _, alias := range c.Aliases {
		idx.put(alias, c)
	}
}

func (idx *nameIndex) put(name string, c *model.VsCharacter) {
	key := normalizeLookupKey(name)
	if key == "" || c == nil {
		return
	}
	if old, ok := idx.byKey[key]; ok {
		idx.byKey[key] = pickBetterCharacter(old, c)
		return
	}
	idx.byKey[key] = c
}

func (idx *nameIndex) lookupExcept(name string, exceptID uint64) *model.VsCharacter {
	c := idx.byKey[normalizeLookupKey(name)]
	if c == nil || c.CharacterID == exceptID {
		return nil
	}
	return c
}

func normalizeLookupKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// splitParenName 拆「许七安（三号）」「三号(许七安)」为外层与括号内文本
func splitParenName(name string) (outer, inner string, ok bool) {
	name = strings.TrimSpace(name)
	for _, pair := range parenPairs {
		left, right := pair[0], pair[1]
		i := strings.Index(name, left)
		if i <= 0 || !strings.HasSuffix(name, right) {
			continue
		}
		inner = strings.TrimSpace(name[i+len(left) : len(name)-len(right)])
		outer = strings.TrimSpace(name[:i])
		if outer == "" || inner == "" || strings.Contains(inner, left) {
			continue
		}
		return outer, inner, true
	}
	return "", "", false
}

func stripUnnamedPrefix(name string) string {
	name = strings.TrimSpace(name)
	if strings.HasPrefix(name, unnamedPrefix) {
		if rest := strings.TrimSpace(strings.TrimPrefix(name, unnamedPrefix)); rest != "" {
			return rest
		}
	}
	return name
}

func stripNumberedSuffix(name string) string {
	name = strings.TrimSpace(name)
	runes := []rune(name)
	if len(runes) < 2 {
		return name
	}
	last := string(runes[len(runes)-1])
	for _, suffix := range numberedSuffixes {
		if last == suffix {
			return strings.TrimSpace(string(runes[:len(runes)-1]))
		}
	}
	r := runes[len(runes)-1]
	if (r >= 'A' && r <= 'E') || (r >= 'a' && r <= 'e') || (r >= '1' && r <= '9') {
		return strings.TrimSpace(string(runes[:len(runes)-1]))
	}
	return name
}

func isCodeLikeName(name string) bool {
	name = strings.TrimSpace(name)
	return codeLikeNameRe.MatchString(name)
}

func pickBetterCharacter(a, b *model.VsCharacter) *model.VsCharacter {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if characterLevelRank[a.Level] != characterLevelRank[b.Level] {
		if characterLevelRank[a.Level] > characterLevelRank[b.Level] {
			return a
		}
		return b
	}
	if isCodeLikeName(a.Name) != isCodeLikeName(b.Name) {
		if isCodeLikeName(a.Name) {
			return b
		}
		return a
	}
	if utf8.RuneCountInString(a.Name) != utf8.RuneCountInString(b.Name) {
		if utf8.RuneCountInString(a.Name) > utf8.RuneCountInString(b.Name) {
			return a
		}
		return b
	}
	if a.CharacterID > 0 && b.CharacterID > 0 && a.CharacterID != b.CharacterID {
		if a.CharacterID < b.CharacterID {
			return a
		}
		return b
	}
	return a
}

func canBeMergeSource(c *model.VsCharacter) bool {
	if c == nil || c.Name == narratorName || isPoolCharacterName(c.Name) {
		return false
	}
	if c.Level == "main" {
		return false
	}
	if c.MergedIntoID != nil {
		return false
	}
	return true
}

// resolveSpeaker 按规则把角色名归一到已有角色或路人池。未命中时 Existing/PoolName 都为空。
func resolveSpeaker(idx *nameIndex, raw, gender string, exceptID uint64) speakerResolve {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == narratorName || isPoolCharacterName(raw) {
		return speakerResolve{}
	}

	if hit := idx.lookupExcept(raw, exceptID); hit != nil {
		return speakerResolve{Existing: hit, Rule: "exact", Reason: "名称或别名完全相同"}
	}

	if outer, inner, ok := splitParenName(raw); ok {
		left := idx.lookupExcept(outer, exceptID)
		right := idx.lookupExcept(inner, exceptID)
		if target := pickBetterCharacter(left, right); target != nil {
			return speakerResolve{
				Existing:   target,
				Rule:       "paren_alias",
				Reason:     "括号内外与已有角色同名",
				AddAliases: uniqueAliasAdds(target, raw, outer, inner),
			}
		}
	}

	stripped := stripUnnamedPrefix(raw)
	if stripped != raw {
		if hit := idx.lookupExcept(stripped, exceptID); hit != nil && !isPoolCharacterName(hit.Name) {
			return speakerResolve{
				Existing:   hit,
				Rule:       "unnamed_to_named",
				Reason:     "去掉「不具名」后对应已有角色",
				AddAliases: uniqueAliasAdds(hit, raw),
			}
		}
		stem := stripNumberedSuffix(stripped)
		if pool, ok := resolveGenericPoolName(stem, gender); ok {
			return speakerResolve{PoolName: pool, Rule: "unnamed_pool", Reason: "不具名泛称，归入路人池"}
		}
		if pool, ok := resolveGenericPoolName(stripped, gender); ok {
			return speakerResolve{PoolName: pool, Rule: "unnamed_pool", Reason: "不具名泛称，归入路人池"}
		}
		return speakerResolve{
			PoolName: poolNameByCategory("young", normalizeGender(gender)),
			Rule:     "unnamed_pool",
			Reason:   "不具名路人，归入路人池",
		}
	}

	stem := stripNumberedSuffix(raw)
	if stem != raw {
		if pool, ok := resolveGenericPoolName(stem, gender); ok {
			return speakerResolve{PoolName: pool, Rule: "numbered_pool", Reason: "带甲乙丙等编号的泛称，归入路人池"}
		}
	}

	if pool, ok := resolveGenericPoolName(raw, gender); ok {
		return speakerResolve{PoolName: pool, Rule: "generic_pool", Reason: "泛称角色，归入路人池"}
	}
	return speakerResolve{}
}

func uniqueAliasAdds(target *model.VsCharacter, names ...string) []string {
	if target == nil {
		return nil
	}
	seen := map[string]bool{normalizeLookupKey(target.Name): true}
	for _, a := range target.Aliases {
		seen[normalizeLookupKey(a)] = true
	}
	var result []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		key := normalizeLookupKey(name)
		if name == "" || seen[key] || isPoolCharacterName(name) || isCodeLikeName(name) {
			continue
		}
		// 括号拆出的外层/内层若本身是已有角色名，不必再当别名
		seen[key] = true
		result = append(result, name)
	}
	return result
}

func appendUniqueAliases(current model.StringList, adds []string, selfName string) model.StringList {
	seen := map[string]bool{normalizeLookupKey(selfName): true}
	result := make(model.StringList, 0, len(current)+len(adds))
	for _, a := range current {
		key := normalizeLookupKey(a)
		if a == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, a)
	}
	for _, a := range adds {
		key := normalizeLookupKey(a)
		if a == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, a)
	}
	return result
}
