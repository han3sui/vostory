package handler

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysApiHandler struct {
	*Handler
	sysApiService service.SysApiService
}

func NewSysApiHandler(
	handler *Handler,
	sysApiService service.SysApiService,
) *SysApiHandler {
	return &SysApiHandler{
		Handler:       handler,
		sysApiService: sysApiService,
	}
}

// ListSysApi godoc
// @Summary      获取API列表
// @Description  分页获取系统API列表
// @Tags         API管理
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "页码"
// @Param        size     query     int     false  "每页数量"
// @Param        method   query     string  false  "请求方法"
// @Param        path     query     string  false  "API路径"
// @Param        name     query     string  false  "API名称"
// @Param        desc     query     string  false  "API描述"
// @Param        tag      query     string  false  "API标签"
// @Success      200      {object}  v1.Response[v1.PageResponse[v1.SysApiResponse]]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/api/list [get]
// @Id        system:api:list
func (h *SysApiHandler) ListSysApi(c *gin.Context) {
	query := &v1.SysApiListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	page := c.Query("page")
	size := c.Query("size")

	query.Page = cast.ToInt(page)
	query.Size = cast.ToInt(size)

	query.Method = c.Query("method")
	query.Path = c.Query("path")
	query.Name = c.Query("name")
	query.Desc = c.Query("desc")
	query.Tag = c.Query("tag")
	query.Perms = c.Query("perms")

	apis, total, err := h.sysApiService.FindWithPagination(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取API列表失败"), err.Error())
		return
	}

	v1.HandleSuccess(c, v1.NewPageResponse(query.Page, query.Size, total, apis))
}

// ListTag godoc
// @Summary      获取API标签列表
// @Description  获取系统API标签列表
// @Tags         API管理
// @Accept       json
// @Produce      json
// @Success      200      {object}  v1.Response[[]string]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/api/tag/list [get]
// @Id system:api:tag:list
func (h *SysApiHandler) ListTag(c *gin.Context) {
	tags, err := h.sysApiService.ListTag(c)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取API标签列表失败"), err.Error())
		return
	}
	v1.HandleSuccess(c, tags)
}
