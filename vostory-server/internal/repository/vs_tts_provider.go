package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsTTSProviderRepository interface {
	Create(ctx context.Context, provider *model.VsTTSProvider) error
	Update(ctx context.Context, provider *model.VsTTSProvider) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsTTSProvider, error)
	FindWithPagination(ctx context.Context, query *v1.VsTTSProviderListQuery) ([]*model.VsTTSProvider, int64, error)
	FindAllEnabled(ctx context.Context) ([]*model.VsTTSProvider, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsTTSProviderRepository(repository *Repository) VsTTSProviderRepository {
	return &vsTTSProviderRepository{Repository: repository}
}

type vsTTSProviderRepository struct {
	*Repository
}

func (r *vsTTSProviderRepository) Create(ctx context.Context, provider *model.VsTTSProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *vsTTSProviderRepository) Update(ctx context.Context, provider *model.VsTTSProvider) error {
	return r.db.WithContext(ctx).Model(provider).
		Where("provider_id = ?", provider.ProviderID).
		Omit("created_by", "created_at", "provider_id").
		Updates(provider).Error
}

func (r *vsTTSProviderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("provider_id = ?", id).Delete(&model.VsTTSProvider{}).Error
}

func (r *vsTTSProviderRepository) FindByID(ctx context.Context, id uint64) (*model.VsTTSProvider, error) {
	var provider model.VsTTSProvider
	if err := r.db.WithContext(ctx).Where("provider_id = ?", id).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *vsTTSProviderRepository) FindWithPagination(ctx context.Context, query *v1.VsTTSProviderListQuery) ([]*model.VsTTSProvider, int64, error) {
	var providers []*model.VsTTSProvider
	db := r.db.WithContext(ctx).Model(&model.VsTTSProvider{})

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.ProviderType != "" {
		db = db.Where("provider_type = ?", query.ProviderType)
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

func (r *vsTTSProviderRepository) FindAllEnabled(ctx context.Context) ([]*model.VsTTSProvider, error) {
	var providers []*model.VsTTSProvider
	if err := r.db.WithContext(ctx).Where("status = '0'").Order("sort_order ASC").Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (r *vsTTSProviderRepository) Enable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsTTSProvider{}).
		Where("provider_id = ?", id).Update("status", "0").Error
}

func (r *vsTTSProviderRepository) Disable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsTTSProvider{}).
		Where("provider_id = ?", id).Update("status", "1").Error
}
