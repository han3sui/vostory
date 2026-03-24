package handler

import (
	"net/http"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysLogininforHandler struct {
	sysLogininforService service.SysLogininforService
}

func NewSysLogininforHandler(sysLogininforService service.SysLogininforService) *SysLogininforHandler {
	return &SysLogininforHandler{sysLogininforService: sysLogininforService}
}

// Get godoc
// @Summary      获取登录日志详情
// @Description  根据ID获取登录日志详情
// @Tags         登录日志
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "登录日志ID"
// @Success      200  {object}  v1.Response[v1.SysLogininforResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/logininfor/{id} [get]
// @Id        system:logininfor:detail
func (h *SysLogininforHandler) Get(c *gin.Context) {
	id := c.Param("id")
	idUint := cast.ToUint(id)

	logininfor, err := h.sysLogininforService.FindByID(c, idUint)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(c, logininfor)
}

// List godoc
// @Summary      获取登录日志列表
// @Description  分页获取登录日志列表
// @Tags         登录日志
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "当前页"
// @Param        size       query     int     false  "每页数量"
// @Param        login_name query     string  false  "登录账号"
// @Param        ip_addr    query     string  false  "登录IP地址"
// @Param        status     query     string  false  "登录状态"
// @Param        start_time query     string  false  "开始时间"
// @Param        end_time   query     string  false  "结束时间"
// @Success      200        {object}  v1.Response[v1.PageResponse[v1.SysLogininforResponse]]
// @Failure      400        {object}  v1.Response[any]
// @Failure      500        {object}  v1.Response[any]
// @Router       /api/v1/system/logininfor/list [get]
// @Id        system:logininfor:list
func (h *SysLogininforHandler) List(c *gin.Context) {
	query := &v1.SysLogininforQueryParams{}
	query.BasePageQuery = &v1.BasePageQuery{}

	query.LoginName = c.Query("login_name")
	query.IPAddr = c.Query("ip_addr")
	query.Status = c.Query("status")
	query.StartTime = c.Query("start_time")
	query.EndTime = c.Query("end_time")

	query.Page = cast.ToInt(c.Query("page"))
	query.Size = cast.ToInt(c.Query("size"))

	logininfor, total, err := h.sysLogininforService.FindWithPagination(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(c, v1.NewPageResponse(query.Page, query.Size, total, logininfor))
}
