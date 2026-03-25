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
