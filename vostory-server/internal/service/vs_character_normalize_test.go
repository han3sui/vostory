package service

import (
	"testing"

	"iot-alert-center/internal/model"
)

func TestSplitParenName(t *testing.T) {
	cases := []struct {
		in           string
		outer, inner string
		ok           bool
	}{
		{"许七安（三号）", "许七安", "三号", true},
		{"三号(许七安)", "三号", "许七安", true},
		{"周公子（周立）", "周公子", "周立", true},
		{"女子木偶（许七安操纵）", "女子木偶", "许七安操纵", true},
		{"许七安和许新年", "", "", false},
		{"许七安", "", "", false},
	}
	for _, c := range cases {
		outer, inner, ok := splitParenName(c.in)
		if ok != c.ok || outer != c.outer || inner != c.inner {
			t.Fatalf("%s => (%s,%s,%v), want (%s,%s,%v)", c.in, outer, inner, ok, c.outer, c.inner, c.ok)
		}
	}
}

func TestResolveSpeaker(t *testing.T) {
	xu := &model.VsCharacter{CharacterID: 1, Name: "许七安", Level: "main", Gender: "male"}
	pingzhi := &model.VsCharacter{CharacterID: 16, Name: "许平志", Level: "supporting", Gender: "male"}
	ershu := &model.VsCharacter{CharacterID: 110, Name: "许二叔", Level: "minor", Gender: "male"}
	zhouli := &model.VsCharacter{CharacterID: 105, Name: "周立", Level: "minor", Gender: "male"}
	sanhao := &model.VsCharacter{CharacterID: 261, Name: "三号", Level: "minor", Gender: "male"}
	shushi := &model.VsCharacter{CharacterID: 396, Name: "白衣术士", Level: "minor", Gender: "male"}
	idx := newNameIndex([]*model.VsCharacter{xu, pingzhi, ershu, zhouli, sanhao, shushi})

	cases := []struct {
		name     string
		gender   string
		except   uint64
		wantID   uint64
		wantPool string
		wantRule string
	}{
		{"许七安（三号）", "male", 286, 1, "", "paren_alias"},
		{"三号（许七安）", "male", 264, 1, "", "paren_alias"},
		{"许二叔（许平志）", "male", 84, 16, "", "paren_alias"},
		{"周公子（周立）", "male", 104, 105, "", "paren_alias"},
		{"女子木偶（许七安操纵）", "female", 651, 0, "", ""},
		{"许七安和许新年", "male", 123, 0, "", ""},
		{"许七安父母", "unknown", 1890, 0, "", ""},
		{"不具名白衣术士", "male", 111, 396, "", "unnamed_to_named"},
		{"不具名士卒甲", "male", 2679, 0, "路人-青年男", "unnamed_pool"},
		{"学子甲", "male", 77, 0, "路人-青年男", "numbered_pool"},
		{"众人", "unknown", 559, 0, "路人-群体", "generic_pool"},
		{"店小二", "male", 260, 0, "路人-青年男", "generic_pool"},
		{"魏公", "male", 115, 0, "", ""},
	}
	for _, c := range cases {
		got := resolveSpeaker(idx, c.name, c.gender, c.except)
		var gotID uint64
		if got.Existing != nil {
			gotID = got.Existing.CharacterID
		}
		if gotID != c.wantID || got.PoolName != c.wantPool || got.Rule != c.wantRule {
			t.Fatalf("%s => id=%d pool=%s rule=%s, want id=%d pool=%s rule=%s",
				c.name, gotID, got.PoolName, got.Rule, c.wantID, c.wantPool, c.wantRule)
		}
	}
}

func TestSuggestMergeGroups(t *testing.T) {
	chars := []*model.VsCharacter{
		{CharacterID: 1, Name: "许七安", Level: "main", Gender: "male"},
		{CharacterID: 264, Name: "三号（许七安）", Level: "minor", Gender: "male"},
		{CharacterID: 286, Name: "许七安（三号）", Level: "minor", Gender: "male"},
		{CharacterID: 260, Name: "店小二", Level: "minor", Gender: "male"},
		{CharacterID: 77, Name: "学子甲", Level: "minor", Gender: "male"},
		{CharacterID: 123, Name: "许七安和许新年", Level: "minor", Gender: "male"},
	}
	groups := suggestMergeGroups(chars)
	if len(groups) != 2 {
		t.Fatalf("groups=%d, want 2: %+v", len(groups), groups)
	}
	var toXu, toPool int
	for _, g := range groups {
		if g.TargetKind == "character" && g.TargetID == 1 {
			toXu = len(g.SourceIDs)
		}
		if g.TargetKind == "pool" {
			toPool += len(g.SourceIDs)
		}
	}
	if toXu != 2 {
		t.Fatalf("merge to 许七安 got %d, want 2", toXu)
	}
	if toPool != 2 {
		t.Fatalf("merge to pool got %d, want 2", toPool)
	}
}
