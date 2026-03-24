package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RobotNotifier 钉钉群机器人通知器
type RobotNotifier struct {
	webhookURL string
	secret     string // 用于签名的secret
}

// NewRobotNotifier 创建钉钉群机器人通知器
func NewRobotNotifier(webhookURL string, secret string) *RobotNotifier {
	return &RobotNotifier{
		webhookURL: webhookURL,
		secret:     secret,
	}
}

// Notify 发送钉钉群机器人通知
func (n *RobotNotifier) Notify(message string) error {
	// 构造签名
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	stringToSign := timestamp + "\n" + n.secret

	mac := hmac.New(sha256.New, []byte(n.secret))
	mac.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	// 构造请求URL
	requestURL := fmt.Sprintf("%s&timestamp=%s&sign=%s", n.webhookURL, timestamp, sign)

	// 钉钉API格式
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

	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode DingTalk robot response: %w", err)
	}

	// 钉钉API返回errcode为0表示成功
	if errcode, exists := result["errcode"]; exists && errcode.(float64) != 0 {
		return fmt.Errorf("DingTalk robot notification failed with code: %v, message: %v",
			errcode, result["errmsg"])
	}

	return nil
}
