package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysDictDataHandler struct {
	*Handler
	sysDictDataService service.SysDictDataService
}

func NewSysDictDataHandler(
	handler *Handler,
	sysDictDataService service.SysDictDataService,
) *SysDictDataHandler {
	return &SysDictDataHandler{
		Handler:            handler,
		sysDictDataService: sysDictDataService,
	}
}

// Create godoc
// @Summary      创建字典数据
// @Description  创建新的字典数据
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysDictDataCreateRequest  true  "创建字典数据请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data [post]
// @Id        system:dict:data:add
func (h *SysDictDataHandler) Create(ctx *gin.Context) {
	request := &v1.SysDictDataCreateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	err := h.sysDictDataService.Create(ctx, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新字典数据
// @Description  更新字典数据信息
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "字典数据ID"
// @Param        request  body      v1.SysDictDataUpdateRequest  true  "更新字典数据请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/{id} [put]
// @Id        system:dict:data:edit
func (h *SysDictDataHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	request := &v1.SysDictDataUpdateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}

	request.ID = cast.ToUint(id)

	err := h.sysDictDataService.Update(ctx, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除字典数据
// @Description  删除指定字典数据
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典数据ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/{id} [delete]
// @Id        system:dict:data:remove
func (h *SysDictDataHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysDictDataService.Delete(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Get godoc
// @Summary      获取字典数据详情
// @Description  根据ID获取字典数据详情
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典数据ID"
// @Success      200  {object}  v1.Response[v1.SysDictDataDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/{id} [get]
// @Id        system:dict:data:detail
func (h *SysDictDataHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	dictData, err := h.sysDictDataService.FindByID(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, dictData)
}

// List godoc
// @Summary      获取字典数据列表
// @Description  分页获取字典数据列表
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "当前页"
// @Param        size       query     int     false  "每页数量"
// @Param        dict_type  query     string  false  "字典类型"
// @Param        dict_label query     string  false  "字典标签"
// @Param        status     query     string  false  "状态"
// @Success      200        {object}  v1.Response[v1.PageResponse[v1.SysDictDataDetailResponse]]
// @Failure      400        {object}  v1.Response[any]
// @Failure      500        {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/list [get]
// @Id        system:dict:data:list
func (h *SysDictDataHandler) List(ctx *gin.Context) {
	query := &v1.SysDictDataListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.DictType = ctx.Query("dict_type")
	query.DictLabel = ctx.Query("dict_label")
	query.Status = ctx.Query("status")

	dictDataList, total, err := h.sysDictDataService.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, dictDataList))
}

// ListByType godoc
// @Summary      根据字典类型获取字典数据
// @Description  根据字典类型获取所有启用的字典数据
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        dictType  path      string  true  "字典类型"
// @Success      200       {object}  v1.Response[[]v1.SysDictDataDetailResponse]
// @Failure      400       {object}  v1.Response[any]
// @Failure      500       {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/type/{dictType} [get]
// @Id        system:dict:data:type
func (h *SysDictDataHandler) ListByType(ctx *gin.Context) {
	dictType := ctx.Param("dictType")
	if dictType == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "dictType is required"), nil)
		return
	}

	dictDataList, err := h.sysDictDataService.FindByDictType(ctx, dictType)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, dictDataList)
}

// Enable godoc
// @Summary      启用字典数据
// @Description  启用指定字典数据
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典数据ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/{id}/enable [put]
// @Id        system:dict:data:enable
func (h *SysDictDataHandler) Enable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysDictDataService.Enable(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Disable godoc
// @Summary      禁用字典数据
// @Description  禁用指定字典数据
// @Tags         字典数据
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "字典数据ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/dict/data/{id}/disable [put]
// @Id        system:dict:data:disable
func (h *SysDictDataHandler) Disable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysDictDataService.Disable(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}
