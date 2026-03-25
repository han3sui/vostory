package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsProjectRepository interface {
	Create(ctx context.Context, project *model.VsProject) error
	Update(ctx context.Context, project *model.VsProject) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsProject, error)
	FindWithPagination(ctx context.Context, query *v1.VsProjectListQuery) ([]*model.VsProject, int64, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*model.VsProject, error)
}

func NewVsProjectRepository(repository *Repository) VsProjectRepository {
	return &vsProjectRepository{Repository: repository}
}

type vsProjectRepository struct {
	*Repository
}

func (r *vsProjectRepository) Create(ctx context.Context, project *model.VsProject) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *vsProjectRepository) Update(ctx context.Context, project *model.VsProject) error {
	return r.db.WithContext(ctx).Model(project).
		Where("project_id = ?", project.ProjectID).
		Omit("created_by", "created_at", "project_id", "workspace_id").
		Updates(project).Error
}

func (r *vsProjectRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("project_id = ?", id).Delete(&model.VsProject{}).Error
}

func (r *vsProjectRepository) FindByID(ctx context.Context, id uint64) (*model.VsProject, error) {
	var project model.VsProject
	if err := r.db.WithContext(ctx).
		Preload("Workspace").
		Preload("LLMProvider").
		Preload("TTSProvider").
		Where("project_id = ?", id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *vsProjectRepository) FindWithPagination(ctx context.Context, query *v1.VsProjectListQuery) ([]*model.VsProject, int64, error) {
	var projects []*model.VsProject
	db := r.db.WithContext(ctx).Model(&model.VsProject{})

	if query.WorkspaceID > 0 {
		db = db.Where("workspace_id = ?", query.WorkspaceID)
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.SourceType != "" {
		db = db.Where("source_type = ?", query.SourceType)
	}

	db = db.Scopes(model.WithDataScope(ctx))

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Preload("Workspace").Preload("LLMProvider").Preload("TTSProvider").
		Order("project_id DESC").Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *vsProjectRepository) FindByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*model.VsProject, error) {
	var projects []*model.VsProject
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Scopes(model.WithDataScope(ctx)).
		Order("project_id DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}
