package repository

import (
	"context"

	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysRoleMenuRepository interface {
	Create(ctx context.Context, roleMenu *model.SysRoleMenu) error
	CreateBatch(ctx context.Context, roleMenus []*model.SysRoleMenu) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
	FindMenuIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error)
	//根据角色ID，查找关联的菜单
	FindMenusByRoleIDs(ctx context.Context, roleIDs []uint) ([]*model.SysMenu, error)
}

type sysRoleMenuRepository struct {
	db *gorm.DB
}

func NewSysRoleMenuRepository(db *gorm.DB) SysRoleMenuRepository {
	return &sysRoleMenuRepository{db: db}
}

func (r *sysRoleMenuRepository) Create(ctx context.Context, roleMenu *model.SysRoleMenu) error {
	return r.db.WithContext(ctx).Create(roleMenu).Error
}

func (r *sysRoleMenuRepository) CreateBatch(ctx context.Context, roleMenus []*model.SysRoleMenu) error {
	if len(roleMenus) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&roleMenus).Error
}

func (r *sysRoleMenuRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&model.SysRoleMenu{}).Error
}

func (r *sysRoleMenuRepository) FindMenuIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	var menuIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysRoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *sysRoleMenuRepository) FindMenusByRoleIDs(ctx context.Context, roleIDs []uint) ([]*model.SysMenu, error) {
	var menuIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysRoleMenu{}).
		Where("role_id IN ?", roleIDs).
		Pluck("menu_id", &menuIDs).Error

	if err != nil {
		return nil, err
	}

	var menus []*model.SysMenu
	err = r.db.WithContext(ctx).
		Model(&model.SysMenu{}).
		Where("menu_id IN ?", menuIDs).
		Find(&menus).Error

	if err != nil {
		return nil, err
	}

	return menus, nil
}
