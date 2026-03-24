package phone

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AliyunNotifier 阿里云电话通知器
type AliyunNotifier struct {
	endpoint         string
	accessKeyId      string
	accessKeySecret  string
	callTtsCode      string
	phoneList        []string
	calledShowNumber string
}

// NewAliyunNotifier 创建阿里云电话通知器
func NewAliyunNotifier(endpoint, accessKeyId, accessKeySecret, callTtsCode string, phoneList []string, calledShowNumber string) *AliyunNotifier {
	return &AliyunNotifier{
		endpoint:         endpoint,
		accessKeyId:      accessKeyId,
		accessKeySecret:  accessKeySecret,
		callTtsCode:      callTtsCode,
		phoneList:        phoneList,
		calledShowNumber: calledShowNumber,
	}
}

// Notify 发送阿里云电话通知
func (n *AliyunNotifier) Notify(message string) error {
	// 阿里云API支持同时发给多个号码，但我们这里循环处理，便于处理错误
	for _, phone := range n.phoneList {
		// 阿里云语音API需要的参数
		params := map[string]string{
			"AccessKeyId":      n.accessKeyId,
			"Action":           "SingleCallByTts",
			"Format":           "JSON",
			"RegionId":         "cn-hangzhou",
			"SignatureMethod":  "HMAC-SHA1",
			"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
			"SignatureVersion": "1.0",
			"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			"Version":          "2017-05-25",
			"CalledNumber":     phone,
			"CalledShowNumber": n.calledShowNumber,
			"TtsCode":          n.callTtsCode,
			"TtsParam":         fmt.Sprintf("{\"message\":\"%s\"}", message),
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
			return fmt.Errorf("Aliyun Phone call failed for %s with status: %d", phone, resp.StatusCode)
		}

		// 读取响应体，检查返回的JSON是否包含错误
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode Aliyun response: %w", err)
		}

		// 检查是否有错误码
		if code, exists := result["Code"]; exists {
			return fmt.Errorf("Aliyun Phone call failed with code: %v, message: %v", code, result["Message"])
		}

		// 为避免对阿里云API过度请求，添加短暂延迟
		time.Sleep(200 * time.Millisecond)
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
