package service

import (
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/eventbus"
	"iot-alert-center/pkg/jwt"
	"iot-alert-center/pkg/log"
	"iot-alert-center/pkg/sid"
)

type Service struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
	bus    *eventbus.EventBus
}

func NewService(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
	jwt *jwt.JWT,
	bus *eventbus.EventBus,
) *Service {
	return &Service{
		logger: logger,
		sid:    sid,
		jwt:    jwt,
		tm:     tm,
		bus:    bus,
	}
}
