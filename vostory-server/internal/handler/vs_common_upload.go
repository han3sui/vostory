package handler

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsCommonUploadHandler struct {
	*Handler
	voiceAssetRepo   repository.VsVoiceAssetRepository
	voiceProfileRepo repository.VsVoiceProfileRepository
	voiceEmotionRepo repository.VsVoiceEmotionRepository
}

func NewVsCommonUploadHandler(
	handler *Handler,
	voiceAssetRepo repository.VsVoiceAssetRepository,
	voiceProfileRepo repository.VsVoiceProfileRepository,
	voiceEmotionRepo repository.VsVoiceEmotionRepository,
) *VsCommonUploadHandler {
	return &VsCommonUploadHandler{
		Handler:          handler,
		voiceAssetRepo:   voiceAssetRepo,
		voiceProfileRepo: voiceProfileRepo,
		voiceEmotionRepo: voiceEmotionRepo,
	}
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
// @Description  上传参考音频文件（mp3/wav/flac/ogg），返回存储路径和原始文件名
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

	storedName := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), file.Filename)
	filePath := filepath.Join(uploadDir, storedName)

	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, "保存文件失败"), nil)
		return
	}

	v1.HandleSuccess(ctx, gin.H{
		"path":     filePath,
		"filename": file.Filename,
	})
}

// StreamReferenceAudio godoc
// @Summary      流式获取参考音频
// @Description  根据来源类型和ID以流的方式返回参考音频文件
// @Tags         通用上传
// @Param        source  query  string  true  "来源类型（voice-asset/voice-profile/voice-emotion）"
// @Param        id      query  int     true  "资源ID"
// @Produce      application/octet-stream
// @Success      200  {file}  audio
// @Failure      400  {object}  v1.Response
// @Failure      404  {object}  v1.Response
// @Router       /api/v1/common/reference-audio/stream [get]
// @Id        common:referenceAudio:stream
func (h *VsCommonUploadHandler) StreamReferenceAudio(ctx *gin.Context) {
	source := ctx.Query("source")
	id := cast.ToUint64(ctx.Query("id"))
	if source == "" || id == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "source 和 id 参数不能为空"), nil)
		return
	}

	var storedPath string
	switch source {
	case "voice-asset":
		asset, err := h.voiceAssetRepo.FindByID(ctx, id)
		if err != nil {
			v1.HandleError(ctx, http.StatusNotFound, v1.NewError(404, "音色资源不存在"), nil)
			return
		}
		storedPath = asset.ReferenceAudioURL
	case "voice-profile":
		profile, err := h.voiceProfileRepo.FindByID(ctx, id)
		if err != nil {
			v1.HandleError(ctx, http.StatusNotFound, v1.NewError(404, "声音配置不存在"), nil)
			return
		}
		storedPath = profile.ReferenceAudioURL
	case "voice-emotion":
		emotion, err := h.voiceEmotionRepo.FindByID(ctx, id)
		if err != nil {
			v1.HandleError(ctx, http.StatusNotFound, v1.NewError(404, "情绪音频不存在"), nil)
			return
		}
		storedPath = emotion.ReferenceAudioURL
	default:
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "不支持的 source 类型"), nil)
		return
	}

	if storedPath == "" {
		v1.HandleError(ctx, http.StatusNotFound, v1.NewError(404, "该资源未配置参考音频"), nil)
		return
	}

	if _, err := os.Stat(storedPath); os.IsNotExist(err) {
		v1.HandleError(ctx, http.StatusNotFound, v1.NewError(404, "音频文件不存在"), nil)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(storedPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Content-Disposition", "inline")
	ctx.File(storedPath)
}
