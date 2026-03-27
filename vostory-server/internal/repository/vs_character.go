package repository

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type VsCharacterRepository interface {
	Create(ctx context.Context, character *model.VsCharacter) error
	Update(ctx context.Context, character *model.VsCharacter) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.VsCharacter, error)
	FindWithPagination(ctx context.Context, query *v1.VsCharacterListQuery) ([]*model.VsCharacter, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsCharacter, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsCharacterRepository(repository *Repository) VsCharacterRepository {
	return &vsCharacterRepository{Repository: repository}
}

type vsCharacterRepository struct {
	*Repository
}

func (r *vsCharacterRepository) Create(ctx context.Context, character *model.VsCharacter) error {
	return r.db.WithContext(ctx).Create(character).Error
}

func (r *vsCharacterRepository) Update(ctx context.Context, character *model.VsCharacter) error {
	return r.db.WithContext(ctx).Model(character).
		Where("character_id = ?", character.CharacterID).
		Omit("created_by", "created_at", "character_id", "project_id").
		Updates(character).Error
}

func (r *vsCharacterRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Where("character_id = ?", id).Delete(&model.VsCharacter{}).Error
}

func (r *vsCharacterRepository) FindByID(ctx context.Context, id uint64) (*model.VsCharacter, error) {
	var character model.VsCharacter
	if err := r.db.WithContext(ctx).Where("character_id = ?", id).First(&character).Error; err != nil {
		return nil, err
	}
	return &character, nil
}

func (r *vsCharacterRepository) FindWithPagination(ctx context.Context, query *v1.VsCharacterListQuery) ([]*model.VsCharacter, int64, error) {
	var characters []*model.VsCharacter
	db := r.db.WithContext(ctx).Model(&model.VsCharacter{})

	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Gender != "" {
		db = db.Where("gender = ?", query.Gender)
	}
	if query.Level != "" {
		db = db.Where("level = ?", query.Level)
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

	if err := db.Order("sort_order ASC, character_id DESC").Find(&characters).Error; err != nil {
		return nil, 0, err
	}

	return characters, total, nil
}

func (r *vsCharacterRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]*model.VsCharacter, error) {
	var characters []*model.VsCharacter
	if err := r.db.WithContext(ctx).Where("project_id = ? AND status = '0'", projectID).
		Order("sort_order ASC, character_id DESC").Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *vsCharacterRepository) Enable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsCharacter{}).
		Where("character_id = ?", id).Update("status", "0").Error
}

func (r *vsCharacterRepository) Disable(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.VsCharacter{}).
		Where("character_id = ?", id).Update("status", "1").Error
}
