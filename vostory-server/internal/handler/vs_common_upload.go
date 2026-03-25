package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"

	"github.com/gin-gonic/gin"
)

type VsCommonUploadHandler struct {
	*Handler
}

func NewVsCommonUploadHandler(handler *Handler) *VsCommonUploadHandler {
	return &VsCommonUploadHandler{Handler: handler}
}

var allowedAudioExts = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".ogg":  true,
}

const maxAudioSize = 20 << 20 // 20MB

// UploadReferenceAudio godoc
// @Summary      上传参考音频
// @Description  上传参考音频文件（mp3/wav/flac/ogg），返回存储路径
// @Tags         通用上传
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "音频文件"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/common/upload/reference-audio [post]
// @Id        common:upload:referenceAudio
func (h *VsCommonUploadHandler) UploadReferenceAudio(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "请选择音频文件"), nil)
		return
	}

	if file.Size > maxAudioSize {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "文件大小不能超过 20MB"), nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAudioExts[ext] {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "不支持的音频格式，仅支持 mp3/wav/flac/ogg"), nil)
		return
	}

	uploadDir := filepath.Join("storage", "reference-audio")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, "创建上传目录失败"), nil)
		return
	}

	fileName := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), file.Filename)
	filePath := filepath.Join(uploadDir, fileName)

	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, "保存文件失败"), nil)
		return
	}

	v1.HandleSuccess(ctx, gin.H{"url": filePath})
}
