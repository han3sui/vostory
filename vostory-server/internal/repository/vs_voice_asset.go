package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsVoiceAssetRepository interface {
	Create(ctx context.Context, asset *model.VsVoiceAsset) error
	Update(ctx context.Context, asset *model.VsVoiceAsset) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsVoiceAsset, error)
	FindByIDs(ctx context.Context, ids []uint64) ([]*model.VsVoiceAsset, error)
	FindWithPagination(ctx context.Context, query *v1.VsVoiceAssetListQuery) ([]*model.VsVoiceAsset, int64, error)
	FindAllEnabled(ctx context.Context) ([]*model.VsVoiceAsset, error)
}

func NewVsVoiceAssetRepository(repository *Repository) VsVoiceAssetRepository {
	return &vsVoiceAssetRepository{Repository: repository}
}

type vsVoiceAssetRepository struct {
	*Repository
}

func (r *vsVoiceAssetRepository) Create(ctx context.Context, asset *model.VsVoiceAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *vsVoiceAssetRepository) Update(ctx context.Context, asset *model.VsVoiceAsset) error {
	return r.db.WithContext(ctx).Model(asset).
		Where("voice_asset_id = ?", asset.VoiceAssetID).
		Omit("created_by", "created_at", "voice_asset_id", "workspace_id").
		Updates(asset).Error
}

func (r *vsVoiceAssetRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("voice_asset_id = ?", id).Delete(&model.VsVoiceAsset{}).Error
}

func (r *vsVoiceAssetRepository) FindByID(ctx context.Context, id uint64) (*model.VsVoiceAsset, error) {
	var asset model.VsVoiceAsset
	if err := r.db.WithContext(ctx).Preload("TTSProvider").
		Where("voice_asset_id = ?", id).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *vsVoiceAssetRepository) FindByIDs(ctx context.Context, ids []uint64) ([]*model.VsVoiceAsset, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var assets []*model.VsVoiceAsset
	if err := r.db.WithContext(ctx).
		Where("voice_asset_id IN ? AND status = '0'", ids).
		Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *vsVoiceAssetRepository) FindWithPagination(ctx context.Context, query *v1.VsVoiceAssetListQuery) ([]*model.VsVoiceAsset, int64, error) {
	var assets []*model.VsVoiceAsset
	db := r.db.WithContext(ctx).Model(&model.VsVoiceAsset{})

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Gender != "" {
		db = db.Where("gender = ?", query.Gender)
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

	if err := db.Preload("TTSProvider").Order("voice_asset_id DESC").Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *vsVoiceAssetRepository) FindAllEnabled(ctx context.Context) ([]*model.VsVoiceAsset, error) {
	var assets []*model.VsVoiceAsset
	if err := r.db.WithContext(ctx).
		Where("status = '0'").
		Order("voice_asset_id ASC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}
