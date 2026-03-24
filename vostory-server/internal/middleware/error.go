package middleware

import (
	"net/http"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/pkg/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recover(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "系统未知错误"), err)
				logger.Error("[Recovery from panic]",
					zap.Any("error", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.RequestURI),
					zap.String("ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.String("token", c.GetHeader("Authorization")),
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func NotFound(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		v1.HandleError(c, http.StatusNotFound, v1.NewError(404, "NOT FOUND"), nil)
		c.Abort()
	}
}
