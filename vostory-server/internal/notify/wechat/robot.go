package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RobotNotifier 企业微信群机器人通知器
type RobotNotifier struct {
	webhookURL string
}

// NewRobotNotifier 创建企业微信群机器人通知器
func NewRobotNotifier(webhookURL string) *RobotNotifier {
	return &RobotNotifier{
		webhookURL: webhookURL,
	}
}

// Notify 发送企业微信群机器人通知
func (n *RobotNotifier) Notify(message string) error {
	// 企业微信群机器人API格式
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode WeChat robot response: %w", err)
	}

	// 企业微信API返回码为0表示成功
	if errcode, exists := result["errcode"]; exists && errcode.(float64) != 0 {
		return fmt.Errorf("WeChat robot notification failed with code: %v, message: %v", errcode, result["errmsg"])
	}

	return nil
}
