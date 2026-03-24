package mqtt

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"iot-alert-center/pkg/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// MqttClient MQTT客户端结构体
type MqttClient struct {
	log         *zap.Logger
	conf        *viper.Viper
	client      mqtt.Client
	mutex       sync.RWMutex
	isRunning   bool
	isConnected int32 // 使用原子操作，0=未连接，1=已连接
}

// NewMqttClient 创建MQTT客户端
func NewMqttClient(
	logger *log.Logger,
	conf *viper.Viper,
) *MqttClient {
	return &MqttClient{
		log:         logger.Logger, // 从log.Logger中获取内部的*zap.Logger
		conf:        conf,
		isConnected: 0, // 显式初始化为未连接状态
	}
}

func (m *MqttClient) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		m.log.Info("MQTT客户端已经在运行")
		return nil
	}

	// 从配置中获取MQTT连接参数
	host := m.conf.GetString("emqx.host")
	mqttPort := m.conf.GetString("emqx.mqtt_port")
	brokerURL := fmt.Sprintf("tcp://%s:%s", host, mqttPort)
	clientID := m.conf.GetString("emqx.client_id")
	username := m.conf.GetString("emqx.username")
	password := m.conf.GetString("emqx.password")

	if brokerURL == "" || clientID == "" || username == "" || password == "" {
		m.log.Error("MQTT连接参数不能为空")
		return fmt.Errorf("MQTT连接参数不能为空")
	}

	m.log.Info("启动MQTT客户端",
		zap.String("broker", brokerURL),
		zap.String("client_id", clientID),
		zap.String("username", username))

	// 创建MQTT客户端选项（优化版本）
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			atomic.StoreInt32(&m.isConnected, 0) // 原子设置为未连接
			m.log.Error("MQTT连接丢失", zap.Error(err))
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			atomic.StoreInt32(&m.isConnected, 0) // 重连时先设为未连接
			m.log.Info("正在尝试重新连接MQTT...")
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			atomic.StoreInt32(&m.isConnected, 1) // 原子设置为已连接
			m.log.Info("MQTT客户端已连接到broker")
		})

	// 创建MQTT客户端
	m.client = mqtt.NewClient(opts)

	// 连接到MQTT broker，带重试机制
	if err := m.connectWithRetry(ctx, brokerURL); err != nil {
		return err
	}

	atomic.StoreInt32(&m.isConnected, 1)
	m.log.Info("MQTT客户端启动成功", zap.String("broker", brokerURL))
	m.isRunning = true

	// 保持运行直到上下文取消
	<-ctx.Done()
	return nil
}

// connectWithRetry 带重试的连接方法
func (m *MqttClient) connectWithRetry(ctx context.Context, brokerURL string) error {
	maxRetries := 15
	baseDelay := 2 * time.Second
	maxDelay := 30 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m.log.Info("尝试连接MQTT broker",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.String("broker", brokerURL))

		// 尝试连接
		if token := m.client.Connect(); token.Wait() && token.Error() != nil {
			atomic.StoreInt32(&m.isConnected, 0)

			if attempt == maxRetries {
				// 最后一次重试失败，返回错误
				m.log.Error("连接MQTT broker失败，已达到最大重试次数",
					zap.Error(token.Error()),
					zap.Int("max_retries", maxRetries),
					zap.String("broker", brokerURL))
				return fmt.Errorf("连接MQTT broker失败: %w", token.Error())
			}

			// 计算重试延迟（指数退避）
			delay := time.Duration(attempt) * baseDelay
			if delay > maxDelay {
				delay = maxDelay
			}

			m.log.Warn("MQTT连接失败，准备重试",
				zap.Error(token.Error()),
				zap.Int("attempt", attempt),
				zap.Duration("retry_delay", delay))

			// 等待重试延迟
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			// 连接成功
			m.log.Info("MQTT连接成功",
				zap.Int("attempt", attempt),
				zap.String("broker", brokerURL))
			return nil
		}
	}

	return fmt.Errorf("连接MQTT broker失败，已达到最大重试次数")
}

func (m *MqttClient) Stop(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return nil
	}

	// 断开MQTT连接
	if m.client != nil && atomic.LoadInt32(&m.isConnected) == 1 {
		m.client.Disconnect(250)
		atomic.StoreInt32(&m.isConnected, 0)
		m.log.Info("MQTT客户端已断开连接")
	}

	m.isRunning = false
	m.log.Info("MQTT客户端已停止")
	return nil
}

// IsConnected 检查连接状态（完全优化版本）
func (m *MqttClient) IsConnected() bool {
	// 直接使用原子操作读取连接状态，完全避免锁和任何可能的阻塞
	// 这是一个无锁、非阻塞的操作
	return atomic.LoadInt32(&m.isConnected) == 1
}

// PublishMessage 发布消息到指定主题
func (m *MqttClient) PublishMessage(ctx context.Context, topic string, payload []byte, qos byte) error {
	if !m.IsConnected() {
		return fmt.Errorf("MQTT客户端未连接")
	}

	// m.log.Debug("发布MQTT消息",
	// 	zap.String("topic", topic),
	// 	zap.Int("payload_size", len(payload)),
	// 	zap.Uint8("qos", qos))

	token := m.client.Publish(topic, qos, false, payload)
	if token.Wait() && token.Error() != nil {
		m.log.Error("发布MQTT消息失败",
			zap.String("topic", topic),
			zap.Error(token.Error()))
		return token.Error()
	}

	// m.log.Debug("MQTT消息发布成功", zap.String("topic", topic))
	return nil
}

// PublishDeviceResponse 发布设备响应消息
func (m *MqttClient) PublishDeviceResponse(ctx context.Context, deviceID, productID, messageType string, payload []byte) error {
	var topic string

	// 根据消息类型构建响应主题
	switch messageType {
	case "property":
		// 属性上报响应
		topic = fmt.Sprintf("/sys/%s/%s/thing/property/post_reply", productID, deviceID)
	case "event":
		// 事件上报响应
		topic = fmt.Sprintf("/sys/%s/%s/thing/event/post_reply", productID, deviceID)
	default:
		return fmt.Errorf("不支持的消息类型: %s", messageType)
	}

	// m.log.Info("发布设备响应消息",
	// 	zap.String("device_id", deviceID),
	// 	zap.String("product_id", productID),
	// 	zap.String("message_type", messageType),
	// 	zap.String("topic", topic))

	return m.PublishMessage(ctx, topic, payload, 1) // 使用QoS 1确保消息送达
}
