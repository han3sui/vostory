package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysLogininforRepository interface {
	Create(ctx context.Context, logininfor *model.SysLogininfor) error
	FindByID(ctx context.Context, id uint) (*model.SysLogininfor, error)
	FindWithPagination(ctx context.Context, query *v1.SysLogininforQueryParams) ([]*model.SysLogininfor, int64, error)
}

type sysLogininforRepository struct {
	db *gorm.DB
}

func NewSysLogininforRepository(db *gorm.DB) SysLogininforRepository {
	return &sysLogininforRepository{
		db: db,
	}
}

func (r *sysLogininforRepository) Create(ctx context.Context, logininfor *model.SysLogininfor) error {
	return r.db.WithContext(ctx).Create(logininfor).Error
}

func (r *sysLogininforRepository) FindByID(ctx context.Context, id uint) (*model.SysLogininfor, error) {
	var logininfor model.SysLogininfor
	if err := r.db.WithContext(ctx).Where("info_id = ?", id).First(&logininfor).Error; err != nil {
		return nil, err
	}
	return &logininfor, nil
}

func (r *sysLogininforRepository) FindWithPagination(ctx context.Context, query *v1.SysLogininforQueryParams) ([]*model.SysLogininfor, int64, error) {
	var logininfor []*model.SysLogininfor
	db := r.db.WithContext(ctx).Model(&model.SysLogininfor{})

	if query.LoginName != "" {
		db = db.Where("login_name LIKE ?", "%"+query.LoginName+"%")
	}

	if query.IPAddr != "" {
		db = db.Where("ipaddr LIKE ?", "%"+query.IPAddr+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.StartTime != "" {
		db = db.Where("login_time >= ?", query.StartTime)
	}

	if query.EndTime != "" {
		db = db.Where("login_time <= ?", query.EndTime)
	}

	// 查询总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 排序
	if err := db.
		Order("login_time DESC").
		Find(&logininfor).Error; err != nil {
		return nil, 0, err
	}
	return logininfor, total, nil
}
