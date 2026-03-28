package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsScriptSegmentHandler struct {
	*Handler
	svc service.VsScriptSegmentService
}

func NewVsScriptSegmentHandler(handler *Handler, svc service.VsScriptSegmentService) *VsScriptSegmentHandler {
	return &VsScriptSegmentHandler{Handler: handler, svc: svc}
}

// Create godoc
// @Summary      创建脚本片段
// @Description  创建新的脚本片段
// @Tags         脚本片段管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.VsScriptSegmentCreateRequest  true  "创建请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/script-segment [post]
// @Id        script-segment:add
func (h *VsScriptSegmentHandler) Create(ctx *gin.Context) {
	request := &v1.VsScriptSegmentCreateRequest{}
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
// @Summary      更新脚本片段
// @Description  更新脚本片段信息（修正说话人、调整类型/情绪/强度等）
// @Tags         脚本片段管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                              true  "片段ID"
// @Param        request  body      v1.VsScriptSegmentUpdateRequest  true  "更新请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/script-segment/{id} [put]
// @Id        script-segment:edit
func (h *VsScriptSegmentHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	request := &v1.VsScriptSegmentUpdateRequest{}
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
// @Summary      删除脚本片段
// @Description  删除指定脚本片段
// @Tags         脚本片段管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "片段ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/script-segment/{id} [delete]
// @Id        script-segment:remove
func (h *VsScriptSegmentHandler) Delete(ctx *gin.Context) {
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
// @Summary      获取脚本片段详情
// @Description  根据ID获取脚本片段详情
// @Tags         脚本片段管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "片段ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/script-segment/{id} [get]
// @Id        script-segment:detail
func (h *VsScriptSegmentHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	segment, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, segment)
}

// List godoc
// @Summary      获取脚本片段列表
// @Description  分页获取脚本片段列表
// @Tags         脚本片段管理
// @Accept       json
// @Produce      json
// @Param        page         query     int     false  "当前页"
// @Param        size         query     int     false  "每页数量"
// @Param        chapter_id   query     int     false  "章节ID"
// @Param        scene_id     query     int     false  "场景ID"
// @Param        segment_type query     string  false  "片段类型"
// @Param        character_id query     int     false  "说话人ID"
// @Param        status       query     string  false  "状态"
// @Success      200          {object}  v1.Response
// @Failure      500          {object}  v1.Response
// @Router       /api/v1/script-segment/list [get]
// @Id        script-segment:list
func (h *VsScriptSegmentHandler) List(ctx *gin.Context) {
	query := &v1.VsScriptSegmentListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ChapterID = cast.ToUint64(ctx.Query("chapter_id"))
	query.SceneID = cast.ToUint64(ctx.Query("scene_id"))
	query.SegmentType = ctx.Query("segment_type")
	query.CharacterID = cast.ToUint64(ctx.Query("character_id"))
	query.Status = ctx.Query("status")
	query.ProjectID = cast.ToUint64(ctx.Query("project_id"))

	segments, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, segments))
}

// GetByChapter godoc
// @Summary      获取章节下所有脚本片段
// @Description  获取指定章节的全部脚本片段（不分页，按序号排列）
// @Tags         脚本片段管理
// @Accept       json
// @Produce      json
// @Param        chapter_id  path      int  true  "章节ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/common/script-segment/chapter/{chapter_id} [get]
// @Id        common:script-segment:chapter
func (h *VsScriptSegmentHandler) GetByChapter(ctx *gin.Context) {
	chapterID := ctx.Param("chapter_id")
	if chapterID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}
	segments, err := h.svc.FindByChapterID(ctx, cast.ToUint64(chapterID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, segments)
}
