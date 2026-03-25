package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsWorkspaceHandler struct {
	*Handler
	svc service.VsWorkspaceService
}

func NewVsWorkspaceHandler(
	handler *Handler,
	svc service.VsWorkspaceService,
) *VsWorkspaceHandler {
	return &VsWorkspaceHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建工作空间
// @Description  创建新的工作空间
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsWorkspaceCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/workspace [post]
// @Id        workspace:add
func (h *VsWorkspaceHandler) Create(ctx *gin.Context) {
	request := &v1.VsWorkspaceCreateRequest{}
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
// @Summary      更新工作空间
// @Description  更新工作空间信息
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "工作空间ID"
// @Param        request  body      v1.VsWorkspaceUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/workspace/{id} [put]
// @Id        workspace:edit
func (h *VsWorkspaceHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.VsWorkspaceUpdateRequest{}
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
// @Summary      删除工作空间
// @Description  删除指定工作空间
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "工作空间ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/workspace/{id} [delete]
// @Id        workspace:remove
func (h *VsWorkspaceHandler) Delete(ctx *gin.Context) {
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
// @Summary      获取工作空间详情
// @Description  根据ID获取工作空间详情
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "工作空间ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/workspace/{id} [get]
// @Id        workspace:detail
func (h *VsWorkspaceHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	workspace, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, workspace)
}

// List godoc
// @Summary      获取工作空间列表
// @Description  分页获取工作空间列表
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        page   query     int     false  "当前页"
// @Param        size   query     int     false  "每页数量"
// @Param        name   query     string  false  "空间名称"
// @Param        status query     string  false  "状态"
// @Success      200    {object}  v1.Response
// @Failure      500    {object}  v1.Response
// @Router       /api/v1/workspace/list [get]
// @Id        workspace:list
func (h *VsWorkspaceHandler) List(ctx *gin.Context) {
	query := &v1.VsWorkspaceListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.Name = ctx.Query("name")
	query.Status = ctx.Query("status")

	workspaces, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, workspaces))
}

// Enable godoc
// @Summary      启用工作空间
// @Description  启用指定工作空间
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "工作空间ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/workspace/{id}/enable [put]
// @Id        workspace:enable
func (h *VsWorkspaceHandler) Enable(ctx *gin.Context) {
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
// @Summary      禁用工作空间
// @Description  禁用指定工作空间
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "工作空间ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/workspace/{id}/disable [put]
// @Id        workspace:disable
func (h *VsWorkspaceHandler) Disable(ctx *gin.Context) {
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

// GetOptions godoc
// @Summary      获取工作空间选项列表
// @Description  获取启用的工作空间列表（下拉选择用）
// @Tags         工作空间管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/common/workspace/options [get]
// @Id        common:workspace:options
func (h *VsWorkspaceHandler) GetOptions(ctx *gin.Context) {
	options, err := h.svc.FindAllEnabled(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, options)
}
