package repository

import (
	"context"
	"time"

	"iot-alert-center/internal/model"
)

// TaskProgressRow 项目活跃任务进度，只含进度展示需要的列（不带 segment_ids、章节正文）
type TaskProgressRow struct {
	TaskID           uint64
	ChapterID        *uint64
	ChapterTitle     string
	TaskType         string
	Status           string
	Progress         int
	TotalBatches     int
	CompletedBatches int
	FailedBatches    int
}

type VsGenerationTaskRepository interface {
	Create(ctx context.Context, task *model.VsGenerationTask) error
	FindByID(ctx context.Context, id uint64) (*model.VsGenerationTask, error)
	FindActiveByChapterID(ctx context.Context, chapterID uint64) (*model.VsGenerationTask, error)
	FindActiveListByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsGenerationTask, error)
	FindActiveByProjectID(ctx context.Context, projectID uint64) ([]*model.VsGenerationTask, error)
	FindActiveProgressByProjectID(ctx context.Context, projectID uint64) ([]*TaskProgressRow, error)
	FindAllByStatuses(ctx context.Context, statuses []string) ([]*model.VsGenerationTask, error)
	FindAllByStatusesAndType(ctx context.Context, statuses []string, taskType string) ([]*model.VsGenerationTask, error)
	UpdateProgress(ctx context.Context, id uint64, completed int, progress int) error
	UpdateStatus(ctx context.Context, id uint64, status string, errMsg string) error
	ResetStatus(ctx context.Context, id uint64, status string) error
	SetStarted(ctx context.Context, id uint64) error
	SetCompleted(ctx context.Context, id uint64) error
	SetFailed(ctx context.Context, id uint64, errMsg string) error
	SetCancelled(ctx context.Context, id uint64, errMsg string) error
	IncrementCompleted(ctx context.Context, id uint64) (int64, error)
	IncrementFailed(ctx context.Context, id uint64) (int64, error)
	ReduceTotalBatches(ctx context.Context, id uint64, count int) error
	SetTotalBatches(ctx context.Context, id uint64, total int) error
}

func NewVsGenerationTaskRepository(repository *Repository) VsGenerationTaskRepository {
	return &vsGenerationTaskRepository{Repository: repository}
}

type vsGenerationTaskRepository struct {
	*Repository
}

func (r *vsGenerationTaskRepository) Create(ctx context.Context, task *model.VsGenerationTask) error {
	return r.DB(ctx).Create(task).Error
}

func (r *vsGenerationTaskRepository) FindByID(ctx context.Context, id uint64) (*model.VsGenerationTask, error) {
	var task model.VsGenerationTask
	if err := r.db.WithContext(ctx).Preload("Chapter").Where("task_id = ?", id).First(&task).Error; err != nil {
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

func (r *vsGenerationTaskRepository) FindActiveListByChapterID(ctx context.Context, chapterID uint64) ([]*model.VsGenerationTask, error) {
	var tasks []*model.VsGenerationTask
	err := r.db.WithContext(ctx).
		Preload("Chapter").
		Where("chapter_id = ? AND status IN ?", chapterID, []string{"pending", "running"}).
		Order("task_id ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *vsGenerationTaskRepository) FindAllByStatuses(ctx context.Context, statuses []string) ([]*model.VsGenerationTask, error) {
	var tasks []*model.VsGenerationTask
	err := r.db.WithContext(ctx).
		Where("status IN ?", statuses).
		Order("task_id ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *vsGenerationTaskRepository) FindAllByStatusesAndType(ctx context.Context, statuses []string, taskType string) ([]*model.VsGenerationTask, error) {
	var tasks []*model.VsGenerationTask
	err := r.db.WithContext(ctx).
		Where("status IN ? AND task_type = ?", statuses, taskType).
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

func (r *vsGenerationTaskRepository) SetFailed(ctx context.Context, id uint64, errMsg string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errMsg,
			"completed_at":  &now,
		}).Error
}

func (r *vsGenerationTaskRepository) SetCancelled(ctx context.Context, id uint64, errMsg string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Updates(map[string]interface{}{
			"status":        "cancelled",
			"error_message": errMsg,
			"completed_at":  &now,
		}).Error
}

func (r *vsGenerationTaskRepository) IncrementFailed(ctx context.Context, id uint64) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Update("failed_batches", r.db.Raw("failed_batches + 1"))
	if result.Error != nil {
		return 0, result.Error
	}
	var task model.VsGenerationTask
	if err := r.db.WithContext(ctx).Select("failed_batches").Where("task_id = ?", id).First(&task).Error; err != nil {
		return 0, err
	}
	return int64(task.FailedBatches), nil
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

func (r *vsGenerationTaskRepository) FindActiveByProjectID(ctx context.Context, projectID uint64) ([]*model.VsGenerationTask, error) {
	var tasks []*model.VsGenerationTask
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND status IN ?", projectID, []string{"pending", "running"}).
		Order("task_id ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *vsGenerationTaskRepository) FindActiveProgressByProjectID(ctx context.Context, projectID uint64) ([]*TaskProgressRow, error) {
	var rows []*TaskProgressRow
	err := r.db.WithContext(ctx).
		Table("vs_generation_task AS t").
		Select("t.task_id, t.chapter_id, COALESCE(c.title, '') AS chapter_title, t.task_type, t.status, t.progress, "+
			"t.total_batches, t.completed_batches, t.failed_batches").
		Joins("LEFT JOIN vs_chapter AS c ON c.chapter_id = t.chapter_id").
		Where("t.project_id = ? AND t.status IN ? AND t.deleted_at IS NULL", projectID, []string{"pending", "running"}).
		Order("t.task_id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *vsGenerationTaskRepository) ReduceTotalBatches(ctx context.Context, id uint64, count int) error {
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Update("total_batches", r.db.Raw("GREATEST(total_batches - ?, 0)", count)).Error
}

func (r *vsGenerationTaskRepository) SetTotalBatches(ctx context.Context, id uint64, total int) error {
	return r.db.WithContext(ctx).Model(&model.VsGenerationTask{}).
		Where("task_id = ?", id).
		Update("total_batches", total).Error
}
