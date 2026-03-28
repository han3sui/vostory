package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsPronunciationDictHandler struct {
	*Handler
	svc service.VsPronunciationDictService
}

func NewVsPronunciationDictHandler(handler *Handler, svc service.VsPronunciationDictService) *VsPronunciationDictHandler {
	return &VsPronunciationDictHandler{Handler: handler, svc: svc}
}

// Get godoc
// @Summary      获取发音词典条目详情
// @Description  根据ID获取发音词典条目详情
// @Tags         发音词典
// @Param        id   path      int  true  "词典条目ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/pronunciation-dict/{id} [get]
// @Id        pronunciation-dict:detail
func (h *VsPronunciationDictHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	dict, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, dict)
}

// List godoc
// @Summary      获取发音词典列表
// @Description  分页获取发音词典列表
// @Tags         发音词典
// @Param        page         query     int     false  "当前页"
// @Param        size         query     int     false  "每页数量"
// @Param        project_id   query     int     true   "项目ID"
// @Param        word         query     string  false  "原始词"
// @Success      200          {object}  v1.Response
// @Failure      500          {object}  v1.Response
// @Router       /api/v1/pronunciation-dict/list [get]
// @Id        pronunciation-dict:list
func (h *VsPronunciationDictHandler) List(ctx *gin.Context) {
	query := &v1.VsPronunciationDictListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ProjectID = cast.ToUint64(ctx.Query("project_id"))
	query.Word = ctx.Query("word")

	dicts, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, dicts))
}

// Create godoc
// @Summary      创建发音词典条目
// @Description  创建新的发音词典条目
// @Tags         发音词典
// @Accept       json
// @Produce      json
// @Param        body  body      v1.VsPronunciationDictCreateRequest  true  "词典条目信息"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/pronunciation-dict [post]
// @Id        pronunciation-dict:add
func (h *VsPronunciationDictHandler) Create(ctx *gin.Context) {
	var request v1.VsPronunciationDictCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.svc.Create(ctx, &request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新发音词典条目
// @Description  更新发音词典条目信息
// @Tags         发音词典
// @Accept       json
// @Produce      json
// @Param        id    path      int                                   true  "词典条目ID"
// @Param        body  body      v1.VsPronunciationDictUpdateRequest   true  "词典条目信息"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/pronunciation-dict/{id} [put]
// @Id        pronunciation-dict:edit
func (h *VsPronunciationDictHandler) Update(ctx *gin.Context) {
	var request v1.VsPronunciationDictUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	request.ID = cast.ToUint64(ctx.Param("id"))
	if err := h.svc.Update(ctx, &request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除发音词典条目
// @Description  删除指定发音词典条目
// @Tags         发音词典
// @Param        id   path      int  true  "词典条目ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/pronunciation-dict/{id} [delete]
// @Id        pronunciation-dict:remove
func (h *VsPronunciationDictHandler) Delete(ctx *gin.Context) {
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
