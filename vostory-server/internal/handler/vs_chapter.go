package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsChapterHandler struct {
	*Handler
	svc service.VsChapterService
}

func NewVsChapterHandler(handler *Handler, svc service.VsChapterService) *VsChapterHandler {
	return &VsChapterHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建章节
// @Description  在指定项目下创建新章节
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsChapterCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/chapter [post]
// @Id        chapter:add
func (h *VsChapterHandler) Create(ctx *gin.Context) {
	request := &v1.VsChapterCreateRequest{}
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
// @Summary      更新章节
// @Description  更新章节信息
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                        true  "章节ID"
// @Param        request  body      v1.VsChapterUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/chapter/{id} [put]
// @Id        chapter:edit
func (h *VsChapterHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	request := &v1.VsChapterUpdateRequest{}
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
// @Summary      删除章节
// @Description  删除指定章节
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "章节ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/chapter/{id} [delete]
// @Id        chapter:remove
func (h *VsChapterHandler) Delete(ctx *gin.Context) {
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
// @Summary      获取章节详情
// @Description  根据ID获取章节详情
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "章节ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/chapter/{id} [get]
// @Id        chapter:detail
func (h *VsChapterHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	chapter, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, chapter)
}

// List godoc
// @Summary      获取章节列表
// @Description  分页获取章节列表
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "当前页"
// @Param        size       query     int     false  "每页数量"
// @Param        project_id query     int     false  "项目ID"
// @Param        title      query     string  false  "章节标题"
// @Param        status     query     string  false  "状态"
// @Success      200        {object}  v1.Response
// @Failure      500        {object}  v1.Response
// @Router       /api/v1/chapter/list [get]
// @Id        chapter:list
func (h *VsChapterHandler) List(ctx *gin.Context) {
	query := &v1.VsChapterListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ProjectID = cast.ToUint64(ctx.Query("project_id"))
	query.Title = ctx.Query("title")
	query.Status = ctx.Query("status")

	chapters, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, chapters))
}

// GetByProject godoc
// @Summary      获取项目下所有章节
// @Description  获取指定项目的全部章节列表（不分页）
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/common/chapter/project/{project_id} [get]
// @Id        common:chapter:project
func (h *VsChapterHandler) GetByProject(ctx *gin.Context) {
	projectID := ctx.Param("project_id")
	if projectID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	chapters, err := h.svc.FindByProjectID(ctx, cast.ToUint64(projectID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, chapters)
}
