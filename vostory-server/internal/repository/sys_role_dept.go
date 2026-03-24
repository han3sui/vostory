package repository

import (
	"context"

	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysRoleDeptRepository interface {
	Create(ctx context.Context, roleDept *model.SysRoleDept) error
	CreateBatch(ctx context.Context, roleDepts []*model.SysRoleDept) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
	FindDeptIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error)
	FindRoleIDsByDeptID(ctx context.Context, deptID uint) ([]uint, error)
}

type sysRoleDeptRepository struct {
	db *gorm.DB
}

func NewSysRoleDeptRepository(db *gorm.DB) SysRoleDeptRepository {
	return &sysRoleDeptRepository{db: db}
}

func (r *sysRoleDeptRepository) Create(ctx context.Context, roleDept *model.SysRoleDept) error {
	return r.db.WithContext(ctx).Create(roleDept).Error
}

func (r *sysRoleDeptRepository) CreateBatch(ctx context.Context, roleDepts []*model.SysRoleDept) error {
	if len(roleDepts) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&roleDepts).Error
}

func (r *sysRoleDeptRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Delete(&model.SysRoleDept{}).Error
}

func (r *sysRoleDeptRepository) FindDeptIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	var deptIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysRoleDept{}).
		Where("role_id = ?", roleID).
		Pluck("dept_id", &deptIDs).Error
	return deptIDs, err
}

func (r *sysRoleDeptRepository) FindRoleIDsByDeptID(ctx context.Context, deptID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysRoleDept{}).
		Where("dept_id = ?", deptID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}
