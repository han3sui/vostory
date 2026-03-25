package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsCharacterHandler struct {
	*Handler
	svc service.VsCharacterService
}

func NewVsCharacterHandler(handler *Handler, svc service.VsCharacterService) *VsCharacterHandler {
	return &VsCharacterHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建角色
// @Description  在指定项目下创建角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsCharacterCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character [post]
// @Id        character:add
func (h *VsCharacterHandler) Create(ctx *gin.Context) {
	request := &v1.VsCharacterCreateRequest{}
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
// @Summary      更新角色
// @Description  更新角色信息（编辑别名、绑定声音配置等）
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "角色ID"
// @Param        request  body      v1.VsCharacterUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character/{id} [put]
// @Id        character:edit
func (h *VsCharacterHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	request := &v1.VsCharacterUpdateRequest{}
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
// @Summary      删除角色
// @Description  删除指定角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/character/{id} [delete]
// @Id        character:remove
func (h *VsCharacterHandler) Delete(ctx *gin.Context) {
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
// @Summary      获取角色详情
// @Description  根据ID获取角色详情
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/character/{id} [get]
// @Id        character:detail
func (h *VsCharacterHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	character, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, character)
}

// List godoc
// @Summary      获取角色列表
// @Description  分页获取角色列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "当前页"
// @Param        size       query     int     false  "每页数量"
// @Param        project_id query     int     false  "项目ID"
// @Param        name       query     string  false  "角色名称"
// @Param        gender     query     string  false  "性别"
// @Param        level      query     string  false  "角色层级"
// @Param        status     query     string  false  "状态"
// @Success      200        {object}  v1.Response
// @Failure      500        {object}  v1.Response
// @Router       /api/v1/character/list [get]
// @Id        character:list
func (h *VsCharacterHandler) List(ctx *gin.Context) {
	query := &v1.VsCharacterListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ProjectID = cast.ToUint64(ctx.Query("project_id"))
	query.Name = ctx.Query("name")
	query.Gender = ctx.Query("gender")
	query.Level = ctx.Query("level")
	query.Status = ctx.Query("status")

	characters, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, characters))
}

// Enable godoc
// @Summary      启用角色
// @Description  启用指定角色
// @Tags         角色管理
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/character/{id}/enable [put]
// @Id        character:enable
func (h *VsCharacterHandler) Enable(ctx *gin.Context) {
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
// @Summary      禁用角色
// @Description  禁用指定角色
// @Tags         角色管理
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/character/{id}/disable [put]
// @Id        character:disable
func (h *VsCharacterHandler) Disable(ctx *gin.Context) {
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

// GetByProject godoc
// @Summary      获取项目下的角色选项
// @Description  获取指定项目下启用的角色列表（下拉选择用）
// @Tags         角色管理
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/common/character/project/{project_id} [get]
// @Id        common:character:project
func (h *VsCharacterHandler) GetByProject(ctx *gin.Context) {
	projectID := ctx.Param("project_id")
	if projectID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	characters, err := h.svc.FindByProjectID(ctx, cast.ToUint64(projectID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, characters)
}
