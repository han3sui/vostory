//go:build wireinject
// +build wireinject

package wire

import (
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/server"
	"iot-alert-center/internal/task"
	"iot-alert-center/pkg/app"
	"iot-alert-center/pkg/log"
	"iot-alert-center/pkg/sid"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	//repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
)

var taskSet = wire.NewSet(
	task.NewTask,
)
var serverSet = wire.NewSet(
	server.NewTaskServer,
)

// build App
func newApp(
	task *server.TaskServer,
) *app.App {
	return app.NewApp(
		app.WithServer(task),
		app.WithName("demo-task"),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		taskSet,
		serverSet,
		newApp,
		sid.NewSid,
	))
}
