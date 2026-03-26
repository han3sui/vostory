package llm

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
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

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) buildConfig(baseURL, apiKey string) openai.ClientConfig {
	config := openai.DefaultConfig(apiKey)
	base := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(base, "/v1") {
		base += "/v1"
	}
	config.BaseURL = base
	return config
}

func (c *Client) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	config := c.buildConfig(req.BaseURL, req.APIKey)
	client := openai.NewClientWithConfig(config)

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	chatReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   false,
	}

	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		chatReq.Temperature = float32(req.Temperature)
	}
	if req.ResponseFormat != nil {
		chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatType(req.ResponseFormat.Type),
		}
	}

	resp, err := client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned empty choices")
	}

	return &ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}, nil
}
