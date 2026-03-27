package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsVoiceMatchHandler struct {
	*Handler
	svc service.VsVoiceMatchService
}

func NewVsVoiceMatchHandler(handler *Handler, svc service.VsVoiceMatchService) *VsVoiceMatchHandler {
	return &VsVoiceMatchHandler{Handler: handler, svc: svc}
}

// MatchVoices godoc
// @Summary      LLM 自动匹配角色声音
// @Description  调用 LLM 根据角色描述和声音描述自动匹配，已绑定声音的角色自动跳过
// @Tags         角色管理
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/character/voice-match/{project_id} [post]
// @Id        character:voice_match
func (h *VsVoiceMatchHandler) MatchVoices(ctx *gin.Context) {
	projectID := ctx.Param("project_id")
	if projectID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	result, err := h.svc.MatchVoices(ctx, cast.ToUint64(projectID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}
