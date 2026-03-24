package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysMenuRepository interface {
	Create(ctx context.Context, menu *model.SysMenu) error
	Update(ctx context.Context, menu *model.SysMenu) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysMenu, error)
	FindByIDs(ctx context.Context, ids []uint) ([]*model.SysMenu, error)
	FindWithPagination(ctx context.Context, query *v1.SysMenuListQuery) ([]*model.SysMenu, int64, error)

	FindChildren(ctx context.Context, parentID uint) ([]*model.SysMenu, error)
	FindByType(ctx context.Context, menuType string) ([]*model.SysMenu, error)
	FindByPerms(ctx context.Context, perms string) (*model.SysMenu, error)
	FindByPermsList(ctx context.Context, permsList []string) ([]*model.SysMenu, error)
	FindByParentIDAndPermsList(ctx context.Context, parentID uint, permsList []string) ([]*model.SysMenu, error)
	ExistsWithParent(ctx context.Context, parentID uint) (bool, error)
	CountChildren(ctx context.Context, id uint) (int64, error)
	CreateMutiByPerms(ctx context.Context, menus []*model.SysMenu) error
}

type SysMenuQueryParams struct {
	MenuName string
	Visible  string
	ParentID *uint
	MenuType string
}

type sysMenuRepository struct {
	db *gorm.DB
}

func NewSysMenuRepository(db *gorm.DB) SysMenuRepository {
	return &sysMenuRepository{db: db}
}

func (r *sysMenuRepository) Create(ctx context.Context, menu *model.SysMenu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *sysMenuRepository) Update(ctx context.Context, menu *model.SysMenu) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Omit("created_by", "created_at", "menu_id").Updates(menu).Error
}

func (r *sysMenuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("menu_id = ?", id).Delete(&model.SysMenu{}).Error
}

func (r *sysMenuRepository) FindByID(ctx context.Context, id uint) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("menu_id = ?", id).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *sysMenuRepository) FindByIDs(ctx context.Context, ids []uint) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	if len(ids) == 0 {
		return menus, nil
	}

	err := r.db.WithContext(ctx).Where("menu_id IN ?", ids).Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *sysMenuRepository) FindWithPagination(ctx context.Context, query *v1.SysMenuListQuery) ([]*model.SysMenu, int64, error) {
	var menus []*model.SysMenu
	var total int64

	db := r.buildQuery(ctx, query)

	// 应用数据权限过滤
	db = db.Scopes(model.WithDataScope(ctx))

	// 获取总数
	if err := db.Model(&model.SysMenu{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 排序
	err := db.Order("order_num ASC, menu_id ASC").Find(&menus).Error
	return menus, total, err
}

func (r *sysMenuRepository) FindChildren(ctx context.Context, parentID uint) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("parent_id = ?", parentID).
		Order("order_num ASC, menu_id ASC").Find(&menus).Error
	return menus, err
}

func (r *sysMenuRepository) FindByType(ctx context.Context, menuType string) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("menu_type = ?", menuType).
		Order("order_num ASC, menu_id ASC").Find(&menus).Error
	return menus, err
}

func (r *sysMenuRepository) FindByPerms(ctx context.Context, perms string) (*model.SysMenu, error) {
	var menu model.SysMenu
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("perms = ?", perms).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *sysMenuRepository) FindByPermsList(ctx context.Context, permsList []string) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	if len(permsList) == 0 {
		return menus, nil
	}
	err := r.db.WithContext(ctx).Where("perms IN ?", permsList).Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *sysMenuRepository) FindByParentIDAndPermsList(ctx context.Context, parentID uint, permsList []string) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	if len(permsList) == 0 {
		return menus, nil
	}
	err := r.db.WithContext(ctx).Where("parent_id = ? AND perms IN ?", parentID, permsList).Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *sysMenuRepository) ExistsWithParent(ctx context.Context, parentID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SysMenu{}).
		Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

func (r *sysMenuRepository) CountChildren(ctx context.Context, id uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SysMenu{}).
		Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

func (r *sysMenuRepository) buildQuery(ctx context.Context, query *v1.SysMenuListQuery) *gorm.DB {
	db := r.db.WithContext(ctx)

	if query == nil {
		return db
	}

	if query.MenuName != "" {
		db = db.Where("menu_name LIKE ?", "%"+query.MenuName+"%")
	}

	if query.Visible != "" {
		db = db.Where("visible = ?", query.Visible)
	}

	if query.ParentID != nil {
		db = db.Where("parent_id = ?", *query.ParentID)
	}

	if query.MenuType != "" {
		db = db.Where("menu_type = ?", query.MenuType)
	}

	return db
}

func (r *sysMenuRepository) CreateMutiByPerms(ctx context.Context, menus []*model.SysMenu) error {
	return r.db.WithContext(ctx).Create(menus).Error
}
