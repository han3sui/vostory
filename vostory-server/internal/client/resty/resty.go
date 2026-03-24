package resty

import (
	"iot-alert-center/pkg/log"
	"time"

	"github.com/spf13/viper"
	"resty.dev/v3"
)

type RestyClient struct {
	log         *log.Logger
	conf        *viper.Viper
	restyClient *resty.Client
}

func NewRestyClient(
	log *log.Logger,
	conf *viper.Viper,
) *RestyClient {
	return &RestyClient{
		log:         log,
		conf:        conf,
		restyClient: resty.New().SetRetryCount(3).SetRetryWaitTime(10 * time.Second).SetRetryMaxWaitTime(30 * time.Second),
	}
}
