package v1

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BasePageQuery struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total,omitempty"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func HandleSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = []interface{}{}
	} else {
		// 检查是否为 nil 切片
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			data = []interface{}{}
		} else if v.Kind() == reflect.Slice && v.IsNil() {
			data = []interface{}{}
		}
	}
	ctx.JSON(http.StatusOK, data)
}

func HandleError(ctx *gin.Context, httpCode int, err error, data interface{}) {
	if data == nil {
		data = map[string]string{}
	}

	if validationErrorData, ok := data.(validator.ValidationErrors); ok {
		var errMsgs []string
		for _, err := range validationErrorData {
			errMsgs = append(errMsgs, err.Error())
		}
		data = errMsgs
	}

	// 如果data是error类型，则将data转换为string
	if err, ok := data.(error); ok {
		data = err.Error()
	}

	// Handle validation errors specifically
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		resp := Response{
			Code:    400,
			Message: "Validation failed: " + validationErrors.Error(),
			Data:    data,
		}
		ctx.JSON(httpCode, resp)
		return
	}

	// Handle other errors
	resp := Response{Code: errorCodeMap[err], Message: err.Error(), Data: data}
	if _, ok := errorCodeMap[err]; !ok {
		resp = Response{Code: 500, Message: "unknown error", Data: data}
	}
	ctx.JSON(httpCode, resp)
}

type Error struct {
	Code    int
	Message string
}

var errorCodeMap = map[error]int{}

func NewError(code int, msg string) error {
	err := errors.New(msg)
	errorCodeMap[err] = code
	return err
}
func (e Error) Error() string {
	return e.Message
}

// PageResponse 通用分页响应结构
type PageResponse struct {
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

// NewPageResponse 创建分页响应
func NewPageResponse(page, size int, total int64, data interface{}) PageResponse {
	// 处理nil数据
	if data == nil {
		data = []interface{}{}
	} else {
		// 处理空切片
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Slice && v.Len() == 0 {
			data = []interface{}{}
		}
	}
	return PageResponse{
		Page:  page,
		Size:  size,
		Total: total,
		Data:  data,
	}
}
