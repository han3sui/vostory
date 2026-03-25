package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

type SynthesizeRequest struct {
	Text      string    `json:"text"`
	AudioPath string    `json:"audio_path"`
	EmoVector []float64 `json:"emo_vector,omitempty"`
	EmoText   string    `json:"emo_text,omitempty"`
}

// Synthesize calls POST /v2/synthesize and returns raw audio bytes.
func (c *Client) Synthesize(text, audioPath string, emoVector []float64, emoText string) ([]byte, error) {
	reqBody := SynthesizeRequest{
		Text:      text,
		AudioPath: audioPath,
	}
	if len(emoVector) > 0 {
		reqBody.EmoVector = emoVector
	} else if emoText != "" {
		reqBody.EmoText = emoText
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/v2/synthesize", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("synthesize request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("synthesize failed (status %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

type checkAudioResponse struct {
	Exists bool `json:"exists"`
}

// CheckAudioExists calls GET /v1/check/audio?file_name=xxx
func (c *Client) CheckAudioExists(fileName string) (bool, error) {
	u := fmt.Sprintf("%s/v1/check/audio?file_name=%s", c.baseURL, url.QueryEscape(fileName))

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return false, fmt.Errorf("check audio request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("check audio failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result checkAudioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode check audio response: %w", err)
	}
	return result.Exists, nil
}

// UploadAudio calls POST /v1/upload_audio with multipart form data.
func (c *Client) UploadAudio(filePath, fullPath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("copy file data: %w", err)
	}

	if fullPath != "" {
		if err := writer.WriteField("full_path", fullPath); err != nil {
			return fmt.Errorf("write full_path field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/v1/upload_audio", writer.FormDataContentType(), &buf)
	if err != nil {
		return fmt.Errorf("upload audio request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload audio failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// EnsureAudioUploaded checks if the reference audio exists on the TTS server,
// and uploads it if not. Returns the key (fullPath) used on the server side.
func (c *Client) EnsureAudioUploaded(localPath, remoteKey string) error {
	exists, err := c.CheckAudioExists(remoteKey)
	if err != nil {
		return fmt.Errorf("check audio: %w", err)
	}
	if exists {
		return nil
	}
	return c.UploadAudio(localPath, remoteKey)
}

// TestConnection calls GET /v1/models to verify the TTS service is reachable.
func (c *Client) TestConnection() error {
	testClient := &http.Client{Timeout: 15 * time.Second}

	resp, err := testClient.Get(c.baseURL + "/v1/models")
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error (status %d)", resp.StatusCode)
	}
	return nil
}
