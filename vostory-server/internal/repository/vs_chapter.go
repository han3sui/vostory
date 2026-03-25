package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsChapterRepository interface {
	Create(ctx context.Context, chapter *model.VsChapter) error
	Update(ctx context.Context, chapter *model.VsChapter) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsChapter, error)
	FindWithPagination(ctx context.Context, query *v1.VsChapterListQuery) ([]*model.VsChapter, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsChapter, error)
	CountByProjectID(ctx context.Context, projectID uint64) (int64, error)
}

func NewVsChapterRepository(repository *Repository) VsChapterRepository {
	return &vsChapterRepository{Repository: repository}
}

type vsChapterRepository struct {
	*Repository
}

func (r *vsChapterRepository) Create(ctx context.Context, chapter *model.VsChapter) error {
	return r.db.WithContext(ctx).Create(chapter).Error
}

func (r *vsChapterRepository) Update(ctx context.Context, chapter *model.VsChapter) error {
	return r.db.WithContext(ctx).Model(chapter).
		Where("chapter_id = ?", chapter.ChapterID).
		Omit("created_by", "created_at", "chapter_id", "project_id").
		Updates(chapter).Error
}

func (r *vsChapterRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("chapter_id = ?", id).Delete(&model.VsChapter{}).Error
}

func (r *vsChapterRepository) FindByID(ctx context.Context, id uint64) (*model.VsChapter, error) {
	var chapter model.VsChapter
	if err := r.db.WithContext(ctx).Where("chapter_id = ?", id).First(&chapter).Error; err != nil {
		return nil, err
	}
	return &chapter, nil
}

func (r *vsChapterRepository) FindWithPagination(ctx context.Context, query *v1.VsChapterListQuery) ([]*model.VsChapter, int64, error) {
	var chapters []*model.VsChapter
	db := r.db.WithContext(ctx).Model(&model.VsChapter{})

	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.Title != "" {
		db = db.Where("title LIKE ?", "%"+query.Title+"%")
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

	if err := db.Order("chapter_num ASC").Find(&chapters).Error; err != nil {
		return nil, 0, err
	}

	return chapters, total, nil
}

func (r *vsChapterRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsChapter, error) {
	var chapters []*model.VsChapter
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).
		Order("chapter_num ASC").Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}

func (r *vsChapterRepository) CountByProjectID(ctx context.Context, projectID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.VsChapter{}).
		Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
