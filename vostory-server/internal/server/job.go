package server

import (
	"context"
	"iot-alert-center/internal/job"
	"iot-alert-center/internal/service"
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
	vdWorker  *worker.VoiceDesignWorker
	importSvc service.VsFileImportService
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
	vdWorker *worker.VoiceDesignWorker,
	importSvc service.VsFileImportService,
) *JobServer {
	return &JobServer{
		log:       log,
		conf:      conf,
		userJob:   userJob,
		ttsWorker: ttsWorker,
		llmWorker: llmWorker,
		vdWorker:  vdWorker,
		importSvc: importSvc,
	}
}

func (j *JobServer) Start(ctx context.Context) error {
	j.jobsMutex.Lock()
	defer j.jobsMutex.Unlock()

	if !j.workerEnabled() {
		j.log.Info("worker.enabled=false，跳过 TTS / LLM / 音色设计 / 解析恢复")
		<-ctx.Done()
		return nil
	}

	go j.userJob.KafkaConsumer(ctx)

	j.ttsWorker.Start(ctx)
	j.llmWorker.Start(ctx)
	j.vdWorker.Start(ctx)
	go j.importSvc.RecoverParsing(ctx)

	<-ctx.Done()
	return nil
}

func (j *JobServer) Stop(ctx context.Context) error {
	j.jobsMutex.Lock()
	defer j.jobsMutex.Unlock()

	if !j.workerEnabled() {
		return nil
	}

	j.ttsWorker.Stop()
	j.llmWorker.Stop()
	j.vdWorker.Stop()

	j.log.Info("所有Job已停止")
	return nil
}

// workerEnabled 未配置时默认开启，避免漏配导致线上 worker 静默停掉
func (j *JobServer) workerEnabled() bool {
	if !j.conf.IsSet("worker.enabled") {
		return true
	}
	return j.conf.GetBool("worker.enabled")
}
