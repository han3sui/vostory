package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsPreciseFillHandler struct {
	*Handler
	svc service.VsPreciseFillService
}

func NewVsPreciseFillHandler(handler *Handler, svc service.VsPreciseFillService) *VsPreciseFillHandler {
	return &VsPreciseFillHandler{Handler: handler, svc: svc}
}

// AlignChapter godoc
// @Summary      精准填充对齐
// @Description  将章节下所有脚本片段的文本对齐回章节原文，确保不丢字不加字
// @Tags         精准填充
// @Accept       json
// @Produce      json
// @Param        chapter_id  path      int  true  "章节ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/chapter/{chapter_id}/align [post]
// @Id        chapter:align
func (h *VsPreciseFillHandler) AlignChapter(ctx *gin.Context) {
	chapterID := cast.ToUint64(ctx.Param("chapter_id"))
	if chapterID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}

	count, err := h.svc.AlignChapter(ctx, chapterID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, gin.H{
		"aligned_count": count,
		"message":       "精准填充完成",
	})
}
