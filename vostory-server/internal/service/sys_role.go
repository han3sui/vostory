package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"

	"gorm.io/gorm"
)

type SysRoleService interface {
	Create(ctx context.Context, req *v1.SysRoleCreateRequest) error
	Update(ctx context.Context, req *v1.SysRoleUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysRoleDetailResponse, error)
	FindByIDs(ctx context.Context, ids []uint) ([]*v1.SysRoleDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysRoleListQuery) ([]*v1.SysRoleDetailResponse, int64, error)
	// 角色菜单关联方法
	GetRoleMenus(ctx context.Context, roleID uint) (*v1.SysRoleMenuResponse, error)
	UpdateRoleMenus(ctx context.Context, req *v1.SysRoleMenuUpdateRequest) error
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

type sysRoleService struct {
	db           *gorm.DB
	roleRepo     repository.SysRoleRepository
	roleDeptRepo repository.SysRoleDeptRepository
	roleMenuRepo repository.SysRoleMenuRepository
}

func NewSysRoleService(db *gorm.DB, roleRepo repository.SysRoleRepository, roleDeptRepo repository.SysRoleDeptRepository, roleMenuRepo repository.SysRoleMenuRepository) SysRoleService {
	return &sysRoleService{
		db:           db,
		roleRepo:     roleRepo,
		roleDeptRepo: roleDeptRepo,
		roleMenuRepo: roleMenuRepo,
	}
}

func (s *sysRoleService) Create(ctx context.Context, req *v1.SysRoleCreateRequest) error {
	// 检查角色名称是否已存在
	exists, err := s.roleRepo.ExistsByRoleName(ctx, req.RoleName, 0)
	if err != nil {
		return fmt.Errorf("检查角色名称失败: %w", err)
	}
	if exists {
		return fmt.Errorf("角色名称已存在")
	}

	// 检查角色权限字符串是否已存在
	exists, err = s.roleRepo.ExistsByRoleKey(ctx, req.RoleKey, 0)
	if err != nil {
		return fmt.Errorf("检查角色权限字符串失败: %w", err)
	}
	if exists {
		return fmt.Errorf("角色权限字符串已存在")
	}

	// 使用事务处理
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		role := &model.SysRole{
			RoleName:  req.RoleName,
			RoleKey:   req.RoleKey,
			RoleSort:  req.RoleSort,
			DataScope: req.DataScope,
			Status:    req.Status,
			Remark:    req.Remark,
			BaseModel: model.BaseModel{
				CreatedBy: ctx.Value("login_name").(string),
				DeptID:    ctx.Value("dept_id").(uint),
			},
		}

		// 创建角色
		if err := tx.Create(role).Error; err != nil {
			return err
		}

		// 如果是自定数据权限，需要保存角色部门关联
		if req.DataScope == "2" && len(req.DataScopeIds) > 0 {
			var roleDepts []*model.SysRoleDept
			for _, deptID := range req.DataScopeIds {
				roleDepts = append(roleDepts, &model.SysRoleDept{
					RoleID: role.RoleID,
					DeptID: deptID,
				})
			}
			if err := tx.Create(&roleDepts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sysRoleService) Update(ctx context.Context, req *v1.SysRoleUpdateRequest) error {
	// 检查角色是否存在
	_, err := s.roleRepo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 检查角色名称是否已存在（排除当前角色）
	exists, err := s.roleRepo.ExistsByRoleName(ctx, req.RoleName, req.ID)
	if err != nil {
		return fmt.Errorf("检查角色名称失败: %w", err)
	}
	if exists {
		return fmt.Errorf("角色名称已存在")
	}

	// 检查角色权限字符串是否已存在（排除当前角色）
	exists, err = s.roleRepo.ExistsByRoleKey(ctx, req.RoleKey, req.ID)
	if err != nil {
		return fmt.Errorf("检查角色权限字符串失败: %w", err)
	}
	if exists {
		return fmt.Errorf("角色权限字符串已存在")
	}

	// 使用事务处理
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 构建更新的角色信息
		role := &model.SysRole{
			RoleID:    req.ID,
			RoleName:  req.RoleName,
			RoleKey:   req.RoleKey,
			RoleSort:  req.RoleSort,
			DataScope: req.DataScope,
			Status:    req.Status,
			Remark:    req.Remark,
			BaseModel: model.BaseModel{
				UpdatedBy: ctx.Value("login_name").(string),
			},
		}

		if err := tx.Model(&model.SysRole{}).Where("role_id = ?", req.ID).
			Omit("created_by", "created_at", "role_id").
			Updates(role).Error; err != nil {
			return err
		}

		// 删除原有的角色部门关联
		if err := tx.Where("role_id = ?", req.ID).Unscoped().Delete(&model.SysRoleDept{}).Error; err != nil {
			return err
		}

		// 如果是自定数据权限，重新保存角色部门关联
		if req.DataScope == "2" && len(req.DataScopeIds) > 0 {
			var roleDepts []*model.SysRoleDept
			for _, deptID := range req.DataScopeIds {
				roleDepts = append(roleDepts, &model.SysRoleDept{
					RoleID: req.ID,
					DeptID: deptID,
				})
			}
			if err := tx.Create(&roleDepts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sysRoleService) Delete(ctx context.Context, id uint) error {
	// 检查角色是否存在
	_, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 使用事务删除
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除角色部门关联
		if err := tx.Where("role_id = ?", id).Unscoped().Delete(&model.SysRoleDept{}).Error; err != nil {
			return err
		}

		// 删除角色
		if err := tx.Delete(&model.SysRole{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *sysRoleService) FindByID(ctx context.Context, id uint) (*v1.SysRoleDetailResponse, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取关联的部门ID
	deptIDs, err := s.roleDeptRepo.FindDeptIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertToDetailResponse(role, deptIDs), nil
}

func (s *sysRoleService) FindByIDs(ctx context.Context, ids []uint) ([]*v1.SysRoleDetailResponse, error) {
	roles, err := s.roleRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var result []*v1.SysRoleDetailResponse
	for _, role := range roles {
		// 获取关联的部门ID
		deptIDs, err := s.roleDeptRepo.FindDeptIDsByRoleID(ctx, role.RoleID)
		if err != nil {
			continue
		}
		result = append(result, s.convertToDetailResponse(role, deptIDs))
	}

	return result, nil
}

func (s *sysRoleService) FindWithPagination(ctx context.Context, query *v1.SysRoleListQuery) ([]*v1.SysRoleDetailResponse, int64, error) {

	// 查询数据
	roles, total, err := s.roleRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var records []*v1.SysRoleDetailResponse
	for _, role := range roles {
		// 获取关联的部门ID
		deptIDs, err := s.roleDeptRepo.FindDeptIDsByRoleID(ctx, role.RoleID)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, s.convertToDetailResponse(role, deptIDs))
	}

	return records, total, nil
}

// 辅助方法
func (s *sysRoleService) convertToDetailResponse(role *model.SysRole, deptIDs []uint) *v1.SysRoleDetailResponse {
	return &v1.SysRoleDetailResponse{
		RoleID:       role.RoleID,
		RoleName:     role.RoleName,
		RoleKey:      role.RoleKey,
		RoleSort:     role.RoleSort,
		DataScope:    role.DataScope,
		DataScopeIds: deptIDs,
		Status:       role.Status,
		Remark:       role.Remark,
		CreatedAt:    role.CreatedAt,
		UpdatedAt:    role.UpdatedAt,
	}
}

func (s *sysRoleService) GetRoleMenus(ctx context.Context, roleID uint) (*v1.SysRoleMenuResponse, error) {
	// 检查角色是否存在
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}

	// 获取角色关联的菜单ID列表
	menuIDs, err := s.roleMenuRepo.FindMenuIDsByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("获取角色菜单关联失败: %w", err)
	}

	return &v1.SysRoleMenuResponse{
		RoleID:  roleID,
		MenuIDs: menuIDs,
	}, nil
}

func (s *sysRoleService) UpdateRoleMenus(ctx context.Context, req *v1.SysRoleMenuUpdateRequest) error {
	// 检查角色是否存在
	_, err := s.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 使用事务处理
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除原有的角色菜单关联
		if err := tx.Where("role_id = ?", req.RoleID).Delete(&model.SysRoleMenu{}).Error; err != nil {
			return fmt.Errorf("删除原有角色菜单关联失败: %w", err)
		}

		// 如果有新的菜单ID，批量创建关联
		if len(req.MenuIDs) > 0 {
			var roleMenus []*model.SysRoleMenu
			for _, menuID := range req.MenuIDs {
				roleMenus = append(roleMenus, &model.SysRoleMenu{
					RoleID: req.RoleID,
					MenuID: menuID,
				})
			}
			if err := tx.Create(&roleMenus).Error; err != nil {
				return fmt.Errorf("创建角色菜单关联失败: %w", err)
			}
		}

		return nil
	})
}

func (s *sysRoleService) Enable(ctx context.Context, id uint) error {
	return s.roleRepo.Enable(ctx, id)
}

func (s *sysRoleService) Disable(ctx context.Context, id uint) error {
	return s.roleRepo.Disable(ctx, id)
}
