package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysPostHandler struct {
	*Handler
	sysPostService service.SysPostService
}

func NewSysPostHandler(
	handler *Handler,
	sysPostService service.SysPostService,
) *SysPostHandler {
	return &SysPostHandler{
		Handler:        handler,
		sysPostService: sysPostService,
	}
}

// Create godoc
// @Summary      创建岗位
// @Description  创建新的岗位
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysPostCreateRequest  true  "创建岗位请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/post [post]
// @Id        system:post:add
// Create 创建岗位
func (h *SysPostHandler) Create(ctx *gin.Context) {
	request := &v1.SysPostCreateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	err := h.sysPostService.Create(ctx, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新岗位
// @Description  更新岗位信息
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "岗位ID"
// @Param        request  body      v1.SysPostUpdateRequest  true  "更新岗位请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/post/{id} [put]
// @Id        system:post:edit
// Update 更新岗位
func (h *SysPostHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.SysPostUpdateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	request.ID = cast.ToUint(id)

	err := h.sysPostService.Update(ctx, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除岗位
// @Description  删除指定岗位
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "岗位ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/post/{id} [delete]
// @Id        system:post:remove
// Delete 删除岗位
func (h *SysPostHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysPostService.Delete(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Get godoc
// @Summary      获取岗位详情
// @Description  根据ID获取岗位详情
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "岗位ID"
// @Success      200  {object}  v1.Response[v1.SysPostDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/post/{id} [get]
// @Id        system:post:detail
// Get 获取岗位详情
func (h *SysPostHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	post, err := h.sysPostService.FindByID(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, post)
}

// List godoc
// @Summary      获取岗位列表
// @Description  分页获取岗位列表
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "当前页"
// @Param        size      query     int     false  "每页数量"
// @Param        post_code query     string  false  "岗位编码"
// @Param        post_name query     string  false  "岗位名称"
// @Param        status    query     string  false  "状态"
// @Success      200       {object}  v1.Response[v1.PageResponse[v1.SysPostDetailResponse]]
// @Failure      400       {object}  v1.Response[any]
// @Failure      500       {object}  v1.Response[any]
// @Router       /api/v1/system/post/list [get]
// @Id        system:post:list
// List 分页获取岗位列表
func (h *SysPostHandler) List(ctx *gin.Context) {
	query := &v1.SysPostListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.PostCode = ctx.Query("post_code")
	query.PostName = ctx.Query("post_name")
	query.Status = ctx.Query("status")

	posts, total, err := h.sysPostService.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, posts))
}

// Enable godoc
// @Summary      启用岗位
// @Description  启用指定岗位
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "岗位ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/post/{id}/enable [put]
// @Id        system:post:enable
// Enable 启用岗位
func (h *SysPostHandler) Enable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysPostService.Enable(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Disable godoc
// @Summary      禁用岗位
// @Description  禁用指定岗位
// @Tags         岗位管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "岗位ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/post/{id}/disable [put]
// @Id        system:post:disable
// Disable 禁用岗位
func (h *SysPostHandler) Disable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysPostService.Disable(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}
