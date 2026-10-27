package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VsScriptSegmentRepository interface {
	Create(ctx context.Context, segment *model.VsScriptSegment) error
	BatchCreate(ctx context.Context, segments []*model.VsScriptSegment) error
	Update(ctx context.Context, segment *model.VsScriptSegment) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsScriptSegment, error)
	FindWithPagination(ctx context.Context, query *v1.VsScriptSegmentListQuery) ([]*model.VsScriptSegment, int64, error)
	FindByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsScriptSegment, error)
	FindBySceneID(ctx context.Context, sceneID uint64) ([]*model.VsScriptSegment, error)
	CountByChapterID(ctx context.Context, chapterID uint64) (int64, error)
	DeleteByChapterID(ctx context.Context, chapterID uint64) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
	UpdateStatusWithError(ctx context.Context, id uint64, status string, errMsg string) error
	FindByChapterIDAndStatus(ctx context.Context, chapterID uint64, status string) ([]*model.VsScriptSegment, error)
	BatchUpdateStatus(ctx context.Context, ids []uint64, fromStatus, toStatus string) (int64, error)
	BatchUpdateStatusByChapter(ctx context.Context, chapterID uint64, fromStatus, toStatus string) (int64, error)
	IncrementSegmentNumAfter(ctx context.Context, chapterID uint64, afterNum int) error
	ReorderSegmentNums(ctx context.Context, chapterID uint64) error
	FindByCharacterIDs(ctx context.Context, characterIDs []uint64) ([]*model.VsScriptSegment, error)
	CountByCharacterIDs(ctx context.Context, characterIDs []uint64) (int64, error)
	FindRegenerableByCharacterIDs(ctx context.Context, characterIDs []uint64) ([]SegmentRef, error)
	ClaimForGenerate(ctx context.Context, ids []uint64) ([]uint64, error)
}

// SegmentRef 入队只需要的几列，不拉正文
type SegmentRef struct {
	SegmentID uint64
	ProjectID uint64
	ChapterID uint64
}

// nonRegenerableStatuses 排队/合成中的片段再入队会被合成两次，锁定的片段不允许重新生成
var nonRegenerableStatuses = []string{"queued", "processing", "locked"}

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
		Omit("created_by", "created_at", "segment_id", "project_id", "scene_id", "chapter_id").
		Updates(segment).Error
}

func (r *vsScriptSegmentRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("segment_id = ?", id).
		Updates(fields).Error
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

	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
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

func (r *vsScriptSegmentRepository) DeleteByChapterID(ctx context.Context, chapterID uint64) error {
	return r.db.WithContext(ctx).Where("chapter_id = ?", chapterID).Delete(&model.VsScriptSegment{}).Error
}

func (r *vsScriptSegmentRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("segment_id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"error_message": "",
	}).Error
}

func (r *vsScriptSegmentRepository) UpdateStatusWithError(ctx context.Context, id uint64, status string, errMsg string) error {
	return r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("segment_id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
	}).Error
}

func (r *vsScriptSegmentRepository) FindByChapterIDAndStatus(ctx context.Context, chapterID uint64, status string) ([]*model.VsScriptSegment, error) {
	var segments []*model.VsScriptSegment
	if err := r.db.WithContext(ctx).
		Where("chapter_id = ? AND status = ?", chapterID, status).
		Order("segment_num ASC").Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *vsScriptSegmentRepository) BatchUpdateStatus(ctx context.Context, ids []uint64, fromStatus, toStatus string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("segment_id IN ? AND status = ?", ids, fromStatus).
		Update("status", toStatus)
	return result.RowsAffected, result.Error
}

func (r *vsScriptSegmentRepository) BatchUpdateStatusByChapter(ctx context.Context, chapterID uint64, fromStatus, toStatus string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("chapter_id = ? AND status = ?", chapterID, fromStatus).
		Update("status", toStatus)
	return result.RowsAffected, result.Error
}

func (r *vsScriptSegmentRepository) IncrementSegmentNumAfter(ctx context.Context, chapterID uint64, afterNum int) error {
	return r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("chapter_id = ? AND segment_num > ?", chapterID, afterNum).
		UpdateColumn("segment_num", gorm.Expr("segment_num + 1")).Error
}

func (r *vsScriptSegmentRepository) ReorderSegmentNums(ctx context.Context, chapterID uint64) error {
	segments, err := r.FindByChapterID(ctx, chapterID)
	if err != nil {
		return err
	}
	for i, seg := range segments {
		newNum := i + 1
		if seg.SegmentNum != newNum {
			r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
				Where("segment_id = ?", seg.SegmentID).
				Update("segment_num", newNum)
		}
	}
	return nil
}

func (r *vsScriptSegmentRepository) FindByCharacterIDs(ctx context.Context, characterIDs []uint64) ([]*model.VsScriptSegment, error) {
	if len(characterIDs) == 0 {
		return nil, nil
	}
	var segments []*model.VsScriptSegment
	if err := r.db.WithContext(ctx).
		Where("character_id IN ?", characterIDs).
		Order("chapter_id ASC, segment_num ASC").
		Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

func (r *vsScriptSegmentRepository) CountByCharacterIDs(ctx context.Context, characterIDs []uint64) (int64, error) {
	if len(characterIDs) == 0 {
		return 0, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Where("character_id IN ?", characterIDs).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *vsScriptSegmentRepository) FindRegenerableByCharacterIDs(ctx context.Context, characterIDs []uint64) ([]SegmentRef, error) {
	if len(characterIDs) == 0 {
		return nil, nil
	}
	var refs []SegmentRef
	if err := r.db.WithContext(ctx).Model(&model.VsScriptSegment{}).
		Select("segment_id", "project_id", "chapter_id").
		Where("character_id IN ? AND status NOT IN ? AND content <> ''", characterIDs, nonRegenerableStatuses).
		Order("chapter_id ASC, segment_num ASC").
		Scan(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

// ClaimForGenerate 把可生成的片段原子地标为 queued，只返回本次真正改到的片段。
// 并发的两次认领不会拿到同一个片段，调用方只入队返回的这些，就不会重复合成
func (r *vsScriptSegmentRepository) ClaimForGenerate(ctx context.Context, ids []uint64) ([]uint64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var claimed []model.VsScriptSegment
	if err := r.DB(ctx).Model(&claimed).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "segment_id"}}}).
		Where("segment_id IN ? AND status NOT IN ?", ids, nonRegenerableStatuses).
		Updates(map[string]interface{}{
			"status":        "queued",
			"error_message": "",
		}).Error; err != nil {
		return nil, err
	}
	claimedIDs := make([]uint64, len(claimed))
	for i, seg := range claimed {
		claimedIDs[i] = seg.SegmentID
	}
	return claimedIDs, nil
}
