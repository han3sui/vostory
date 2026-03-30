package middleware

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LicenseCheckMiddleware(licenseService service.LicenseService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodGet {
			ctx.Next()
			return
		}
		if !licenseService.IsActivated() {
			ctx.JSON(http.StatusForbidden, v1.Response{
				Code:    4031,
				Message: "系统未激活，请先完成授权激活",
				Data:    nil,
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
