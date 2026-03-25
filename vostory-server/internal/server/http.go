package server

import (
	"iot-alert-center/docs"
	"iot-alert-center/internal/cache"
	"iot-alert-center/internal/handler"
	"iot-alert-center/internal/middleware"
	"iot-alert-center/internal/service"
	"iot-alert-center/pkg/eventbus"
	"iot-alert-center/pkg/jwt"
	"iot-alert-center/pkg/log"
	"iot-alert-center/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	db *gorm.DB,

	sysPostHandler *handler.SysPostHandler,
	sysDeptHandler *handler.SysDeptHandler,
	sysMenuHandler *handler.SysMenuHandler,
	sysRoleHandler *handler.SysRoleHandler,
	sysUserHandler *handler.SysUserHandler,
	eventBus *eventbus.EventBus,
	sysLogininforHandler *handler.SysLogininforHandler,
	userCache cache.UserCache,
	sysApiHandler *handler.SysApiHandler,
	sysDictTypeHandler *handler.SysDictTypeHandler,
	sysDictDataHandler *handler.SysDictDataHandler,
	sysOperLogHandler *handler.SysOperLogHandler,
	sysOperLogService service.SysOperLogService,
	vsLLMProviderHandler *handler.VsLLMProviderHandler,
	vsTTSProviderHandler *handler.VsTTSProviderHandler,
	vsPromptTemplateHandler *handler.VsPromptTemplateHandler,
	vsWorkspaceHandler *handler.VsWorkspaceHandler,

) *http.Server {
	if conf.GetString("env") == "local" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	s := http.NewServer(
		gin.New(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)

	s.NoMethod(middleware.NotFound(logger))
	s.NoRoute(middleware.NotFound(logger))

	// swagger doc
	docs.SwaggerInfo.BasePath = "/api/v1"
	s.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		//ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", conf.GetInt("app.http.port"))),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	))

	s.Use(
		middleware.CORSMiddleware(),
		middleware.Recover(logger),

		// middleware.ResponseLogMiddleware(logger),
		// middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)

	v1 := s.Group("/api/v1")
	{
		// 1. 公开路由：无需任何验证
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/user/login", sysUserHandler.Login)
			noAuthRouter.POST("/user/logout", sysUserHandler.Logout)

		}

		// 2. 认证路由：只需登录，不需要特定权限（白名单接口）
		noStrictAuthRouter := v1.Group("/").Use(middleware.TokenCacheAuthMiddleware(conf, logger, userCache))
		{
			noStrictAuthRouter.GET("/user/info", sysUserHandler.GetUserInfo)
			noStrictAuthRouter.PUT("/user/update-password", sysUserHandler.UpdateCurrentPassword)
			// 通用选项接口（用于下拉选择，如转办、指定审批人等场景）
			noStrictAuthRouter.GET("/common/user/options", sysUserHandler.GetUserOptions)
			noStrictAuthRouter.GET("/common/role/options", sysRoleHandler.GetRoleOptions)
			noStrictAuthRouter.GET("/common/dept/options", sysDeptHandler.GetDeptOptionsTree)
			noStrictAuthRouter.GET("/common/dict/data/type/:dictType", sysDictDataHandler.ListByType)
			noStrictAuthRouter.GET("/common/prompt-template/type/:type", vsPromptTemplateHandler.ListByType)
			noStrictAuthRouter.GET("/common/workspace/options", vsWorkspaceHandler.GetOptions)
		}

		// 3. 授权路由：需要登录 + 权限验证（白名单模式）
		strictAuthRouter := v1.Group("/")
		strictAuthRouter.Use(middleware.TokenCacheAuthMiddleware(conf, logger, userCache), middleware.APIPermissionMiddleware(logger, userCache), middleware.OperLogMiddleware(sysOperLogService))
		{

			systemRouter := strictAuthRouter.Group("/system")
			{
				postRouter := systemRouter.Group("/post")
				{
					postRouter.GET("/:id", sysPostHandler.Get)
					postRouter.GET("/list", sysPostHandler.List)
					postRouter.POST("", sysPostHandler.Create)
					postRouter.PUT("/:id", sysPostHandler.Update)
					postRouter.DELETE("/:id", sysPostHandler.Delete)
					postRouter.PUT("/:id/enable", sysPostHandler.Enable)
					postRouter.PUT("/:id/disable", sysPostHandler.Disable)
				}

				deptRouter := systemRouter.Group("/dept")
				{
					deptRouter.GET("/:id", sysDeptHandler.GetDept)
					deptRouter.GET("/list", sysDeptHandler.ListDepts)
					deptRouter.GET("/tree", sysDeptHandler.GetDeptTree)
					deptRouter.POST("", sysDeptHandler.CreateDept)
					deptRouter.PUT("/:id", sysDeptHandler.UpdateDept)
					deptRouter.DELETE("/:id", sysDeptHandler.DeleteDept)
					deptRouter.PUT("/:id/enable", sysDeptHandler.Enable)
					deptRouter.PUT("/:id/disable", sysDeptHandler.Disable)
				}

				menuRouter := systemRouter.Group("/menu")
				{
					menuRouter.GET("/:id", sysMenuHandler.GetMenu)
					menuRouter.GET("/list", sysMenuHandler.ListMenus)
					menuRouter.GET("/tree", sysMenuHandler.GetMenuTree)
					menuRouter.GET("/type/:type", sysMenuHandler.GetMenusByType)
					menuRouter.POST("", sysMenuHandler.CreateMenu)
					menuRouter.PUT("/:id", sysMenuHandler.UpdateMenu)
					menuRouter.DELETE("/:id", sysMenuHandler.DeleteMenu)
					menuRouter.POST("/perms/muti", sysMenuHandler.CreatePermsMuti)
				}

				roleRouter := systemRouter.Group("/role")
				{
					roleRouter.GET("/:id", sysRoleHandler.GetRole)
					roleRouter.GET("/list", sysRoleHandler.ListRoles)
					roleRouter.POST("", sysRoleHandler.CreateRole)
					roleRouter.PUT("/:id", sysRoleHandler.UpdateRole)
					roleRouter.DELETE("/:id", sysRoleHandler.DeleteRole)
					// 角色菜单关联接口
					roleRouter.GET("/:id/menus", sysRoleHandler.GetRoleMenus)
					roleRouter.PUT("/:id/menus", sysRoleHandler.UpdateRoleMenus)
					roleRouter.PUT("/:id/enable", sysRoleHandler.EnableRole)
					roleRouter.PUT("/:id/disable", sysRoleHandler.DisableRole)
				}

				userGroup := systemRouter.Group("/user")
				{
					userGroup.GET("/:id", sysUserHandler.GetUser)
					userGroup.GET("/list", sysUserHandler.ListUsers)
					userGroup.POST("", sysUserHandler.CreateUser)
					userGroup.PUT("/:id", sysUserHandler.UpdateUser)
					userGroup.DELETE("/:id", sysUserHandler.DeleteUser)
					userGroup.PUT("/:id/reset-password", sysUserHandler.ResetPassword)
					userGroup.PUT("/:id/status", sysUserHandler.ChangeStatus)
					userGroup.PUT("/:id/update-password", sysUserHandler.UpdatePassword)
					userGroup.PUT("/:id/enable", sysUserHandler.EnableUser)
					userGroup.PUT("/:id/disable", sysUserHandler.DisableUser)
					userGroup.POST("/import", sysUserHandler.ImportUsers)
					userGroup.GET("/import/template", sysUserHandler.DownloadImportTemplate)
				}

				logininforRouter := systemRouter.Group("/logininfor")
				{
					logininforRouter.GET("/:id", sysLogininforHandler.Get)
					logininforRouter.GET("/list", sysLogininforHandler.List)
				}

			sysApiRouter := systemRouter.Group("/api")
			{
				sysApiRouter.GET("/list", sysApiHandler.ListSysApi)
				sysApiRouter.GET("/tag/list", sysApiHandler.ListTag)
			}

			dictTypeRouter := systemRouter.Group("/dict/type")
			{
				dictTypeRouter.GET("/:id", sysDictTypeHandler.Get)
				dictTypeRouter.GET("/list", sysDictTypeHandler.List)
				dictTypeRouter.POST("", sysDictTypeHandler.Create)
				dictTypeRouter.PUT("/:id", sysDictTypeHandler.Update)
				dictTypeRouter.DELETE("/:id", sysDictTypeHandler.Delete)
				dictTypeRouter.PUT("/:id/enable", sysDictTypeHandler.Enable)
				dictTypeRouter.PUT("/:id/disable", sysDictTypeHandler.Disable)
			}

			dictDataRouter := systemRouter.Group("/dict/data")
			{
				dictDataRouter.GET("/:id", sysDictDataHandler.Get)
				dictDataRouter.GET("/list", sysDictDataHandler.List)
				dictDataRouter.GET("/type/:dictType", sysDictDataHandler.ListByType)
				dictDataRouter.POST("", sysDictDataHandler.Create)
				dictDataRouter.PUT("/:id", sysDictDataHandler.Update)
				dictDataRouter.DELETE("/:id", sysDictDataHandler.Delete)
				dictDataRouter.PUT("/:id/enable", sysDictDataHandler.Enable)
				dictDataRouter.PUT("/:id/disable", sysDictDataHandler.Disable)
			}

			operlogRouter := systemRouter.Group("/operlog")
			{
				operlogRouter.GET("/:id", sysOperLogHandler.Get)
				operlogRouter.GET("/list", sysOperLogHandler.List)
				operlogRouter.DELETE("/:id", sysOperLogHandler.Delete)
				operlogRouter.DELETE("/clean", sysOperLogHandler.Clean)
			}

		}

			workspaceRouter := strictAuthRouter.Group("/workspace")
			{
				workspaceRouter.GET("/:id", vsWorkspaceHandler.Get)
				workspaceRouter.GET("/list", vsWorkspaceHandler.List)
				workspaceRouter.POST("", vsWorkspaceHandler.Create)
				workspaceRouter.PUT("/:id", vsWorkspaceHandler.Update)
				workspaceRouter.DELETE("/:id", vsWorkspaceHandler.Delete)
				workspaceRouter.PUT("/:id/enable", vsWorkspaceHandler.Enable)
				workspaceRouter.PUT("/:id/disable", vsWorkspaceHandler.Disable)
			}

			aiRouter := strictAuthRouter.Group("/ai")
			{
				llmProviderRouter := aiRouter.Group("/llm-provider")
				{
					llmProviderRouter.GET("/:id", vsLLMProviderHandler.Get)
					llmProviderRouter.GET("/list", vsLLMProviderHandler.List)
					llmProviderRouter.POST("", vsLLMProviderHandler.Create)
					llmProviderRouter.PUT("/:id", vsLLMProviderHandler.Update)
					llmProviderRouter.DELETE("/:id", vsLLMProviderHandler.Delete)
					llmProviderRouter.PUT("/:id/enable", vsLLMProviderHandler.Enable)
					llmProviderRouter.PUT("/:id/disable", vsLLMProviderHandler.Disable)
					llmProviderRouter.POST("/test", vsLLMProviderHandler.TestConnection)
				}

				ttsProviderRouter := aiRouter.Group("/tts-provider")
				{
					ttsProviderRouter.GET("/:id", vsTTSProviderHandler.Get)
					ttsProviderRouter.GET("/list", vsTTSProviderHandler.List)
					ttsProviderRouter.POST("", vsTTSProviderHandler.Create)
					ttsProviderRouter.PUT("/:id", vsTTSProviderHandler.Update)
					ttsProviderRouter.DELETE("/:id", vsTTSProviderHandler.Delete)
					ttsProviderRouter.PUT("/:id/enable", vsTTSProviderHandler.Enable)
					ttsProviderRouter.PUT("/:id/disable", vsTTSProviderHandler.Disable)
					ttsProviderRouter.POST("/test", vsTTSProviderHandler.TestConnection)
				}

				promptTemplateRouter := aiRouter.Group("/prompt-template")
				{
					promptTemplateRouter.GET("/:id", vsPromptTemplateHandler.Get)
					promptTemplateRouter.GET("/list", vsPromptTemplateHandler.List)
					promptTemplateRouter.POST("", vsPromptTemplateHandler.Create)
					promptTemplateRouter.PUT("/:id", vsPromptTemplateHandler.Update)
					promptTemplateRouter.DELETE("/:id", vsPromptTemplateHandler.Delete)
					promptTemplateRouter.PUT("/:id/enable", vsPromptTemplateHandler.Enable)
					promptTemplateRouter.PUT("/:id/disable", vsPromptTemplateHandler.Disable)
				}
			}
		}

	}

	return s
}
