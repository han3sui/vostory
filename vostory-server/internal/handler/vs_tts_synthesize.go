package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

type VsTTSSynthesizeHandler struct {
	*Handler
	svc service.VsTTSSynthesizeService
	rdb *redis.Client
}

func NewVsTTSSynthesizeHandler(handler *Handler, svc service.VsTTSSynthesizeService, rdb *redis.Client) *VsTTSSynthesizeHandler {
	return &VsTTSSynthesizeHandler{Handler: handler, svc: svc, rdb: rdb}
}

// Synthesize godoc
// @Summary      合成单个片段语音（异步队列）
// @Description  将单个片段加入TTS生成队列，返回任务ID用于轮询进度
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

	result, err := h.svc.SingleGenerate(ctx, segmentID)
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
		if strings.HasPrefix(err.Error(), "CONFLICT:") {
			v1.HandleError(ctx, http.StatusConflict, v1.NewError(409, strings.TrimPrefix(err.Error(), "CONFLICT:")), nil)
			return
		}
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

// GetActiveTask godoc
// @Summary      查询章节活跃任务
// @Description  查询指定章节当前正在运行或等待中的生成任务
// @Tags         TTS合成
// @Param        chapter_id  path  int  true  "章节ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Router       /api/v1/tts/chapter/{chapter_id}/active-task [get]
// @Id        tts:activeTask
func (h *VsTTSSynthesizeHandler) GetActiveTask(ctx *gin.Context) {
	chapterID := cast.ToUint64(ctx.Param("chapter_id"))
	if chapterID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}

	result, err := h.svc.GetActiveTaskByChapter(ctx, chapterID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, result)
}

// GetActiveTasksByProject godoc
// @Summary      查询项目活跃任务列表
// @Description  查询指定项目下所有正在运行或等待中的生成任务
// @Tags         TTS合成
// @Param        project_id  path  int  true  "项目ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Router       /api/v1/tts/project/{project_id}/active-tasks [get]
// @Id        tts:projectActiveTasks
func (h *VsTTSSynthesizeHandler) GetActiveTasksByProject(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	result, err := h.svc.GetActiveTasksByProject(ctx, projectID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// LockSegment godoc
// @Summary      锁定片段
// @Tags         TTS合成
// @Param        segment_id  path  int  true  "片段ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/segment/{segment_id}/lock [put]
// @Id        tts:lockSegment
func (h *VsTTSSynthesizeHandler) LockSegment(ctx *gin.Context) {
	segmentID := cast.ToUint64(ctx.Param("segment_id"))
	if segmentID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "segment_id is required"), nil)
		return
	}
	if err := h.svc.LockSegment(ctx, segmentID); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// UnlockSegment godoc
// @Summary      解锁片段
// @Tags         TTS合成
// @Param        segment_id  path  int  true  "片段ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/segment/{segment_id}/unlock [put]
// @Id        tts:unlockSegment
func (h *VsTTSSynthesizeHandler) UnlockSegment(ctx *gin.Context) {
	segmentID := cast.ToUint64(ctx.Param("segment_id"))
	if segmentID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "segment_id is required"), nil)
		return
	}
	if err := h.svc.UnlockSegment(ctx, segmentID); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// BatchLockByChapter godoc
// @Summary      批量锁定章节下所有已生成片段
// @Tags         TTS合成
// @Param        chapter_id  path  int  true  "章节ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/chapter/{chapter_id}/lock [put]
// @Id        tts:batchLock
func (h *VsTTSSynthesizeHandler) BatchLockByChapter(ctx *gin.Context) {
	chapterID := cast.ToUint64(ctx.Param("chapter_id"))
	if chapterID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}
	affected, err := h.svc.BatchLockByChapter(ctx, chapterID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.BatchLockResponse{AffectedCount: affected})
}

// BatchUnlockByChapter godoc
// @Summary      批量解锁章节下所有已锁定片段
// @Tags         TTS合成
// @Param        chapter_id  path  int  true  "章节ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/chapter/{chapter_id}/unlock [put]
// @Id        tts:batchUnlock
func (h *VsTTSSynthesizeHandler) BatchUnlockByChapter(ctx *gin.Context) {
	chapterID := cast.ToUint64(ctx.Param("chapter_id"))
	if chapterID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}
	affected, err := h.svc.BatchUnlockByChapter(ctx, chapterID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.BatchLockResponse{AffectedCount: affected})
}

// CancelChapterQueue godoc
// @Summary      取消章节生成队列
// @Description  取消指定章节下所有排队中的片段
// @Tags         TTS合成
// @Param        chapter_id  path  int  true  "章节ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/chapter/{chapter_id}/cancel [post]
// @Id        tts:cancelChapter
func (h *VsTTSSynthesizeHandler) CancelChapterQueue(ctx *gin.Context) {
	chapterID := cast.ToUint64(ctx.Param("chapter_id"))
	if chapterID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}
	affected, err := h.svc.CancelChapterQueue(ctx, chapterID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.CancelQueueResponse{CancelledCount: affected})
}

// CancelProjectQueue godoc
// @Summary      取消项目生成队列
// @Description  取消指定项目下所有排队中的片段
// @Tags         TTS合成
// @Param        project_id  path  int  true  "项目ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/project/{project_id}/cancel [post]
// @Id        tts:cancelProject
func (h *VsTTSSynthesizeHandler) CancelProjectQueue(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	affected, err := h.svc.CancelProjectQueue(ctx, projectID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.CancelQueueResponse{CancelledCount: affected})
}

// StreamProjectEvents godoc
// @Summary      SSE 实时推送项目级 TTS 进度
// @Description  通过 Server-Sent Events 实时推送整个项目的 TTS 生成事件（包含所有章节）
// @Tags         TTS合成
// @Produce      text/event-stream
// @Param        project_id  path  int  true  "项目ID"
// @Success      200  {string}  string  "SSE stream"
// @Failure      400  {object}  v1.Response
// @Router       /api/v1/tts/project/{project_id}/events [get]
// @Id        tts:projectEvents
func (h *VsTTSSynthesizeHandler) StreamProjectEvents(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	channel := fmt.Sprintf("vs:tts:events:project:%d", projectID)
	sub := h.rdb.Subscribe(ctx.Request.Context(), channel)
	defer sub.Close()

	connEvent := fmt.Sprintf("event: connected\ndata: {\"project_id\":%d}\n\n", projectID)
	if _, err := ctx.Writer.Write([]byte(connEvent)); err != nil {
		return
	}
	ctx.Writer.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	ch := sub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			event := fmt.Sprintf("event: segment\ndata: %s\n\n", msg.Payload)
			if _, err := ctx.Writer.Write([]byte(event)); err != nil {
				return
			}
			ctx.Writer.Flush()

		case <-heartbeat.C:
			hb := fmt.Sprintf("event: heartbeat\ndata: {\"ts\":%d}\n\n", time.Now().UnixMilli())
			if _, err := ctx.Writer.Write([]byte(hb)); err != nil {
				return
			}
			ctx.Writer.Flush()

		case <-ctx.Request.Context().Done():
			return
		}
	}
}
