package repository

import (
	"context"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SysUserRepository interface {
	Create(ctx context.Context, user *model.SysUser) error
	Update(ctx context.Context, user *model.SysUser) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.SysUser, error)
	FindByIDWithoutScope(ctx context.Context, userID uint) (*model.SysUser, error)
	FindWithPagination(ctx context.Context, query *v1.SysUserListQuery) ([]*model.SysUser, int64, error)

	FindByLoginName(ctx context.Context, loginName string) (*model.SysUser, error)
	FindByEmail(ctx context.Context, email string) (*model.SysUser, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.SysUser, error)
	ExistsByLoginName(ctx context.Context, loginName string, excludeID uint) (bool, error)
	ExistsByEmail(ctx context.Context, email string, excludeID uint) (bool, error)
	ExistsByPhoneNumber(ctx context.Context, phoneNumber string, excludeID uint) (bool, error)
	// 密码相关
	EncryptPassword(password string) string
	ComparePassword(hashedPassword, password string) bool
	// 启用/禁用
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
	// 修改密码
	UpdatePassword(ctx context.Context, id uint, newPassword string) error
	UpdateLoginInfo(ctx context.Context, user *model.SysUser) error
}

type SysUserQueryParams struct {
	LoginName   string
	UserName    string
	Status      string
	DeptID      *uint
	Phonenumber string
	Email       string
}

type sysUserRepository struct {
	db *gorm.DB
}

func NewSysUserRepository(db *gorm.DB) SysUserRepository {
	return &sysUserRepository{
		db: db,
	}
}

func (r *sysUserRepository) Create(ctx context.Context, user *model.SysUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *sysUserRepository) Update(ctx context.Context, user *model.SysUser) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysUser{}).
		Omit("created_by", "created_at", "password", "user_id").
		Updates(user).Error
}

func (r *sysUserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Delete(&model.SysUser{}, id).Error
}

func (r *sysUserRepository) FindByID(ctx context.Context, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).
		Scopes(model.WithDataScope(ctx)).
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Preload("Superior", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "user_name", "login_name", "dept_id", "email", "phonenumber", "avatar", "status")
		}).
		Where("user_id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDWithoutScope 根据ID查询用户（不应用数据权限过滤）
func (r *sysUserRepository) FindByIDWithoutScope(ctx context.Context, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Preload("Superior", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "user_name", "login_name", "dept_id", "email", "phonenumber", "avatar", "status")
		}).
		Where("user_id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *sysUserRepository) FindWithPagination(ctx context.Context, query *v1.SysUserListQuery) ([]*model.SysUser, int64, error) {
	var users []*model.SysUser
	var total int64

	// 构建基础查询
	db := r.buildQuery(ctx, query)

	// 应用数据权限过滤
	db = db.Scopes(model.WithDataScope(ctx))

	// 获取总数
	if err := db.Model(&model.SysUser{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.BasePageQuery != nil && query.Page > 0 && query.Size > 0 {
		db = db.Scopes(model.Paginate(query.Page, query.Size))
	}

	// 获取数据，使用分页和预加载
	err := db.
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Preload("Superior", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "user_name", "login_name", "dept_id", "email", "phonenumber", "avatar", "status")
		}).
		Order("user_id ASC").
		Find(&users).Error

	return users, total, err
}

func (r *sysUserRepository) FindByLoginName(ctx context.Context, loginName string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Preload("Superior", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "user_name", "login_name", "dept_id", "email", "phonenumber", "avatar", "status")
		}).
		Where("login_name = ?", loginName).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *sysUserRepository) FindByEmail(ctx context.Context, email string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Preload("Superior", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "user_name", "login_name", "dept_id", "email", "phonenumber", "avatar", "status")
		}).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *sysUserRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Preload("Superior", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "user_name", "login_name", "dept_id", "email", "phonenumber", "avatar", "status")
		}).
		Where("phonenumber = ?", phoneNumber).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *sysUserRepository) ExistsByLoginName(ctx context.Context, loginName string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("login_name = ?", loginName)

	if excludeID > 0 {
		query = query.Where("user_id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *sysUserRepository) ExistsByEmail(ctx context.Context, email string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("email = ?", email)

	if excludeID > 0 {
		query = query.Where("user_id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (r *sysUserRepository) ExistsByPhoneNumber(ctx context.Context, phoneNumber string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("phonenumber = ?", phoneNumber)

	if excludeID > 0 {
		query = query.Where("user_id != ?", excludeID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// 加密密码
func (r *sysUserRepository) EncryptPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

// 比较密码
func (r *sysUserRepository) ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (r *sysUserRepository) buildQuery(ctx context.Context, query *v1.SysUserListQuery) *gorm.DB {
	db := r.db.WithContext(ctx)

	if query == nil {
		return db
	}

	if query.LoginName != "" {
		db = db.Where("login_name LIKE ?", "%"+query.LoginName+"%")
	}

	if query.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+query.UserName+"%")
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.DeptID != nil {
		db = db.Where("dept_id = ?", *query.DeptID)
	}

	if query.Phonenumber != "" {
		db = db.Where("phonenumber LIKE ?", "%"+query.Phonenumber+"%")
	}

	if query.Email != "" {
		db = db.Where("email LIKE ?", "%"+query.Email+"%")
	}

	return db
}

func (r *sysUserRepository) Enable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysUser{}).Where("user_id = ?", id).Update("status", "0").Error
}

func (r *sysUserRepository) Disable(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysUser{}).Where("user_id = ?", id).Update("status", "1").Error
}

func (r *sysUserRepository) UpdatePassword(ctx context.Context, id uint, newPassword string) error {
	now := time.Now()
	user := &model.SysUser{
		Password:      newPassword,
		PwdUpdateDate: &now,
	}
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysUser{}).Where("user_id = ?", id).Updates(user).Error
}

func (r *sysUserRepository) UpdateLoginInfo(ctx context.Context, user *model.SysUser) error {
	return r.db.WithContext(ctx).Scopes(model.WithDataScope(ctx)).Model(&model.SysUser{}).Where("user_id = ?", user.UserID).Select("login_date", "login_ip").Updates(user).Error
}
