package handler

import (
	"net/http"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsExportJobHandler struct {
	*Handler
	svc service.VsExportJobService
}

func NewVsExportJobHandler(handler *Handler, svc service.VsExportJobService) *VsExportJobHandler {
	return &VsExportJobHandler{Handler: handler, svc: svc}
}

// ExportChapterAudio godoc
// @Summary      导出章节音频
// @Description  将章节下所有已生成的音频片段合并为一个文件
// @Tags         导出
// @Accept       json
// @Produce      json
// @Param        chapter_id  path  int                         true  "章节ID"
// @Param        body        body  v1.ExportChapterAudioRequest true  "导出请求"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/chapter/{chapter_id}/export [post]
// @Id        tts:exportChapter
func (h *VsExportJobHandler) ExportChapterAudio(ctx *gin.Context) {
	chapterID := cast.ToUint64(ctx.Param("chapter_id"))
	if chapterID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "chapter_id is required"), nil)
		return
	}

	var req v1.ExportChapterAudioRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "参数错误: "+err.Error()), nil)
		return
	}

	result, err := h.svc.ExportChapterAudio(ctx, chapterID, req.Format)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// GetExportJob godoc
// @Summary      查询导出任务状态
// @Tags         导出
// @Param        export_job_id  path  int  true  "导出任务ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/tts/export/{export_job_id} [get]
// @Id        tts:getExportJob
func (h *VsExportJobHandler) GetExportJob(ctx *gin.Context) {
	exportJobID := cast.ToUint64(ctx.Param("export_job_id"))
	if exportJobID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "export_job_id is required"), nil)
		return
	}

	result, err := h.svc.GetExportJob(ctx, exportJobID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// DownloadExport godoc
// @Summary      下载导出文件
// @Tags         导出
// @Produce      application/octet-stream
// @Param        export_job_id  path  int  true  "导出任务ID"
// @Success      200  {file}  audio
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/tts/export/{export_job_id}/download [get]
// @Id        tts:downloadExport
func (h *VsExportJobHandler) DownloadExport(ctx *gin.Context) {
	exportJobID := cast.ToUint64(ctx.Param("export_job_id"))
	if exportJobID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "export_job_id is required"), nil)
		return
	}

	filePath, contentType, fileName, err := h.svc.GetExportFilePath(ctx, exportJobID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.File(filePath)
}
