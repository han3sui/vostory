package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsWorkspaceRepository interface {
	Create(ctx context.Context, workspace *model.VsWorkspace) error
	Update(ctx context.Context, workspace *model.VsWorkspace) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsWorkspace, error)
	FindWithPagination(ctx context.Context, query *v1.VsWorkspaceListQuery) ([]*model.VsWorkspace, int64, error)
	FindAllEnabled(ctx context.Context) ([]*model.VsWorkspace, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsWorkspaceRepository(repository *Repository) VsWorkspaceRepository {
	return &vsWorkspaceRepository{Repository: repository}
}

type vsWorkspaceRepository struct {
	*Repository
}

func (r *vsWorkspaceRepository) Create(ctx context.Context, workspace *model.VsWorkspace) error {
	return r.db.WithContext(ctx).Create(workspace).Error
}

func (r *vsWorkspaceRepository) Update(ctx context.Context, workspace *model.VsWorkspace) error {
	return r.db.WithContext(ctx).Model(workspace).
		Where("workspace_id = ?", workspace.WorkspaceID).
		Omit("created_by", "created_at", "workspace_id", "owner_id").
		Updates(workspace).Error
}

func (r *vsWorkspaceRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("workspace_id = ?", id).Delete(&model.VsWorkspace{}).Error
}

func (r *vsWorkspaceRepository) FindByID(ctx context.Context, id uint64) (*model.VsWorkspace, error) {
	var workspace model.VsWorkspace
	if err := r.db.WithContext(ctx).Preload("Owner").Where("workspace_id = ?", id).First(&workspace).Error; err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *vsWorkspaceRepository) FindWithPagination(ctx context.Context, query *v1.VsWorkspaceListQuery) ([]*model.VsWorkspace, int64, error) {
	var workspaces []*model.VsWorkspace
	db := r.db.WithContext(ctx).Model(&model.VsWorkspace{})

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	db = db.Scopes(model.WithDataScope(ctx))

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Preload("Owner").Order("workspace_id DESC").Find(&workspaces).Error; err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

func (r *vsWorkspaceRepository) FindAllEnabled(ctx context.Context) ([]*model.VsWorkspace, error) {
	var workspaces []*model.VsWorkspace
	db := r.db.WithContext(ctx).Where("status = '0'").Scopes(model.WithDataScope(ctx))
	if err := db.Order("workspace_id DESC").Find(&workspaces).Error; err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (r *vsWorkspaceRepository) Enable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsWorkspace{}).
		Where("workspace_id = ?", id).Update("status", "0").Error
}

func (r *vsWorkspaceRepository) Disable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsWorkspace{}).
		Where("workspace_id = ?", id).Update("status", "1").Error
}
