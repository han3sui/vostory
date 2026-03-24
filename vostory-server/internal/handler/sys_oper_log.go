package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysOperLogHandler struct {
	*Handler
	sysOperLogService service.SysOperLogService
}

func NewSysOperLogHandler(
	handler *Handler,
	sysOperLogService service.SysOperLogService,
) *SysOperLogHandler {
	return &SysOperLogHandler{
		Handler:           handler,
		sysOperLogService: sysOperLogService,
	}
}

// Get godoc
// @Summary      获取操作日志详情
// @Description  根据ID获取操作日志详情
// @Tags         操作日志
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "操作日志ID"
// @Success      200  {object}  v1.Response[v1.SysOperLogDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/operlog/{id} [get]
// @Id        system:operlog:detail
func (h *SysOperLogHandler) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	operLog, err := h.sysOperLogService.FindByID(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, operLog)
}

// List godoc
// @Summary      获取操作日志列表
// @Description  分页获取操作日志列表
// @Tags         操作日志
// @Accept       json
// @Produce      json
// @Param        page           query     int     false  "当前页"
// @Param        size           query     int     false  "每页数量"
// @Param        title          query     string  false  "模块标题"
// @Param        business_type  query     int     false  "业务类型"
// @Param        oper_name      query     string  false  "操作人员"
// @Param        status         query     int     false  "操作状态"
// @Param        begin_time     query     string  false  "开始时间"
// @Param        end_time       query     string  false  "结束时间"
// @Success      200            {object}  v1.Response[v1.PageResponse[v1.SysOperLogDetailResponse]]
// @Failure      400            {object}  v1.Response[any]
// @Failure      500            {object}  v1.Response[any]
// @Router       /api/v1/system/operlog/list [get]
// @Id        system:operlog:list
func (h *SysOperLogHandler) List(ctx *gin.Context) {
	query := &v1.SysOperLogListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	query.Page = cast.ToInt(ctx.Query("page"))
	query.Size = cast.ToInt(ctx.Query("size"))
	query.Title = ctx.Query("title")
	query.BusinessType = ctx.Query("business_type")
	query.OperName = ctx.Query("oper_name")
	query.Status = ctx.Query("status")
	query.BeginTime = ctx.Query("begin_time")
	query.EndTime = ctx.Query("end_time")

	operLogs, total, err := h.sysOperLogService.FindWithPagination(ctx, query)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, v1.NewPageResponse(query.Page, query.Size, total, operLogs))
}

// Delete godoc
// @Summary      删除操作日志
// @Description  删除指定操作日志
// @Tags         操作日志
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "操作日志ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/operlog/{id} [delete]
// @Id        system:operlog:remove
func (h *SysOperLogHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		v1.HandleError(ctx, http.StatusBadRequest, v1.NewError(400, "ID is required"), nil)
		return
	}

	err := h.sysOperLogService.Delete(ctx, cast.ToUint(id))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}

// Clean godoc
// @Summary      清空操作日志
// @Description  清空所有操作日志
// @Tags         操作日志
// @Accept       json
// @Produce      json
// @Success      200  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/operlog/clean [delete]
// @Id        system:operlog:clean
func (h *SysOperLogHandler) Clean(ctx *gin.Context) {
	err := h.sysOperLogService.Clean(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.NewError(500, err.Error()), nil)
		return
	}

	v1.HandleSuccess(ctx, nil)
}
