package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsVoiceAssetHandler struct {
	*Handler
	svc service.VsVoiceAssetService
}

func NewVsVoiceAssetHandler(handler *Handler, svc service.VsVoiceAssetService) *VsVoiceAssetHandler {
	return &VsVoiceAssetHandler{Handler: handler, svc: svc}
}

// Get godoc
// @Summary      获取声音资产详情
// @Tags         声音资产
// @Param        id   path      int  true  "声音资产ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-asset/{id} [get]
// @Id        voice-asset:detail
func (h *VsVoiceAssetHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	asset, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, asset)
}

// List godoc
// @Summary      获取声音资产列表
// @Tags         声音资产
// @Param        page   query     int     false  "当前页"
// @Param        size   query     int     false  "每页数量"
// @Param        name   query     string  false  "名称"
// @Param        gender query     string  false  "性别"
// @Param        status query     string  false  "状态"
// @Success      200    {object}  v1.Response
// @Router       /api/v1/voice-asset/list [get]
// @Id        voice-asset:list
func (h *VsVoiceAssetHandler) List(ctx *gin.Context) {
	query := &v1.VsVoiceAssetListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.Name = ctx.Query("name")
	query.Gender = ctx.Query("gender")
	query.Status = ctx.Query("status")

	assets, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, assets))
}

// Create godoc
// @Summary      创建声音资产
// @Tags         声音资产
// @Param        body  body      v1.VsVoiceAssetCreateRequest  true  "声音资产信息"
// @Success      200   {object}  v1.Response
// @Router       /api/v1/voice-asset [post]
// @Id        voice-asset:add
func (h *VsVoiceAssetHandler) Create(ctx *gin.Context) {
	var request v1.VsVoiceAssetCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.svc.Create(ctx, &request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新声音资产
// @Tags         声音资产
// @Param        id    path      int                            true  "声音资产ID"
// @Param        body  body      v1.VsVoiceAssetUpdateRequest   true  "声音资产信息"
// @Success      200   {object}  v1.Response
// @Router       /api/v1/voice-asset/{id} [put]
// @Id        voice-asset:edit
func (h *VsVoiceAssetHandler) Update(ctx *gin.Context) {
	var request v1.VsVoiceAssetUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	request.ID = cast.ToUint64(ctx.Param("id"))
	if err := h.svc.Update(ctx, &request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除声音资产
// @Tags         声音资产
// @Param        id   path      int  true  "声音资产ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-asset/{id} [delete]
// @Id        voice-asset:remove
func (h *VsVoiceAssetHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	if err := h.svc.Delete(ctx, cast.ToUint64(id)); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Enable godoc
// @Summary      启用声音资产
// @Tags         声音资产
// @Param        id   path      int  true  "声音资产ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-asset/{id}/enable [put]
// @Id        voice-asset:enable
func (h *VsVoiceAssetHandler) Enable(ctx *gin.Context) {
	if err := h.svc.Enable(ctx, cast.ToUint64(ctx.Param("id"))); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Disable godoc
// @Summary      停用声音资产
// @Tags         声音资产
// @Param        id   path      int  true  "声音资产ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-asset/{id}/disable [put]
// @Id        voice-asset:disable
func (h *VsVoiceAssetHandler) Disable(ctx *gin.Context) {
	if err := h.svc.Disable(ctx, cast.ToUint64(ctx.Param("id"))); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// GetAllEnabled godoc
// @Summary      获取所有启用的声音资产（下拉选择用）
// @Tags         声音资产
// @Success      200  {object}  v1.Response
// @Router       /api/v1/common/voice-asset/options [get]
// @Id        common:voice-asset:options
func (h *VsVoiceAssetHandler) GetAllEnabled(ctx *gin.Context) {
	options, err := h.svc.FindAllEnabled(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, options)
}
