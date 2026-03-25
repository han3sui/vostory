package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type VsLLMLogHandler struct {
	*Handler
	svc service.VsLLMLogService
}

func NewVsLLMLogHandler(handler *Handler, svc service.VsLLMLogService) *VsLLMLogHandler {
	return &VsLLMLogHandler{Handler: handler, svc: svc}
}

// Get godoc
// @Summary      获取LLM调用日志详情
// @Description  根据ID获取LLM调用日志详情
// @Tags         LLM调用日志
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "日志ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/llm-log/{id} [get]
// @Id        ai:llm-log:detail
func (h *VsLLMLogHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}
	log, err := h.svc.FindByID(ctx, cast.ToUint64(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, log)
}

// List godoc
// @Summary      获取LLM调用日志列表
// @Description  分页获取LLM调用日志列表
// @Tags         LLM调用日志
// @Accept       json
// @Produce      json
// @Param        page        query     int     false  "当前页"
// @Param        size        query     int     false  "每页数量"
// @Param        project_id  query     int     false  "项目ID"
// @Param        provider_id query     int     false  "提供商ID"
// @Param        model_name  query     string  false  "模型名称"
// @Param        status      query     int     false  "状态（0成功 1失败, -1全部）"
// @Success      200         {object}  v1.Response
// @Failure      500         {object}  v1.Response
// @Router       /api/v1/ai/llm-log/list [get]
// @Id        ai:llm-log:list
func (h *VsLLMLogHandler) List(ctx *gin.Context) {
	query := &v1.VsLLMLogListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.ProjectID = cast.ToUint64(ctx.Query("project_id"))
	query.ProviderID = cast.ToUint64(ctx.Query("provider_id"))
	query.ModelName = ctx.Query("model_name")
	query.Status = cast.ToInt(ctx.DefaultQuery("status", "-1"))

	logs, total, err := h.svc.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}
	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, logs))
}

// Delete godoc
// @Summary      删除LLM调用日志
// @Description  删除指定LLM调用日志
// @Tags         LLM调用日志
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "日志ID"
// @Success      200  {object}  v1.Response
// @Failure      400  {object}  v1.Response
// @Failure      500  {object}  v1.Response
// @Router       /api/v1/ai/llm-log/{id} [delete]
// @Id        ai:llm-log:remove
func (h *VsLLMLogHandler) Delete(ctx *gin.Context) {
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
