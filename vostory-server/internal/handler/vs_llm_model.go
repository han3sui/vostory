package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsLLMModelHandler struct {
	*Handler
	svc service.VsLLMModelService
}

func NewVsLLMModelHandler(
	handler *Handler,
	svc service.VsLLMModelService,
) *VsLLMModelHandler {
	return &VsLLMModelHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建LLM模型
// @Description  创建新的LLM模型
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMModelCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/model [post]
// @Id        ai:model:add
func (h *VsLLMModelHandler) Create(ctx *gin.Context) {
	request := &v1.VsLLMModelCreateRequest{}
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
// @Summary      更新LLM模型
// @Description  更新LLM模型
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "模型ID"
// @Param        request  body      v1.VsLLMModelUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/model/{id} [put]
// @Id        ai:model:edit
func (h *VsLLMModelHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.VsLLMModelUpdateRequest{}
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
// @Summary      删除LLM模型
// @Description  删除指定LLM模型
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模型ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/model/{id} [delete]
// @Id        ai:model:remove
func (h *VsLLMModelHandler) Delete(ctx *gin.Context) {
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
// @Summary      批量删除LLM模型
// @Description  批量删除LLM模型
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsLLMModelDeleteRequest  true  "删除请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/model [delete]
// @Id        ai:model:remove
func (h *VsLLMModelHandler) DeleteByIDs(ctx *gin.Context) {
	req := new(v1.VsLLMModelDeleteRequest)
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
// @Summary      获取LLM模型详情
// @Description  根据ID获取LLM模型详情
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模型ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/model/{id} [get]
// @Id        ai:model:view
func (h *VsLLMModelHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	m, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, m)
}

// List godoc
// @Summary      获取LLM模型列表
// @Description  分页获取LLM模型列表
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        page        query     int     false  "当前页"
// @Param        size        query     int     false  "每页数量"
// @Param        provider_id query     int     false  "厂家ID"
// @Param        model_name  query     string  false  "模型标识"
// @Param        status      query     string  false  "状态"
// @Success      200         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/ai/model/list [get]
// @Id        ai:model:list
func (h *VsLLMModelHandler) List(ctx *gin.Context) {
	query := &v1.VsLLMModelListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ProviderID = cast.ToUint64(ctx.Query("provider_id"))
	query.ModelName = ctx.Query("model_name")
	query.Status = ctx.Query("status")

	models, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, models))
}

// Enable godoc
// @Summary      启用LLM模型
// @Description  启用指定LLM模型
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模型ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/model/{id}/enable [put]
// @Id        ai:model:enable
func (h *VsLLMModelHandler) Enable(ctx *gin.Context) {
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
// @Summary      禁用LLM模型
// @Description  禁用指定LLM模型
// @Tags         LLM模型管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模型ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/model/{id}/disable [put]
// @Id        ai:model:disable
func (h *VsLLMModelHandler) Disable(ctx *gin.Context) {
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
