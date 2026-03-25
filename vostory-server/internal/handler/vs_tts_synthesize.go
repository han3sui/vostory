package handler

import (
	"net/http"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

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

// BatchGenerate godoc
// @Summary      批量生成章节语音
// @Description  异步批量生成指定章节所有可生成片段的语音
// @Tags         TTS合成
// @Accept       json
// @Param        body  body  v1.BatchGenerateRequest  true  "批量生成请求"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/batch-generate [post]
// @Id        tts:batchGenerate
func (h *VsTTSSynthesizeHandler) BatchGenerate(ctx *gin.Context) {
	var req v1.BatchGenerateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "参数错误: "+err.Error()), nil)
		return
	}

	result, err := h.svc.BatchGenerate(ctx, req.ChapterID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// GetTaskProgress godoc
// @Summary      查询任务进度
// @Description  查询异步生成任务的当前进度
// @Tags         TTS合成
// @Param        task_id  path  int  true  "任务ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/task/{task_id} [get]
// @Id        tts:taskProgress
func (h *VsTTSSynthesizeHandler) GetTaskProgress(ctx *gin.Context) {
	taskID := cast.ToUint64(ctx.Param("task_id"))
	if taskID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "task_id is required"), nil)
		return
	}

	result, err := h.svc.GetTaskProgress(ctx, taskID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// StreamAudio godoc
// @Summary      流式获取音频
// @Description  根据音频片段ID以流的方式返回音频数据
// @Tags         TTS合成
// @Param        clip_id  path  int  true  "音频片段ID"
// @Produce      application/octet-stream
// @Success      200  {file}  audio
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/stream/{clip_id} [get]
// @Id        tts:streamAudio
func (h *VsTTSSynthesizeHandler) StreamAudio(ctx *gin.Context) {
	clipID := cast.ToUint64(ctx.Param("clip_id"))
	if clipID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "clip_id is required"), nil)
		return
	}

	filePath, contentType, err := h.svc.GetAudioClipFile(ctx, clipID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Content-Disposition", "inline")
	ctx.File(filePath)
}
