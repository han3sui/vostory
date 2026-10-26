package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SysOSSConfigHandler struct {
	*Handler
	ossConfigService service.SysOSSConfigService
}

func NewSysOSSConfigHandler(handler *Handler, ossConfigService service.SysOSSConfigService) *SysOSSConfigHandler {
	return &SysOSSConfigHandler{Handler: handler, ossConfigService: ossConfigService}
}

// GetConfig godoc
// @Summary      获取对象存储配置
// @Tags         对象存储配置
// @Produce      json
// @Success      200 {object} v1.Response
// @Router       /api/v1/system/oss-config [get]
// @Id           system:oss-config:view
func (h *SysOSSConfigHandler) GetConfig(ctx *gin.Context) {
	resp, err := h.ossConfigService.GetConfig(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// UpdateConfig godoc
// @Summary      保存对象存储配置
// @Tags         对象存储配置
// @Accept       json
// @Produce      json
// @Param        request body v1.OSSConfigUpdateRequest true "配置"
// @Success      200 {object} v1.Response
// @Router       /api/v1/system/oss-config [put]
// @Id           system:oss-config:update
func (h *SysOSSConfigHandler) UpdateConfig(ctx *gin.Context) {
	req := new(v1.OSSConfigUpdateRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.ossConfigService.UpdateConfig(ctx, req); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// TestConnection godoc
// @Summary      测试对象存储连接
// @Tags         对象存储配置
// @Accept       json
// @Produce      json
// @Param        request body v1.OSSConfigUpdateRequest true "配置"
// @Success      200 {object} v1.Response
// @Router       /api/v1/system/oss-config/test [post]
// @Id           system:oss-config:test
func (h *SysOSSConfigHandler) TestConnection(ctx *gin.Context) {
	req := new(v1.OSSConfigUpdateRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.ossConfigService.TestConnection(ctx, req); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, "连接失败: "+err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}
