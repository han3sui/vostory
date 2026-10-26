package model

// VsVoiceDesignItem 批量音色设计任务项：每个角色一条，由后台 worker 逐条处理。
// 需要试听确认的角色生成草稿（draft），确认后才保存为声音配置；其余角色直接保存并绑定。
type VsVoiceDesignItem struct {
	BaseModel
	ItemID         uint64  `json:"item_id" gorm:"primaryKey;autoIncrement;comment:任务项ID"`
	ProjectID      uint64  `json:"project_id" gorm:"not null;index:idx_vdi_project_batch;comment:所属项目"`
	BatchID        uint64  `json:"batch_id" gorm:"not null;index:idx_vdi_project_batch;comment:批次号（创建时的毫秒时间戳）"`
	CharacterID    uint64  `json:"character_id" gorm:"not null;index;comment:角色ID"`
	NeedConfirm    string  `json:"need_confirm" gorm:"size:1;default:'0';comment:是否需要试听确认（1是 0否）"`
	Status         string  `json:"status" gorm:"size:20;not null;index;comment:状态（queued/running/draft/applied/skipped/failed/discarded/cancelled）"`
	Instruction    string  `json:"instruction" gorm:"size:500;comment:音色描述"`
	SampleText     string  `json:"sample_text" gorm:"size:500;comment:参考台词"`
	Seed           int64   `json:"seed" gorm:"comment:随机种子"`
	PreviewPath    string  `json:"preview_path" gorm:"size:500;comment:草稿试听音频路径"`
	VoiceProfileID *uint64 `json:"voice_profile_id" gorm:"comment:保存后的声音配置"`
	ErrorMessage   string  `json:"error_message" gorm:"size:1000;comment:失败原因"`
}

func (VsVoiceDesignItem) TableName() string {
	return "vs_voice_design_item"
}
