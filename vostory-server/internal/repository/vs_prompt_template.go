package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsPromptTemplateRepository interface {
	Create(ctx context.Context, template *model.VsPromptTemplate) error
	Update(ctx context.Context, template *model.VsPromptTemplate) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsPromptTemplate, error)
	FindWithPagination(ctx context.Context, query *v1.VsPromptTemplateListQuery) ([]*model.VsPromptTemplate, int64, error)
	FindByType(ctx context.Context, templateType string) ([]*model.VsPromptTemplate, error)
	FindAllEnabled(ctx context.Context) ([]*model.VsPromptTemplate, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
	CountByType(ctx context.Context, templateType string) (int64, error)
}

func NewVsPromptTemplateRepository(repository *Repository) VsPromptTemplateRepository {
	return &vsPromptTemplateRepository{Repository: repository}
}

type vsPromptTemplateRepository struct {
	*Repository
}

func (r *vsPromptTemplateRepository) Create(ctx context.Context, template *model.VsPromptTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *vsPromptTemplateRepository) Update(ctx context.Context, template *model.VsPromptTemplate) error {
	return r.db.WithContext(ctx).Model(template).
		Where("template_id = ?", template.TemplateID).
		Omit("created_by", "created_at", "template_id", "is_system").
		Updates(template).Error
}

func (r *vsPromptTemplateRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("template_id = ?", id).Delete(&model.VsPromptTemplate{}).Error
}

func (r *vsPromptTemplateRepository) FindByID(ctx context.Context, id uint64) (*model.VsPromptTemplate, error) {
	var template model.VsPromptTemplate
	if err := r.db.WithContext(ctx).Where("template_id = ?", id).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *vsPromptTemplateRepository) FindWithPagination(ctx context.Context, query *v1.VsPromptTemplateListQuery) ([]*model.VsPromptTemplate, int64, error) {
	var templates []*model.VsPromptTemplate
	db := r.db.WithContext(ctx).Model(&model.VsPromptTemplate{})

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.TemplateType != "" {
		db = db.Where("template_type = ?", query.TemplateType)
	}
	if query.IsSystem != "" {
		db = db.Where("is_system = ?", query.IsSystem)
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

	if err := db.Order("sort_order ASC, template_id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *vsPromptTemplateRepository) FindByType(ctx context.Context, templateType string) ([]*model.VsPromptTemplate, error) {
	var templates []*model.VsPromptTemplate
	if err := r.db.WithContext(ctx).
		Where("template_type = ? AND status = '0'", templateType).
		Order("sort_order ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *vsPromptTemplateRepository) FindAllEnabled(ctx context.Context) ([]*model.VsPromptTemplate, error) {
	var templates []*model.VsPromptTemplate
	if err := r.db.WithContext(ctx).Where("status = '0'").Order("template_type ASC, sort_order ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *vsPromptTemplateRepository) Enable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsPromptTemplate{}).
		Where("template_id = ?", id).Update("status", "0").Error
}

func (r *vsPromptTemplateRepository) Disable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsPromptTemplate{}).
		Where("template_id = ?", id).Update("status", "1").Error
}

func (r *vsPromptTemplateRepository) CountByType(ctx context.Context, templateType string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.VsPromptTemplate{}).
		Where("template_type = ?", templateType).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
