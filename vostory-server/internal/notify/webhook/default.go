package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DefaultNotifier 默认Webhook通知器
type DefaultNotifier struct {
	url     string
	headers map[string]string
}

// NewDefaultNotifier 创建默认Webhook通知器
func NewDefaultNotifier(url string, headers map[string]string) *DefaultNotifier {
	return &DefaultNotifier{
		url:     url,
		headers: headers,
	}
}

// Notify 发送Webhook通知
func (n *DefaultNotifier) Notify(message string) error {
	payload := map[string]string{
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", n.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range n.headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook notification failed with status: %d", resp.StatusCode)
	}

	return nil
}
