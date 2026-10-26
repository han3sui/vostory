package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsLLMProviderHandler struct {
	*Handler
	svc service.VsLLMProviderService
}

func NewVsLLMProviderHandler(
	handler *Handler,
	svc service.VsLLMProviderService,
) *VsLLMProviderHandler {
	return &VsLLMProviderHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建LLM厂家
// @Description  创建新的LLM厂家配置
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMProviderCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/provider [post]
// @Id        ai:provider:add
func (h *VsLLMProviderHandler) Create(ctx *gin.Context) {
	request := &v1.VsLLMProviderCreateRequest{}
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
// @Summary      更新LLM厂家
// @Description  更新LLM厂家配置
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "厂家ID"
// @Param        request  body      v1.VsLLMProviderUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/provider/{id} [put]
// @Id        ai:provider:edit
func (h *VsLLMProviderHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.VsLLMProviderUpdateRequest{}
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
// @Summary      删除LLM厂家
// @Description  删除指定LLM厂家
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "厂家ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/provider/{id} [delete]
// @Id        ai:provider:remove
func (h *VsLLMProviderHandler) Delete(ctx *gin.Context) {
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

// DeleteByIDs godoc
// @Summary      批量删除LLM厂家
// @Description  批量删除LLM厂家
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMProviderDeleteRequest  true  "删除请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/provider [delete]
// @Id        ai:provider:remove
func (h *VsLLMProviderHandler) DeleteByIDs(ctx *gin.Context) {
	req := new(v1.VsLLMProviderDeleteRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.svc.DeleteByIDs(ctx, req.IDs); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, err, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Get godoc
// @Summary      获取LLM厂家详情
// @Description  根据ID获取LLM厂家详情
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "厂家ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/provider/{id} [get]
// @Id        ai:provider:view
func (h *VsLLMProviderHandler) Get(ctx *gin.Context) {
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
// @Summary      获取LLM厂家列表
// @Description  分页获取LLM厂家列表
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        page   query     int     false  "当前页"
// @Param        size   query     int     false  "每页数量"
// @Param        name   query     string  false  "名称"
// @Param        status query     string  false  "状态"
// @Success      200    {object}  v1.Response
// @Failure      500    {object}  v1.Response
// @Router       /api/v1/ai/provider/list [get]
// @Id        ai:provider:list
func (h *VsLLMProviderHandler) List(ctx *gin.Context) {
	query := &v1.VsLLMProviderListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.Name = ctx.Query("name")
	query.Status = ctx.Query("status")

	providers, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, providers))
}

// Enable godoc
// @Summary      启用LLM厂家
// @Description  启用指定LLM厂家
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "厂家ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/provider/{id}/enable [put]
// @Id        ai:provider:enable
func (h *VsLLMProviderHandler) Enable(ctx *gin.Context) {
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
// @Summary      禁用LLM厂家
// @Description  禁用指定LLM厂家
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "厂家ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/provider/{id}/disable [put]
// @Id        ai:provider:disable
func (h *VsLLMProviderHandler) Disable(ctx *gin.Context) {
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
// @Summary      测试LLM厂家连通性
// @Description  测试LLM厂家API连通性
// @Tags         LLM厂家配置
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMProviderTestRequest  true  "测试请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Router       /api/v1/ai/provider/test [post]
// @Id        ai:provider:test
func (h *VsLLMProviderHandler) TestConnection(ctx *gin.Context) {
	request := &v1.VsLLMProviderTestRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	result := h.svc.TestConnection(ctx, request)
	v1.HandleSuccess(ctx, result)
}
