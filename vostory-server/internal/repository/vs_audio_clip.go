package repository

import (
	"context"

	"iot-alert-center/internal/model"
)

type VsAudioClipRepository interface {
	Create(ctx context.Context, clip *model.VsAudioClip) error
	FindByID(ctx context.Context, id uint64) (*model.VsAudioClip, error)
	FindCurrentBySegmentID(ctx context.Context, segmentID uint64) (*model.VsAudioClip, error)
	FindCurrentBySegmentIDs(ctx context.Context, segmentIDs []uint64) (map[uint64]*model.VsAudioClip, error)
	FindBySegmentID(ctx context.Context, segmentID uint64) ([]*model.VsAudioClip, error)
	SetAllNonCurrent(ctx context.Context, segmentID uint64) error
	GetMaxVersion(ctx context.Context, segmentID uint64) (int, error)
}

func NewVsAudioClipRepository(repository *Repository) VsAudioClipRepository {
	return &vsAudioClipRepository{Repository: repository}
}

type vsAudioClipRepository struct {
	*Repository
}

func (r *vsAudioClipRepository) Create(ctx context.Context, clip *model.VsAudioClip) error {
	return r.db.WithContext(ctx).Create(clip).Error
}

func (r *vsAudioClipRepository) FindByID(ctx context.Context, id uint64) (*model.VsAudioClip, error) {
	var clip model.VsAudioClip
	if err := r.db.WithContext(ctx).Where("clip_id = ?", id).First(&clip).Error; err != nil {
		return nil, err
	}
	return &clip, nil
}

func (r *vsAudioClipRepository) FindCurrentBySegmentID(ctx context.Context, segmentID uint64) (*model.VsAudioClip, error) {
	var clip model.VsAudioClip
	if err := r.db.WithContext(ctx).
		Where("segment_id = ? AND is_current = '1'", segmentID).
		First(&clip).Error; err != nil {
		return nil, err
	}
	return &clip, nil
}

func (r *vsAudioClipRepository) FindCurrentBySegmentIDs(ctx context.Context, segmentIDs []uint64) (map[uint64]*model.VsAudioClip, error) {
	if len(segmentIDs) == 0 {
		return make(map[uint64]*model.VsAudioClip), nil
	}
	var clips []*model.VsAudioClip
	if err := r.db.WithContext(ctx).
		Where("segment_id IN ? AND is_current = '1'", segmentIDs).
		Find(&clips).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64]*model.VsAudioClip, len(clips))
	for _, c := range clips {
		result[c.SegmentID] = c
	}
	return result, nil
}

func (r *vsAudioClipRepository) FindBySegmentID(ctx context.Context, segmentID uint64) ([]*model.VsAudioClip, error) {
	var clips []*model.VsAudioClip
	if err := r.db.WithContext(ctx).
		Where("segment_id = ?", segmentID).
		Order("version DESC").Find(&clips).Error; err != nil {
		return nil, err
	}
	return clips, nil
}

func (r *vsAudioClipRepository) SetAllNonCurrent(ctx context.Context, segmentID uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsAudioClip{}).
		Where("segment_id = ? AND is_current = '1'", segmentID).
		Update("is_current", "0").Error
}

func (r *vsAudioClipRepository) GetMaxVersion(ctx context.Context, segmentID uint64) (int, error) {
	var maxVersion *int
	if err := r.db.WithContext(ctx).Model(&model.VsAudioClip{}).
		Where("segment_id = ?", segmentID).
		Select("MAX(version)").Scan(&maxVersion).Error; err != nil {
		return 0, err
	}
	if maxVersion == nil {
		return 0, nil
	}
	return *maxVersion, nil
}
