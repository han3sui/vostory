package v1

import (
	"time"
)

// SysUserListQuery 系统用户列表查询参数
type SysUserListQuery struct {
	*BasePageQuery
	LoginName   string `json:"login_name"`  // 登录账号
	UserName    string `json:"user_name"`   // 用户昵称
	Status      string `json:"status"`      // 账号状态
	DeptID      *uint  `json:"dept_id"`     // 部门ID
	Phonenumber string `json:"phonenumber"` // 手机号码
	Email       string `json:"email"`       // 邮箱
}

type BaseSysUserRequest struct {
	DeptID      *uint  `json:"dept_id"`                       // 部门ID
	SuperiorID  *uint  `json:"superior_id"`                   // 直属上级ID
	LoginName   string `json:"login_name" binding:"required"` // 登录账号
	UserName    string `json:"user_name" binding:"required"`  // 用户昵称
	UserType    string `json:"user_type"`                     // 用户类型
	Email       string `json:"email"`                         // 用户邮箱
	Phonenumber string `json:"phonenumber"`                   // 手机号码
	Sex         string `json:"sex"`                           // 用户性别
	Avatar      string `json:"avatar"`                        // 头像路径
	Status      string `json:"status" binding:"required"`     // 账号状态
	RoleIDs     []uint `json:"role_ids"`                      // 角色ID列表
	PostIDs     []uint `json:"post_ids"`                      // 岗位ID列表
	Remark      string `json:"remark"`                        // 备注
}

// SysUserCreateRequest 创建系统用户请求
type SysUserCreateRequest struct {
	BaseSysUserRequest
	Password string `json:"password" binding:"required"` // 密码

}

// SysUserUpdateRequest 更新系统用户请求
type SysUserUpdateRequest struct {
	BaseSysUserRequest
	UserID   uint   `json:"user_id" binding:"required"`
	Password string `json:"password"`
}

// SysUserDetailResponse 系统用户详情响应
type SysUserDetailResponse struct {
	UserID        uint        `json:"user_id"`
	DeptID        *uint       `json:"dept_id"`
	DeptName      string      `json:"dept_name,omitempty"`    // 部门名称
	SuperiorID    *uint       `json:"superior_id"`            // 直属上级ID
	SuperiorName  string      `json:"superior_name,omitempty"` // 直属上级姓名
	LoginName     string      `json:"login_name"`
	UserName      string      `json:"user_name"`
	UserType      string      `json:"user_type"`
	Email         string      `json:"email"`
	Phonenumber   string      `json:"phonenumber"`
	Sex           string      `json:"sex"`
	Avatar        string      `json:"avatar"`
	Status        string      `json:"status"`
	LoginIP       string      `json:"login_ip"`
	LoginDate     *time.Time  `json:"login_date"`
	PwdUpdateDate *time.Time  `json:"pwd_update_date"`
	CreateBy      string      `json:"create_by"`
	UpdateBy      string      `json:"update_by"`
	Remark        string      `json:"remark"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	RoleIDs       []uint      `json:"role_ids,omitempty"` // 用户角色ID列表
	PostIDs       []uint      `json:"post_ids,omitempty"` // 用户岗位ID列表
	Roles         interface{} `json:"roles,omitempty"`    // 用户角色列表
	Posts         interface{} `json:"posts,omitempty"`    // 用户岗位列表
	Dept          interface{} `json:"dept,omitempty"`     // 用户部门
	Superior      interface{} `json:"superior,omitempty"` // 直属上级
}

// SysUserResetPasswordRequest 重置密码请求
type SysUserResetPasswordRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// SysUserChangeStatusRequest 修改用户状态请求
type SysUserChangeStatusRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Status string `json:"status" binding:"required"`
}

type SysUserUpdatePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type SysUserUpdateCurrentPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type SysUserLoginRequest struct {
	LoginName string `json:"login_name" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type SysUserLoginResponse struct {
	Token string `json:"token"`
}

// SysUserImportRequest Excel导入用户请求（单行数据）
type SysUserImportRequest struct {
	LoginName    string `json:"login_name"`    // 登录账号
	UserName     string `json:"user_name"`     // 用户姓名
	DeptName     string `json:"dept_name"`     // 部门名称
	SuperiorName string `json:"superior_name"` // 直属上级（用户名或登录账号）
	Email        string `json:"email"`         // 邮箱
	Phonenumber  string `json:"phonenumber"`   // 手机号码
	Sex          string `json:"sex"`           // 性别（男/女）
	Status       string `json:"status"`        // 状态（正常/停用）
	RoleNames    string `json:"role_names"`    // 角色名称（多个用逗号分隔）
	PostNames    string `json:"post_names"`    // 岗位名称（多个用逗号分隔）
	Password     string `json:"password"`      // 密码（可选，为空则使用默认密码）
}

// SysUserImportError 导入错误信息
type SysUserImportError struct {
	LineNum int      `json:"lineNum"` // 行号
	Errors  []string `json:"errors"`  // 错误信息列表
}

// SysUserImportResponse 导入结果响应
type SysUserImportResponse struct {
	SuccessCount int                  `json:"success_count"` // 成功数量
	FailCount    int                  `json:"fail_count"`    // 失败数量
	Errors       []SysUserImportError `json:"errors"`        // 错误详情
}

type SysUserGetInfoResponse struct {
	User        interface{} `json:"user"`
	Permissions []string    `json:"permissions"`
}

type TokenData struct {
	Exp            int64    `json:"exp"`
	UserId         uint     `json:"user_id"`
	DeptId         uint     `json:"dept_id"`
	RoleIds        []uint   `json:"role_ids"`
	LoginName      string   `json:"login_name"`
	DataScope      string   `json:"data_scope"`
	DataScopeDepts []uint   `json:"data_scope_depts"`
	ApiPermissions []string `json:"api_permissions"`
}

// UserOptionQuery 用户选项查询参数（用于下拉选择）
type UserOptionQuery struct {
	Keyword string `form:"keyword"` // 搜索关键词（用户名/登录名）
	DeptID  *uint  `form:"dept_id"` // 部门ID
	Limit   int    `form:"limit"`   // 返回数量限制，默认50
}

// UserOptionResponse 用户选项响应（简化版，用于下拉选择）
type UserOptionResponse struct {
	UserID    uint   `json:"user_id"`
	UserName  string `json:"user_name"`
	LoginName string `json:"login_name"`
	DeptName  string `json:"dept_name,omitempty"` // 部门名称，方便区分同名用户
}
