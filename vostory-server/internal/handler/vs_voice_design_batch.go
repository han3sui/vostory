package handler

import (
	v1 "iot-alert-center/api/v1"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// StartBatch godoc
// @Summary      批量生成音色
// @Description  为选中的角色创建批量音色设计任务，由后台逐个生成；confirm_important 为 true 时主角、配角、旁白先生成草稿待试听确认
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int                         true  "项目ID"
// @Param        request     body      v1.VoiceDesignBatchRequest  true  "请求参数"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/character/voice-design/batch/{project_id} [post]
// @Id        character:voice_design_batch
func (h *VsVoiceDesignHandler) StartBatch(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	request := &v1.VoiceDesignBatchRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	resp, err := h.svc.StartBatch(ctx, projectID, request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// GetBatch godoc
// @Summary      批量音色设计进度
// @Description  返回项目最近一批音色设计任务的进度，以及更早批次中仍待确认的草稿
// @Tags         角色管理
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/character/voice-design/batch/{project_id} [get]
// @Id        character:voice_design_batch_status
func (h *VsVoiceDesignHandler) GetBatch(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	resp, err := h.svc.GetBatch(ctx, projectID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// CancelBatch godoc
// @Summary      停止批量音色设计
// @Description  取消项目中尚未开始的音色设计任务项，正在处理的那一个会继续完成
// @Tags         角色管理
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/character/voice-design/batch/{project_id}/cancel [post]
// @Id        character:voice_design_batch_cancel
func (h *VsVoiceDesignHandler) CancelBatch(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	cancelled, err := h.svc.CancelBatch(ctx, projectID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, gin.H{"cancelled": cancelled})
}

// ApplyItem godoc
// @Summary      确认音色草稿
// @Description  将批量生成的音色草稿保存为声音配置并绑定角色
// @Tags         角色管理
// @Produce      json
// @Param        item_id  path      int  true  "任务项ID"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character/voice-design/item/{item_id}/apply [post]
// @Id        character:voice_design_item_apply
func (h *VsVoiceDesignHandler) ApplyItem(ctx *gin.Context) {
	itemID := cast.ToUint64(ctx.Param("item_id"))
	if itemID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "item_id is required"), nil)
		return
	}
	resp, err := h.svc.ApplyItem(ctx, itemID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, resp)
}

// RegenerateItem godoc
// @Summary      重新生成音色草稿
// @Description  丢弃当前草稿或失败结果，重新排队生成
// @Tags         角色管理
// @Produce      json
// @Param        item_id  path      int  true  "任务项ID"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character/voice-design/item/{item_id}/regenerate [post]
// @Id        character:voice_design_item_regenerate
func (h *VsVoiceDesignHandler) RegenerateItem(ctx *gin.Context) {
	itemID := cast.ToUint64(ctx.Param("item_id"))
	if itemID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "item_id is required"), nil)
		return
	}
	if err := h.svc.RegenerateItem(ctx, itemID); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// DiscardItem godoc
// @Summary      放弃音色草稿
// @Description  删除草稿音频，角色保持未绑定声音
// @Tags         角色管理
// @Produce      json
// @Param        item_id  path      int  true  "任务项ID"
// @Success      200      {object}  v1.Response
// @Failure      400      {object}  v1.Response
// @Failure      500      {object}  v1.Response
// @Router       /api/v1/character/voice-design/item/{item_id}/discard [post]
// @Id        character:voice_design_item_discard
func (h *VsVoiceDesignHandler) DiscardItem(ctx *gin.Context) {
	itemID := cast.ToUint64(ctx.Param("item_id"))
	if itemID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "item_id is required"), nil)
		return
	}
	if err := h.svc.DiscardItem(ctx, itemID); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// ItemAudio godoc
// @Summary      试听音色草稿
// @Description  返回音色草稿的 WAV 音频
// @Tags         角色管理
// @Produce      audio/wav
// @Param        item_id  path      int  true  "任务项ID"
// @Success      200      {file}    binary
// @Failure      400      {object}  v1.Response
// @Failure      404      {object}  v1.Response
// @Router       /api/v1/character/voice-design/item/{item_id}/audio [get]
// @Id        character:voice_design_item_audio
func (h *VsVoiceDesignHandler) ItemAudio(ctx *gin.Context) {
	itemID := cast.ToUint64(ctx.Param("item_id"))
	if itemID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "item_id is required"), nil)
		return
	}
	r, err := h.svc.OpenItemAudio(ctx, itemID)
	if err != nil {
		v1.HandleError(ctx, http.StatusNotFound, v1.NewError(404, err.Error()), nil)
		return
	}
	streamAttachment(ctx, r, "audio/wav", "inline")
}
