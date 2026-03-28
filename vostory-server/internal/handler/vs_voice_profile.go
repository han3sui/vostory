package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsVoiceProfileHandler struct {
	*Handler
	svc service.VsVoiceProfileService
}

func NewVsVoiceProfileHandler(handler *Handler, svc service.VsVoiceProfileService) *VsVoiceProfileHandler {
	return &VsVoiceProfileHandler{Handler: handler, svc: svc}
}

// Get godoc
// @Summary      获取声音配置详情
// @Description  根据ID获取声音配置详情
// @Tags         声音配置
// @Param        id   path      int  true  "声音配置ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/voice-profile/{id} [get]
// @Id        voice-profile:detail
func (h *VsVoiceProfileHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	profile, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, profile)
}

// List godoc
// @Summary      获取声音配置列表
// @Description  分页获取声音配置列表
// @Tags         声音配置
// @Param        page       query     int     false  "当前页"
// @Param        size       query     int     false  "每页数量"
// @Param        project_id query     int     false  "项目ID"
// @Param        name       query     string  false  "配置名称"
// @Param        status     query     string  false  "状态"
// @Success      200        {object}  v1.Response
// @Failure      500        {object}  v1.Response
// @Router       /api/v1/voice-profile/list [get]
// @Id        voice-profile:list
func (h *VsVoiceProfileHandler) List(ctx *gin.Context) {
	query := &v1.VsVoiceProfileListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ProjectID = cast.ToUint64(ctx.Query("project_id"))
	query.Name = ctx.Query("name")
	query.Status = ctx.Query("status")

	profiles, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, profiles))
}

// Create godoc
// @Summary      创建声音配置
// @Description  创建新的声音配置
// @Tags         声音配置
// @Accept       json
// @Produce      json
// @Param        body  body      v1.VsVoiceProfileCreateRequest  true  "声音配置信息"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/voice-profile [post]
// @Id        voice-profile:add
func (h *VsVoiceProfileHandler) Create(ctx *gin.Context) {
	var request v1.VsVoiceProfileCreateRequest
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
// @Summary      更新声音配置
// @Description  更新声音配置信息
// @Tags         声音配置
// @Accept       json
// @Produce      json
// @Param        id    path      int                              true  "声音配置ID"
// @Param        body  body      v1.VsVoiceProfileUpdateRequest   true  "声音配置信息"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/voice-profile/{id} [put]
// @Id        voice-profile:edit
func (h *VsVoiceProfileHandler) Update(ctx *gin.Context) {
	var request v1.VsVoiceProfileUpdateRequest
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
// @Summary      删除声音配置
// @Description  删除指定声音配置
// @Tags         声音配置
// @Param        id   path      int  true  "声音配置ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/voice-profile/{id} [delete]
// @Id        voice-profile:remove
func (h *VsVoiceProfileHandler) Delete(ctx *gin.Context) {
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

// Enable godoc
// @Summary      启用声音配置
// @Tags         声音配置
// @Param        id   path      int  true  "声音配置ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-profile/{id}/enable [put]
// @Id        voice-profile:enable
func (h *VsVoiceProfileHandler) Enable(ctx *gin.Context) {
	if err := h.svc.Enable(ctx, cast.ToUint64(ctx.Param("id"))); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Disable godoc
// @Summary      停用声音配置
// @Tags         声音配置
// @Param        id   path      int  true  "声音配置ID"
// @Success      200  {object}  v1.Response
// @Router       /api/v1/voice-profile/{id}/disable [put]
// @Id        voice-profile:disable
func (h *VsVoiceProfileHandler) Disable(ctx *gin.Context) {
	if err := h.svc.Disable(ctx, cast.ToUint64(ctx.Param("id"))); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Import godoc
// @Summary      从音色库批量导入声音配置
// @Description  根据音色资产ID列表批量创建声音配置
// @Tags         声音配置
// @Accept       json
// @Produce      json
// @Param        body  body      v1.VsVoiceProfileImportRequest  true  "导入请求"
// @Success      200   {object}  v1.Response
// @Failure      400   {object}  v1.Response
// @Failure      500   {object}  v1.Response
// @Router       /api/v1/voice-profile/import [post]
// @Id        voice-profile:import
func (h *VsVoiceProfileHandler) Import(ctx *gin.Context) {
	var request v1.VsVoiceProfileImportRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, err.Error()), nil)
		return
	}
	count, err := h.svc.ImportFromAssets(ctx, &request)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, map[string]int{"imported": count})
}

// GetByProject godoc
// @Summary      获取项目下的声音配置选项
// @Description  获取指定项目下启用的声音配置列表（下拉选择用）
// @Tags         声音配置
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  v1.Response
// @Failure      400         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/common/voice-profile/project/{project_id} [get]
// @Id        common:voice-profile:project
func (h *VsVoiceProfileHandler) GetByProject(ctx *gin.Context) {
	projectID := cast.ToUint64(ctx.Param("project_id"))
	if projectID == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "project_id is required"), nil)
		return
	}
	options, err := h.svc.FindByProjectID(ctx, projectID)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, options)
}
