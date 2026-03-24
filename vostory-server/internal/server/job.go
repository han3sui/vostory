package server

import (
	"context"
	"iot-alert-center/internal/job"
	"iot-alert-center/pkg/log"
	"sync"

	"github.com/spf13/viper"
)

type JobServer struct {
	log       *log.Logger
	conf      *viper.Viper
	userJob   job.UserJob
	jobsMutex sync.Mutex
}

type Job interface {
	Start(ctx context.Context) error
	Stop() error
}

func NewJobServer(
	log *log.Logger,
	conf *viper.Viper,
	userJob job.UserJob,
) *JobServer {
	return &JobServer{
		log:     log,
		conf:    conf,
		userJob: userJob,
	}
}

func (j *JobServer) Start(ctx context.Context) error {
	j.jobsMutex.Lock()
	defer j.jobsMutex.Unlock()

	go j.userJob.KafkaConsumer(ctx)

	// 等待上下文取消
	<-ctx.Done()
	return nil
}

func (j *JobServer) Stop(ctx context.Context) error {
	j.jobsMutex.Lock()
	defer j.jobsMutex.Unlock()

	j.log.Info("所有Job已停止")
	return nil
}
