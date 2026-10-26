package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsLLMModelRepository interface {
	Create(ctx context.Context, m *model.VsLLMModel) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsLLMModel, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMModelListQuery) ([]*model.VsLLMModel, int64, error)
	CountByProviderID(ctx context.Context, providerID uint64) (int64, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsLLMModelRepository(repository *Repository) VsLLMModelRepository {
	return &vsLLMModelRepository{Repository: repository}
}

type vsLLMModelRepository struct {
	*Repository
}

func (r *vsLLMModelRepository) Create(ctx context.Context, m *model.VsLLMModel) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *vsLLMModelRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.VsLLMModel{}).
		Where("model_id = ?", id).
		Updates(fields).Error
}

func (r *vsLLMModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("model_id = ?", id).Delete(&model.VsLLMModel{}).Error
}

func (r *vsLLMModelRepository) DeleteByIDs(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Where("model_id IN ?", ids).Delete(&model.VsLLMModel{}).Error
}

func (r *vsLLMModelRepository) FindByID(ctx context.Context, id uint64) (*model.VsLLMModel, error) {
	var m model.VsLLMModel
	if err := r.db.WithContext(ctx).Where("model_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *vsLLMModelRepository) FindWithPagination(ctx context.Context, query *v1.VsLLMModelListQuery) ([]*model.VsLLMModel, int64, error) {
	var models []*model.VsLLMModel
	db := r.db.WithContext(ctx).Model(&model.VsLLMModel{})

	if query.ProviderID > 0 {
		db = db.Where("provider_id = ?", query.ProviderID)
	}
	if query.ModelName != "" {
		db = db.Where("model_name LIKE ?", "%"+query.ModelName+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Order("sort_order ASC, model_id DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return models, total, nil
}

func (r *vsLLMModelRepository) CountByProviderID(ctx context.Context, providerID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.VsLLMModel{}).
		Where("provider_id = ?", providerID).Count(&count).Error
	return count, err
}

func (r *vsLLMModelRepository) Enable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsLLMModel{}).
		Where("model_id = ?", id).Update("status", "0").Error
}

func (r *vsLLMModelRepository) Disable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsLLMModel{}).
		Where("model_id = ?", id).Update("status", "1").Error
}
