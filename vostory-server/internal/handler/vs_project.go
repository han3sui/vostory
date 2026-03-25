package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsProjectHandler struct {
	*Handler
	svc service.VsProjectService
}

func NewVsProjectHandler(
	handler *Handler,
	svc service.VsProjectService,
) *VsProjectHandler {
	return &VsProjectHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建项目
// @Description  在指定工作空间下创建新项目
// @Tags         项目管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsProjectCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/project [post]
// @Id        project:add
func (h *VsProjectHandler) Create(ctx *gin.Context) {
	request := &v1.VsProjectCreateRequest{}
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
// @Summary      更新项目
// @Description  更新项目信息
// @Tags         项目管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                        true  "项目ID"
// @Param        request  body      v1.VsProjectUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/project/{id} [put]
// @Id        project:edit
func (h *VsProjectHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.VsProjectUpdateRequest{}
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
// @Summary      删除项目
// @Description  删除指定项目
// @Tags         项目管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "项目ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/project/{id} [delete]
// @Id        project:remove
func (h *VsProjectHandler) Delete(ctx *gin.Context) {
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
// @Summary      获取项目详情
// @Description  根据ID获取项目详情
// @Tags         项目管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "项目ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/project/{id} [get]
// @Id        project:detail
func (h *VsProjectHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	project, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, project)
}

// List godoc
// @Summary      获取项目列表
// @Description  分页获取项目列表
// @Tags         项目管理
// @Accept       json
// @Produce      json
// @Param        page         query     int     false  "当前页"
// @Param        size         query     int     false  "每页数量"
// @Param        workspace_id query     int     false  "工作空间ID"
// @Param        name         query     string  false  "项目名称"
// @Param        status       query     string  false  "项目状态"
// @Param        source_type  query     string  false  "导入来源"
// @Success      200          {object}  v1.Response
// @Failure      500          {object}  v1.Response
// @Router       /api/v1/project/list [get]
// @Id        project:list
func (h *VsProjectHandler) List(ctx *gin.Context) {
	query := &v1.VsProjectListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.WorkspaceID = cast.ToUint64(ctx.Query("workspace_id"))
	query.Name = ctx.Query("name")
	query.Status = ctx.Query("status")
	query.SourceType = ctx.Query("source_type")

	projects, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, projects))
}

// GetByWorkspace godoc
// @Summary      获取工作空间下的项目选项
// @Description  获取指定工作空间下的项目列表（下拉选择用）
// @Tags         项目管理
// @Accept       json
// @Produce      json
// @Param        workspace_id  path      int  true  "工作空间ID"
// @Success      200           {object}  v1.Response
// @Failure      400           {object}  v1.Response
// @Failure      500           {object}  v1.Response
// @Router       /api/v1/common/project/workspace/{workspace_id} [get]
// @Id        common:project:workspace
func (h *VsProjectHandler) GetByWorkspace(ctx *gin.Context) {
	wsID := ctx.Param("workspace_id")
	if wsID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "workspace_id is required"), nil)
		return
	}

	projects, err := h.svc.FindByWorkspaceID(ctx, cast.ToUint64(wsID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, projects)
}
