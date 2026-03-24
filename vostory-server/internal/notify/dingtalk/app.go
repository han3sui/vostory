package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// AppNotifier 钉钉工作通知器
type AppNotifier struct {
	appKey       string
	appSecret    string
	agentId      int
	userIdList   []string
	deptIdList   []string
	accessToken  string
	tokenExpires time.Time
}

// NewAppNotifier 创建钉钉工作通知器
func NewAppNotifier(appKey, appSecret string, agentId int, userIdList, deptIdList []string) *AppNotifier {
	return &AppNotifier{
		appKey:     appKey,
		appSecret:  appSecret,
		agentId:    agentId,
		userIdList: userIdList,
		deptIdList: deptIdList,
	}
}

// Notify 发送钉钉工作通知
func (n *AppNotifier) Notify(message string) error {
	// 检查token是否过期，过期则重新获取
	if n.accessToken == "" || time.Now().After(n.tokenExpires) {
		err := n.getAccessToken()
		if err != nil {
			return err
		}
	}

	// 构造API请求
	apiURL := "https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2"

	// 构造请求参数
	form := url.Values{}
	form.Add("access_token", n.accessToken)
	form.Add("agent_id", strconv.Itoa(n.agentId))

	if len(n.userIdList) > 0 {
		form.Add("userid_list", strings.Join(n.userIdList, ","))
	}

	if len(n.deptIdList) > 0 {
		form.Add("dept_id_list", strings.Join(n.deptIdList, ","))
	}

	// 构造消息内容
	msgContent := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	msgBytes, err := json.Marshal(msgContent)
	if err != nil {
		return err
	}
	form.Add("msg", string(msgBytes))

	// 发送请求
	resp, err := http.PostForm(apiURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode DingTalk app response: %w", err)
	}

	// 钉钉API返回errcode为0表示成功
	if errcode, exists := result["errcode"]; exists && errcode.(float64) != 0 {
		return fmt.Errorf("DingTalk app notification failed with code: %v, message: %v",
			errcode, result["errmsg"])
	}

	return nil
}

// getAccessToken 获取钉钉访问令牌
func (n *AppNotifier) getAccessToken() error {
	timestamp := time.Now().UnixMilli()
	// 签名计算
	signStr := strconv.FormatInt(timestamp, 10)
	signData := fmt.Sprintf("%s\n%s", signStr, n.appSecret)

	// 计算签名
	h := hmac.New(sha256.New, []byte(n.appSecret))
	h.Write([]byte(signData))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	// 构造获取token的URL
	tokenURL := fmt.Sprintf(
		"https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s&timestamp=%d&signature=%s",
		n.appKey, n.appSecret, timestamp, signature)

	// 发送请求
	resp, err := http.Get(tokenURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode DingTalk token response: %w", err)
	}

	// 检查返回值
	if errcode, exists := result["errcode"]; exists && errcode.(float64) != 0 {
		return fmt.Errorf("failed to get DingTalk access token: %v", result["errmsg"])
	}

	// 获取token和过期时间
	if token, exists := result["access_token"]; exists {
		n.accessToken = token.(string)
		expires, _ := result["expires_in"].(float64)
		// 比官方过期时间提前5分钟过期，确保安全
		n.tokenExpires = time.Now().Add(time.Duration(expires-300) * time.Second)
		return nil
	}

	return fmt.Errorf("access_token not found in response")
}
