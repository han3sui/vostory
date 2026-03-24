package v1

import "time"

// SysMenuListQuery 菜单列表查询参数
type SysMenuListQuery struct {
	*BasePageQuery
	MenuName string `json:"menu_name"` // 菜单名称
	Visible  string `json:"visible"`   // 菜单状态（0显示 1隐藏）
	ParentID *uint  `json:"parent_id"` // 父菜单ID
	MenuType string `json:"menu_type"` // 菜单类型（M目录 C菜单 F按钮）
}

// SysMenuCreateRequest 创建菜单请求
type SysMenuCreateRequest struct {
	ParentID  uint   `json:"parent_id"`                    // 父菜单ID
	MenuName  string `json:"menu_name" binding:"required"` // 菜单名称
	OrderNum  int    `json:"order_num"`                    // 显示顺序
	URL       string `json:"url"`                          // 请求地址
	Target    string `json:"target"`                       // 打开方式
	MenuType  string `json:"menu_type" binding:"required"` // 菜单类型（M目录 C菜单 F按钮）
	Visible   string `json:"visible" binding:"required"`   // 菜单状态（0显示 1隐藏）
	IsRefresh string `json:"is_refresh" `                  // 是否刷新（0刷新 1不刷新）
	Perms     string `json:"perms"`                        // 权限标识
	Icon      string `json:"icon"`                         // 菜单图标
	Remark    string `json:"remark"`                       // 备注
}

type MenuPerms struct {
	ParentID uint   `json:"parent_id"` // 父菜单ID
	MenuName string `json:"menu_name"` // 菜单名称
	Perms    string `json:"perms"`     // 权限标识
}

type SysPermsMenuMutiCreateRequest []MenuPerms

// SysMenuUpdateRequest 更新菜单请求
type SysMenuUpdateRequest struct {
	SysMenuCreateRequest
	ID uint `json:"id" binding:"required"`
}

// SysMenuDetailResponse 菜单详情响应
type SysMenuDetailResponse struct {
	ID        uint                     `json:"id"`
	ParentID  uint                     `json:"parent_id"`
	MenuName  string                   `json:"menu_name"`
	OrderNum  int                      `json:"order_num"`
	URL       string                   `json:"url"`
	Target    string                   `json:"target"`
	MenuType  string                   `json:"menu_type"`
	Visible   string                   `json:"visible"`
	IsRefresh string                   `json:"is_refresh"`
	Perms     string                   `json:"perms"`
	Icon      string                   `json:"icon"`
	CreateBy  string                   `json:"create_by"`
	UpdateBy  string                   `json:"update_by"`
	Remark    string                   `json:"remark"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	Children  []*SysMenuDetailResponse `json:"children,omitempty"` // 子菜单
}

// SysMenuTreeResponse 菜单树形响应
type SysMenuTreeResponse struct {
	ID       uint                   `json:"id"`
	ParentID uint                   `json:"parent_id"`
	MenuName string                 `json:"menu_name"`
	OrderNum int                    `json:"order_num"`
	URL      string                 `json:"url"`
	MenuType string                 `json:"menu_type"`
	Visible  string                 `json:"visible"`
	Perms    string                 `json:"perms"`
	Icon     string                 `json:"icon"`
	Method   string                 `json:"method,omitempty"`
	Children []*SysMenuTreeResponse `json:"children,omitempty"`
}
