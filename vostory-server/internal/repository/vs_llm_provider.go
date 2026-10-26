package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsLLMProviderRepository interface {
	Create(ctx context.Context, provider *model.VsLLMProvider) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsLLMProvider, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMProviderListQuery) ([]*model.VsLLMProvider, int64, error)
	FindAll(ctx context.Context) ([]*model.VsLLMProvider, error)
	FindAllEnabled(ctx context.Context) ([]*model.VsLLMProvider, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsLLMProviderRepository(repository *Repository) VsLLMProviderRepository {
	return &vsLLMProviderRepository{Repository: repository}
}

type vsLLMProviderRepository struct {
	*Repository
}

func (r *vsLLMProviderRepository) Create(ctx context.Context, provider *model.VsLLMProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *vsLLMProviderRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.VsLLMProvider{}).
		Where("provider_id = ?", id).
		Updates(fields).Error
}

func (r *vsLLMProviderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("provider_id = ?", id).Delete(&model.VsLLMProvider{}).Error
}

func (r *vsLLMProviderRepository) DeleteByIDs(ctx context.Context, ids []uint64) error {
	return r.db.WithContext(ctx).Where("provider_id IN ?", ids).Delete(&model.VsLLMProvider{}).Error
}

func (r *vsLLMProviderRepository) FindByID(ctx context.Context, id uint64) (*model.VsLLMProvider, error) {
	var provider model.VsLLMProvider
	if err := r.db.WithContext(ctx).Where("provider_id = ?", id).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *vsLLMProviderRepository) FindWithPagination(ctx context.Context, query *v1.VsLLMProviderListQuery) ([]*model.VsLLMProvider, int64, error) {
	var providers []*model.VsLLMProvider
	db := r.db.WithContext(ctx).Model(&model.VsLLMProvider{})

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
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

	if err := db.Order("sort_order ASC, provider_id DESC").Find(&providers).Error; err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *vsLLMProviderRepository) FindAll(ctx context.Context) ([]*model.VsLLMProvider, error) {
	var providers []*model.VsLLMProvider
	if err := r.db.WithContext(ctx).Order("sort_order ASC, provider_id DESC").Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *vsLLMProviderRepository) FindAllEnabled(ctx context.Context) ([]*model.VsLLMProvider, error) {
	var providers []*model.VsLLMProvider
	if err := r.db.WithContext(ctx).Where("status = '0'").Order("sort_order ASC").Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *vsLLMProviderRepository) Enable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsLLMProvider{}).
		Where("provider_id = ?", id).Update("status", "0").Error
}

func (r *vsLLMProviderRepository) Disable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsLLMProvider{}).
		Where("provider_id = ?", id).Update("status", "1").Error
}
