package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsScriptSegmentRepository interface {
	Create(ctx context.Context, segment *model.VsScriptSegment) error
	BatchCreate(ctx context.Context, segments []*model.VsScriptSegment) error
	Update(ctx context.Context, segment *model.VsScriptSegment) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsScriptSegment, error)
	FindWithPagination(ctx context.Context, query *v1.VsScriptSegmentListQuery) ([]*model.VsScriptSegment, int64, error)
	FindByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsScriptSegment, error)
	FindBySceneID(ctx context.Context, sceneID uint64) ([]*model.VsScriptSegment, error)
	CountByChapterID(ctx context.Context, chapterID uint64) (int64, error)
}

func NewVsScriptSegmentRepository(repository *Repository) VsScriptSegmentRepository {
	return &vsScriptSegmentRepository{Repository: repository}
}

type vsScriptSegmentRepository struct {
	*Repository
}

func (r *vsScriptSegmentRepository) Create(ctx context.Context, segment *model.VsScriptSegment) error {
	return r.db.WithContext(ctx).Create(segment).Error
}

func (r *vsScriptSegmentRepository) BatchCreate(ctx context.Context, segments []*model.VsScriptSegment) error {
	return r.db.WithContext(ctx).CreateInBatches(segments, 100).Error
}

func (r *vsScriptSegmentRepository) Update(ctx context.Context, segment *model.VsScriptSegment) error {
	return r.db.WithContext(ctx).Model(segment).
		Where("segment_id = ?", segment.SegmentID).
		Omit("created_by", "created_at", "segment_id", "scene_id", "chapter_id").
		Updates(segment).Error
}

func (r *vsScriptSegmentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("segment_id = ?", id).Delete(&model.VsScriptSegment{}).Error
}

func (r *vsScriptSegmentRepository) FindByID(ctx context.Context, id uint64) (*model.VsScriptSegment, error) {
	var segment model.VsScriptSegment
	if err := r.db.WithContext(ctx).Preload("Character").
		Where("segment_id = ?", id).First(&segment).Error; err != nil {
		return nil, err
	}
	return &segment, nil
}

func (r *vsScriptSegmentRepository) FindWithPagination(ctx context.Context, query *v1.VsScriptSegmentListQuery) ([]*model.VsScriptSegment, int64, error) {
	var segments []*model.VsScriptSegment
	db := r.db.WithContext(ctx).Model(&model.VsScriptSegment{})

	if query.ChapterID > 0 {
		db = db.Where("chapter_id = ?", query.ChapterID)
	}
	if query.SceneID > 0 {
		db = db.Where("scene_id = ?", query.SceneID)
	}
	if query.SegmentType != "" {
		db = db.Where("segment_type = ?", query.SegmentType)
	}
	if query.CharacterID > 0 {
		db = db.Where("character_id = ?", query.CharacterID)
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

	if err := db.Preload("Character").Order("segment_num ASC").Find(&segments).Error; err != nil {
		return nil, 0, err
	}

	return segments, total, nil
}

func (r *vsScriptSegmentRepository) FindByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsScriptSegment, error) {
	var segments []*model.VsScriptSegment
	if err := r.db.WithContext(ctx).Preload("Character").
		Where("chapter_id = ?", chapterID).
		Order("segment_num ASC").Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *vsScriptSegmentRepository) FindBySceneID(ctx context.Context, sceneID uint64) ([]*model.VsScriptSegment, error) {
	var segments []*model.VsScriptSegment
	if err := r.db.WithContext(ctx).Preload("Character").
		Where("scene_id = ?", sceneID).
		Order("segment_num ASC").Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *vsScriptSegmentRepository) CountByChapterID(ctx context.Context, chapterID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("chapter_id = ?", chapterID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
