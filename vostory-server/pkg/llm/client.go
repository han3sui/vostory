package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func (c *Client) buildRequest(req *ChatRequest) openai.ChatCompletionRequest {
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
	return chatReq
}

func (c *Client) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	config := c.buildConfig(req.BaseURL, req.APIKey)
	client := openai.NewClientWithConfig(config)

	chatReq := c.buildRequest(req)
	chatReq.Stream = false

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

// ChatCompletionStream 使用 SSE 流式接收，拼接完整内容后返回。
// 流式传输可避免 Cloudflare 等反向代理的超时断连（如 524）。
func (c *Client) ChatCompletionStream(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	config := c.buildConfig(req.BaseURL, req.APIKey)
	client := openai.NewClientWithConfig(config)

	chatReq := c.buildRequest(req)
	chatReq.Stream = true
	chatReq.StreamOptions = &openai.StreamOptions{IncludeUsage: true}

	stream, err := client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}
	defer stream.Close()

	var buf strings.Builder
	var inputTokens, outputTokens int

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("LLM stream error: %w", err)
		}
		if len(chunk.Choices) > 0 {
			buf.WriteString(chunk.Choices[0].Delta.Content)
		}
		if chunk.Usage != nil {
			inputTokens = chunk.Usage.PromptTokens
			outputTokens = chunk.Usage.CompletionTokens
		}
	}

	return &ChatResponse{
		Content:      buf.String(),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}, nil
}
