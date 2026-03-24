package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysRoleRepository interface {
	Create(ctx context.Context, role *model.SysRole) error
	Update(ctx context.Context, role *model.SysRole) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysRole, error)
	FindByIDs(ctx context.Context, ids []uint) ([]*model.SysRole, error)
	FindWithPagination(ctx context.Context, query *v1.SysRoleListQuery) ([]*model.SysRole, int64, error)

	FindByRoleKey(ctx context.Context, roleKey string) (*model.SysRole, error)
	FindByRoleName(ctx context.Context, roleName string) (*model.SysRole, error)
	FindByRoleNames(ctx context.Context, roleNames []string) ([]*model.SysRole, error)
	ExistsByRoleKey(ctx context.Context, roleKey string, excludeID uint) (bool, error)
	ExistsByRoleName(ctx context.Context, roleName string, excludeID uint) (bool, error)

	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

type SysRoleQueryParams struct {
	RoleName string
	RoleKey  string
	Status   string
}

type sysRoleRepository struct {
	db *gorm.DB
}

func NewSysRoleRepository(db *gorm.DB) SysRoleRepository {
	return &sysRoleRepository{db: db}
}

func (r *sysRoleRepository) Create(ctx context.Context, role *model.SysRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *sysRoleRepository) Update(ctx context.Context, role *model.SysRole) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Omit("created_by", "created_at", "role_id").Save(role).Error
}

func (r *sysRoleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Delete(&model.SysRole{}, id).Error
}

func (r *sysRoleRepository) FindByID(ctx context.Context, id uint) (*model.SysRole, error) {
	var role model.SysRole
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("role_id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *sysRoleRepository) FindByIDs(ctx context.Context, ids []uint) ([]*model.SysRole, error) {
	var roles []*model.SysRole
	if len(ids) == 0 {
		return roles, nil
	}

	err := r.db.WithContext(ctx).Where("role_id IN ?", ids).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *sysRoleRepository) FindWithPagination(ctx context.Context, query *v1.SysRoleListQuery) ([]*model.SysRole, int64, error) {
	var roles []*model.SysRole
	var total int64

	db := r.buildQuery(ctx, query)

	// 应用数据权限过滤
	db = db.Scopes(model.WithDataScope(ctx))

	// 获取总数
	if err := db.Model(&model.SysRole{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 获取数据
	err := db.Order("role_sort ASC, role_id ASC").Find(&roles).Error
	return roles, total, err
}

func (r *sysRoleRepository) FindByRoleKey(ctx context.Context, roleKey string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("role_key = ?", roleKey).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *sysRoleRepository) FindByRoleName(ctx context.Context, roleName string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.db.WithContext(ctx).Where("role_name = ? AND status = '0'", roleName).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *sysRoleRepository) FindByRoleNames(ctx context.Context, roleNames []string) ([]*model.SysRole, error) {
	var roles []*model.SysRole
	if len(roleNames) == 0 {
		return roles, nil
	}
	err := r.db.WithContext(ctx).Where("role_name IN ? AND status = '0'", roleNames).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *sysRoleRepository) ExistsByRoleKey(ctx context.Context, roleKey string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.SysRole{}).
		Where("role_key = ?", roleKey)

	if excludeID > 0 {
		query = query.Where("role_id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *sysRoleRepository) ExistsByRoleName(ctx context.Context, roleName string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.SysRole{}).
		Where("role_name = ?", roleName)

	if excludeID > 0 {
		query = query.Where("role_id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *sysRoleRepository) buildQuery(ctx context.Context, query *v1.SysRoleListQuery) *gorm.DB {
	db := r.db.WithContext(ctx)

	if query == nil {
		return db
	}

	if query.RoleName != "" {
		db = db.Where("role_name LIKE ?", "%"+query.RoleName+"%")
	}

	if query.RoleKey != "" {
		db = db.Where("role_key LIKE ?", "%"+query.RoleKey+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	return db
}

func (r *sysRoleRepository) Enable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysRole{}).Where("role_id = ?", id).Update("status", "0").Error
}

func (r *sysRoleRepository) Disable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysRole{}).Where("role_id = ?", id).Update("status", "1").Error
}
