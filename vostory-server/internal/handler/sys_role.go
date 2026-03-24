package handler

import (
	"net/http"
	"strconv"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type SysRoleHandler struct {
	roleService service.SysRoleService
	db          *gorm.DB
}

func NewSysRoleHandler(roleService service.SysRoleService, db *gorm.DB) *SysRoleHandler {
	return &SysRoleHandler{
		roleService: roleService,
		db:          db,
	}
}

// GetRole godoc
// @Summary      获取角色详情
// @Description  根据ID获取角色详情
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response[v1.SysRoleDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id} [get]
// @Id        system:role:detail
func (h *SysRoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	role, err := h.roleService.FindByID(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取角色失败"), err)
		return
	}

	v1.HandleSuccess(c, role)
}

// ListRoles godoc
// @Summary      获取角色列表
// @Description  分页获取角色列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        current    query     int     false  "当前页"
// @Param        size       query     int     false  "每页数量"
// @Param        role_name  query     string  false  "角色名称"
// @Param        role_key   query     string  false  "角色权限字符串"
// @Param        status     query     string  false  "状态"
// @Success      200        {object}  v1.Response[v1.PageResponse[v1.SysRoleDetailResponse]]
// @Failure      400        {object}  v1.Response[any]
// @Failure      500        {object}  v1.Response[any]
// @Router       /api/v1/system/role/list [get]
// @Id        system:role:list
func (h *SysRoleHandler) ListRoles(c *gin.Context) {
	query := &v1.SysRoleListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	page := c.Query("page")
	size := c.Query("size")

	query.Page = cast.ToInt(page)
	query.Size = cast.ToInt(size)

	query.RoleName = c.Query("role_name")
	query.RoleKey = c.Query("role_key")
	query.Status = c.Query("status")

	result, total, err := h.roleService.FindWithPagination(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取角色列表失败"), err)
		return
	}

	v1.HandleSuccess(c, v1.PageResponse{
		Total: total,
		Size:  query.Size,
		Page:  query.Page,
		Data:  result,
	})
}

// CreateRole godoc
// @Summary      创建角色
// @Description  创建新角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysRoleCreateRequest  true  "创建角色请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/role [post]
// @Id        system:role:add
func (h *SysRoleHandler) CreateRole(c *gin.Context) {
	var req v1.SysRoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	err := h.roleService.Create(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "创建角色失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// UpdateRole godoc
// @Summary      更新角色
// @Description  更新角色信息
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "角色ID"
// @Param        request  body      v1.SysRoleCreateRequest  true  "更新角色请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id} [put]
// @Id        system:role:edit
func (h *SysRoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	var req v1.SysRoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	updateReq := v1.SysRoleUpdateRequest{
		SysRoleCreateRequest: req,
		ID:                   uint(id),
	}

	err = h.roleService.Update(c, &updateReq)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "更新角色失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// DeleteRole godoc
// @Summary      删除角色
// @Description  删除指定角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id} [delete]
// @Id        system:role:remove
func (h *SysRoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	err = h.roleService.Delete(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "删除角色失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// GetRoleMenus godoc
// @Summary      获取角色菜单关联
// @Description  获取指定角色关联的菜单ID列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response[v1.SysRoleMenuResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id}/menus [get]
// @Id        system:role:menus
func (h *SysRoleHandler) GetRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	result, err := h.roleService.GetRoleMenus(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取角色菜单关联失败"), err)
		return
	}

	v1.HandleSuccess(c, result.MenuIDs)
}

// UpdateRoleMenus godoc
// @Summary      更新角色菜单关联
// @Description  更新指定角色关联的菜单ID列表
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "角色ID"
// @Param        request  body      v1.SysRoleMenuUpdateRequest  true  "更新角色菜单关联请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id}/menus [put]
// @Id        system:role:edit:menus
func (h *SysRoleHandler) UpdateRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	var req v1.SysRoleMenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	// 确保请求中的角色ID与路径参数一致
	req.RoleID = uint(id)

	err = h.roleService.UpdateRoleMenus(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "更新角色菜单关联失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// EnableRole godoc
// @Summary      启用角色
// @Description  启用指定角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id}/enable [put]
// @Id        system:role:enable
func (h *SysRoleHandler) EnableRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	err = h.roleService.Enable(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "启用角色失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// DisableRole godoc
// @Summary      禁用角色
// @Description  禁用指定角色
// @Tags         角色管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "角色ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/role/{id}/disable [put]
// @Id        system:role:disable
func (h *SysRoleHandler) DisableRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的角色ID"), err)
		return
	}

	err = h.roleService.Disable(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "禁用角色失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// GetRoleOptions godoc
// @Summary      获取角色选项列表
// @Description  获取简化的角色列表，用于下拉选择（指定审批人角色等场景）
// @Tags         通用接口
// @Accept       json
// @Produce      json
// @Param        keyword  query     string  false  "搜索关键词（角色名称）"
// @Param        limit    query     int     false  "返回数量限制"  default(100)
// @Success      200      {object}  v1.Response[[]v1.RoleOptionResponse]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/common/role/options [get]
// @Whitelist    true
func (h *SysRoleHandler) GetRoleOptions(c *gin.Context) {
	query := &v1.RoleOptionQuery{}

	query.Keyword = c.Query("keyword")

	limit := c.Query("limit")
	if limit != "" {
		query.Limit = cast.ToInt(limit)
	}
	if query.Limit <= 0 || query.Limit > 500 {
		query.Limit = 100 // 默认100，最大500
	}

	db := h.db.WithContext(c).Table("sys_role").
		Select("role_id, role_name").
		Where("status = ?", "0") // 只查询启用状态的角色

	// 关键词搜索（角色名称）
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("role_name LIKE ?", keyword)
	}

	var roles []*v1.RoleOptionResponse
	err := db.Order("role_sort ASC").Limit(query.Limit).Find(&roles).Error
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取角色列表失败"), err)
		return
	}

	v1.HandleSuccess(c, roles)
}
