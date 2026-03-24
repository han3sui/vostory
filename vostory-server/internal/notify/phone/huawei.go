package phone

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

// HuaweiNotifier 华为云电话通知器
type HuaweiNotifier struct {
	apiURL     string
	appKey     string
	appSecret  string
	callFrom   string
	phoneList  []string
	templateId string
}

// NewHuaweiNotifier 创建华为云电话通知器
func NewHuaweiNotifier(apiURL, appKey, appSecret, callFrom string, phoneList []string, templateId string) *HuaweiNotifier {
	return &HuaweiNotifier{
		apiURL:     apiURL,
		appKey:     appKey,
		appSecret:  appSecret,
		callFrom:   callFrom,
		phoneList:  phoneList,
		templateId: templateId,
	}
}

// Notify 发送华为云电话通知
func (n *HuaweiNotifier) Notify(message string) error {
	// 华为云API WSSE认证头
	xWSSEHeader := buildHuaweiWSSEHeader(n.appKey, n.appSecret)

	// 对每个号码进行呼叫
	for _, phone := range n.phoneList {
		// 构造API请求体
		payload := map[string]interface{}{
			"call_type":          "VOICE_TTS",
			"caller":             n.callFrom,
			"callee":             phone,
			"display_name":       "告警通知",
			"status_callback":    "",
			"tts_template_id":    n.templateId,
			"tts_template_param": fmt.Sprintf("[%q]", message),
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
			return fmt.Errorf("Huawei Phone call failed for %s with status: %d", phone, resp.StatusCode)
		}

		// 为避免对华为云API过度请求，添加短暂延迟
		time.Sleep(200 * time.Millisecond)
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
