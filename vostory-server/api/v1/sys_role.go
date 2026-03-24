package v1

import "time"

// SysRoleListQuery 角色列表查询参数
type SysRoleListQuery struct {
	*BasePageQuery
	RoleName string `json:"role_name"` // 角色名称
	RoleKey  string `json:"role_key"`  // 角色权限字符串
	Status   string `json:"status"`    // 角色状态（0正常 1停用）
}

// SysRoleCreateRequest 创建角色请求
type SysRoleCreateRequest struct {
	RoleName     string `json:"role_name" binding:"required"`  // 角色名称
	RoleKey      string `json:"role_key" binding:"required"`   // 角色权限字符串
	RoleSort     int    `json:"role_sort" binding:"required"`  // 显示顺序
	DataScope    string `json:"data_scope" binding:"required"` // 数据范围
	DataScopeIds []uint `json:"data_scope_ids"`                // 数据范围ID,关联sys_dept表
	Status       string `json:"status" binding:"required"`     // 角色状态（0正常 1停用）
	Remark       string `json:"remark"`                        // 备注
}

// SysRoleUpdateRequest 更新角色请求
type SysRoleUpdateRequest struct {
	SysRoleCreateRequest
	ID uint `json:"id" binding:"required"`
}

// SysRoleDetailResponse 角色详情响应
type SysRoleDetailResponse struct {
	RoleID       uint      `json:"role_id"`
	RoleName     string    `json:"role_name"`
	RoleKey      string    `json:"role_key"`
	RoleSort     int       `json:"role_sort"`
	DataScope    string    `json:"data_scope"`
	DataScopeIds []uint    `json:"data_scope_ids"` // 数据范围ID，关联的部门ID列表
	Status       string    `json:"status"`
	CreateBy     string    `json:"create_by"`
	UpdateBy     string    `json:"update_by"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SysRoleMenuUpdateRequest 更新角色菜单关联请求
type SysRoleMenuUpdateRequest struct {
	RoleID  uint   `json:"role_id" binding:"required"`  // 角色ID
	MenuIDs []uint `json:"menu_ids" binding:"required"` // 菜单ID列表
}

// SysRoleMenuResponse 角色菜单关联响应
type SysRoleMenuResponse struct {
	RoleID  uint   `json:"role_id"`  // 角色ID
	MenuIDs []uint `json:"menu_ids"` // 菜单ID列表
}

// RoleOptionQuery 角色选项查询参数
type RoleOptionQuery struct {
	Keyword string `json:"keyword"` // 搜索关键词（角色名称）
	Limit   int    `json:"limit"`   // 返回数量限制
}

// RoleOptionResponse 角色选项响应（简化版，用于下拉选择）
type RoleOptionResponse struct {
	RoleID   uint   `json:"role_id" gorm:"column:role_id"`
	RoleName string `json:"role_name" gorm:"column:role_name"`
}
