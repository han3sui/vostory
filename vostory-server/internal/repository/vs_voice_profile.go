package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsVoiceProfileRepository interface {
	Create(ctx context.Context, profile *model.VsVoiceProfile) error
	Update(ctx context.Context, profile *model.VsVoiceProfile) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsVoiceProfile, error)
	FindByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.VsVoiceProfile, error)
	FindWithPagination(ctx context.Context, query *v1.VsVoiceProfileListQuery) ([]*model.VsVoiceProfile, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsVoiceProfile, error)
}

func NewVsVoiceProfileRepository(repository *Repository) VsVoiceProfileRepository {
	return &vsVoiceProfileRepository{Repository: repository}
}

type vsVoiceProfileRepository struct {
	*Repository
}

func (r *vsVoiceProfileRepository) Create(ctx context.Context, profile *model.VsVoiceProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *vsVoiceProfileRepository) Update(ctx context.Context, profile *model.VsVoiceProfile) error {
	return r.db.WithContext(ctx).Model(profile).
		Where("voice_profile_id = ?", profile.VoiceProfileID).
		Omit("created_by", "created_at", "voice_profile_id", "project_id").
		Updates(profile).Error
}

func (r *vsVoiceProfileRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("voice_profile_id = ?", id).Delete(&model.VsVoiceProfile{}).Error
}

func (r *vsVoiceProfileRepository) FindByID(ctx context.Context, id uint64) (*model.VsVoiceProfile, error) {
	var profile model.VsVoiceProfile
	if err := r.db.WithContext(ctx).Preload("TTSProvider").
		Where("voice_profile_id = ?", id).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *vsVoiceProfileRepository) FindByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.VsVoiceProfile, error) {
	if len(ids) == 0 {
		return make(map[uint64]*model.VsVoiceProfile), nil
	}
	var profiles []*model.VsVoiceProfile
	if err := r.db.WithContext(ctx).Where("voice_profile_id IN ?", ids).Find(&profiles).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64]*model.VsVoiceProfile, len(profiles))
	for _, p := range profiles {
		result[p.VoiceProfileID] = p
	}
	return result, nil
}

func (r *vsVoiceProfileRepository) FindWithPagination(ctx context.Context, query *v1.VsVoiceProfileListQuery) ([]*model.VsVoiceProfile, int64, error) {
	var profiles []*model.VsVoiceProfile
	db := r.db.WithContext(ctx).Model(&model.VsVoiceProfile{})

	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
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

	if err := db.Preload("TTSProvider").Order("voice_profile_id DESC").Find(&profiles).Error; err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

func (r *vsVoiceProfileRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsVoiceProfile, error) {
	var profiles []*model.VsVoiceProfile
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND status = '0'", projectID).
		Order("voice_profile_id ASC").Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}
