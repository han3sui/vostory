package main

import (
	"context"
	"flag"
	"iot-alert-center/cmd/task/wire"
	"iot-alert-center/pkg/config"
	"iot-alert-center/pkg/log"
)

func main() {
	var envConf = flag.String("conf", "config/dev.yml", "config path, eg: -conf ./config/dev.yml")
	flag.Parse()
	conf := config.NewConfig(*envConf)

	logger := log.NewLog(conf)
	logger.Info("start task")
	app, cleanup, err := wire.NewWire(conf, logger)
	defer cleanup()
	if err != nil {
		panic(err)
	}
	if err = app.Run(context.Background()); err != nil {
		panic(err)
	}

}
