package model

// VsScriptSegment 脚本片段
type VsScriptSegment struct {
	BaseModel
	SegmentID       uint64  `json:"segment_id" gorm:"primaryKey;comment:片段ID"`
	SceneID         uint64  `json:"scene_id" gorm:"not null;index;comment:所属场景"`
	ChapterID       uint64  `json:"chapter_id" gorm:"not null;index;index:idx_chapter_segment_num;comment:所属章节（冗余，便于按章节查询）"`
	SegmentNum      int     `json:"segment_num" gorm:"not null;index:idx_chapter_segment_num;comment:片段序号"`
	SegmentType     string  `json:"segment_type" gorm:"size:20;not null;comment:片段类型（dialogue/narration/monologue/description）"`
	Content         string  `json:"content" gorm:"type:text;not null;comment:片段文本内容"`
	OriginalContent string  `json:"original_content" gorm:"type:text;comment:原始文本（精准填充对齐前）"`
	CharacterID     *uint64 `json:"character_id" gorm:"comment:说话人角色ID（旁白/描述时为空）"`
	EmotionType     string  `json:"emotion_type" gorm:"size:50;comment:情绪类型（happy/sad/angry/fear/surprise/neutral等）"`
	EmotionStrength string  `json:"emotion_strength" gorm:"size:20;default:'medium';comment:情绪强度（light/medium/strong）"`
	Status          string  `json:"status" gorm:"size:20;default:'raw';comment:片段状态（raw/edited/generated）"`
	Version         int     `json:"version" gorm:"default:1;comment:版本号"`

	Scene     *VsScene     `json:"scene" gorm:"foreignKey:SceneID;references:SceneID"`
	Chapter   *VsChapter   `json:"chapter" gorm:"foreignKey:ChapterID;references:ChapterID"`
	Character *VsCharacter `json:"character" gorm:"foreignKey:CharacterID;references:CharacterID"`
}

func (VsScriptSegment) TableName() string {
	return "vs_script_segment"
}
