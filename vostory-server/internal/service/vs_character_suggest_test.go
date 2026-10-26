package service

import (
	"context"
	"testing"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newSuggestTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// 内存库每个连接各自独立，限制为单连接保证事务内外看到同一份数据
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.VsCharacter{}, &model.VsScriptSegment{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func newSuggestTestService(db *gorm.DB) *vsCharacterService {
	repo := repository.NewVsCharacterRepository(repository.NewRepository(nil, db, nil))
	return &vsCharacterService{db: db, repo: repo}
}

func createTestCharacter(t *testing.T, db *gorm.DB, name, level string) *model.VsCharacter {
	t.Helper()
	c := &model.VsCharacter{ProjectID: 1, Name: name, Gender: "male", Level: level, Status: "0"}
	if err := db.Create(c).Error; err != nil {
		t.Fatal(err)
	}
	return c
}

func loadTestCharacter(t *testing.T, db *gorm.DB, id uint64) model.VsCharacter {
	t.Helper()
	var c model.VsCharacter
	if err := db.First(&c, id).Error; err != nil {
		t.Fatal(err)
	}
	return c
}

func TestSuggestMergeGroupsResolvesChain(t *testing.T) {
	chars := []*model.VsCharacter{
		{CharacterID: 1, Name: "长老", Level: "minor", Gender: "male"},
		{CharacterID: 2, Name: "不具名长老", Level: "minor", Gender: "male"},
	}
	groups := suggestMergeGroups(chars)
	if len(groups) != 1 {
		t.Fatalf("groups=%d, want 1: %+v", len(groups), groups)
	}
	g := groups[0]
	if g.TargetKind != "pool" || g.TargetName != "路人-老年男" || len(g.SourceIDs) != 2 {
		t.Fatalf("group=%+v, want both merged into 路人-老年男", g)
	}
}

// 前端可能提交旧预览里的链式分组：目标在同一批次里先被合并掉，应顺着 merged_into_id 找到最终目标
func TestApplySuggestGroupsFollowsMergedTarget(t *testing.T) {
	for _, reversed := range []bool{false, true} {
		db := newSuggestTestDB(t)
		svc := newSuggestTestService(db)
		elder := createTestCharacter(t, db, "长老", "minor")
		unnamed := createTestCharacter(t, db, "不具名长老", "minor")

		toElder := v1.VsCharacterMergeSuggestGroup{
			Rule: "unnamed_to_named", TargetKind: "character", TargetID: elder.CharacterID, TargetName: elder.Name,
			SourceIDs: []uint64{unnamed.CharacterID},
		}
		toPool := v1.VsCharacterMergeSuggestGroup{
			Rule: "generic_pool", TargetKind: "pool", TargetName: "路人-老年男",
			SourceIDs: []uint64{elder.CharacterID},
		}
		groups := []v1.VsCharacterMergeSuggestGroup{toElder, toPool}
		if reversed {
			groups = []v1.VsCharacterMergeSuggestGroup{toPool, toElder}
		}

		resp, err := svc.SuggestMerge(context.Background(), 1, &v1.VsCharacterMergeSuggestRequest{Groups: groups})
		if err != nil {
			t.Fatalf("reversed=%v: %v", reversed, err)
		}
		if resp.MergedCount != 2 {
			t.Fatalf("reversed=%v: merged=%d, want 2", reversed, resp.MergedCount)
		}

		var pool model.VsCharacter
		if err := db.Where("name = ?", "路人-老年男").First(&pool).Error; err != nil {
			t.Fatal(err)
		}
		for _, id := range []uint64{elder.CharacterID, unnamed.CharacterID} {
			c := loadTestCharacter(t, db, id)
			if c.MergedIntoID == nil || *c.MergedIntoID != pool.CharacterID {
				t.Fatalf("reversed=%v: %s merged_into=%v, want %d", reversed, c.Name, c.MergedIntoID, pool.CharacterID)
			}
		}
	}
}

func TestSuggestMergeRequiresGroups(t *testing.T) {
	db := newSuggestTestDB(t)
	svc := newSuggestTestService(db)
	createTestCharacter(t, db, "店小二", "minor")

	if _, err := svc.SuggestMerge(context.Background(), 1, &v1.VsCharacterMergeSuggestRequest{DryRun: false}); err == nil {
		t.Fatal("expected error when confirming without groups")
	}
	var merged int64
	db.Model(&model.VsCharacter{}).Where("merged_into_id IS NOT NULL").Count(&merged)
	if merged != 0 {
		t.Fatalf("merged=%d, want 0", merged)
	}

	resp, err := svc.SuggestMerge(context.Background(), 1, &v1.VsCharacterMergeSuggestRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Groups) != 1 || resp.MergedCount != 1 {
		t.Fatalf("dry run=%+v, want 1 group", resp)
	}
}

func TestEnsureCharactersUnnamedSameChapter(t *testing.T) {
	db := newSuggestTestDB(t)
	split := &vsChapterSplitService{db: db}
	charMap := map[string]uint64{}
	result := &llmSplitResult{Scenes: []llmScene{{Segments: []llmSegment{
		{Type: "dialogue", Character: "不具名白衣术士", CharacterGender: "male"},
		{Type: "dialogue", Character: "白衣术士", CharacterGender: "male"},
		{Type: "dialogue", Character: "不具名士卒", CharacterGender: "male"},
	}}}}
	if _, err := split.ensureCharacters(context.Background(), result, 1, nil, charMap, "test", 0); err != nil {
		t.Fatal(err)
	}

	named := resolveCharacterID(charMap, "白衣术士")
	unnamed := resolveCharacterID(charMap, "不具名白衣术士")
	if named == nil || unnamed == nil || *named != *unnamed {
		t.Fatalf("不具名白衣术士=%v, 白衣术士=%v, want same character", unnamed, named)
	}
	c := loadTestCharacter(t, db, *named)
	if len(c.Aliases) != 1 || c.Aliases[0] != "不具名白衣术士" {
		t.Fatalf("aliases=%v, want [不具名白衣术士]", c.Aliases)
	}

	// 本章没有对应本名的不具名角色仍归入路人池
	soldier := resolveCharacterID(charMap, "不具名士卒")
	pool := resolveCharacterID(charMap, "路人-青年男")
	if soldier == nil || pool == nil || *soldier != *pool {
		t.Fatalf("不具名士卒=%v, 路人-青年男=%v, want pool", soldier, pool)
	}
	var count int64
	db.Model(&model.VsCharacter{}).Where("name = ?", "不具名白衣术士").Count(&count)
	if count != 0 {
		t.Fatal("不具名白衣术士 should not be created as a separate character")
	}
}
