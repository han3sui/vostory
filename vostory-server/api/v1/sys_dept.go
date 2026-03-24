package v1

import "time"

// SysDeptListQuery 部门列表查询参数
type SysDeptListQuery struct {
	*BasePageQuery
	DeptName string `json:"dept_name"` // 部门名称
	Status   string `json:"status"`    // 状态（0正常 1停用）
	ParentID *uint  `json:"parent_id"` // 父部门ID
}

// SysDeptCreateRequest 创建部门请求
type SysDeptCreateRequest struct {
	ParentID *uint  `json:"parent_id" binding:"required"` // 父部门ID
	DeptName string `json:"dept_name" binding:"required"` // 部门名称
	OrderNum int    `json:"order_num" binding:"required"` // 显示顺序
	LeaderID *uint  `json:"leader_id"`                    // 负责人ID
	Leader   string `json:"leader"`                       // 负责人姓名（冗余字段）
	Phone    string `json:"phone"`                        // 联系电话
	Email    string `json:"email"`                        // 邮箱
	Status   string `json:"status" binding:"required"`    // 部门状态（0正常 1停用）
	Remark   string `json:"remark"`                       // 备注
}

// SysDeptUpdateRequest 更新部门请求
type SysDeptUpdateRequest struct {
	SysDeptCreateRequest
	ID uint `json:"id" binding:"required"`
}

// SysDeptDetailResponse 部门详情响应
type SysDeptDetailResponse struct {
	ID         uint                     `json:"id"`
	ParentID   uint                     `json:"parent_id"`
	Ancestors  string                   `json:"ancestors"`
	DeptName   string                   `json:"dept_name"`
	OrderNum   int                      `json:"order_num"`
	LeaderID   *uint                    `json:"leader_id"`
	Leader     string                   `json:"leader"`
	LeaderUser *SysUserBriefResponse    `json:"leader_user,omitempty"` // 负责人详情
	Phone      string                   `json:"phone"`
	Email      string                   `json:"email"`
	Status     string                   `json:"status"`
	Remark     string                   `json:"remark"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	Children   []*SysDeptDetailResponse `json:"children,omitempty"` // 子部门
}

// SysDeptTreeResponse 部门树形响应
type SysDeptTreeResponse struct {
	ID         uint                   `json:"id"`
	ParentID   uint                   `json:"parent_id"`
	DeptName   string                 `json:"dept_name"`
	OrderNum   int                    `json:"order_num"`
	LeaderID   *uint                  `json:"leader_id"`
	Leader     string                 `json:"leader"`
	LeaderUser *SysUserBriefResponse  `json:"leader_user,omitempty"` // 负责人详情
	Status     string                 `json:"status"`
	Children   []*SysDeptTreeResponse `json:"children,omitempty"`
}

// SysUserBriefResponse 用户简要信息响应
type SysUserBriefResponse struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar"`
}

// DeptOptionTreeResponse 部门选项树形响应（简化版，用于下拉选择）
type DeptOptionTreeResponse struct {
	DeptID   uint                      `json:"dept_id"`
	DeptName string                    `json:"dept_name"`
	Children []*DeptOptionTreeResponse `json:"children,omitempty"`
}
