package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsChapterSplitHandler struct {
	*Handler
	svc service.VsChapterSplitService
}

func NewVsChapterSplitHandler(handler *Handler, svc service.VsChapterSplitService) *VsChapterSplitHandler {
	return &VsChapterSplitHandler{Handler: handler, svc: svc}
}

// Split godoc
// @Summary      LLM 智能切割章节
// @Description  调用 LLM 将章节内容切割为场景和脚本片段
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        chapter_id  path      int  true  "章节ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/chapter/{chapter_id}/split [post]
// @Id        chapter:split
func (h *VsChapterSplitHandler) Split(ctx *gin.Context) {
	chapterID := ctx.Param("chapter_id")
	if chapterID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}
	result, err := h.svc.SplitChapter(ctx, cast.ToUint64(chapterID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// BatchSplit godoc
// @Summary      批量 LLM 智能切割
// @Description  将多个章节加入切割队列，后台异步依次执行
// @Tags         章节管理
// @Accept       json
// @Produce      json
// @Param        body  body      v1.BatchSplitRequest  true  "批量切割请求"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/chapter/batch-split [post]
// @Id        chapter:batch-split
func (h *VsChapterSplitHandler) BatchSplit(ctx *gin.Context) {
	var req v1.BatchSplitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "参数错误: "+err.Error()), nil)
		return
	}

	loginName := ""
	if v, ok := ctx.Get("login_name"); ok {
		if s, ok := v.(string); ok {
			loginName = s
		}
	}

	result, err := h.svc.BatchSplitChapters(ctx, req.ProjectID, req.ChapterIDs, loginName)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}
