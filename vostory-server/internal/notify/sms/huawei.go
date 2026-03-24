package sms

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HuaweiConfig 华为云短信通知器配置
type HuaweiConfig struct {
	From            string   `json:"from"`
	To              []string `json:"to"`
	TemplateId      string   `json:"templateId"`
	TemplateParams  []string `json:"templateParams"`
	Endpoint        string   `json:"endpoint"`
	AccessKeyId     string   `json:"accessKeyId"`
	AccessKeySecret string   `json:"accessKeySecret"`
}

// NewHuaweiNotifier 创建华为云短信通知器
func NewHuaweiNotifier(config HuaweiConfig) *HuaweiConfig {
	return &HuaweiConfig{
		From:            config.From,
		To:              config.To,
		TemplateId:      config.TemplateId,
		TemplateParams:  config.TemplateParams,
		Endpoint:        config.Endpoint,
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
	}
}

type HuaweiHttpResult struct {
	Code        string               `json:"code"`
	DesCription string               `json:"description"`
	Result      []HuaweiNotifyResult `json:"result"`
}

type HuaweiNotifyResult struct {
	SmsMsgId   string `json:"smsMsgId"`
	From       string `json:"from"`
	OriginTo   string `json:"originTo"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	CountryId  string `json:"countryId"`
	Total      int    `json:"total"`
}

// Notify 发送华为云短信通知
func (n *HuaweiConfig) Notify(message string) ([]HuaweiNotifyResult, error) {
	// 华为云API格式
	xWSSEHeader := buildWSSEHeader(n.AccessKeyId, n.AccessKeySecret)

	// 构造短信参数，华为云需要使用templateId和模板参数
	templateParams := []string{message}

	values := url.Values{}
	values.Set("from", n.From)
	values.Set("to", strings.Join(n.To, ","))
	values.Set("templateId", n.TemplateId)
	templateParasJSON, _ := json.Marshal(templateParams)
	values.Set("templateParas", string(templateParasJSON))
	values.Set("statusCallback", "")

	req, err := http.NewRequest("POST", n.Endpoint+"/sms/batchSendSms/v1", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "WSSE realm=\"SDP\",profile=\"UsernameToken\",type=\"Appkey\"")
	req.Header.Set("X-WSSE", xWSSEHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%v", string(body))
	}

	var huaweiHttpResult HuaweiHttpResult
	err = json.Unmarshal(body, &huaweiHttpResult)
	if err != nil {
		return nil, err
	}

	if huaweiHttpResult.Code != "200" {
		return nil, fmt.Errorf("%v", huaweiHttpResult.DesCription)
	}

	return huaweiHttpResult.Result, nil
}

// buildWSSEHeader 构建华为云API的WSSE认证头
func buildWSSEHeader(appKey, appSecret string) string {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	nonce := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)

	// 计算PasswordDigest
	h := sha256.New()
	h.Write([]byte(nonce + now + appSecret))
	passwordDigest := fmt.Sprintf("%x", h.Sum(nil))

	return fmt.Sprintf("UsernameToken Username=\"%s\",PasswordDigest=\"%s\",Nonce=\"%s\",Created=\"%s\"",
		appKey, passwordDigest, nonce, now)
}
