package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BotConfig Telegram机器人通知器配置
type BotConfig struct {
	BotToken  string
	ChatID    string
	ParseMode string
}

// NewBotNotifier 创建Telegram机器人通知器
func NewBotNotifier(config BotConfig) *BotConfig {
	return &BotConfig{
		BotToken:  config.BotToken,
		ChatID:    config.ChatID,
		ParseMode: config.ParseMode,
	}
}

// Notify 发送Telegram通知
func (n *BotConfig) Notify(message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.BotToken)

	// Telegram API格式
	payload := map[string]string{
		"chat_id": n.ChatID,
		"text":    message,
	}

	// 如果指定了解析模式，添加到请求中
	if n.ParseMode != "" {
		payload["parse_mode"] = n.ParseMode
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		return fmt.Errorf("%v", string(body))
	}

	return nil
}
