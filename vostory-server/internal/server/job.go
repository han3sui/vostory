package server

import (
	"context"
	"iot-alert-center/internal/job"
	"iot-alert-center/internal/worker"
	"iot-alert-center/pkg/log"
	"sync"

	"github.com/spf13/viper"
)

type JobServer struct {
	log       *log.Logger
	conf      *viper.Viper
	userJob   job.UserJob
	ttsWorker *worker.TTSWorker
	llmWorker *worker.LLMWorker
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
	ttsWorker *worker.TTSWorker,
	llmWorker *worker.LLMWorker,
) *JobServer {
	return &JobServer{
		log:       log,
		conf:      conf,
		userJob:   userJob,
		ttsWorker: ttsWorker,
		llmWorker: llmWorker,
	}
}

func (j *JobServer) Start(ctx context.Context) error {
	j.jobsMutex.Lock()
	defer j.jobsMutex.Unlock()

	go j.userJob.KafkaConsumer(ctx)

	j.ttsWorker.Start(ctx)
	j.llmWorker.Start(ctx)

	<-ctx.Done()
	return nil
}

func (j *JobServer) Stop(ctx context.Context) error {
	j.jobsMutex.Lock()
	defer j.jobsMutex.Unlock()

	j.ttsWorker.Stop()
	j.llmWorker.Stop()

	j.log.Info("所有Job已停止")
	return nil
}
