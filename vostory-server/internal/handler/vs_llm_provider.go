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
// @Summary      创建LLM提供商
// @Description  创建新的LLM提供商配置
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMProviderCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/llm-provider [post]
// @Id        ai:llm-provider:add
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
// @Summary      更新LLM提供商
// @Description  更新LLM提供商配置
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "提供商ID"
// @Param        request  body      v1.VsLLMProviderUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/{id} [put]
// @Id        ai:llm-provider:edit
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
// @Summary      删除LLM提供商
// @Description  删除指定LLM提供商
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/{id} [delete]
// @Id        ai:llm-provider:remove
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

// Get godoc
// @Summary      获取LLM提供商详情
// @Description  根据ID获取LLM提供商详情
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/{id} [get]
// @Id        ai:llm-provider:detail
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
// @Summary      获取LLM提供商列表
// @Description  分页获取LLM提供商列表
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        page          query     int     false  "当前页"
// @Param        size          query     int     false  "每页数量"
// @Param        name          query     string  false  "名称"
// @Param        provider_type query     string  false  "提供商类型"
// @Param        status        query     string  false  "状态"
// @Success      200           {object}  v1.Response
// @Failure      500           {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/list [get]
// @Id        ai:llm-provider:list
func (h *VsLLMProviderHandler) List(ctx *gin.Context) {
	query := &v1.VsLLMProviderListQuery{}
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
// @Summary      启用LLM提供商
// @Description  启用指定LLM提供商
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/{id}/enable [put]
// @Id        ai:llm-provider:enable
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
// @Summary      禁用LLM提供商
// @Description  禁用指定LLM提供商
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "提供商ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/{id}/disable [put]
// @Id        ai:llm-provider:disable
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
// @Summary      测试LLM连通性
// @Description  测试LLM提供商API连通性
// @Tags         LLM提供商管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMProviderTestRequest  true  "测试请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Router       /api/v1/ai/llm-provider/test [post]
// @Id        ai:llm-provider:test
func (h *VsLLMProviderHandler) TestConnection(ctx *gin.Context) {
	request := &v1.VsLLMProviderTestRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	result := h.svc.TestConnection(ctx, request)
	v1.HandleSuccess(ctx, result)
}
