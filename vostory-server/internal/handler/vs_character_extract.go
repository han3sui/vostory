package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsCharacterExtractHandler struct {
	*Handler
	svc service.VsCharacterExtractService
}

func NewVsCharacterExtractHandler(handler *Handler, svc service.VsCharacterExtractService) *VsCharacterExtractHandler {
	return &VsCharacterExtractHandler{Handler: handler, svc: svc}
}

// Extract godoc
// @Summary      LLM 智能提取角色
// @Description  调用 LLM 从全书文本中自动提取角色信息
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/character/extract/{project_id} [post]
// @Id        character:extract
func (h *VsCharacterExtractHandler) Extract(ctx *gin.Context) {
	projectID := ctx.Param("project_id")
	if projectID == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	result, err := h.svc.ExtractCharacters(ctx, cast.ToUint64(projectID))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}

// ExtractFromText godoc
// @Summary      LLM 智能录入角色
// @Description  用户提供一段文字描述，调用 LLM 提取角色信息并录入角色库
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        body  body      v1.CharacterExtractFromTextRequest  true  "请求参数"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/character/extract-from-text [post]
// @Id        character:extract_from_text
func (h *VsCharacterExtractHandler) ExtractFromText(ctx *gin.Context) {
	var req v1.CharacterExtractFromTextRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "参数错误: "+err.Error()), nil)
		return
	}
	result, err := h.svc.ExtractFromText(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, result)
}
