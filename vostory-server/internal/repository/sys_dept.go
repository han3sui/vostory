package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysDeptRepository interface {
	Create(ctx context.Context, dept *model.SysDept) error
	Update(ctx context.Context, dept *model.SysDept) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysDept, error)
	FindByDeptName(ctx context.Context, deptName string) (*model.SysDept, error)
	FindWithPagination(ctx context.Context, query *v1.SysDeptListQuery) ([]*model.SysDept, int64, error)
	FindChildren(ctx context.Context, parentID uint) ([]*model.SysDept, error)
	FindByAncestors(ctx context.Context, ancestors string) ([]*model.SysDept, error)
	ExistsWithParent(ctx context.Context, parentID uint) (bool, error)
	CountChildren(ctx context.Context, id uint) (int64, error)

	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

type sysDeptRepository struct {
	db *gorm.DB
}

func NewSysDeptRepository(db *gorm.DB) SysDeptRepository {
	return &sysDeptRepository{
		db: db,
	}
}

func (r *sysDeptRepository) Create(ctx context.Context, dept *model.SysDept) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *sysDeptRepository) Update(ctx context.Context, dept *model.SysDept) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Omit("created_by", "created_at", "dept_id").Save(dept).Error
}

func (r *sysDeptRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("dept_id = ?", id).Delete(&model.SysDept{}).Error
}

func (r *sysDeptRepository) FindByID(ctx context.Context, id uint) (*model.SysDept, error) {
	var dept model.SysDept
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Preload("LeaderUser").
		Where("dept_id = ?", id).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *sysDeptRepository) FindByDeptName(ctx context.Context, deptName string) (*model.SysDept, error) {
	var dept model.SysDept
	err := r.db.WithContext(ctx).
		Where("dept_name = ? AND status = '0'", deptName).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *sysDeptRepository) FindWithPagination(ctx context.Context, query *v1.SysDeptListQuery) ([]*model.SysDept, int64, error) {
	var depts []*model.SysDept
	var total int64

	// 构建基础查询
	db := r.buildQuery(ctx, query)

	// 应用数据权限过滤
	db = db.Scopes(model.WithDataScope(ctx))

	// 查询总数
	if err := db.Model(&model.SysDept{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 排序并预加载负责人信息
	if err := db.Preload("LeaderUser").Order("order_num ASC, dept_id ASC").Find(&depts).Error; err != nil {
		return nil, 0, err
	}
	return depts, total, nil
}

func (r *sysDeptRepository) FindChildren(ctx context.Context, parentID uint) ([]*model.SysDept, error) {
	var depts []*model.SysDept
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Preload("LeaderUser").
		Where("parent_id = ?", parentID).
		Order("order_num ASC, dept_id ASC").Find(&depts).Error
	return depts, err
}

func (r *sysDeptRepository) FindByAncestors(ctx context.Context, ancestors string) ([]*model.SysDept, error) {
	var depts []*model.SysDept
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("ancestors LIKE ?", ancestors+"%").
		Order("order_num ASC, dept_id ASC").Find(&depts).Error
	return depts, err
}

func (r *sysDeptRepository) ExistsWithParent(ctx context.Context, parentID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SysDept{}).
		Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

func (r *sysDeptRepository) CountChildren(ctx context.Context, id uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SysDept{}).
		Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

func (r *sysDeptRepository) buildQuery(ctx context.Context, query *v1.SysDeptListQuery) *gorm.DB {
	db := r.db.WithContext(ctx)

	if query == nil {
		return db
	}

	if query.DeptName != "" {
		db = db.Where("dept_name LIKE ?", "%"+query.DeptName+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.ParentID != nil {
		db = db.Where("parent_id = ?", *query.ParentID)
	}

	return db
}

func (r *sysDeptRepository) Enable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysDept{}).Where("dept_id = ?", id).Update("status", "0").Error
}

func (r *sysDeptRepository) Disable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysDept{}).Where("dept_id = ?", id).Update("status", "1").Error
}
