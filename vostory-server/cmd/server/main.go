package main

import (
	"context"
	"flag"
	"fmt"
	"iot-alert-center/cmd/server/wire"

	"iot-alert-center/pkg/config"
	"iot-alert-center/pkg/log"
	"net/http"

	"iot-alert-center/pkg/eventbus"

	"go.uber.org/zap"

	_ "net/http/pprof"
)

// @title           Nunu Example API
// @version         1.0.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8000
// @securityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	var envConf = flag.String("conf", "config/dev.yml", "config path, eg: -conf ./config/dev.yml")
	flag.Parse()
	conf := config.NewConfig(*envConf)

	logger := log.NewLog(conf)

	// 创建事件总线
	bus := eventbus.NewEventBus()

	app, cleanup, err := wire.NewWire(conf, logger, bus)
	defer cleanup()
	if err != nil {
		panic(err)
	}

	if conf.GetString("env") == "local" {
		// pprof端口
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	// 启动服务
	logger.Info("server start", zap.String("host", fmt.Sprintf("http://%s:%d", conf.GetString("http.host"), conf.GetInt("http.port"))))
	logger.Info("docs addr", zap.String("addr", fmt.Sprintf("http://%s:%d/swagger/index.html", conf.GetString("http.host"), conf.GetInt("http.port"))))
	if err = app.Run(context.Background()); err != nil {
		panic(err)
	}

}
