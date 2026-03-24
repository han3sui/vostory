package repository

import (
	"context"

	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysUserRoleRepository interface {
	Create(ctx context.Context, userRole *model.SysUserRole) error
	CreateBatch(ctx context.Context, userRoles []*model.SysUserRole) error
	DeleteByUserID(ctx context.Context, userID uint) error
	FindRoleIDsByUserID(ctx context.Context, userID uint) ([]uint, error)
	FindUserIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error)
}

type sysUserRoleRepository struct {
	db *gorm.DB
}

func NewSysUserRoleRepository(db *gorm.DB) SysUserRoleRepository {
	return &sysUserRoleRepository{db: db}
}

func (r *sysUserRoleRepository) Create(ctx context.Context, userRole *model.SysUserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *sysUserRoleRepository) CreateBatch(ctx context.Context, userRoles []*model.SysUserRole) error {
	if len(userRoles) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&userRoles).Error
}

func (r *sysUserRoleRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.SysUserRole{}).Error
}

func (r *sysUserRoleRepository) FindRoleIDsByUserID(ctx context.Context, userID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysUserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *sysUserRoleRepository) FindUserIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysUserRole{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
