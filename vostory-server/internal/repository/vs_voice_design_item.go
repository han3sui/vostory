package repository

import (
	"context"

	"iot-alert-center/internal/model"
)

// 批量音色设计任务项状态
const (
	VoiceDesignItemQueued    = "queued"
	VoiceDesignItemRunning   = "running"
	VoiceDesignItemDraft     = "draft"
	VoiceDesignItemApplied   = "applied"
	VoiceDesignItemSkipped   = "skipped"
	VoiceDesignItemFailed    = "failed"
	VoiceDesignItemDiscarded = "discarded"
	VoiceDesignItemCancelled = "cancelled"
)

type VsVoiceDesignItemRepository interface {
	CreateBatch(ctx context.Context, items []*model.VsVoiceDesignItem) error
	FindByID(ctx context.Context, id uint64) (*model.VsVoiceDesignItem, error)
	UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error
	// Claim 把 queued 项置为 running，返回是否抢到（防止重复处理）
	Claim(ctx context.Context, id uint64) (bool, error)
	NextQueued(ctx context.Context) (*model.VsVoiceDesignItem, error)
	ResetRunning(ctx context.Context) (int64, error)
	CountActiveByProject(ctx context.Context, projectID uint64) (int64, error)
	FindDraftCharacterIDs(ctx context.Context, projectID uint64) ([]uint64, error)
	// FindLatestByProject 最近一个批次的全部任务项，以及更早批次中仍待确认的草稿
	FindLatestByProject(ctx context.Context, projectID uint64) ([]*model.VsVoiceDesignItem, error)
	CancelQueuedByProject(ctx context.Context, projectID uint64) (int64, error)
}

func NewVsVoiceDesignItemRepository(repository *Repository) VsVoiceDesignItemRepository {
	return &vsVoiceDesignItemRepository{Repository: repository}
}

type vsVoiceDesignItemRepository struct {
	*Repository
}

func (r *vsVoiceDesignItemRepository) CreateBatch(ctx context.Context, items []*model.VsVoiceDesignItem) error {
	return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}

func (r *vsVoiceDesignItemRepository) FindByID(ctx context.Context, id uint64) (*model.VsVoiceDesignItem, error) {
	var item model.VsVoiceDesignItem
	if err := r.db.WithContext(ctx).Where("item_id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *vsVoiceDesignItemRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("item_id = ?", id).
		Updates(fields).Error
}

func (r *vsVoiceDesignItemRepository) Claim(ctx context.Context, id uint64) (bool, error) {
	res := r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("item_id = ? AND status = ?", id, VoiceDesignItemQueued).
		Update("status", VoiceDesignItemRunning)
	return res.RowsAffected == 1, res.Error
}

func (r *vsVoiceDesignItemRepository) NextQueued(ctx context.Context) (*model.VsVoiceDesignItem, error) {
	var items []*model.VsVoiceDesignItem
	// 刚重试的项 updated_at 最新，插到队首；其余仍按创建顺序
	if err := r.db.WithContext(ctx).
		Where("status = ?", VoiceDesignItemQueued).
		Order("updated_at DESC, item_id ASC").Limit(1).
		Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func (r *vsVoiceDesignItemRepository) ResetRunning(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("status = ?", VoiceDesignItemRunning).
		Update("status", VoiceDesignItemQueued)
	return res.RowsAffected, res.Error
}

func (r *vsVoiceDesignItemRepository) CountActiveByProject(ctx context.Context, projectID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("project_id = ? AND status IN ?", projectID, []string{VoiceDesignItemQueued, VoiceDesignItemRunning}).
		Count(&count).Error
	return count, err
}

func (r *vsVoiceDesignItemRepository) FindDraftCharacterIDs(ctx context.Context, projectID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("project_id = ? AND status = ?", projectID, VoiceDesignItemDraft).
		Pluck("character_id", &ids).Error
	return ids, err
}

func (r *vsVoiceDesignItemRepository) FindLatestByProject(ctx context.Context, projectID uint64) ([]*model.VsVoiceDesignItem, error) {
	var latest []uint64
	if err := r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("project_id = ?", projectID).
		Order("batch_id DESC").Limit(1).
		Pluck("batch_id", &latest).Error; err != nil {
		return nil, err
	}
	if len(latest) == 0 {
		return nil, nil
	}
	var items []*model.VsVoiceDesignItem
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND (batch_id = ? OR status = ?)", projectID, latest[0], VoiceDesignItemDraft).
		Order("item_id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *vsVoiceDesignItemRepository) CancelQueuedByProject(ctx context.Context, projectID uint64) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.VsVoiceDesignItem{}).
		Where("project_id = ? AND status = ?", projectID, VoiceDesignItemQueued).
		Update("status", VoiceDesignItemCancelled)
	return res.RowsAffected, res.Error
}
