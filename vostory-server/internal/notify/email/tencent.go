package email

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// TencentNotifier 腾讯云Email通知器
type TencentNotifier struct {
	endpoint  string
	secretId  string
	secretKey string
	from      string
	to        []string
	subject   string
}

// NewTencentNotifier 创建腾讯云Email通知器
func NewTencentNotifier(endpoint, secretId, secretKey, from string, to []string, subject string) *TencentNotifier {
	return &TencentNotifier{
		endpoint:  endpoint,
		secretId:  secretId,
		secretKey: secretKey,
		from:      from,
		to:        to,
		subject:   subject,
	}
}

// Notify 发送腾讯云Email通知
func (n *TencentNotifier) Notify(message string) error {
	// 腾讯云API公共参数
	timestamp := time.Now().Unix()
	nonce := int(time.Now().UnixNano() / 1000000)

	// 请求参数
	params := map[string]interface{}{
		"Action":    "SendEmail",
		"Version":   "2020-10-02",
		"Region":    "ap-guangzhou",
		"Timestamp": timestamp,
		"Nonce":     nonce,
		"SecretId":  n.secretId,
		"From":      n.from,
		"Recipient": n.to,
		"Subject":   n.subject,
		"Html":      message,
	}

	// 构造签名
	signature := buildTencentSignature("POST", "ses.tencentcloudapi.com", "/", params, n.secretKey)

	// 添加签名到请求
	params["Signature"] = signature

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
		return fmt.Errorf("Tencent Email notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// buildTencentSignature 构建腾讯云API签名
func buildTencentSignature(method, host, path string, params map[string]interface{}, secretKey string) string {
	// 构造规范请求串
	httpRequestMethod := strings.ToLower(method)
	canonicalURI := path
	canonicalQueryString := ""

	// 参数排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 组合查询参数
	var paramStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			paramStr.WriteString("&")
		}
		var paramValue string
		switch v := params[k].(type) {
		case string:
			paramValue = v
		case int64:
			paramValue = fmt.Sprintf("%d", v)
		case int:
			paramValue = fmt.Sprintf("%d", v)
		case []string:
			paramValue = strings.Join(v, ",")
		default:
			paramValue = fmt.Sprintf("%v", v)
		}
		paramStr.WriteString(fmt.Sprintf("%s=%s", k, paramValue))
	}

	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:%s\n", host)
	signedHeaders := "content-type;host"

	// 组合请求字符串
	requestPayload := paramStr.String()
	hashedRequestPayload := getSHA256Hex(requestPayload)

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)

	// 计算签名
	algorithm := "TC3-HMAC-SHA256"
	requestTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	date := time.Now().UTC().Format("2006-01-02")
	service := strings.Split(host, ".")[0]

	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := getSHA256Hex(canonicalRequest)

	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		requestTimestamp,
		credentialScope,
		hashedCanonicalRequest)

	// 计算签名
	secretDate := hmacSHA256(fmt.Sprintf("TC3%s", secretKey), date)
	secretService := hmacSHA256(secretDate, service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256Binary(secretSigning, stringToSign))

	return signature
}

// getSHA256Hex 计算字符串的SHA256哈希值
func getSHA256Hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

// hmacSHA256 计算HMAC SHA256 并返回字符串
func hmacSHA256(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacSHA256Binary 计算HMAC SHA256 并返回二进制
func hmacSHA256Binary(key, data string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return h.Sum(nil)
}
