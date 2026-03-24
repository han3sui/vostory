package repository

import (
	"context"

	"iot-alert-center/internal/model"

	"gorm.io/gorm"
)

type SysUserPostRepository interface {
	Create(ctx context.Context, userPost *model.SysUserPost) error
	CreateBatch(ctx context.Context, userPosts []*model.SysUserPost) error
	DeleteByUserID(ctx context.Context, userID uint) error
	FindPostIDsByUserID(ctx context.Context, userID uint) ([]uint, error)
	FindUserIDsByPostID(ctx context.Context, postID uint) ([]uint, error)
}

type sysUserPostRepository struct {
	db *gorm.DB
}

func NewSysUserPostRepository(db *gorm.DB) SysUserPostRepository {
	return &sysUserPostRepository{db: db}
}

func (r *sysUserPostRepository) Create(ctx context.Context, userPost *model.SysUserPost) error {
	return r.db.WithContext(ctx).Create(userPost).Error
}

func (r *sysUserPostRepository) CreateBatch(ctx context.Context, userPosts []*model.SysUserPost) error {
	if len(userPosts) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&userPosts).Error
}

func (r *sysUserPostRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.SysUserPost{}).Error
}

func (r *sysUserPostRepository) FindPostIDsByUserID(ctx context.Context, userID uint) ([]uint, error) {
	var postIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysUserPost{}).
		Where("user_id = ?", userID).
		Pluck("post_id", &postIDs).Error
	return postIDs, err
}

func (r *sysUserPostRepository) FindUserIDsByPostID(ctx context.Context, postID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.SysUserPost{}).
		Where("post_id = ?", postID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
