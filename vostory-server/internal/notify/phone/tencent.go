package phone

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// TencentNotifier 腾讯云电话通知器
type TencentNotifier struct {
	endpoint   string
	secretId   string
	secretKey  string
	appId      string
	templateId string
	phoneList  []string
	playTimes  int
}

// NewTencentNotifier 创建腾讯云电话通知器
func NewTencentNotifier(endpoint, secretId, secretKey, appId, templateId string, phoneList []string, playTimes int) *TencentNotifier {
	return &TencentNotifier{
		endpoint:   endpoint,
		secretId:   secretId,
		secretKey:  secretKey,
		appId:      appId,
		templateId: templateId,
		phoneList:  phoneList,
		playTimes:  playTimes,
	}
}

// Notify 发送腾讯云电话通知
func (n *TencentNotifier) Notify(message string) error {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 腾讯云语音API参数
	timestamp := time.Now().Unix()

	// 为每个电话号码发送通知
	for _, phone := range n.phoneList {
		// 生成随机数
		nonce := rand.Int31()

		// 构造请求参数
		params := map[string]interface{}{
			"Action":           "DescribeInstances",
			"Version":          "2019-03-01",
			"Timestamp":        timestamp,
			"Nonce":            nonce,
			"SecretId":         n.secretId,
			"Region":           "ap-guangzhou",
			"SmsSdkAppId":      n.appId,
			"TemplateId":       n.templateId,
			"PlayTimes":        n.playTimes,
			"TemplateParamSet": []string{message},
			"PhoneNumberSet":   []string{phone},
		}

		// 构造请求签名
		signature := buildTencentSignature("POST", "vms.tencentcloudapi.com", "/", params, n.secretKey)

		// 构造请求体
		requestBody := map[string]interface{}{
			"SmsSdkAppId":      n.appId,
			"TemplateId":       n.templateId,
			"PlayTimes":        n.playTimes,
			"TemplateParamSet": []string{message},
			"PhoneNumberSet":   []string{phone},
		}

		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}

		// 创建请求
		req, err := http.NewRequest("POST", n.endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization",
			fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s/vms/tc3_request, SignedHeaders=content-type;host, Signature=%s",
				n.secretId, time.Now().UTC().Format("2006-01-02"), signature))
		req.Header.Set("Host", "vms.tencentcloudapi.com")

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			return fmt.Errorf("Tencent Phone call failed for %s with status: %d", phone, resp.StatusCode)
		}

		// 读取响应体，检查是否有错误
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode Tencent response: %w", err)
		}

		// 检查错误信息
		if errObj, exists := result["Error"]; exists {
			errMap, ok := errObj.(map[string]interface{})
			if ok {
				return fmt.Errorf("Tencent Phone call failed with code: %v, message: %v",
					errMap["Code"], errMap["Message"])
			}
		}

		// 为避免对腾讯云API过度请求，添加短暂延迟
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

// buildTencentSignature 构建腾讯云API签名
func buildTencentSignature(method, host, path string, params map[string]interface{}, secretKey string) string {
	// 计算当前UTC时间
	timestamp := time.Now()
	dateStr := timestamp.UTC().Format("2006-01-02")

	// 准备签名内容
	canonicalRequest := buildCanonicalRequest(method, path, params, host)

	// 构造待签名字符串
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s/vms/tc3_request\n%s",
		timestamp.Unix(), dateStr, sha256hex(canonicalRequest))

	// 计算签名
	signature := sign(secretKey, dateStr, "vms", stringToSign)

	return signature
}

// buildCanonicalRequest 构建规范化请求字符串
func buildCanonicalRequest(method, path string, params map[string]interface{}, host string) string {
	// 将参数序列化为JSON字符串
	jsonBytes, _ := json.Marshal(params)
	payload := string(jsonBytes)

	// 构造规范请求
	httpRequestMethod := strings.ToLower(method)
	canonicalURI := path
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:%s\n", host)
	signedHeaders := "content-type;host"
	hashedRequestPayload := sha256hex(payload)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)
}

// sha256hex 计算字符串的SHA256哈希值并返回十六进制表示
func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

// sign 计算签名
func sign(secretKey, date, service, stringToSign string) string {
	// 计算派生密钥
	dateKey := hmacSHA256(fmt.Sprintf("TC3%s", secretKey), date)
	serviceKey := hmacSHA256(dateKey, service)
	secretSigning := hmacSHA256(serviceKey, "tc3_request")

	// 签名
	return hmacSHA256(secretSigning, stringToSign)
}

// hmacSHA256 计算HMAC SHA256并返回十六进制表示
func hmacSHA256(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
