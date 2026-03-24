package handler

import (
	"context"

	"github.com/gin-gonic/gin"
)

// HasPermission 检查当前请求是否有指定权限
func HasPermission(c *gin.Context, permission string) bool {
	// 获取上下文中的权限列表
	perms, exists := c.Get("permissions")
	if !exists {
		return false
	}

	permList, ok := perms.([]string)
	if !ok {
		return false
	}

	for _, p := range permList {
		if p == permission {
			return true
		}
	}

	return false
}

// GetAccessibleDeptIDs 获取当前用户可访问的部门ID
func GetAccessibleDeptIDs(c *gin.Context) []uint {
	deptIDs, exists := c.Get("accessible_dept_ids")
	if !exists {
		return nil
	}

	result, ok := deptIDs.([]uint)
	if !ok {
		return nil
	}

	return result
}

// GetUserID 获取当前用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	result, ok := userID.(uint)
	if !ok {
		return 0
	}

	return result
}

// GetUserDeptID 获取当前用户部门ID
func GetUserDeptID(c *gin.Context) uint {
	deptID, exists := c.Get("dept_id")
	if !exists {
		return 0
	}

	result, ok := deptID.(uint)
	if !ok {
		return 0
	}

	return result
}

// InjectUserPermissions 注入用户权限信息到上下文
func InjectUserPermissions(c *gin.Context, permissionService interface{}) {
	// 如果已经注入了权限，则跳过
	if _, exists := c.Get("permissions"); exists {
		return
	}

	// 获取用户ID和角色ID
	userID := GetUserID(c)
	roleIDs := GetUserRoleIDs(c)

	// 使用permissionService获取用户权限
	if permService, ok := permissionService.(interface {
		GetUserPermissions(ctx context.Context, userID uint, roleIDs []uint) ([]string, error)
	}); ok {
		permissions, err := permService.GetUserPermissions(c, userID, roleIDs)
		if err == nil {
			c.Set("permissions", permissions)
		}
	}
}

// GetUserRoleIDs 获取当前用户角色ID列表
func GetUserRoleIDs(c *gin.Context) []uint {
	roleIDs, exists := c.Get("role_ids")
	if !exists {
		return []uint{}
	}

	result, ok := roleIDs.([]uint)
	if !ok {
		return []uint{}
	}

	return result
}
