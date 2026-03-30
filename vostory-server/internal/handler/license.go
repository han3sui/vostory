package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LicenseHandler struct {
	*Handler
	licenseService service.LicenseService
}

func NewLicenseHandler(handler *Handler, licenseService service.LicenseService) *LicenseHandler {
	return &LicenseHandler{Handler: handler, licenseService: licenseService}
}

// GetStatus 获取授权状态
func (h *LicenseHandler) GetStatus(ctx *gin.Context) {
	resp := h.licenseService.GetStatus()
	v1.HandleSuccess(ctx, resp)
}

// ActivateOnline 在线激活
func (h *LicenseHandler) ActivateOnline(ctx *gin.Context) {
	req := &v1.LicenseActivateOnlineRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.licenseService.ActivateOnline(req.LicenseCode); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, h.licenseService.GetStatus())
}

// ActivateOffline 离线激活
func (h *LicenseHandler) ActivateOffline(ctx *gin.Context) {
	req := &v1.LicenseActivateOfflineRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.licenseService.ActivateOffline(req.LicenseFileContent); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, h.licenseService.GetStatus())
}

// Deactivate 注销授权
func (h *LicenseHandler) Deactivate(ctx *gin.Context) {
	if err := h.licenseService.Deactivate(); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}
