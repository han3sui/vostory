package sms

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AliyunNotifier 阿里云短信通知器
type AliyunNotifier struct {
	endpoint        string
	accessKeyId     string
	accessKeySecret string
	signName        string
	templateCode    string
	phoneList       []string
}

// NewAliyunNotifier 创建阿里云短信通知器
func NewAliyunNotifier(endpoint, accessKeyId, accessKeySecret, signName, templateCode string, phoneList []string) *AliyunNotifier {
	return &AliyunNotifier{
		endpoint:        endpoint,
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
		signName:        signName,
		templateCode:    templateCode,
		phoneList:       phoneList,
	}
}

// Notify 发送阿里云短信通知
func (n *AliyunNotifier) Notify(message string) error {
	// 阿里云短信API需要的公共参数
	params := map[string]string{
		"AccessKeyId":      n.accessKeyId,
		"Action":           "SendSms",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
		"RegionId":         "cn-hangzhou",
		"PhoneNumbers":     strings.Join(n.phoneList, ","),
		"SignName":         n.signName,
		"TemplateCode":     n.templateCode,
		"TemplateParam":    fmt.Sprintf("{\"content\":\"%s\"}", message),
	}

	// 构造签名
	signature := buildAliyunSignature(params, n.accessKeySecret)
	params["Signature"] = signature

	// 构造请求URL
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	requestURL := n.endpoint + "?" + query.Encode()

	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Aliyun SMS notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// buildAliyunSignature 构建阿里云API签名
func buildAliyunSignature(params map[string]string, secret string) string {
	// 参数排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构造规范化请求字符串
	var queryStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			queryStr.WriteString("&")
		}
		queryStr.WriteString(url.QueryEscape(k) + "=" + url.QueryEscape(params[k]))
	}

	// 构造待签名字符串
	stringToSign := "GET&" + url.QueryEscape("/") + "&" + url.QueryEscape(queryStr.String())

	// 计算签名
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}
