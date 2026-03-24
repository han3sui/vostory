package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type SysUserHandler struct {
	userService service.SysUserService
}

func NewSysUserHandler(userService service.SysUserService) *SysUserHandler {
	return &SysUserHandler{
		userService: userService,
	}
}

// GetUser godoc
// @Summary      获取用户详情
// @Description  根据ID获取用户详情
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  v1.Response[v1.SysUserDetailResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id} [get]
// @Id        system:user:detail
func (h *SysUserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), err)
		return
	}

	user, err := h.userService.FindByID(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取用户失败"), err)
		return
	}

	v1.HandleSuccess(c, user)
}

// ListUsers godoc
// @Summary      获取用户列表
// @Description  分页获取用户列表
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        current     query     int     false  "当前页"
// @Param        size        query     int     false  "每页数量"
// @Param        login_name  query     string  false  "登录账号"
// @Param        user_name   query     string  false  "用户昵称"
// @Param        status      query     string  false  "账号状态"
// @Param        dept_id     query     int     false  "部门ID"
// @Param        phonenumber query     string  false  "手机号码"
// @Param        email       query     string  false  "邮箱"
// @Success      200         {object}  v1.Response[v1.PageResponse[v1.SysUserDetailResponse]]
// @Failure      400         {object}  v1.Response[any]
// @Failure      500         {object}  v1.Response[any]
// @Router       /api/v1/system/user/list [get]
// @Id        system:user:list
func (h *SysUserHandler) ListUsers(c *gin.Context) {
	query := &v1.SysUserListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	page := c.Query("page")
	size := c.Query("size")

	query.Page = cast.ToInt(page)
	query.Size = cast.ToInt(size)

	query.LoginName = c.Query("login_name")
	query.UserName = c.Query("user_name")
	query.Status = c.Query("status")
	query.Phonenumber = c.Query("phonenumber")
	query.Email = c.Query("email")

	deptID := c.Query("dept_id")
	if deptID != "" {
		deptIDUint := cast.ToUint(deptID)
		query.DeptID = &deptIDUint
	}

	result, total, err := h.userService.FindWithPagination(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取用户列表失败"), err)
		return
	}

	v1.HandleSuccess(c, v1.PageResponse{
		Total: total,
		Size:  query.Size,
		Page:  query.Page,
		Data:  result,
	})
}

// CreateUser godoc
// @Summary      创建用户
// @Description  创建新用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysUserCreateRequest  true  "创建用户请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/user [post]
// @Id        system:user:add
func (h *SysUserHandler) CreateUser(c *gin.Context) {
	var req v1.SysUserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	if req.UserType == "99" {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "非法用户类型"), errors.New("非法用户类型"))
		return
	}

	err := h.userService.Create(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "创建用户失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// UpdateUser godoc
// @Summary      更新用户
// @Description  更新用户信息
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "用户ID"
// @Param        request  body      v1.SysUserCreateRequest  true  "更新用户请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id} [put]
// @Id        system:user:edit
func (h *SysUserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), err)
		return
	}

	var req v1.SysUserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	if req.UserType == "99" {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "非法用户类型"), errors.New("非法用户类型"))
		return
	}

	req.UserID = uint(id)

	err = h.userService.Update(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "更新用户失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// DeleteUser godoc
// @Summary      删除用户
// @Description  删除指定用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id} [delete]
// @Id        system:user:remove
func (h *SysUserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), err)
		return
	}

	err = h.userService.Delete(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "删除用户失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// ResetPassword godoc
// @Summary      重置用户密码
// @Description  重置指定用户的密码
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                              true  "用户ID"
// @Param        request  body      v1.SysUserResetPasswordRequest  true  "重置密码请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id}/reset-password [put]
// @Id        system:user:resetpwd
func (h *SysUserHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), err)
		return
	}

	var req v1.SysUserResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	// 确保请求中的用户ID与路径参数一致
	req.UserID = uint(id)

	err = h.userService.ResetPassword(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "重置密码失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// ChangeStatus godoc
// @Summary      修改用户状态
// @Description  修改指定用户的启用/停用状态
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                             true  "用户ID"
// @Param        request  body      v1.SysUserChangeStatusRequest  true  "修改状态请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id}/status [put]
// @Id        system:user:change:status
func (h *SysUserHandler) ChangeStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), err)
		return
	}

	var req v1.SysUserChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	// 确保请求中的用户ID与路径参数一致
	req.UserID = uint(id)

	err = h.userService.ChangeStatus(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "修改用户状态失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// EnableUser godoc
// @Summary      启用用户
// @Description  启用指定用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id}/enable [put]
// @Id        system:user:enable
func (h *SysUserHandler) EnableUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), errors.New("无效的用户ID"))
		return
	}
	id := cast.ToUint(idStr)

	err := h.userService.Enable(c, id)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "启用用户失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// DisableUser godoc
// @Summary      禁用用户
// @Description  禁用指定用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id}/disable [put]
// @Id        system:user:disable
func (h *SysUserHandler) DisableUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), errors.New("无效的用户ID"))
		return
	}
	id := cast.ToUint(idStr)

	err := h.userService.Disable(c, id)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "禁用用户失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// UpdatePassword godoc
// @Summary      更新用户密码
// @Description  更新指定用户的密码
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                             true  "用户ID"
// @Param        request  body      v1.SysUserUpdatePasswordRequest  true  "更新密码请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/system/user/{id}/update-password [put]
// @Id        system:user:updatepwd
func (h *SysUserHandler) UpdatePassword(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "无效的用户ID"), errors.New("无效的用户ID"))
		return
	}
	id := cast.ToUint(idStr)

	var req v1.SysUserUpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	err := h.userService.UpdatePassword(c, id, req.Password)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "更新密码失败"), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// UpdateCurrentPassword godoc
// @Summary      修改当前用户密码
// @Description  当前登录用户修改自己的密码
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysUserUpdateCurrentPasswordRequest  true  "修改密码请求"
// @Success      200      {object}  v1.Response[any]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/user/update-password [put]
// @Whitelist    true
func (h *SysUserHandler) UpdateCurrentPassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		v1.HandleError(c, http.StatusUnauthorized, v1.NewError(401, "未授权"), errors.New("未授权"))
		return
	}

	var req v1.SysUserUpdateCurrentPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	err := h.userService.UpdateCurrentPassword(c, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, err.Error()), err)
		return
	}

	v1.HandleSuccess(c, nil)
}

// GetUserInfo godoc
// @Summary      获取当前用户信息
// @Description  获取当前登录用户的详细信息和权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  v1.Response[v1.SysUserGetInfoResponse]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/user/info [get]
// @Id        user:info:profile
func (h *SysUserHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, menus, err := h.userService.GetUserInfo(c, userID)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取用户信息失败"), err)
		return
	}

	//删除密码
	user.Password = ""

	permissions := make([]string, 0)
	for _, menu := range menus {
		if menu.MenuType == "F" {
			permissions = append(permissions, menu.Perms)
		} else {
			permissions = append(permissions, menu.URL)
		}
	}

	resp := v1.SysUserGetInfoResponse{
		User:        user,
		Permissions: permissions,
	}

	v1.HandleSuccess(c, resp)
}

// Login godoc
// @Summary      用户登录
// @Description  用户登录接口
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        request  body      v1.SysUserLoginRequest  true  "登录请求"
// @Success      200      {object}  v1.Response[map[string]string]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/user/login [post]
// @Whitelist    true
func (h *SysUserHandler) Login(c *gin.Context) {
	var req v1.SysUserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "参数错误"), err)
		return
	}

	token, err := h.userService.Login(c, &req)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "登录失败"), err)
		return
	}

	v1.HandleSuccess(c, map[string]string{
		"token": token,
	})
}

// Logout godoc
// @Summary      退出登录
// @Description  退出登录
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  v1.Response[any]
// @Failure      400  {object}  v1.Response[any]
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/user/logout [post]
// @Whitelist    true
func (h *SysUserHandler) Logout(c *gin.Context) {
	// 从请求头获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(400, "token不能为空"), nil)
		return
	}

	// 去除Bearer前缀
	token = strings.Replace(token, "Bearer ", "", 1)
	err := h.userService.Logout(c, token)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "退出登录失败"), err)
		return
	}
	v1.HandleSuccess(c, nil)
}

// ImportUsers godoc
// @Summary      导入用户
// @Description  通过Excel文件批量导入用户
// @Tags         用户管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Excel文件"
// @Success      200   {object}  v1.Response[v1.SysUserImportResponse]
// @Failure      400   {object}  v1.Response[any]
// @Failure      500   {object}  v1.Response[any]
// @Router       /api/v1/system/user/import [post]
// @Id        system:user:import
func (h *SysUserHandler) ImportUsers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "请上传文件"), err)
		return
	}

	// 检查文件类型
	if !strings.HasSuffix(file.Filename, ".xlsx") && !strings.HasSuffix(file.Filename, ".xls") {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "仅支持Excel文件(.xlsx, .xls)"), nil)
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "读取文件失败"), err)
		return
	}
	defer src.Close()

	fileData := make([]byte, file.Size)
	_, err = src.Read(fileData)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "读取文件内容失败"), err)
		return
	}

	// 调用服务导入用户
	result, err := h.userService.ImportUsers(c, fileData)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "导入失败"), err)
		return
	}

	// 如果有失败的记录，返回错误详情
	if result.FailCount > 0 {
		v1.HandleError(c, http.StatusBadRequest, v1.NewError(400, "部分数据导入失败"), result.Errors)
		return
	}

	v1.HandleSuccess(c, result)
}

// DownloadImportTemplate godoc
// @Summary      下载用户导入模板
// @Description  下载用户导入Excel模板
// @Tags         用户管理
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success      200  {file}    file
// @Failure      500  {object}  v1.Response[any]
// @Router       /api/v1/system/user/import/template [get]
// @Id        system:user:import:template
func (h *SysUserHandler) DownloadImportTemplate(c *gin.Context) {
	data, err := h.userService.GenerateImportTemplate()
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "生成模板失败"), err)
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=user_import_template.xlsx")
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// GetUserOptions godoc
// @Summary      获取用户选项列表
// @Description  获取简化的用户列表，用于下拉选择（转办、指定审批人等场景）
// @Tags         通用接口
// @Accept       json
// @Produce      json
// @Param        keyword  query     string  false  "搜索关键词（用户名/登录名）"
// @Param        dept_id  query     int     false  "部门ID"
// @Param        limit    query     int     false  "返回数量限制"  default(50)
// @Success      200      {object}  v1.Response[[]v1.UserOptionResponse]
// @Failure      400      {object}  v1.Response[any]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/common/user/options [get]
// @Whitelist    true
func (h *SysUserHandler) GetUserOptions(c *gin.Context) {
	query := &v1.UserOptionQuery{}

	query.Keyword = c.Query("keyword")

	deptID := c.Query("dept_id")
	if deptID != "" {
		deptIDUint := cast.ToUint(deptID)
		query.DeptID = &deptIDUint
	}

	limit := c.Query("limit")
	if limit != "" {
		query.Limit = cast.ToInt(limit)
	}
	if query.Limit <= 0 || query.Limit > 200 {
		query.Limit = 50 // 默认50，最大200
	}

	result, err := h.userService.GetUserOptions(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取用户列表失败"), err)
		return
	}

	v1.HandleSuccess(c, result)
}
