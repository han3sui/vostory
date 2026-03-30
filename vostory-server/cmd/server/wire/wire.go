//go:build wireinject
// +build wireinject

package wire

import (
	"iot-alert-center/internal/cache"
	"iot-alert-center/internal/client/kafka"
	"iot-alert-center/internal/client/mqtt"
	"iot-alert-center/internal/client/resty"
	"iot-alert-center/internal/handler"
	"iot-alert-center/internal/job"
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/server"
	"iot-alert-center/internal/service"
	"iot-alert-center/internal/worker"
	"iot-alert-center/pkg/app"
	"iot-alert-center/pkg/eventbus"
	"iot-alert-center/pkg/jwt"
	"iot-alert-center/pkg/log"
	"iot-alert-center/pkg/server/http"
	"iot-alert-center/pkg/sid"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

var repositorySet = wire.NewSet(
	repository.NewDB,
	repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewSysPostRepository,
	repository.NewSysDeptRepository,
	repository.NewSysMenuRepository,
	repository.NewSysRoleRepository,
	repository.NewSysRoleDeptRepository,
	repository.NewSysRoleMenuRepository,
	repository.NewSysUserRoleRepository,
	repository.NewSysUserPostRepository,
	repository.NewSysUserRepository,
	repository.NewSysLogininforRepository,
	repository.NewSysApiRepository,
	repository.NewSysDictTypeRepository,
	repository.NewSysDictDataRepository,
	repository.NewSysOperLogRepository,
	repository.NewVsLLMProviderRepository,
	repository.NewVsTTSProviderRepository,
	repository.NewVsPromptTemplateRepository,
	repository.NewVsWorkspaceRepository,
	repository.NewVsProjectRepository,
	repository.NewVsChapterRepository,
	repository.NewVsScriptSegmentRepository,
	repository.NewVsCharacterRepository,
	repository.NewVsLLMLogRepository,
	repository.NewVsVoiceProfileRepository,
	repository.NewVsPronunciationDictRepository,
	repository.NewVsSceneRepository,
	repository.NewVsAudioClipRepository,
	repository.NewVsVoiceEmotionRepository,
	repository.NewVsVoiceAssetRepository,
	repository.NewVsGenerationTaskRepository,
	repository.NewVsExportJobRepository,
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewSysPostService,
	service.NewSysDeptService,
	service.NewSysMenuService,
	service.NewSysRoleService,
	service.NewSysUserService,
	service.NewSysLogininforService,
	service.NewSysApiService,
	service.NewSysDictTypeService,
	service.NewSysDictDataService,
	service.NewSysOperLogService,
	service.NewVsLLMProviderService,
	service.NewVsTTSProviderService,
	service.NewVsPromptTemplateService,
	service.NewVsWorkspaceService,
	service.NewVsProjectService,
	service.NewVsChapterService,
	service.NewVsScriptSegmentService,
	service.NewVsCharacterService,
	service.NewVsFileImportService,
	service.NewVsLLMLogService,
	service.NewVsVoiceProfileService,
	service.NewVsPronunciationDictService,
	service.NewVsPreciseFillService,
	service.NewVsChapterSplitService,
	service.NewVsCharacterExtractService,
	service.NewVsVoiceEmotionService,
	service.NewVsTTSSynthesizeService,
	service.NewVsVoiceAssetService,
	service.NewVsVoiceMatchService,
	service.NewVsExportJobService,
	service.NewLicenseService,
	cache.NewUserCache,
	mqtt.NewMqttClient,
	kafka.NewKafkaProducer,
	resty.NewRestyClient,
)

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewSysPostHandler,
	handler.NewSysDeptHandler,
	handler.NewSysMenuHandler,
	handler.NewSysRoleHandler,
	handler.NewSysUserHandler,
	handler.NewSysLogininforHandler,
	handler.NewSysApiHandler,
	handler.NewSysDictTypeHandler,
	handler.NewSysDictDataHandler,
	handler.NewSysOperLogHandler,
	handler.NewVsLLMProviderHandler,
	handler.NewVsTTSProviderHandler,
	handler.NewVsPromptTemplateHandler,
	handler.NewVsWorkspaceHandler,
	handler.NewVsProjectHandler,
	handler.NewVsChapterHandler,
	handler.NewVsScriptSegmentHandler,
	handler.NewVsCharacterHandler,
	handler.NewVsFileImportHandler,
	handler.NewVsLLMLogHandler,
	handler.NewVsVoiceProfileHandler,
	handler.NewVsPronunciationDictHandler,
	handler.NewVsPreciseFillHandler,
	handler.NewVsChapterSplitHandler,
	handler.NewVsCharacterExtractHandler,
	handler.NewVsVoiceEmotionHandler,
	handler.NewVsTTSSynthesizeHandler,
	handler.NewVsVoiceAssetHandler,
	handler.NewVsCommonUploadHandler,
	handler.NewVsVoiceMatchHandler,
	handler.NewVsExportJobHandler,
	handler.NewLicenseHandler,
)

var workerSet = wire.NewSet(
	worker.NewTTSWorker,
	worker.NewLLMWorker,
)

var jobSet = wire.NewSet(
	job.NewJob,
	job.NewUserJob,
)
var serverSet = wire.NewSet(
	server.NewHTTPServer,
	server.NewJobServer,
	// server.NewClientServer,
)

// build App
func newApp(
	httpServer *http.Server,
	jobServer *server.JobServer,
	// clientServer *server.ClientServer,
) *app.App {
	return app.NewApp(
		app.WithServer(
			httpServer,
			jobServer,
			// clientServer,
		),
		app.WithName("gin-vue3-admin"),
	)
}

func NewWire(*viper.Viper, *log.Logger, *eventbus.EventBus) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		workerSet,
		jobSet,
		serverSet,
		sid.NewSid,
		jwt.NewJwt,
		newApp,
	))
}
