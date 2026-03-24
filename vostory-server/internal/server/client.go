package server

import (
	"context"
	"iot-alert-center/internal/client/kafka"
	"iot-alert-center/internal/client/mqtt"
	"iot-alert-center/internal/client/resty"
	"iot-alert-center/pkg/log"
)

type ClientServer struct {
	log           *log.Logger
	mqttClient    *mqtt.MqttClient
	kafkaProducer *kafka.KafkaProducer
	restyClient   *resty.RestyClient
}

func NewClientServer(
	log *log.Logger,
	mqttClient *mqtt.MqttClient,
	kafkaProducer *kafka.KafkaProducer,
	restyClient *resty.RestyClient,
) *ClientServer {
	return &ClientServer{
		log:           log,
		mqttClient:    mqttClient,
		kafkaProducer: kafkaProducer,
		restyClient:   restyClient,
	}
}

func (c *ClientServer) Start(ctx context.Context) error {
	c.log.Info("启动ClientServer...")

	go func() {
		if c.mqttClient != nil {
			c.mqttClient.Start(ctx)
		}
	}()

	go func() {
		if c.kafkaProducer != nil {
			c.kafkaProducer.Start(ctx)
		}
	}()

	<-ctx.Done()
	c.log.Info("ClientServer收到停止信号")
	return nil
}

func (c *ClientServer) Stop(ctx context.Context) error {
	c.log.Info("开始停止ClientServer...")

	c.mqttClient.Stop(ctx)
	c.kafkaProducer.Stop(ctx)

	c.log.Info("ClientServer已停止")
	return nil
}
