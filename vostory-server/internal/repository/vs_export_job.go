package repository

import (
	"context"

	"iot-alert-center/internal/model"
)

type VsExportJobRepository interface {
	Create(ctx context.Context, job *model.VsExportJob) error
	FindByID(ctx context.Context, id uint64) (*model.VsExportJob, error)
	Update(ctx context.Context, job *model.VsExportJob) error
}

func NewVsExportJobRepository(repository *Repository) VsExportJobRepository {
	return &vsExportJobRepository{Repository: repository}
}

type vsExportJobRepository struct {
	*Repository
}

func (r *vsExportJobRepository) Create(ctx context.Context, job *model.VsExportJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *vsExportJobRepository) FindByID(ctx context.Context, id uint64) (*model.VsExportJob, error) {
	var job model.VsExportJob
	if err := r.db.WithContext(ctx).Where("export_job_id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *vsExportJobRepository) Update(ctx context.Context, job *model.VsExportJob) error {
	return r.db.WithContext(ctx).Model(job).
		Where("export_job_id = ?", job.ExportJobID).
		Updates(job).Error
}
