package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

// Synthesize sends the reference audio file along with the text in a single
// multipart request. The TTS server does not persist any files.
func (c *Client) Synthesize(audioFilePath, text string, emoVector []float64, emoText string) ([]byte, error) {
	f, err := os.Open(audioFilePath)
	if err != nil {
		return nil, fmt.Errorf("open audio file %s: %w", audioFilePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("audio", filepath.Base(audioFilePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copy audio data: %w", err)
	}

	if err := writer.WriteField("text", text); err != nil {
		return nil, fmt.Errorf("write text field: %w", err)
	}

	if len(emoVector) > 0 {
		vecJSON, _ := json.Marshal(emoVector)
		if err := writer.WriteField("emo_vector", string(vecJSON)); err != nil {
			return nil, fmt.Errorf("write emo_vector field: %w", err)
		}
	} else if emoText != "" {
		if err := writer.WriteField("emo_text", emoText); err != nil {
			return nil, fmt.Errorf("write emo_text field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/v2/synthesize", &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
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

// TestConnection calls GET /health to verify the TTS service is reachable.
func (c *Client) TestConnection() error {
	testClient := &http.Client{Timeout: 15 * time.Second}

	resp, err := testClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error (status %d)", resp.StatusCode)
	}
	return nil
}
