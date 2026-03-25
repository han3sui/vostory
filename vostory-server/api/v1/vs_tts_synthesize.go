package v1

// TTSSynthesizeResponse TTS 合成响应
type TTSSynthesizeResponse struct {
	ClipID   uint64 `json:"clip_id"`
	AudioURL string `json:"audio_url"`
	Duration int    `json:"duration"`
	Version  int    `json:"version"`
}
