package repository

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type SysApiRepository interface {
	Create(ctx context.Context, api *model.SysApi) error
	Update(ctx context.Context, api *model.SysApi) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysApi, error)
	FindWithPagination(ctx context.Context, query *v1.SysApiListQuery) ([]*model.SysApi, int64, error)
	ListTag(ctx context.Context) ([]string, error)
	FindByPerms(ctx context.Context, perms []string) (map[string]*model.SysApi, error)
}

func NewSysApiRepository(
	repository *Repository,
) SysApiRepository {
	return &sysApiRepository{
		Repository: repository,
	}
}

type sysApiRepository struct {
	*Repository
}

func (r *sysApiRepository) Create(ctx context.Context, api *model.SysApi) error {
	return r.db.WithContext(ctx).Create(api).Error
}

func (r *sysApiRepository) Update(ctx context.Context, api *model.SysApi) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Omit("created_by", "created_at", "id").Updates(api).Error
}

func (r *sysApiRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("id = ?", id).Delete(&model.SysApi{}).Error
}

func (r *sysApiRepository) FindByID(ctx context.Context, id uint) (*model.SysApi, error) {
	var api model.SysApi
	if err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Where("id = ?", id).First(&api).Error; err != nil {
		return nil, err
	}
	return &api, nil
}

func (r *sysApiRepository) FindWithPagination(ctx context.Context, query *v1.SysApiListQuery) ([]*model.SysApi, int64, error) {
	var apis []*model.SysApi
	var total int64
	db := r.db.WithContext(ctx).Model(&model.SysApi{})

	if query.Method != "" {
		db = db.Where("method = ?", query.Method)
	}
	if query.Path != "" {
		db = db.Where("path LIKE ?", "%"+query.Path+"%")
	}

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}

	if query.Desc != "" {
		db = db.Where("description LIKE ?", "%"+query.Desc+"%")
	}

	if query.Tag != "" {
		db = db.Where("tag = ?", query.Tag)
	}

	if query.Perms != "" {
		db = db.Where("perms LIKE ?", "%"+query.Perms+"%")
	}

	// 查询总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 排序
	if err := db.
		Order("created_at DESC").
		Find(&apis).Error; err != nil {
		return nil, 0, err
	}

	return apis, total, nil
}

func (r *sysApiRepository) ListTag(ctx context.Context) ([]string, error) {
	var tags []string
	if err := r.db.WithContext(ctx).Model(&model.SysApi{}).Distinct().Pluck("tag", &tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// FindByPerms 根据权限标识列表查找API，返回perms到API的映射
func (r *sysApiRepository) FindByPerms(ctx context.Context, perms []string) (map[string]*model.SysApi, error) {
	if len(perms) == 0 {
		return make(map[string]*model.SysApi), nil
	}

	var apis []*model.SysApi
	if err := r.db.WithContext(ctx).Where("perms IN ?", perms).Find(&apis).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*model.SysApi)
	for _, api := range apis {
		if api.Perms != "" {
			result[api.Perms] = api
		}
	}
	return result, nil
}
