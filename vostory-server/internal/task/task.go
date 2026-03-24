package task

import (
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/jwt"
	"iot-alert-center/pkg/log"
	"iot-alert-center/pkg/sid"
)

type Task struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

func NewTask(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
) *Task {
	return &Task{
		logger: logger,
		sid:    sid,
		tm:     tm,
	}
}
