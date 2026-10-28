package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type VsChapterRepository interface {
	Create(ctx context.Context, chapter *model.VsChapter) error
	// CreateBatch 批量创建章节，在一个事务里完成，失败时不会留下部分章节
	CreateBatch(ctx context.Context, chapters []*model.VsChapter) error
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

func (r *vsChapterRepository) CreateBatch(ctx context.Context, chapters []*model.VsChapter) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(chapters, 100).Error
	})
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
	db := r.db.WithContext(ctx).Model(&model.VsChapter{}).
		Select("chapter_id, project_id, title, chapter_num, word_count, status, split_status, remark, created_by, created_at, updated_by, updated_at, dept_id")

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
	// 一次聚合出切割/配音进度，避免按章节多次子查询
	stats := r.db.WithContext(ctx).
		Table("vs_script_segment AS s").
		Select(`s.chapter_id,
			COUNT(*) AS segment_count,
			COUNT(c.clip_id) AS audio_count,
			COUNT(*) FILTER (WHERE s.status = 'failed') AS failed_count,
			COUNT(*) FILTER (WHERE s.status = 'queued') AS queued_count,
			COUNT(*) FILTER (WHERE s.status = 'processing') AS processing_count`).
		Joins("LEFT JOIN vs_audio_clip c ON c.segment_id = s.segment_id AND c.deleted_at IS NULL AND c.is_current = '1'").
		Where("s.deleted_at IS NULL AND s.project_id = ?", projectID).
		Group("s.chapter_id")

	var chapters []*model.VsChapter
	if err := r.db.WithContext(ctx).
		Table("vs_chapter").
		Select(`vs_chapter.chapter_id, vs_chapter.project_id, vs_chapter.title, vs_chapter.chapter_num, vs_chapter.word_count, vs_chapter.status, vs_chapter.split_status, vs_chapter.remark, vs_chapter.created_by, vs_chapter.created_at, vs_chapter.updated_by, vs_chapter.updated_at, vs_chapter.dept_id,
			COALESCE(stats.segment_count, 0) AS segment_count,
			COALESCE(stats.audio_count, 0) AS audio_count,
			COALESCE(stats.failed_count, 0) AS failed_count,
			COALESCE(stats.queued_count, 0) AS queued_count,
			COALESCE(stats.processing_count, 0) AS processing_count`).
		Joins("LEFT JOIN (?) AS stats ON stats.chapter_id = vs_chapter.chapter_id", stats).
		Where("vs_chapter.project_id = ? AND vs_chapter.deleted_at IS NULL", projectID).
		Order("vs_chapter.chapter_num ASC").
		Find(&chapters).Error; err != nil {
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
