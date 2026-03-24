package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysDictTypeHandler struct {
	*Handler
	sysDictTypeService service.SysDictTypeService
}

func NewSysDictTypeHandler(
	handler *Handler,
	sysDictTypeService service.SysDictTypeService,
) *SysDictTypeHandler {
	return &SysDictTypeHandler{
		Handler:            handler,
		sysDictTypeService: sysDictTypeService,
	}
}

// Create godoc
// @Summary      创建字典类型
// @Description  创建新的字典类型
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysDictTypeCreateRequest  true  "创建字典类型请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type [post]
// @Id        system:dict:add
func (h *SysDictTypeHandler) Create(ctx *gin.Context) {
	request := &v1.SysDictTypeCreateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	err := h.sysDictTypeService.Create(ctx, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新字典类型
// @Description  更新字典类型信息
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "字典类型ID"
// @Param        request  body      v1.SysDictTypeUpdateRequest  true  "更新字典类型请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type/{id} [put]
// @Id        system:dict:edit
func (h *SysDictTypeHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.SysDictTypeUpdateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	request.ID = cast.ToUint(id)

	err := h.sysDictTypeService.Update(ctx, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除字典类型
// @Description  删除指定字典类型
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典类型ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type/{id} [delete]
// @Id        system:dict:remove
func (h *SysDictTypeHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysDictTypeService.Delete(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Get godoc
// @Summary      获取字典类型详情
// @Description  根据ID获取字典类型详情
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典类型ID"
// @Success      200  {object}  v1.Response[v1.SysDictTypeDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type/{id} [get]
// @Id        system:dict:detail
func (h *SysDictTypeHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	dictType, err := h.sysDictTypeService.FindByID(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, dictType)
}

// List godoc
// @Summary      获取字典类型列表
// @Description  分页获取字典类型列表
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "当前页"
// @Param        size      query     int     false  "每页数量"
// @Param        dict_name query     string  false  "字典名称"
// @Param        dict_type query     string  false  "字典类型"
// @Param        status    query     string  false  "状态"
// @Success      200       {object}  v1.Response[v1.PageResponse[v1.SysDictTypeDetailResponse]]
// @Failure      400       {object}  v1.Response[any]
// @Failure      500       {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type/list [get]
// @Id        system:dict:list
func (h *SysDictTypeHandler) List(ctx *gin.Context) {
	query := &v1.SysDictTypeListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.DictName = ctx.Query("dict_name")
	query.DictType = ctx.Query("dict_type")
	query.Status = ctx.Query("status")

	dictTypes, total, err := h.sysDictTypeService.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, dictTypes))
}

// Enable godoc
// @Summary      启用字典类型
// @Description  启用指定字典类型
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典类型ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type/{id}/enable [put]
// @Id        system:dict:enable
func (h *SysDictTypeHandler) Enable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysDictTypeService.Enable(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Disable godoc
// @Summary      禁用字典类型
// @Description  禁用指定字典类型
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典类型ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/type/{id}/disable [put]
// @Id        system:dict:disable
func (h *SysDictTypeHandler) Disable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysDictTypeService.Disable(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}
