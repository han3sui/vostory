package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AppMarkDown struct {
	Content string `json:"content"`
}

// AppConfig 企业微信应用消息配置
type AppConfig struct {
	CorpID     string      `json:"corpId"`     // 企业ID
	CorpSecret string      `json:"corpSecret"` // 应用密钥
	AgentID    int64       `json:"agentId"`    // 应用ID
	ToUser     string      `json:"toUser"`     // 接收者用户列表
	ToParty    string      `json:"toParty"`    // 接收者部门列表
	ToTag      string      `json:"toTag"`      // 接收者标签列表
	Markdown   AppMarkDown `json:"markdown"`   // 消息内容
}

// AppNotifier 企业微信应用消息通知器
type AppNotifier struct {
	config       AppConfig
	accessToken  string    // 访问令牌
	tokenExpires time.Time // 令牌过期时间
}

// AppNotifierResult 企业微信应用消息通知器返回结果 https://developer.work.weixin.qq.com/document/path/90236
type AppNotifierResult struct {
	Errcode        int    `json:"errcode"`
	Errmsg         string `json:"errmsg"`
	Invaliduser    string `json:"invaliduser"`    // 不合法的成员帐号列表，|分隔
	Invalidparty   string `json:"invalidparty"`   // 不合法的部门id列表，|分隔
	Invalidtag     string `json:"invalidtag"`     // 不合法的标签id列表，|分隔
	Unlicenseduser string `json:"unlicenseduser"` // 没有操作权限的成员帐号列表，|分隔
	Msgid          string `json:"msgid"`          // 消息id
	ResponseCode   int    `json:"response_code"`  // 响应码
}

// NewAppNotifier 创建企业微信应用消息通知器
func NewAppNotifier(config AppConfig) (*AppNotifier, error) {
	// 验证必要参数
	if config.CorpID == "" {
		return nil, fmt.Errorf("missing corpId for WeChat App notifier")
	}
	if config.CorpSecret == "" {
		return nil, fmt.Errorf("missing corpSecret for WeChat App notifier")
	}
	if config.AgentID <= 0 {
		return nil, fmt.Errorf("invalid agentId for WeChat App notifier")
	}
	if config.ToUser == "" && config.ToParty == "" && config.ToTag == "" {
		return nil, fmt.Errorf("at least one of toUser, toParty, or toTag must be specified")
	}

	return &AppNotifier{
		config: config,
	}, nil
}

// Notify 发送企业微信应用消息通知
func (n *AppNotifier) Notify(message string) error {
	// 检查token是否过期，过期则重新获取
	if n.accessToken == "" || time.Now().After(n.tokenExpires) {
		err := n.getAccessToken()
		if err != nil {
			return err
		}
	}

	// 构造消息内容
	payload := map[string]interface{}{
		"touser":  n.config.ToUser,
		"agentid": n.config.AgentID,
	}
	if n.config.Markdown.Content != "" {
		payload["msgtype"] = "markdown"
		payload["markdown"] = n.config.Markdown
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 发送消息
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", n.accessToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应
	var result AppNotifierResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode WeChat response: %w", err)
	}

	// 企业微信API返回码为0表示成功
	if result.Errcode != 0 {
		return fmt.Errorf("WeChat App notification failed with code: %v, message: %v", result.Errcode, result.Errmsg)
	}

	return nil
}

// getAccessToken 获取企业微信访问令牌
func (n *AppNotifier) getAccessToken() error {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		n.config.CorpID, n.config.CorpSecret)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode WeChat token response: %w", err)
	}

	if errcode, exists := result["errcode"]; exists && errcode.(float64) != 0 {
		return fmt.Errorf("failed to get WeChat access token: %v", result["errmsg"])
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
