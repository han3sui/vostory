package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsPronunciationDictRepository interface {
	Create(ctx context.Context, dict *model.VsPronunciationDict) error
	Update(ctx context.Context, dict *model.VsPronunciationDict) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsPronunciationDict, error)
	FindWithPagination(ctx context.Context, query *v1.VsPronunciationDictListQuery) ([]*model.VsPronunciationDict, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsPronunciationDict, error)
	FindGlobalByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*model.VsPronunciationDict, error)
}

func NewVsPronunciationDictRepository(repository *Repository) VsPronunciationDictRepository {
	return &vsPronunciationDictRepository{Repository: repository}
}

type vsPronunciationDictRepository struct {
	*Repository
}

func (r *vsPronunciationDictRepository) Create(ctx context.Context, dict *model.VsPronunciationDict) error {
	return r.db.WithContext(ctx).Create(dict).Error
}

func (r *vsPronunciationDictRepository) Update(ctx context.Context, dict *model.VsPronunciationDict) error {
	return r.db.WithContext(ctx).Model(dict).
		Where("dict_id = ?", dict.DictID).
		Omit("created_by", "created_at", "dict_id", "workspace_id", "project_id").
		Updates(dict).Error
}

func (r *vsPronunciationDictRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("dict_id = ?", id).Delete(&model.VsPronunciationDict{}).Error
}

func (r *vsPronunciationDictRepository) FindByID(ctx context.Context, id uint64) (*model.VsPronunciationDict, error) {
	var dict model.VsPronunciationDict
	if err := r.db.WithContext(ctx).Preload("Project").Preload("Workspace").
		Where("dict_id = ?", id).First(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (r *vsPronunciationDictRepository) FindWithPagination(ctx context.Context, query *v1.VsPronunciationDictListQuery) ([]*model.VsPronunciationDict, int64, error) {
	var dicts []*model.VsPronunciationDict
	db := r.db.WithContext(ctx).Model(&model.VsPronunciationDict{})

	if query.WorkspaceID > 0 {
		db = db.Where("workspace_id = ?", query.WorkspaceID)
	}
	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	} else if query.ProjectID == 0 {
		db = db.Where("project_id IS NULL")
	}
	if query.Word != "" {
		db = db.Where("word LIKE ?", "%"+query.Word+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	if err := db.Preload("Project").Preload("Workspace").
		Order("dict_id DESC").Find(&dicts).Error; err != nil {
		return nil, 0, err
	}

	return dicts, total, nil
}

func (r *vsPronunciationDictRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsPronunciationDict, error) {
	var dicts []*model.VsPronunciationDict
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("word ASC").Find(&dicts).Error; err != nil {
		return nil, err
	}
	return dicts, nil
}

func (r *vsPronunciationDictRepository) FindGlobalByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*model.VsPronunciationDict, error) {
	var dicts []*model.VsPronunciationDict
	if err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND project_id IS NULL", workspaceID).
		Order("word ASC").Find(&dicts).Error; err != nil {
		return nil, err
	}
	return dicts, nil
}
