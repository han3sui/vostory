package server

import (
	"context"
	"fmt"
	"iot-alert-center/internal/model"
	"iot-alert-center/pkg/log"
	"os"

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
		// VoStory 业务表
		&model.VsWorkspace{},
		&model.VsProject{},
		&model.VsChapter{},
		&model.VsScene{},
		&model.VsScriptSegment{},
		&model.VsCharacter{},
		&model.VsVoiceAsset{},
		&model.VsVoiceProfile{},
		&model.VsVoiceEmotion{},
		&model.VsPronunciationDict{},
		&model.VsGenerationTask{},
		&model.VsTaskBatch{},
		&model.VsAudioClip{},
		&model.VsExportJob{},
		&model.VsLLMProvider{},
		&model.VsTTSProvider{},
		&model.VsPromptTemplate{},
		&model.VsLLMLog{},
	); err != nil {
		m.log.Error("数据库表结构迁移失败", zap.Error(err))
		return err
	}
	m.log.Info("数据库表结构迁移成功")

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
		"vs_workspaces":          "workspace_id",
		"vs_projects":            "project_id",
		"vs_chapters":            "chapter_id",
		"vs_scenes":              "scene_id",
		"vs_script_segments":     "segment_id",
		"vs_characters":          "character_id",
		"vs_voice_assets":        "voice_asset_id",
		"vs_voice_profiles":      "voice_profile_id",
		"vs_voice_emotions":      "voice_emotion_id",
		"vs_pronunciation_dicts": "dict_id",
		"vs_generation_tasks":    "task_id",
		"vs_task_batches":        "batch_id",
		"vs_audio_clips":         "clip_id",
		"vs_export_jobs":         "export_job_id",
		"vs_llm_providers":       "provider_id",
		"vs_tts_providers":       "provider_id",
		"vs_prompt_templates":    "template_id",
		"vs_llm_logs":            "log_id",
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
