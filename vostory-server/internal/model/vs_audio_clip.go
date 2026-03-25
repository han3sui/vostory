package model

// VsAudioClip 音频片段
type VsAudioClip struct {
	BaseModel
	ClipID          uint64  `json:"clip_id" gorm:"primaryKey;comment:音频片段ID"`
	SegmentID       uint64  `json:"segment_id" gorm:"not null;index;comment:关联脚本片段"`
	TaskID          *uint64 `json:"task_id" gorm:"comment:生成该音频的任务ID"`
	AudioURL        string  `json:"audio_url" gorm:"size:500;not null;comment:音频文件路径"`
	Duration        int     `json:"duration" gorm:"default:0;comment:时长（毫秒）"`
	FileSize        int64   `json:"file_size" gorm:"default:0;comment:文件大小（字节）"`
	Format          string  `json:"format" gorm:"size:10;comment:音频格式（wav/mp3/flac）"`
	TTSProviderID   *uint64 `json:"tts_provider_id" gorm:"comment:使用的TTS提供商"`
	VoiceProfileID  *uint64 `json:"voice_profile_id" gorm:"comment:使用的声音配置"`
	EmotionType     string  `json:"emotion_type" gorm:"size:50;comment:生成时的情绪类型"`
	EmotionStrength string  `json:"emotion_strength" gorm:"size:20;comment:生成时的情绪强度"`
	Version         int     `json:"version" gorm:"default:1;comment:版本号"`
	IsCurrent       string  `json:"is_current" gorm:"size:1;default:'1';comment:是否当前版本（1是 0否）"`

	Segment      *VsScriptSegment `json:"segment" gorm:"foreignKey:SegmentID;references:SegmentID"`
	Task         *VsGenerationTask `json:"task" gorm:"foreignKey:TaskID;references:TaskID"`
	TTSProvider  *VsTTSProvider    `json:"tts_provider" gorm:"foreignKey:TTSProviderID;references:ProviderID"`
	VoiceProfile *VsVoiceProfile   `json:"voice_profile" gorm:"foreignKey:VoiceProfileID;references:VoiceProfileID"`
}

func (VsAudioClip) TableName() string {
	return "vs_audio_clip"
}
