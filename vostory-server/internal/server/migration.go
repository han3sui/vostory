package server

import (
	"context"
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
		m.log.Error("user migrate error", zap.Error(err))
		return err
	}
	m.log.Info("AutoMigrate success")
	os.Exit(0)
	return nil
}
func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info("AutoMigrate stop")
	return nil
}
