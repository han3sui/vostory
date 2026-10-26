// Package tts 封装 Breeze-TTS-2 流式 HTTP API。
//
// Breeze-TTS-2 一个接口覆盖三种模式，由传入字段决定：
//   - Voice Clone：ref_audio + ref_text + text，按参考音频克隆音色
//   - Voice Direction：ref_audio + ref_text + text + instruction，保持参考音色，按指令调整语气/语速/情绪
//   - Voice Design：text + instruction，不传参考音频，按文字描述生成新音色
package tts

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Breeze-TTS-2 输出固定为 24kHz 单声道 16bit 小端 PCM
const (
	SampleRate    = 24000
	Channels      = 1
	BitsPerSample = 16
)

// DefaultCfgScale 官方建议使用 4 以增强指令遵循
const DefaultCfgScale = 4.0

// busyRetryTimeout 服务繁忙（409）时最长等待时间，足够等完一条长句或一次音色设计
const busyRetryTimeout = 3 * time.Minute

// busyRetryDelay 409 重试间隔：0.5s 起翻倍，封顶 5s
func busyRetryDelay(attempt int) time.Duration {
	return min(500*time.Millisecond<<min(attempt, 4), 5*time.Second)
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL string, apiKey ...string) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
	if len(apiKey) > 0 {
		c.apiKey = apiKey[0]
	}
	return c
}

func (c *Client) setAuth(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// SpeechRequest 合成请求。RefAudioPath 为空时即 Voice Design 模式，此时 Instruction 必填。
type SpeechRequest struct {
	Text         string
	RefAudioPath string
	RefText      string
	Instruction  string
	CfgScale     float64
	Seed         *int64
}

// Speech 调用 POST /v1/audio/speech，返回封装好 WAV 头的音频数据。
func (c *Client) Speech(sr SpeechRequest) ([]byte, error) {
	if strings.TrimSpace(sr.Text) == "" {
		return nil, fmt.Errorf("合成文本为空")
	}
	if sr.RefAudioPath == "" && strings.TrimSpace(sr.Instruction) == "" {
		return nil, fmt.Errorf("音色设计模式需要提供音色描述")
	}
	if sr.RefAudioPath != "" && strings.TrimSpace(sr.RefText) == "" {
		return nil, fmt.Errorf("参考音频缺少逐字转写文本（ref_text）")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if sr.RefAudioPath != "" {
		f, err := os.Open(sr.RefAudioPath)
		if err != nil {
			return nil, fmt.Errorf("open audio file %s: %w", sr.RefAudioPath, err)
		}
		part, err := writer.CreateFormFile("ref_audio", filepath.Base(sr.RefAudioPath))
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("create form file: %w", err)
		}
		_, err = io.Copy(part, f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("copy audio data: %w", err)
		}
	}

	fields := map[string]string{"text": sr.Text}
	if sr.RefAudioPath != "" {
		fields["ref_text"] = sr.RefText
	}
	if inst := strings.TrimSpace(sr.Instruction); inst != "" {
		fields["instruction"] = inst
		cfg := sr.CfgScale
		if cfg <= 0 {
			cfg = DefaultCfgScale
		}
		fields["cfg_scale"] = strconv.FormatFloat(cfg, 'f', -1, 64)
	}
	if sr.Seed != nil {
		fields["seed"] = strconv.FormatInt(*sr.Seed, 10)
	}
	for k, v := range fields {
		if v == "" {
			continue
		}
		if err := writer.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("write %s field: %w", k, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	body := buf.Bytes()
	contentType := writer.FormDataContentType()

	// Breeze 服务单并发，已有请求在跑时返回 409；等待后重试（如批量生成时又在试听音色设计）
	var (
		status int
		data   []byte
	)
	deadline := time.Now().Add(busyRetryTimeout)
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest("POST", c.baseURL+"/v1/audio/speech", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", contentType)
		c.setAuth(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("synthesize request failed: %w", err)
		}
		data, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		status = resp.StatusCode

		if status != http.StatusConflict || time.Now().After(deadline) {
			break
		}
		time.Sleep(busyRetryDelay(attempt))
	}

	if status == http.StatusConflict {
		return nil, fmt.Errorf("TTS 服务持续繁忙（409），请确认没有其他程序在使用该服务")
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("synthesize failed (status %d): %s", status, truncateBody(data))
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("TTS 服务返回空音频")
	}

	// 兼容返回 WAV 的部署（例如自行封装的服务）
	if bytes.HasPrefix(data, []byte("RIFF")) {
		return data, nil
	}
	return PCMToWAV(data, SampleRate, Channels, BitsPerSample), nil
}

// ModelInfo GET /v1/model 返回的模型信息
type ModelInfo struct {
	ModelDigest   string  `json:"model_digest"`
	FrameRate     float64 `json:"frame_rate"`
	MaxNewTokens  int     `json:"max_new_tokens"`
	MaxBatchTexts int     `json:"max_batch_texts"`
}

// TestConnection 调用 GET /v1/model 验证服务可用并返回模型信息。
func (c *Client) TestConnection() (*ModelInfo, error) {
	testClient := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", c.baseURL+"/v1/model", nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}
	c.setAuth(req)

	resp, err := testClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error (status %d): %s", resp.StatusCode, truncateBody(data))
	}

	var info ModelInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("不是 Breeze-TTS-2 服务（/v1/model 返回无法解析）: %w", err)
	}
	return &info, nil
}

// PCMToWAV 为裸 PCM 数据加上 44 字节 WAV 头
func PCMToWAV(pcm []byte, sampleRate, channels, bitsPerSample int) []byte {
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	var buf bytes.Buffer
	buf.Grow(44 + len(pcm))
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(36+len(pcm)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(channels))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(byteRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(blockAlign))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample))
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(pcm)))
	buf.Write(pcm)
	return buf.Bytes()
}

func truncateBody(data []byte) string {
	s := string(data)
	if len(s) > 500 {
		return s[:500] + "..."
	}
	return s
}
