package job

import (
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/jwt"
	"iot-alert-center/pkg/log"
	"iot-alert-center/pkg/sid"

	"github.com/spf13/viper"
)

type Job struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
	conf   *viper.Viper
}

func NewJob(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
	conf *viper.Viper,
) *Job {
	return &Job{
		logger: logger,
		sid:    sid,
		tm:     tm,
		conf:   conf,
	}
}
