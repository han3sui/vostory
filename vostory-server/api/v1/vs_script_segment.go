package v1

import "time"

// VsScriptSegmentListQuery 脚本片段列表查询参数
type VsScriptSegmentListQuery struct {
	*BasePageQuery
	ChapterID   uint64 `json:"chapter_id"`   // 所属章节
	SceneID     uint64 `json:"scene_id"`     // 所属场景
	SegmentType string `json:"segment_type"` // 片段类型
	CharacterID uint64 `json:"character_id"` // 说话人
	Status      string `json:"status"`       // 片段状态
}

// VsScriptSegmentCreateRequest 创建脚本片段请求
type VsScriptSegmentCreateRequest struct {
	SceneID         uint64  `json:"scene_id" binding:"required"`       // 所属场景
	ChapterID       uint64  `json:"chapter_id" binding:"required"`     // 所属章节
	SegmentNum      int     `json:"segment_num" binding:"required"`    // 片段序号
	SegmentType     string  `json:"segment_type" binding:"required"`   // 片段类型
	Content         string  `json:"content" binding:"required"`        // 片段文本
	OriginalContent string  `json:"original_content"`                  // 原始文本
	CharacterID     *uint64 `json:"character_id"`                      // 说话人角色ID
	EmotionType     string  `json:"emotion_type"`                      // 情绪类型
	EmotionStrength string  `json:"emotion_strength"`                  // 情绪强度
}

// VsScriptSegmentUpdateRequest 更新脚本片段请求
type VsScriptSegmentUpdateRequest struct {
	ID              uint64  `json:"id" binding:"required"`
	SegmentNum      int     `json:"segment_num"`       // 片段序号
	SegmentType     string  `json:"segment_type"`      // 片段类型
	Content         string  `json:"content"`           // 片段文本
	CharacterID     *uint64 `json:"character_id"`      // 说话人角色ID
	EmotionType     string  `json:"emotion_type"`      // 情绪类型
	EmotionStrength string  `json:"emotion_strength"`  // 情绪强度
	Status          string  `json:"status"`            // 片段状态
}

// VsScriptSegmentDetailResponse 脚本片段详情响应
type VsScriptSegmentDetailResponse struct {
	ID              uint64    `json:"id"`
	SceneID         uint64    `json:"scene_id"`
	ChapterID       uint64    `json:"chapter_id"`
	SegmentNum      int       `json:"segment_num"`
	SegmentType     string    `json:"segment_type"`
	Content         string    `json:"content"`
	OriginalContent string    `json:"original_content"`
	CharacterID     *uint64   `json:"character_id"`
	CharacterName   string    `json:"character_name"`
	EmotionType     string    `json:"emotion_type"`
	EmotionStrength string    `json:"emotion_strength"`
	Status          string    `json:"status"`
	Version         int       `json:"version"`
	HasAudio        bool      `json:"has_audio"`
	AudioURL        string    `json:"audio_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
