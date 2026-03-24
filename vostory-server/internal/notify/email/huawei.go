package email

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HuaweiNotifier 华为云Email通知器
type HuaweiNotifier struct {
	apiURL    string
	appKey    string
	appSecret string
	from      string
	to        []string
	subject   string
}

// NewHuaweiNotifier 创建华为云Email通知器
func NewHuaweiNotifier(apiURL, appKey, appSecret, from string, to []string, subject string) *HuaweiNotifier {
	return &HuaweiNotifier{
		apiURL:    apiURL,
		appKey:    appKey,
		appSecret: appSecret,
		from:      from,
		to:        to,
		subject:   subject,
	}
}

// Notify 发送华为云Email通知
func (n *HuaweiNotifier) Notify(message string) error {
	// 构造WSSE认证头
	xWSSEHeader := buildHuaweiWSSEHeader(n.appKey, n.appSecret)

	// 构造华为云邮件API请求
	payload := map[string]interface{}{
		"from":    n.from,
		"to":      n.to,
		"subject": n.subject,
		"content": map[string]string{
			"html": message,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", n.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "WSSE realm=\"SDP\",profile=\"UsernameToken\",type=\"Appkey\"")
	req.Header.Set("X-WSSE", xWSSEHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Huawei Email notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// buildHuaweiWSSEHeader 构建华为云API的WSSE认证头
func buildHuaweiWSSEHeader(appKey, appSecret string) string {
	now := time.Now().Format(time.RFC3339)
	nonce := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)

	// 计算PasswordDigest
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(nonce + now))
	passwordDigest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("UsernameToken Username=\"%s\",PasswordDigest=\"%s\",Nonce=\"%s\",Created=\"%s\"",
		appKey, passwordDigest, nonce, now)
}
