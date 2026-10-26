package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsTTSProviderHandler struct {
	*Handler
	svc service.VsTTSProviderService
}

func NewVsTTSProviderHandler(
	handler *Handler,
	svc service.VsTTSProviderService,
) *VsTTSProviderHandler {
	return &VsTTSProviderHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建TTS提供商
// @Description  创建新的TTS提供商配置
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsTTSProviderCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/tts-provider [post]
// @Id        ai:tts-provider:add
func (h *VsTTSProviderHandler) Create(ctx *gin.Context) {
	request := &v1.VsTTSProviderCreateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	if err := h.svc.Create(ctx, request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新TTS提供商
// @Description  更新TTS提供商配置
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "提供商ID"
// @Param        request  body      v1.VsTTSProviderUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/{id} [put]
// @Id        ai:tts-provider:edit
func (h *VsTTSProviderHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.VsTTSProviderUpdateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	request.ID = cast.ToUint64(id)

	if err := h.svc.Update(ctx, request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除TTS提供商
// @Description  删除指定TTS提供商
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/{id} [delete]
// @Id        ai:tts-provider:remove
func (h *VsTTSProviderHandler) Delete(ctx *gin.Context) {
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

// Get godoc
// @Summary      获取TTS提供商详情
// @Description  根据ID获取TTS提供商详情
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/{id} [get]
// @Id        ai:tts-provider:detail
func (h *VsTTSProviderHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	provider, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, provider)
}

// List godoc
// @Summary      获取TTS提供商列表
// @Description  分页获取TTS提供商列表
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        page          query     int     false  "当前页"
// @Param        size          query     int     false  "每页数量"
// @Param        name          query     string  false  "名称"
// @Param        provider_type query     string  false  "提供商类型"
// @Param        status        query     string  false  "状态"
// @Success      200           {object}  v1.Response
// @Failure      500           {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/list [get]
// @Id        ai:tts-provider:list
func (h *VsTTSProviderHandler) List(ctx *gin.Context) {
	query := &v1.VsTTSProviderListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.Name = ctx.Query("name")
	query.ProviderType = ctx.Query("provider_type")
	query.Status = ctx.Query("status")

	providers, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, providers))
}

// Enable godoc
// @Summary      启用TTS提供商
// @Description  启用指定TTS提供商
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/{id}/enable [put]
// @Id        ai:tts-provider:enable
func (h *VsTTSProviderHandler) Enable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	if err := h.svc.Enable(ctx, cast.ToUint64(id)); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Disable godoc
// @Summary      禁用TTS提供商
// @Description  禁用指定TTS提供商
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/{id}/disable [put]
// @Id        ai:tts-provider:disable
func (h *VsTTSProviderHandler) Disable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	if err := h.svc.Disable(ctx, cast.ToUint64(id)); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// TestConnection godoc
// @Summary      测试TTS连通性
// @Description  测试TTS提供商API连通性
// @Tags         TTS提供商管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsTTSProviderTestRequest  true  "测试请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Router       /api/v1/ai/tts-provider/test [post]
// @Id        ai:tts-provider:test
func (h *VsTTSProviderHandler) TestConnection(ctx *gin.Context) {
	request := &v1.VsTTSProviderTestRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	result := h.svc.TestConnection(ctx, request)
	v1.HandleSuccess(ctx, result)
}
