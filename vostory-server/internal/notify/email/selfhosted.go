package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type SelfHostedConfig struct {
	Host      string   `json:"host"`
	Port      int64    `json:"port"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	EnableTLS bool     `json:"enableTLS"`
}

// SelfHostedNotifier 自建邮件服务器通知器
type SelfHostedNotifier struct {
	config SelfHostedConfig
}

// NewSelfHostedNotifier 创建自建邮件服务器通知器
func NewSelfHostedNotifier(config SelfHostedConfig) *SelfHostedNotifier {
	return &SelfHostedNotifier{
		config: config,
	}
}

// Notify 发送自建邮件服务器通知
func (n *SelfHostedNotifier) Notify(message string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.config.From)
	m.SetHeader("To", n.config.To...)
	m.SetHeader("Subject", n.config.Subject)
	m.SetBody("text/html", message)

	d := gomail.NewDialer(n.config.Host, int(n.config.Port), n.config.Username, n.config.Password)
	if !n.config.EnableTLS {
		d.SSL = false
		d.TLSConfig = nil
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
