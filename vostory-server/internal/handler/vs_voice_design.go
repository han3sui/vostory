package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsVoiceDesignHandler struct {
	*Handler
	svc service.VsVoiceDesignService
}

func NewVsVoiceDesignHandler(handler *Handler, svc service.VsVoiceDesignService) *VsVoiceDesignHandler {
	return &VsVoiceDesignHandler{Handler: handler, svc: svc}
}

// Suggest godoc
// @Summary      LLM 生成音色描述
// @Description  根据角色信息与台词，调用 LLM 生成音色描述（instruction）与参考台词（sample_text）
// @Tags         角色管理
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/character/{id}/voice-design/suggest [post]
// @Id        character:voice_design
func (h *VsVoiceDesignHandler) Suggest(ctx *gin.Context) {
	id := cast.ToUint64(ctx.Param("id"))
	if id == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	resp, err := h.svc.Suggest(ctx, id)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// Preview godoc
// @Summary      音色设计试听
// @Description  按音色描述调用 Breeze-TTS-2 Voice Design 生成试听音频，未传 seed 时随机生成
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                           true  "角色ID"
// @Param        request  body      v1.VoiceDesignPreviewRequest  true  "试听请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character/{id}/voice-design/preview [post]
// @Id        character:voice_design_preview
func (h *VsVoiceDesignHandler) Preview(ctx *gin.Context) {
	id := cast.ToUint64(ctx.Param("id"))
	if id == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	request := &v1.VoiceDesignPreviewRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	resp, err := h.svc.Preview(ctx, id, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// Apply godoc
// @Summary      保存音色设计
// @Description  将试听音频保存为声音配置（参考音频 + 参考台词 + 种子），并绑定到角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "角色ID"
// @Param        request  body      v1.VoiceDesignApplyRequest  true  "保存请求"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character/{id}/voice-design/apply [post]
// @Id        character:voice_design_apply
func (h *VsVoiceDesignHandler) Apply(ctx *gin.Context) {
	id := cast.ToUint64(ctx.Param("id"))
	if id == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	request := &v1.VoiceDesignApplyRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	resp, err := h.svc.Apply(ctx, id, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// Auto godoc
// @Summary      一键自动生成音色
// @Description  LLM 生成音色描述 → Voice Design 生成参考音频 → 保存为声音配置并绑定角色
// @Tags         角色管理
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/character/{id}/voice-design/auto [post]
// @Id        character:voice_design_auto
func (h *VsVoiceDesignHandler) Auto(ctx *gin.Context) {
	id := cast.ToUint64(ctx.Param("id"))
	if id == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	resp, err := h.svc.Auto(ctx, id)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}
