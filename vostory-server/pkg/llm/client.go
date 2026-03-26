package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	BaseURL        string
	APIKey         string
	Model          string
	Messages       []Message
	MaxTokens      int
	Temperature    float64
	ResponseFormat *ResponseFormat
	CustomParams   map[string]interface{}
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type ChatResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 180 * time.Second},
	}
}

func (c *Client) buildURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.Contains(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	return baseURL + "/v1/chat/completions"
}

func (c *Client) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	chatURL := c.buildURL(req.BaseURL)

	model := req.Model
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}

	messages := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	reqBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}

	if req.MaxTokens > 0 {
		reqBody["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		reqBody["temperature"] = req.Temperature
	}
	if req.ResponseFormat != nil {
		reqBody["response_format"] = req.ResponseFormat
	}

	for k, v := range req.CustomParams {
		if k != "model" && k != "messages" && k != "stream" {
			reqBody[k] = v
		}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if req.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API error HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		preview := string(respBody)
		if len(preview) > 500 {
			preview = preview[:500]
		}
		return nil, fmt.Errorf("parse response JSON: %w\nresponse preview: %s", err, preview)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned empty choices")
	}

	return &ChatResponse{
		Content:      result.Choices[0].Message.Content,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
	}, nil
}
