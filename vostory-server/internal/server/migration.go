package server

import (
	"context"
	"encoding/json"
	"fmt"
	"iot-alert-center/internal/model"
	"iot-alert-center/pkg/log"
	"os"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MigrateServer struct {
	db  *gorm.DB
	log *log.Logger
}

func NewMigrateServer(db *gorm.DB, log *log.Logger) *MigrateServer {
	return &MigrateServer{
		db:  db,
		log: log,
	}
}
func (m *MigrateServer) Start(ctx context.Context) error {
	// 步骤 1: 数据库表结构迁移
	m.log.Info("步骤 1/3: 数据库表结构迁移...")
	if err := m.db.AutoMigrate(
		// 系统管理表
		&model.SysDept{},
		&model.SysMenu{},
		&model.SysPost{},
		&model.SysRole{},
		&model.SysRoleDept{},
		&model.SysRoleMenu{},
		&model.SysUser{},
		&model.SysUserPost{},
		&model.SysUserRole{},
		&model.SysLogininfor{},
		&model.SysApi{},
		&model.SysDictType{},
		&model.SysDictData{},
		&model.SysOperLog{},
		&model.SysConfig{},
		// VoStory 业务表
		&model.VsWorkspace{},
		&model.VsProject{},
		&model.VsChapter{},
		&model.VsScene{},
		&model.VsScriptSegment{},
		&model.VsCharacter{},
		&model.VsVoiceAsset{},
		&model.VsVoiceProfile{},
		&model.VsVoiceDesignItem{},
		&model.VsPronunciationDict{},
		&model.VsGenerationTask{},
		&model.VsTaskBatch{},
		&model.VsAudioClip{},
		&model.VsExportJob{},
		&model.VsLLMProvider{},
		&model.VsLLMModel{},
		&model.VsTTSProvider{},
		&model.VsPromptTemplate{},
		&model.VsLLMLog{},
	); err != nil {
		m.log.Error("数据库表结构迁移失败", zap.Error(err))
		return err
	}
	m.log.Info("数据库表结构迁移成功")

	if err := m.backfillLLMVendorModel(); err != nil {
		m.log.Error("回填 LLM 厂家/模型失败", zap.Error(err))
		return err
	}
	m.dropLegacyLLMColumns()

	m.dropLegacyEmotion()
	m.stripEmotionFromPromptTemplates()

	// 步骤 2: 重置序列
	m.log.Info("步骤 2/3: 重置序列...")
	if err := m.resetSequences(); err != nil {
		m.log.Error("重置序列失败", zap.Error(err))
		return err
	}
	m.log.Info("重置序列成功")

	// 步骤 3: 初始化种子数据
	m.log.Info("步骤 3/3: 初始化种子数据...")
	if err := m.seedData(); err != nil {
		m.log.Error("初始化种子数据失败", zap.Error(err))
		return err
	}
	m.log.Info("初始化种子数据成功")

	m.log.Info("数据库迁移和初始化完成")
	os.Exit(0)
	return nil
}
func (m *MigrateServer) resetSequences() error {
	seqMap := map[string]string{
		"vs_workspace":          "workspace_id",
		"vs_project":            "project_id",
		"vs_chapter":            "chapter_id",
		"vs_scene":              "scene_id",
		"vs_script_segment":     "segment_id",
		"vs_character":          "character_id",
		"vs_voice_design_item":  "item_id",
		"vs_voice_asset":        "voice_asset_id",
		"vs_voice_profile":      "voice_profile_id",
		"vs_pronunciation_dict": "dict_id",
		"vs_generation_task":    "task_id",
		"vs_task_batch":         "batch_id",
		"vs_audio_clip":         "clip_id",
		"vs_export_job":         "export_job_id",
		"vs_llm_provider":       "provider_id",
		"vs_llm_model":          "model_id",
		"vs_tts_provider":       "provider_id",
		"vs_prompt_template":    "template_id",
		"vs_llm_log":            "log_id",
	}

	for table, col := range seqMap {
		sql := fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 1))`,
			table, col, col, table,
		)
		if err := m.db.Exec(sql).Error; err != nil {
			m.log.Warn("reset sequence skipped", zap.String("table", table), zap.Error(err))
		}
	}
	return nil
}

// hasColumn 检查表是否仍有指定列（用于兼容旧结构回填）
func (m *MigrateServer) hasColumn(table, column string) bool {
	var count int64
	err := m.db.Raw(
		`SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?`,
		table, column,
	).Scan(&count).Error
	return err == nil && count > 0
}

// backfillLLMVendorModel 把旧厂家上的 default_model / model_list 拆成模型，并回填项目绑定
func (m *MigrateServer) backfillLLMVendorModel() error {
	var modelCount int64
	if err := m.db.Model(&model.VsLLMModel{}).Count(&modelCount).Error; err != nil {
		return err
	}

	if modelCount == 0 && m.hasColumn("vs_llm_provider", "default_model") {
		type oldProviderRow struct {
			ProviderID   uint64
			DefaultModel string
			ModelList    string
			CreatedBy    string
		}
		var rows []oldProviderRow
		if err := m.db.Raw(
			`SELECT provider_id,
				COALESCE(default_model, '') AS default_model,
				COALESCE(model_list, '[]') AS model_list,
				COALESCE(created_by, '') AS created_by
			 FROM vs_llm_provider
			 WHERE deleted_at IS NULL`,
		).Scan(&rows).Error; err != nil {
			return fmt.Errorf("读取旧厂家模型字段失败: %w", err)
		}

		for _, row := range rows {
			names := collectLegacyModelNames(row.DefaultModel, row.ModelList)
			for i, name := range names {
				mdl := &model.VsLLMModel{
					ProviderID: row.ProviderID,
					ModelName:  name,
					SortOrder:  i,
					Status:     "0",
					BaseModel:  model.BaseModel{CreatedBy: row.CreatedBy},
				}
				if err := m.db.Create(mdl).Error; err != nil {
					return fmt.Errorf("回填模型[%s]失败: %w", name, err)
				}
			}
			if len(names) > 0 {
				m.log.Info("回填厂家模型", zap.Uint64("provider_id", row.ProviderID), zap.Int("count", len(names)))
			}
		}
	}

	if !m.hasColumn("vs_project", "llm_provider_id") {
		return nil
	}

	if err := m.db.Exec(`
		UPDATE vs_project p
		SET llm_model_id = (
			SELECT m.model_id FROM vs_llm_model m
			WHERE m.provider_id = p.llm_provider_id AND m.deleted_at IS NULL
			ORDER BY m.model_id ASC
			LIMIT 1
		)
		WHERE p.llm_provider_id IS NOT NULL
		  AND (p.llm_model_id IS NULL OR p.llm_model_id = 0)
		  AND p.deleted_at IS NULL
	`).Error; err != nil {
		return fmt.Errorf("回填项目 llm_model_id 失败: %w", err)
	}
	m.log.Info("已按旧厂家回填项目绑定的 LLM 模型")
	return nil
}

func collectLegacyModelNames(defaultModel, modelListJSON string) []string {
	seen := map[string]struct{}{}
	var names []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	add(defaultModel)
	var listed []string
	if strings.TrimSpace(modelListJSON) != "" {
		_ = json.Unmarshal([]byte(modelListJSON), &listed)
	}
	for _, name := range listed {
		add(name)
	}
	return names
}

// dropLegacyLLMColumns 删除厂家/项目上已废弃的旧列
func (m *MigrateServer) dropLegacyLLMColumns() {
	legacyProviderCols := []string{"model_list", "default_model", "custom_params", "provider_type"}
	for _, col := range legacyProviderCols {
		if !m.hasColumn("vs_llm_provider", col) {
			continue
		}
		if err := m.db.Exec(fmt.Sprintf(`ALTER TABLE vs_llm_provider DROP COLUMN IF EXISTS %s`, col)).Error; err != nil {
			m.log.Warn("删除旧列失败", zap.String("table", "vs_llm_provider"), zap.String("column", col), zap.Error(err))
		} else {
			m.log.Info("已删除旧列", zap.String("table", "vs_llm_provider"), zap.String("column", col))
		}
	}
	if m.hasColumn("vs_project", "llm_provider_id") {
		if err := m.db.Exec(`ALTER TABLE vs_project DROP COLUMN IF EXISTS llm_provider_id`).Error; err != nil {
			m.log.Warn("删除旧列 vs_project.llm_provider_id 失败", zap.Error(err))
		} else {
			m.log.Info("已删除旧列 vs_project.llm_provider_id")
		}
	}
}

// dropLegacyEmotion 删除情绪参考音频表，以及片段/音频版本上的情绪类型、强度列。
// 用 Migrator 判断列（兼容 SQLite / Postgres），删列用原生 DROP COLUMN：SQLite 的 Migrator().DropColumn 会重建表并丢掉索引。
func (m *MigrateServer) dropLegacyEmotion() {
	if m.db.Migrator().HasTable("vs_voice_emotion") {
		if err := m.db.Migrator().DropTable("vs_voice_emotion"); err != nil {
			m.log.Warn("删除旧表 vs_voice_emotion 失败", zap.Error(err))
		} else {
			m.log.Info("已删除旧表 vs_voice_emotion")
		}
	}
	tables := map[string]interface{}{
		"vs_script_segment": &model.VsScriptSegment{},
		"vs_audio_clip":     &model.VsAudioClip{},
	}
	for table, mdl := range tables {
		for _, col := range []string{"emotion_type", "emotion_strength"} {
			if !m.db.Migrator().HasColumn(mdl, col) {
				continue
			}
			if err := m.db.Exec(fmt.Sprintf(`ALTER TABLE %s DROP COLUMN %s`, table, col)).Error; err != nil {
				m.log.Warn("删除旧列失败", zap.String("table", table), zap.String("column", col), zap.Error(err))
			} else {
				m.log.Info("已删除旧列", zap.String("table", table), zap.String("column", col))
			}
		}
	}
}

// legacyEmotionPromptFragments 旧版切分模板中要求输出情绪的片段 → 替换内容；自定义过的模板只要不含这些原文就不受影响
var legacyEmotionPromptFragments = []struct{ old, new string }{
	{"，不要包含当前情绪状态，情绪由emotion字段单独表达", "，不要包含当前情绪状态"},
	{"7. 标注每个片段的情绪：neutral/happy/sad/angry/fear/surprise/disgust\n8. 标注情绪强度：light/medium/strong\n9. 标注演绎指令", "7. 标注演绎指令"},
	{"\n10. content字段中的特殊字符", "\n8. content字段中的特殊字符"},
	{`,"emotion":"neutral|happy|sad|angry|fear|surprise|disgust","emotion_strength":"light|medium|strong"`, ""},
}

// stripEmotionFromPromptTemplates 去掉切分模板里的情绪输出要求，并删除不再使用的 emotion_tag 模板
func (m *MigrateServer) stripEmotionFromPromptTemplates() {
	var templates []model.VsPromptTemplate
	if err := m.db.Where("template_type = ?", "dialogue_parse").Find(&templates).Error; err != nil {
		m.log.Warn("读取切分模板失败", zap.Error(err))
	}
	for _, t := range templates {
		content := t.Content
		for _, f := range legacyEmotionPromptFragments {
			content = strings.ReplaceAll(content, f.old, f.new)
		}
		description := strings.ReplaceAll(t.Description, "识别类型、说话人和情绪", "识别类型、说话人和演绎指令")
		if content == t.Content && description == t.Description {
			continue
		}
		if err := m.db.Model(&model.VsPromptTemplate{}).Where("template_id = ?", t.TemplateID).
			Updates(map[string]interface{}{"content": content, "description": description}).Error; err != nil {
			m.log.Warn("更新切分模板失败", zap.Uint64("template_id", t.TemplateID), zap.Error(err))
		} else {
			m.log.Info("已移除切分模板中的情绪要求", zap.Uint64("template_id", t.TemplateID), zap.String("name", t.Name))
		}
	}

	res := m.db.Where("template_type = ?", "emotion_tag").Delete(&model.VsPromptTemplate{})
	if res.Error != nil {
		m.log.Warn("删除 emotion_tag 模板失败", zap.Error(res.Error))
	} else if res.RowsAffected > 0 {
		m.log.Info("已删除 emotion_tag 模板", zap.Int64("count", res.RowsAffected))
	}
}

func (m *MigrateServer) seedData() error {
	for _, s := range model.DefaultPromptTemplateSeeds {
		var count int64
		m.db.Model(&model.VsPromptTemplate{}).Where("template_type = ?", s.TemplateType).Count(&count)
		if count > 0 {
			continue
		}
		tpl := &model.VsPromptTemplate{
			Name:         s.Name,
			TemplateType: s.TemplateType,
			Content:      s.Content,
			Description:  s.Description,
			IsSystem:     "1",
			Version:      1,
			SortOrder:    0,
			Status:       "0",
			BaseModel:    model.BaseModel{CreatedBy: "system"},
		}
		if err := m.db.Create(tpl).Error; err != nil {
			return fmt.Errorf("创建默认模板[%s]失败: %w", s.Name, err)
		}
		m.log.Info("创建默认提示词模板", zap.String("name", s.Name), zap.String("type", s.TemplateType))
	}
	return nil
}

func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info("AutoMigrate stop")
	return nil
}
