package repository

import (
	"context"

	"iot-alert-center/internal/model"
)

type VsSceneRepository interface {
	Create(ctx context.Context, scene *model.VsScene) error
	BatchCreate(ctx context.Context, scenes []*model.VsScene) error
	DeleteByChapterID(ctx context.Context, chapterID uint64) error
	FindByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsScene, error)
}

func NewVsSceneRepository(repository *Repository) VsSceneRepository {
	return &vsSceneRepository{Repository: repository}
}

type vsSceneRepository struct {
	*Repository
}

func (r *vsSceneRepository) Create(ctx context.Context, scene *model.VsScene) error {
	return r.db.WithContext(ctx).Create(scene).Error
}

func (r *vsSceneRepository) BatchCreate(ctx context.Context, scenes []*model.VsScene) error {
	return r.db.WithContext(ctx).CreateInBatches(scenes, 100).Error
}

func (r *vsSceneRepository) DeleteByChapterID(ctx context.Context, chapterID uint64) error {
	return r.db.WithContext(ctx).Where("chapter_id = ?", chapterID).Delete(&model.VsScene{}).Error
}

func (r *vsSceneRepository) FindByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsScene, error) {
	var scenes []*model.VsScene
	if err := r.db.WithContext(ctx).
		Where("chapter_id = ?", chapterID).
		Order("scene_num ASC").Find(&scenes).Error; err != nil {
		return nil, err
	}
	return scenes, nil
}
