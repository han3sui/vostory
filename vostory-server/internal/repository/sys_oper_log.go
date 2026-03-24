package repository

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type SysOperLogRepository interface {
	Create(ctx context.Context, operLog *model.SysOperLog) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysOperLog, error)
	FindWithPagination(ctx context.Context, query *v1.SysOperLogListQuery) ([]*model.SysOperLog, int64, error)
	Clean(ctx context.Context) error
}

func NewSysOperLogRepository(
	repository *Repository,
) SysOperLogRepository {
	return &sysOperLogRepository{
		Repository: repository,
	}
}

type sysOperLogRepository struct {
	*Repository
}

func (r *sysOperLogRepository) Create(ctx context.Context, operLog *model.SysOperLog) error {
	return r.db.WithContext(ctx).Create(operLog).Error
}

func (r *sysOperLogRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SysOperLog{}, id).Error
}

func (r *sysOperLogRepository) FindByID(ctx context.Context, id uint) (*model.SysOperLog, error) {
	var operLog model.SysOperLog
	if err := r.db.WithContext(ctx).First(&operLog, id).Error; err != nil {
		return nil, err
	}
	return &operLog, nil
}

func (r *sysOperLogRepository) FindWithPagination(ctx context.Context, query *v1.SysOperLogListQuery) ([]*model.SysOperLog, int64, error) {
	var operLogs []*model.SysOperLog
	db := r.db.WithContext(ctx).Model(&model.SysOperLog{})

	if query.Title != "" {
		db = db.Where("title LIKE ?", "%"+query.Title+"%")
	}

	if query.BusinessType != "" {
		db = db.Where("business_type = ?", query.BusinessType)
	}

	if query.OperName != "" {
		db = db.Where("oper_name LIKE ?", "%"+query.OperName+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.BeginTime != "" {
		db = db.Where("oper_time >= ?", query.BeginTime)
	}

	if query.EndTime != "" {
		db = db.Where("oper_time <= ?", query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Order("oper_id DESC").Find(&operLogs).Error; err != nil {
		return nil, 0, err
	}

	return operLogs, total, nil
}

func (r *sysOperLogRepository) Clean(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&model.SysOperLog{}).Error
}
