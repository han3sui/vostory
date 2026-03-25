package repository

import (
	"context"
	"time"

	"iot-alert-center/internal/model"
)

type VsGenerationTaskRepository interface {
	Create(ctx context.Context, task *model.VsGenerationTask) error
	FindByID(ctx context.Context, id uint64) (*model.VsGenerationTask, error)
	FindActiveByChapterID(ctx context.Context, chapterID uint64) (*model.VsGenerationTask, error)
	FindAllByStatuses(ctx context.Context, statuses []string) ([]*model.VsGenerationTask, error)
	UpdateProgress(ctx context.Context, id uint64, completed int, progress int) error
	UpdateStatus(ctx context.Context, id uint64, status string, errMsg string) error
	ResetStatus(ctx context.Context, id uint64, status string) error
	SetStarted(ctx context.Context, id uint64) error
	SetCompleted(ctx context.Context, id uint64) error
	IncrementCompleted(ctx context.Context, id uint64) (int64, error)
}

func NewVsGenerationTaskRepository(repository *Repository) VsGenerationTaskRepository {
	return &vsGenerationTaskRepository{Repository: repository}
}

type vsGenerationTaskRepository struct {
	*Repository
}

func (r *vsGenerationTaskRepository) Create(ctx context.Context, task *model.VsGenerationTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *vsGenerationTaskRepository) FindByID(ctx context.Context, id uint64) (*model.VsGenerationTask, error) {
	var task model.VsGenerationTask
	if err := r.db.WithContext(ctx).Where("task_id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *vsGenerationTaskRepository) FindActiveByChapterID(ctx context.Context, chapterID uint64) (*model.VsGenerationTask, error) {
	var task model.VsGenerationTask
	err := r.db.WithContext(ctx).
		Where("chapter_id = ? AND status IN ?", chapterID, []string{"pending", "running"}).
		Order("task_id DESC").
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *vsGenerationTaskRepository) FindAllByStatuses(ctx context.Context, statuses []string) ([]*model.VsGenerationTask, error) {
	var tasks []*model.VsGenerationTask
	err := r.db.WithContext(ctx).
		Where("status IN ?", statuses).
		Order("task_id ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *vsGenerationTaskRepository) ResetStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Update("status", status).Error
}

func (r *vsGenerationTaskRepository) UpdateProgress(ctx context.Context, id uint64, completed int, progress int) error {
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Updates(map[string]interface{}{
			"completed_batches": completed,
			"progress":          progress,
		}).Error
}

func (r *vsGenerationTaskRepository) UpdateStatus(ctx context.Context, id uint64, status string, errMsg string) error {
	updates := map[string]interface{}{"status": status}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).Updates(updates).Error
}

func (r *vsGenerationTaskRepository) SetStarted(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Updates(map[string]interface{}{
			"status":     "running",
			"started_at": &now,
		}).Error
}

func (r *vsGenerationTaskRepository) SetCompleted(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"progress":     100,
			"completed_at": &now,
		}).Error
}

func (r *vsGenerationTaskRepository) IncrementCompleted(ctx context.Context, id uint64) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Update("completed_batches", r.db.Raw("completed_batches + 1"))
	if result.Error != nil {
		return 0, result.Error
	}
	var task model.VsGenerationTask
	if err := r.db.WithContext(ctx).Select("completed_batches").Where("task_id = ?", id).First(&task).Error; err != nil {
		return 0, err
	}
	return int64(task.CompletedBatches), nil
}
