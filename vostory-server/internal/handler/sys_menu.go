package handler

import (
	"net/http"
	"strconv"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysMenuHandler struct {
	menuService service.SysMenuService
}

func NewSysMenuHandler(menuService service.SysMenuService) *SysMenuHandler {
	return &SysMenuHandler{
		menuService: menuService,
	}
}

// CreateMenu godoc
// @Summary      创建菜单
// @Description  创建新的系统菜单
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysMenuCreateRequest  true  "创建菜单请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/menu [post]
// @Id        system:menu:add
func (h *SysMenuHandler) CreateMenu(c *gin.Context) {
	var req v1.SysMenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, err.Error())
		return
	}

	if err := h.menuService.Create(c, &req); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// UpdateMenu godoc
// @Summary      更新菜单
// @Description  更新系统菜单信息
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "菜单ID"
// @Param        request  body      v1.SysMenuUpdateRequest  true  "更新菜单请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/menu/{id} [put]
// @Id        system:menu:edit
func (h *SysMenuHandler) UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid menu ID")
		return
	}

	var req v1.SysMenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, err.Error())
		return
	}

	req.ID = uint(id)

	if err := h.menuService.Update(c, &req); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// DeleteMenu godoc
// @Summary      删除菜单
// @Description  删除系统菜单
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "菜单ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/menu/{id} [delete]
// @Id        system:menu:remove
func (h *SysMenuHandler) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid menu ID")
		return
	}

	if err := h.menuService.Delete(c, uint(id)); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// GetMenu godoc
// @Summary      获取菜单详情
// @Description  根据ID获取菜单详情
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "菜单ID"
// @Success      200  {object}  v1.Response[v1.SysMenuDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/menu/{id} [get]
// @Id        system:menu:detail
func (h *SysMenuHandler) GetMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid menu ID")
		return
	}

	menu, err := h.menuService.FindByID(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, menu)
}

// ListMenus godoc
// @Summary      获取菜单列表
// @Description  分页获取菜单列表
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "页码"
// @Param        size       query     int     false  "每页数量"
// @Param        menu_name  query     string  false  "菜单名称"
// @Param        visible    query     string  false  "显示状态"
// @Param        menu_type  query     string  false  "菜单类型"
// @Success      200        {object}  v1.Response[v1.PageResponse[v1.SysMenuDetailResponse]]
// @Failure      400        {object}  v1.Response[any]
// @Failure      500        {object}  v1.Response[any]
// @Router       /api/v1/system/menu/list [get]
// @Id        system:menu:list
func (h *SysMenuHandler) ListMenus(c *gin.Context) {
	query := &v1.SysMenuListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	page := c.Query("page")
	size := c.Query("size")

	query.Page = cast.ToInt(page)
	query.Size = cast.ToInt(size)

	query.MenuName = c.Query("menu_name")
	query.Visible = c.Query("visible")
	query.MenuType = c.Query("menu_type")

	result, total, err := h.menuService.FindWithPagination(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, v1.NewPageResponse(query.Page, query.Size, total, result))
}

// GetMenuTree godoc
// @Summary      获取菜单树
// @Description  获取菜单树形结构
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        menu_name  query     string  false  "菜单名称"
// @Param        visible    query     string  false  "显示状态"
// @Param        menu_type  query     string  false  "菜单类型"
// @Success      200        {object}  v1.Response[[]v1.SysMenuTreeResponse]
// @Failure      400        {object}  v1.Response[any]
// @Failure      500        {object}  v1.Response[any]
// @Router       /api/v1/system/menu/tree [get]
// @Id        system:menu:tree
func (h *SysMenuHandler) GetMenuTree(c *gin.Context) {
	query := &v1.SysMenuListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	query.MenuName = c.Query("menu_name")
	query.Visible = c.Query("visible")
	query.MenuType = c.Query("menu_type")

	tree, err := h.menuService.GetMenuTree(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	if len(tree) > 0 {
		v1.HandleSuccess(c, tree)
	} else {
		v1.HandleSuccess(c, []map[string]interface{}{})
	}
}

// GetMenusByType godoc
// @Summary      根据类型获取菜单
// @Description  根据菜单类型获取菜单列表
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        type  path      string  true  "菜单类型"
// @Success      200   {object}  v1.Response[[]v1.SysMenuDetailResponse]
// @Failure      400   {object}  v1.Response[any]
// @Failure      500   {object}  v1.Response[any]
// @Router       /api/v1/system/menu/type/{type} [get]
// @Id        system:menu:type
func (h *SysMenuHandler) GetMenusByType(c *gin.Context) {
	menuType := c.Param("type")
	if menuType == "" {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Menu type is required")
		return
	}

	menus, err := h.menuService.GetMenusByType(c, menuType)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, menus)
}

// CreatePermsMuti godoc
// @Summary      创建多个菜单
// @Description  创建多个菜单
// @Tags         菜单管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysPermsMenuMutiCreateRequest  true  "创建多个菜单请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/menu/perms/muti [post]
// @Id        system:menu:muti:add
func (h *SysMenuHandler) CreatePermsMuti(c *gin.Context) {
	var req v1.SysPermsMenuMutiCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, err.Error())
		return
	}

	if err := h.menuService.CreateMutiByPerms(c, &req); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}
