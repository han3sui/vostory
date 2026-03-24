package middleware

import (
	"bytes"
	"io"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/service"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func OperLogMiddleware(operLogService service.SysOperLogService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodGet {
			ctx.Next()
			return
		}

		startTime := time.Now()

		var reqBody string
		if ctx.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			reqBody = string(bodyBytes)
			if len(reqBody) > 2000 {
				reqBody = reqBody[:2000]
			}
		}

		blw := &operLogBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw

		ctx.Next()

		costTime := time.Since(startTime).Milliseconds()

		operName, _ := ctx.Value("login_name").(string)
		deptName, _ := ctx.Value("dept_name").(string)

		status := 0
		errMsg := ""
		if ctx.Writer.Status() != http.StatusOK {
			status = 1
			errMsg = blw.body.String()
			if len(errMsg) > 2000 {
				errMsg = errMsg[:2000]
			}
		}

		jsonResult := blw.body.String()
		if len(jsonResult) > 2000 {
			jsonResult = jsonResult[:2000]
		}

		operLog := &model.SysOperLog{
			Title:         parseTitle(ctx.FullPath()),
			BusinessType:  parseBusinessType(ctx.Request.Method),
			Method:        ctx.HandlerName(),
			RequestMethod: ctx.Request.Method,
			OperatorType:  1,
			OperName:      operName,
			DeptName:      deptName,
			OperURL:       ctx.Request.URL.String(),
			OperIP:        ctx.ClientIP(),
			OperParam:     reqBody,
			JSONResult:    jsonResult,
			Status:        status,
			ErrorMsg:      errMsg,
			OperTime:      startTime,
			CostTime:      costTime,
		}

		go operLogService.Create(ctx.Copy(), operLog)
	}
}

type operLogBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w operLogBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func parseBusinessType(method string) int {
	switch method {
	case http.MethodPost:
		return 1
	case http.MethodPut:
		return 2
	case http.MethodDelete:
		return 3
	default:
		return 0
	}
}

func parseTitle(path string) string {
	if path == "" {
		return "其他"
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	if len(parts) >= 1 {
		return parts[0]
	}
	return "其他"
}
