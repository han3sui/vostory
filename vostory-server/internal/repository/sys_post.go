package repository

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
)

type SysPostRepository interface {
	Create(ctx context.Context, post *model.SysPost) error
	Update(ctx context.Context, post *model.SysPost) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysPost, error)
	FindWithPagination(ctx context.Context, query *v1.SysPostListQuery) ([]*model.SysPost, int64, error)
	FindByPostName(ctx context.Context, postName string) (*model.SysPost, error)
	FindByPostNames(ctx context.Context, postNames []string) ([]*model.SysPost, error)

	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

func NewSysPostRepository(
	repository *Repository,
) SysPostRepository {
	return &sysPostRepository{
		Repository: repository,
	}
}

type sysPostRepository struct {
	*Repository
}

// Create 创建岗位
func (r *sysPostRepository) Create(ctx context.Context, post *model.SysPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// Update 更新岗位
func (r *sysPostRepository) Update(ctx context.Context, post *model.SysPost) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(post).
		Omit("created_by", "created_at", "post_id").
		Updates(post).Error
}

// Delete 删除岗位
func (r *sysPostRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Delete(&model.SysPost{}, id).Error
}

// FindByID 根据ID查找岗位
func (r *sysPostRepository) FindByID(ctx context.Context, id uint) (*model.SysPost, error) {
	var post model.SysPost
	if err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// FindByPostName 根据岗位名称查找岗位
func (r *sysPostRepository) FindByPostName(ctx context.Context, postName string) (*model.SysPost, error) {
	var post model.SysPost
	if err := r.db.WithContext(ctx).Where("post_name = ? AND status = '0'", postName).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// FindByPostNames 根据岗位名称列表查找岗位
func (r *sysPostRepository) FindByPostNames(ctx context.Context, postNames []string) ([]*model.SysPost, error) {
	var posts []*model.SysPost
	if len(postNames) == 0 {
		return posts, nil
	}
	if err := r.db.WithContext(ctx).Where("post_name IN ? AND status = '0'", postNames).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// FindWithPagination 分页查询岗位
func (r *sysPostRepository) FindWithPagination(ctx context.Context, query *v1.SysPostListQuery) ([]*model.SysPost, int64, error) {
	var posts []*model.SysPost
	db := r.db.WithContext(ctx).Model(&model.SysPost{})

	// 应用过滤条件
	if query.PostCode != "" {
		db = db.Where("post_code LIKE ?", "%"+query.PostCode+"%")
	}

	if query.PostName != "" {
		db = db.Where("post_name LIKE ?", "%"+query.PostName+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	db = db.Scopes(model.WithDataScope(ctx))

	// 查询总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 排序
	if err := db.
		Order("post_sort ASC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *sysPostRepository) Enable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysPost{}).Where("post_id = ?", id).Update("status", "0").Error
}

func (r *sysPostRepository) Disable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysPost{}).Where("post_id = ?", id).Update("status", "1").Error
}
