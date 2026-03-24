package kafka

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"iot-alert-center/pkg/log"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// KafkaProducer Kafka生产者结构体
type KafkaProducer struct {
	log         *zap.Logger
	conf        *viper.Viper
	writer      *kafka.Writer
	mutex       sync.RWMutex
	isRunning   bool
	isConnected int32 // 使用原子操作，0=未连接，1=已连接
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(
	logger *log.Logger,
	conf *viper.Viper,
) *KafkaProducer {
	return &KafkaProducer{
		log:         logger.Logger, // 从log.Logger中获取内部的*zap.Logger
		conf:        conf,
		isConnected: 0, // 显式初始化为未连接状态
	}
}

func (k *KafkaProducer) Start(ctx context.Context) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.isRunning {
		k.log.Info("Kafka生产者已经在运行")
		return nil
	}

	// 从配置中获取Kafka连接参数
	brokerURL := k.conf.GetString("kafka_consumer.broker_url")

	if brokerURL == "" {
		k.log.Error("Kafka连接地址未配置")
		return fmt.Errorf("Kafka连接地址未配置")
	}

	k.log.Info("启动Kafka生产者",
		zap.String("broker", brokerURL))

	// 创建Kafka写入器
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(brokerURL),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireOne,
		Async:        true, // 启用异步写入
		ErrorLogger:  kafka.LoggerFunc(k.log.Sugar().Errorf),
	}

	atomic.StoreInt32(&k.isConnected, 1)
	k.log.Info("Kafka生产者启动成功", zap.String("broker", brokerURL))
	k.isRunning = true

	// 保持运行直到上下文取消
	<-ctx.Done()
	return nil
}

func (k *KafkaProducer) Stop(ctx context.Context) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if !k.isRunning {
		return nil
	}

	// 关闭Kafka写入器
	if k.writer != nil && atomic.LoadInt32(&k.isConnected) == 1 {
		if err := k.writer.Close(); err != nil {
			k.log.Error("关闭Kafka写入器失败", zap.Error(err))
		}
		atomic.StoreInt32(&k.isConnected, 0)
		k.log.Info("Kafka写入器已关闭")
	}

	k.isRunning = false
	k.log.Info("Kafka生产者已停止")
	return nil
}

// IsConnected 检查连接状态
func (k *KafkaProducer) IsConnected() bool {
	// 直接使用原子操作读取连接状态，完全避免锁和任何可能的阻塞
	return atomic.LoadInt32(&k.isConnected) == 1
}

// PublishMessage 发布消息到指定主题
func (k *KafkaProducer) PublishMessage(ctx context.Context, topic string, key string, payload []byte) error {
	if !k.IsConnected() {
		return fmt.Errorf("Kafka生产者未连接")
	}

	message := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	// k.log.Debug("发布Kafka消息",
	// 	zap.String("topic", topic),
	// 	zap.String("key", key),
	// 	zap.Int("payload_size", len(payload)))

	err := k.writer.WriteMessages(ctx, message)
	if err != nil {
		k.log.Error("发布Kafka消息失败",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	// k.log.Debug("Kafka消息发布成功",
	// 	zap.String("topic", topic),
	// 	zap.String("key", key))
	return nil
}

// PublishWebhookMessage 发布EMQX Webhook消息到Kafka
func (k *KafkaProducer) PublishWebhookMessage(ctx context.Context, payload []byte) error {
	// 使用配置中的主题
	topic := k.conf.GetStringSlice("kafka_consumer.topics")[0] // 使用第一个主题
	if topic == "" {
		return fmt.Errorf("Kafka主题未配置")
	}

	// 使用时间戳作为key确保消息分布
	key := fmt.Sprintf("webhook_%d", time.Now().UnixNano())

	return k.PublishMessage(ctx, topic, key, payload)
}
