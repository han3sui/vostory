package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsPromptTemplateHandler struct {
	*Handler
	svc service.VsPromptTemplateService
}

func NewVsPromptTemplateHandler(
	handler *Handler,
	svc service.VsPromptTemplateService,
) *VsPromptTemplateHandler {
	return &VsPromptTemplateHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建Prompt模板
// @Description  创建新的Prompt模板
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsPromptTemplateCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/prompt-template [post]
// @Id        ai:prompt-template:add
func (h *VsPromptTemplateHandler) Create(ctx *gin.Context) {
	request := &v1.VsPromptTemplateCreateRequest{}
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
// @Summary      更新Prompt模板
// @Description  更新Prompt模板
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                               true  "模板ID"
// @Param        request  body      v1.VsPromptTemplateUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/ai/prompt-template/{id} [put]
// @Id        ai:prompt-template:edit
func (h *VsPromptTemplateHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.VsPromptTemplateUpdateRequest{}
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
// @Summary      删除Prompt模板
// @Description  删除指定Prompt模板（系统内置模板不允许删除）
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模板ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/prompt-template/{id} [delete]
// @Id        ai:prompt-template:remove
func (h *VsPromptTemplateHandler) Delete(ctx *gin.Context) {
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
// @Summary      获取Prompt模板详情
// @Description  根据ID获取Prompt模板详情
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模板ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/prompt-template/{id} [get]
// @Id        ai:prompt-template:detail
func (h *VsPromptTemplateHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	template, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, template)
}

// List godoc
// @Summary      获取Prompt模板列表
// @Description  分页获取Prompt模板列表
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        page          query     int     false  "当前页"
// @Param        size          query     int     false  "每页数量"
// @Param        name          query     string  false  "模板名称"
// @Param        template_type query     string  false  "模板类型"
// @Param        is_system     query     string  false  "是否系统内置"
// @Param        status        query     string  false  "状态"
// @Success      200           {object}  v1.Response
// @Failure      500           {object}  v1.Response
// @Router       /api/v1/ai/prompt-template/list [get]
// @Id        ai:prompt-template:list
func (h *VsPromptTemplateHandler) List(ctx *gin.Context) {
	query := &v1.VsPromptTemplateListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.Name = ctx.Query("name")
	query.TemplateType = ctx.Query("template_type")
	query.IsSystem = ctx.Query("is_system")
	query.Status = ctx.Query("status")

	templates, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, templates))
}

// ListByType godoc
// @Summary      按类型获取Prompt模板选项
// @Description  按模板类型获取启用的Prompt模板列表（下拉选择用）
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        type  path      string  true  "模板类型"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/common/prompt-template/type/{type} [get]
// @Id        common:prompt-template:type
func (h *VsPromptTemplateHandler) ListByType(ctx *gin.Context) {
	templateType := ctx.Param("type")
	if templateType == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "type is required"), nil)
		return
	}

	templates, err := h.svc.FindByType(ctx, templateType)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, templates)
}

// Enable godoc
// @Summary      启用Prompt模板
// @Description  启用指定Prompt模板
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模板ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/prompt-template/{id}/enable [put]
// @Id        ai:prompt-template:enable
func (h *VsPromptTemplateHandler) Enable(ctx *gin.Context) {
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
// @Summary      禁用Prompt模板
// @Description  禁用指定Prompt模板
// @Tags         Prompt模板管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "模板ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/prompt-template/{id}/disable [put]
// @Id        ai:prompt-template:disable
func (h *VsPromptTemplateHandler) Disable(ctx *gin.Context) {
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
