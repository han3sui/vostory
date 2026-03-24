package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// TencentNotifier 腾讯云短信通知器
type TencentNotifier struct {
	endpoint   string
	appId      string
	appKey     string
	signName   string
	templateId string
	phoneList  []string
}

// NewTencentNotifier 创建腾讯云短信通知器
func NewTencentNotifier(endpoint, appId, appKey, signName, templateId string, phoneList []string) *TencentNotifier {
	return &TencentNotifier{
		endpoint:   endpoint,
		appId:      appId,
		appKey:     appKey,
		signName:   signName,
		templateId: templateId,
		phoneList:  phoneList,
	}
}

// Notify 发送腾讯云短信通知
func (n *TencentNotifier) Notify(message string) error {
	// 腾讯云短信参数
	timestamp := time.Now().Unix()
	random := rand.New(rand.NewSource(time.Now().UnixNano())).Int31()

	// 构造请求参数
	params := map[string]interface{}{
		"params": []string{message},
		"sig":    calculateTencentSignature(n.appKey, timestamp, random, n.phoneList),
		"sign":   n.signName,
		"tel":    prepareTencentPhones(n.phoneList),
		"time":   timestamp,
		"tpl_id": n.templateId,
	}

	// 添加其他必要的参数
	params["sdkappid"] = n.appId
	params["random"] = random

	jsonData, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", n.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Tencent SMS notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// prepareTencentPhones 准备腾讯云格式的电话号码
func prepareTencentPhones(phoneList []string) []map[string]string {
	phones := make([]map[string]string, len(phoneList))
	for i, phone := range phoneList {
		phones[i] = map[string]string{
			"mobile":     phone,
			"nationcode": "86", // 默认中国区号
		}
	}
	return phones
}

// calculateTencentSignature 计算腾讯云API签名
func calculateTencentSignature(appKey string, timestamp int64, random int32, phoneList []string) string {
	// 构建手机号码字符串
	var mobileStr string
	for i, phone := range phoneList {
		if i > 0 {
			mobileStr += ","
		}
		mobileStr += phone
	}

	// 构建签名内容
	content := fmt.Sprintf("appkey=%s&random=%d&time=%d&mobile=%s",
		appKey, random, timestamp, mobileStr)

	// HMAC-SHA256计算
	h := hmac.New(sha256.New, []byte(appKey))
	h.Write([]byte(content))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
