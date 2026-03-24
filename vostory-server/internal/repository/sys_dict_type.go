package repository

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type SysDictTypeRepository interface {
	Create(ctx context.Context, dictType *model.SysDictType) error
	Update(ctx context.Context, dictType *model.SysDictType) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysDictType, error)
	FindWithPagination(ctx context.Context, query *v1.SysDictTypeListQuery) ([]*model.SysDictType, int64, error)
	FindByDictType(ctx context.Context, dictType string) (*model.SysDictType, error)
	FindAll(ctx context.Context) ([]*model.SysDictType, error)
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

func NewSysDictTypeRepository(
	repository *Repository,
) SysDictTypeRepository {
	return &sysDictTypeRepository{
		Repository: repository,
	}
}

type sysDictTypeRepository struct {
	*Repository
}

func (r *sysDictTypeRepository) Create(ctx context.Context, dictType *model.SysDictType) error {
	return r.db.WithContext(ctx).Create(dictType).Error
}

func (r *sysDictTypeRepository) Update(ctx context.Context, dictType *model.SysDictType) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(dictType).
		Omit("created_by", "created_at", "dict_id").
		Updates(dictType).Error
}

func (r *sysDictTypeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Delete(&model.SysDictType{}, id).Error
}

func (r *sysDictTypeRepository) FindByID(ctx context.Context, id uint) (*model.SysDictType, error) {
	var dictType model.SysDictType
	if err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).First(&dictType, id).Error; err != nil {
		return nil, err
	}
	return &dictType, nil
}

func (r *sysDictTypeRepository) FindByDictType(ctx context.Context, dictType string) (*model.SysDictType, error) {
	var dt model.SysDictType
	if err := r.db.WithContext(ctx).Where("dict_type = ?", dictType).First(&dt).Error; err != nil {
		return nil, err
	}
	return &dt, nil
}

func (r *sysDictTypeRepository) FindAll(ctx context.Context) ([]*model.SysDictType, error) {
	var dictTypes []*model.SysDictType
	if err := r.db.WithContext(ctx).Where("status = '0'").Order("dict_id ASC").Find(&dictTypes).Error; err != nil {
		return nil, err
	}
	return dictTypes, nil
}

func (r *sysDictTypeRepository) FindWithPagination(ctx context.Context, query *v1.SysDictTypeListQuery) ([]*model.SysDictType, int64, error) {
	var dictTypes []*model.SysDictType
	db := r.db.WithContext(ctx).Model(&model.SysDictType{})

	if query.DictName != "" {
		db = db.Where("dict_name LIKE ?", "%"+query.DictName+"%")
	}

	if query.DictType != "" {
		db = db.Where("dict_type LIKE ?", "%"+query.DictType+"%")
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

	if err := db.Order("dict_id ASC").Find(&dictTypes).Error; err != nil {
		return nil, 0, err
	}

	return dictTypes, total, nil
}

func (r *sysDictTypeRepository) Enable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Model(&model.SysDictType{}).Where("dict_id = ?", id).Update("status", "0").Error
}

func (r *sysDictTypeRepository) Disable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Model(&model.SysDictType{}).Where("dict_id = ?", id).Update("status", "1").Error
}
