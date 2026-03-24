package repository

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type SysDictDataRepository interface {
	Create(ctx context.Context, dictData *model.SysDictData) error
	Update(ctx context.Context, dictData *model.SysDictData) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysDictData, error)
	FindWithPagination(ctx context.Context, query *v1.SysDictDataListQuery) ([]*model.SysDictData, int64, error)
	FindByDictType(ctx context.Context, dictType string) ([]*model.SysDictData, error)
	DeleteByDictType(ctx context.Context, dictType string) error
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

func NewSysDictDataRepository(
	repository *Repository,
) SysDictDataRepository {
	return &sysDictDataRepository{
		Repository: repository,
	}
}

type sysDictDataRepository struct {
	*Repository
}

func (r *sysDictDataRepository) Create(ctx context.Context, dictData *model.SysDictData) error {
	return r.db.WithContext(ctx).Create(dictData).Error
}

func (r *sysDictDataRepository) Update(ctx context.Context, dictData *model.SysDictData) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(dictData).
		Omit("created_by", "created_at", "dict_code").
		Updates(dictData).Error
}

func (r *sysDictDataRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Delete(&model.SysDictData{}, id).Error
}

func (r *sysDictDataRepository) FindByID(ctx context.Context, id uint) (*model.SysDictData, error) {
	var dictData model.SysDictData
	if err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).First(&dictData, id).Error; err != nil {
		return nil, err
	}
	return &dictData, nil
}

func (r *sysDictDataRepository) FindByDictType(ctx context.Context, dictType string) ([]*model.SysDictData, error) {
	var dictDataList []*model.SysDictData
	if err := r.db.WithContext(ctx).
		Where("dict_type = ? AND status = '0'", dictType).
		Order("dict_sort ASC").
		Find(&dictDataList).Error; err != nil {
		return nil, err
	}
	return dictDataList, nil
}

func (r *sysDictDataRepository) DeleteByDictType(ctx context.Context, dictType string) error {
	return r.db.WithContext(ctx).Where("dict_type = ?", dictType).Delete(&model.SysDictData{}).Error
}

func (r *sysDictDataRepository) FindWithPagination(ctx context.Context, query *v1.SysDictDataListQuery) ([]*model.SysDictData, int64, error) {
	var dictDataList []*model.SysDictData
	db := r.db.WithContext(ctx).Model(&model.SysDictData{})

	if query.DictType != "" {
		db = db.Where("dict_type = ?", query.DictType)
	}

	if query.DictLabel != "" {
		db = db.Where("dict_label LIKE ?", "%"+query.DictLabel+"%")
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

	if err := db.Order("dict_sort ASC").Find(&dictDataList).Error; err != nil {
		return nil, 0, err
	}

	return dictDataList, total, nil
}

func (r *sysDictDataRepository) Enable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Model(&model.SysDictData{}).Where("dict_code = ?", id).Update("status", "0").Error
}

func (r *sysDictDataRepository) Disable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Model(&model.SysDictData{}).Where("dict_code = ?", id).Update("status", "1").Error
}
