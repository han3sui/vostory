package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsLLMLogRepository interface {
	Create(ctx context.Context, log *model.VsLLMLog) error
	FindByID(ctx context.Context, id uint64) (*model.VsLLMLog, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMLogListQuery) ([]*model.VsLLMLog, int64, error)
	Delete(ctx context.Context, id uint64) error
}

func NewVsLLMLogRepository(repository *Repository) VsLLMLogRepository {
	return &vsLLMLogRepository{Repository: repository}
}

type vsLLMLogRepository struct {
	*Repository
}

func (r *vsLLMLogRepository) Create(ctx context.Context, log *model.VsLLMLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *vsLLMLogRepository) FindByID(ctx context.Context, id uint64) (*model.VsLLMLog, error) {
	var log model.VsLLMLog
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("LLMProvider").
		Preload("PromptTemplate").
		Where("log_id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *vsLLMLogRepository) FindWithPagination(ctx context.Context, query *v1.VsLLMLogListQuery) ([]*model.VsLLMLog, int64, error) {
	var logs []*model.VsLLMLog
	db := r.db.WithContext(ctx).Model(&model.VsLLMLog{})

	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.ProviderID > 0 {
		db = db.Where("provider_id = ?", query.ProviderID)
	}
	if query.ModelName != "" {
		db = db.Where("model_name LIKE ?", "%"+query.ModelName+"%")
	}
	if query.Status >= 0 {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Preload("Project").Preload("LLMProvider").Preload("PromptTemplate").
		Order("log_id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *vsLLMLogRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("log_id = ?", id).Delete(&model.VsLLMLog{}).Error
}
