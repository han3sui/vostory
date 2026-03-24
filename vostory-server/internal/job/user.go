package job

import (
	"context"
	"time"
)

type UserJob interface {
	KafkaConsumer(ctx context.Context) error
}

func NewUserJob(
	job *Job,

) UserJob {
	return &userJob{
		Job: job,
	}
}

type userJob struct {
	*Job
}

func (t userJob) KafkaConsumer(ctx context.Context) error {
	// do something
	for {
		// t.logger.Info("KafkaConsumer")
		time.Sleep(time.Second * 5)
	}
}
