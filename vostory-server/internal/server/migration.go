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
		"vs_workspaces":         "workspace_id",
		"vs_projects":           "project_id",
		"vs_chapters":           "chapter_id",
		"vs_scenes":             "scene_id",
		"vs_script_segments":    "segment_id",
		"vs_characters":         "character_id",
		"vs_voice_assets":       "voice_asset_id",
		"vs_voice_profiles":     "voice_profile_id",
		"vs_voice_emotions":     "voice_emotion_id",
		"vs_pronunciation_dicts": "dict_id",
		"vs_generation_tasks":   "task_id",
		"vs_task_batches":       "batch_id",
		"vs_audio_clips":        "clip_id",
		"vs_export_jobs":        "export_job_id",
		"vs_llm_providers":      "provider_id",
		"vs_tts_providers":      "provider_id",
		"vs_prompt_templates":   "template_id",
		"vs_llm_logs":           "log_id",
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
	seeds := []struct {
		Name         string
		TemplateType string
		Content      string
		Description  string
	}{
		{
			Name:         "默认角色抽取",
			TemplateType: "character_extract",
			Content:      "你是一个专业的小说文本分析助手。请从以下小说文本中抽取所有出现的角色。\n\n要求：\n1. 识别所有有名字的角色（包括只出现一次的）\n2. 不要把地名、物品名当作角色\n3. 同一个角色的不同称呼要合并为一个角色\n\n请严格以JSON格式返回，不要包含任何其他文字，结构如下：\n{\"characters\":[{\"name\":\"角色主要名称\",\"aliases\":[\"别名1\",\"称呼2\"],\"gender\":\"male|female|unknown\",\"level\":\"main|supporting|minor\",\"description\":\"一句话角色描述\"}]}\n\n---\n{{content}}",
			Description:  "从小说文本中自动抽取角色信息",
		},
		{
			Name:         "默认对白解析",
			TemplateType: "dialogue_parse",
			Content:      "你是一个专业的小说文本分析助手。请将以下章节文本进行结构化切分。\n\n要求：\n1. 识别场景切换（基于时间跳跃、地点变化、视角切换）\n2. 在每个场景内，将文本切分为独立片段\n3. 每个片段标注类型：dialogue(对白)、narration(旁白)、monologue(独白)、description(描述)\n4. 对白和独白片段需识别说话人名称\n5. 标注每个片段的情绪：neutral/happy/sad/angry/fear/surprise/disgust\n6. 标注情绪强度：light/medium/strong\n\n请严格以JSON格式返回，不要包含任何其他文字，结构如下：\n{\"scenes\":[{\"title\":\"场景标题\",\"description\":\"场景简述\",\"segments\":[{\"type\":\"dialogue|narration|monologue|description\",\"content\":\"片段文本内容\",\"character\":\"说话人名称（非对白/独白时为空字符串）\",\"emotion\":\"neutral|happy|sad|angry|fear|surprise|disgust\",\"emotion_strength\":\"light|medium|strong\"}]}]}\n\n---\n{{content}}",
			Description:  "将章节文本按场景和片段进行结构化切分，识别类型、说话人和情绪",
		},
		{
			Name:         "默认情绪标注",
			TemplateType: "emotion_tag",
			Content:      "请为以下对白/独白片段标注情绪。对于每个片段，请提供：\n1. 情绪类型：happy/sad/angry/fear/surprise/neutral/disgust/contempt\n2. 情绪强度：light/medium/strong\n\n请以JSON数组格式返回结果。\n\n---\n{{segments}}",
			Description:  "为脚本片段自动标注情绪类型和强度",
		},
		{
			Name:         "默认场景切分",
			TemplateType: "scene_split",
			Content:      "请将以下章节文本按场景进行切分。场景切换的依据包括：\n1. 时间跳跃\n2. 地点变化\n3. 视角切换\n4. 明显的叙事断裂\n\n对于每个场景，请提供：\n1. 场景标题（简要概括）\n2. 场景描述\n3. 场景包含的文本范围（起始和结束位置）\n\n请以JSON数组格式返回结果。\n\n---\n{{content}}",
			Description:  "将章节文本按场景自动切分",
		},
		{
			Name:         "默认文本校正",
			TemplateType: "text_correct",
			Content:      "请对以下文本进行校正，确保：\n1. 不丢失任何原文内容\n2. 不添加原文没有的内容\n3. 修正明显的错别字\n4. 统一标点符号格式\n\n请返回校正后的完整文本。\n\n---\n{{content}}",
			Description:  "精准填充 - 确保LLM输出对齐回原文",
		},
	}

	for _, s := range seeds {
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
