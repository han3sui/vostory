package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsVoiceEmotionHandler struct {
	*Handler
	svc service.VsVoiceEmotionService
}

func NewVsVoiceEmotionHandler(handler *Handler, svc service.VsVoiceEmotionService) *VsVoiceEmotionHandler {
	return &VsVoiceEmotionHandler{Handler: handler, svc: svc}
}

// Get godoc
// @Summary      获取情绪音频详情
// @Tags         情绪音频
// @Param        id   path      int  true  "情绪音频ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-emotion/{id} [get]
// @Id        voice-emotion:detail
func (h *VsVoiceEmotionHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	emotion, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, emotion)
}

// List godoc
// @Summary      获取情绪音频列表
// @Tags         情绪音频
// @Param        page              query  int     false  "当前页"
// @Param        size              query  int     false  "每页数量"
// @Param        voice_profile_id  query  int     false  "声音配置ID"
// @Param        emotion_type      query  string  false  "情绪类型"
// @Param        emotion_strength  query  string  false  "情绪强度"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-emotion/list [get]
// @Id        voice-emotion:list
func (h *VsVoiceEmotionHandler) List(ctx *gin.Context) {
	query := &v1.VsVoiceEmotionListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.VoiceProfileID = cast.ToUint64(ctx.Query("voice_profile_id"))
	query.VoiceAssetID = cast.ToUint64(ctx.Query("voice_asset_id"))
	query.EmotionType = ctx.Query("emotion_type")
	query.EmotionStrength = ctx.Query("emotion_strength")

	emotions, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, emotions))
}

// Create godoc
// @Summary      创建情绪音频
// @Tags         情绪音频
// @Accept       json
// @Param        body  body  v1.VsVoiceEmotionCreateRequest  true  "情绪音频信息"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-emotion [post]
// @Id        voice-emotion:add
func (h *VsVoiceEmotionHandler) Create(ctx *gin.Context) {
	var request v1.VsVoiceEmotionCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	if err := h.svc.Create(ctx, &request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary      更新情绪音频
// @Tags         情绪音频
// @Accept       json
// @Param        id    path  int                              true  "情绪音频ID"
// @Param        body  body  v1.VsVoiceEmotionUpdateRequest   true  "情绪音频信息"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-emotion/{id} [put]
// @Id        voice-emotion:edit
func (h *VsVoiceEmotionHandler) Update(ctx *gin.Context) {
	var request v1.VsVoiceEmotionUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	request.ID = cast.ToUint64(ctx.Param("id"))
	if err := h.svc.Update(ctx, &request); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary      删除情绪音频
// @Tags         情绪音频
// @Param        id   path  int  true  "情绪音频ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-emotion/{id} [delete]
// @Id        voice-emotion:remove
func (h *VsVoiceEmotionHandler) Delete(ctx *gin.Context) {
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

// GetByVoiceProfile godoc
// @Summary      获取声音配置下的情绪音频列表
// @Tags         情绪音频
// @Param        voice_profile_id  path  int  true  "声音配置ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/common/voice-emotion/profile/{voice_profile_id} [get]
// @Id        common:voice-emotion:profile
func (h *VsVoiceEmotionHandler) GetByVoiceProfile(ctx *gin.Context) {
	voiceProfileID := cast.ToUint64(ctx.Param("voice_profile_id"))
	if voiceProfileID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "voice_profile_id is required"), nil)
		return
	}
	emotions, err := h.svc.FindByVoiceProfileID(ctx, voiceProfileID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, emotions)
}

// GetByVoiceAsset godoc
// @Summary      获取音色资产下的情绪音频列表
// @Tags         情绪音频
// @Param        voice_asset_id  path  int  true  "音色资产ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/common/voice-emotion/asset/{voice_asset_id} [get]
// @Id        common:voice-emotion:asset
func (h *VsVoiceEmotionHandler) GetByVoiceAsset(ctx *gin.Context) {
	voiceAssetID := cast.ToUint64(ctx.Param("voice_asset_id"))
	if voiceAssetID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "voice_asset_id is required"), nil)
		return
	}
	emotions, err := h.svc.FindByVoiceAssetID(ctx, voiceAssetID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, emotions)
}
