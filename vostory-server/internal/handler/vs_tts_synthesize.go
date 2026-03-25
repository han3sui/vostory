package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsTTSSynthesizeHandler struct {
	*Handler
	svc service.VsTTSSynthesizeService
}

func NewVsTTSSynthesizeHandler(handler *Handler, svc service.VsTTSSynthesizeService) *VsTTSSynthesizeHandler {
	return &VsTTSSynthesizeHandler{Handler: handler, svc: svc}
}

// Synthesize godoc
// @Summary      合成单个片段语音
// @Description  根据片段ID调用TTS合成语音
// @Tags         TTS合成
// @Param        segment_id  path  int  true  "片段ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/synthesize/{segment_id} [post]
// @Id        tts:synthesize
func (h *VsTTSSynthesizeHandler) Synthesize(ctx *gin.Context) {
	segmentID := cast.ToUint64(ctx.Param("segment_id"))
	if segmentID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "segment_id is required"), nil)
		return
	}

	result, err := h.svc.SynthesizeSegment(ctx, segmentID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// GetAudio godoc
// @Summary      获取片段最新音频
// @Description  获取指定片段的当前版本音频信息
// @Tags         TTS合成
// @Param        segment_id  path  int  true  "片段ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/audio/{segment_id} [get]
// @Id        tts:audio
func (h *VsTTSSynthesizeHandler) GetAudio(ctx *gin.Context) {
	segmentID := cast.ToUint64(ctx.Param("segment_id"))
	if segmentID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "segment_id is required"), nil)
		return
	}

	result, err := h.svc.GetSegmentAudio(ctx, segmentID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}
