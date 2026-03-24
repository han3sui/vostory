package service

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"

	"iot-alert-center/internal/cache"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/internal/utils"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/gin-gonic/gin"
	"github.com/mssola/useragent"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SysUserService interface {
	Create(ctx context.Context, req *v1.SysUserCreateRequest) error
	Update(ctx context.Context, req *v1.SysUserUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysUserDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysUserListQuery) ([]*v1.SysUserDetailResponse, int64, error)
	ResetPassword(ctx context.Context, req *v1.SysUserResetPasswordRequest) error
	ChangeStatus(ctx context.Context, req *v1.SysUserChangeStatusRequest) error
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, newPassword string) error
	GetUserInfo(ctx context.Context, userID uint) (*model.SysUser, []*model.SysMenu, error)
	UpdateCurrentPassword(ctx context.Context, id uint, oldPassword, newPassword string) error
	Login(ctx *gin.Context, req *v1.SysUserLoginRequest) (string, error)
	Logout(ctx *gin.Context, token string) error
	ImportUsers(ctx context.Context, file []byte) (*v1.SysUserImportResponse, error)
	GenerateImportTemplate() ([]byte, error)
	GetUserOptions(ctx context.Context, query *v1.UserOptionQuery) ([]*v1.UserOptionResponse, error)
}

type sysUserService struct {
	db                   *gorm.DB
	conf                 *viper.Viper
	userRepo             repository.SysUserRepository
	userRoleRepo         repository.SysUserRoleRepository
	userPostRepo         repository.SysUserPostRepository
	deptRepo             repository.SysDeptRepository
	roleRepo             repository.SysRoleRepository
	postRepo             repository.SysPostRepository
	roleMenuRepo         repository.SysRoleMenuRepository
	sysLogininforService SysLogininforService
	sysRoleService       SysRoleService
	sysRoleDeptRepo      repository.SysRoleDeptRepository
	userCache            cache.UserCache
}

func NewSysUserService(
	db *gorm.DB,
	conf *viper.Viper,
	userRepo repository.SysUserRepository,
	userRoleRepo repository.SysUserRoleRepository,
	userPostRepo repository.SysUserPostRepository,
	deptRepo repository.SysDeptRepository,
	roleRepo repository.SysRoleRepository,
	postRepo repository.SysPostRepository,
	roleMenuRepo repository.SysRoleMenuRepository,
	sysLogininforService SysLogininforService,
	sysRoleService SysRoleService,
	sysRoleDeptRepo repository.SysRoleDeptRepository,
	userCache cache.UserCache,
) SysUserService {
	return &sysUserService{
		db:                   db,
		conf:                 conf,
		userRepo:             userRepo,
		userRoleRepo:         userRoleRepo,
		userPostRepo:         userPostRepo,
		deptRepo:             deptRepo,
		roleRepo:             roleRepo,
		postRepo:             postRepo,
		roleMenuRepo:         roleMenuRepo,
		sysLogininforService: sysLogininforService,
		sysRoleService:       sysRoleService,
		sysRoleDeptRepo:      sysRoleDeptRepo,
		userCache:            userCache,
	}
}

func (s *sysUserService) Create(ctx context.Context, req *v1.SysUserCreateRequest) error {
	// 检查登录名是否已存在
	exists, err := s.userRepo.ExistsByLoginName(ctx, req.LoginName, 0)
	if err != nil {
		return errors.New("检查登录名失败")
	}
	if exists {
		return errors.New("登录名已存在")
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email, 0)
		if err != nil {
			return errors.New("检查邮箱失败")
		}
		if exists {
			return errors.New("邮箱已存在")
		}
	}

	// 检查手机号是否已存在
	if req.Phonenumber != "" {
		exists, err = s.userRepo.ExistsByPhoneNumber(ctx, req.Phonenumber, 0)
		if err != nil {
			return errors.New("检查手机号失败")
		}
		if exists {
			return errors.New("手机号已存在")
		}
	}

	// 使用事务处理
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 生成盐值和加密密码
		encryptedPassword := s.userRepo.EncryptPassword(req.Password)

		now := time.Now()

		user := &model.SysUser{
			UserDeptID:    req.DeptID,
			SuperiorID:    req.SuperiorID,
			LoginName:     req.LoginName,
			UserName:      req.UserName,
			UserType:      req.UserType,
			Email:         req.Email,
			Phonenumber:   req.Phonenumber,
			Sex:           req.Sex,
			Avatar:        req.Avatar,
			Password:      encryptedPassword,
			Status:        req.Status,
			PwdUpdateDate: &now,
			Remark:        req.Remark,
			BaseModel: model.BaseModel{
				CreatedBy: ctx.Value("login_name").(string),
			},
		}

		// 设置默认值
		if user.UserType == "" {
			user.UserType = "00" // 系统用户
		}
		if user.Sex == "" {
			user.Sex = "0" // 默认男性
		}

		// 创建用户
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 创建用户角色关联
		if len(req.RoleIDs) > 0 {
			var userRoles []*model.SysUserRole
			for _, roleID := range req.RoleIDs {
				userRoles = append(userRoles, &model.SysUserRole{
					UserID: user.UserID,
					RoleID: roleID,
				})
			}
			if err := tx.Create(&userRoles).Error; err != nil {
				return err
			}
		}

		// 创建用户岗位关联
		if len(req.PostIDs) > 0 {
			var userPosts []*model.SysUserPost
			for _, postID := range req.PostIDs {
				userPosts = append(userPosts, &model.SysUserPost{
					UserID: user.UserID,
					PostID: postID,
				})
			}
			if err := tx.Create(&userPosts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sysUserService) Update(ctx context.Context, req *v1.SysUserUpdateRequest) error {
	// 检查用户是否存在
	_, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查登录名是否已存在（排除当前用户）
	exists, err := s.userRepo.ExistsByLoginName(ctx, req.LoginName, req.UserID)
	if err != nil {
		return errors.New("检查登录名失败")
	}
	if exists {
		return errors.New("登录名已存在")
	}

	// 检查邮箱是否已存在（排除当前用户）
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email, req.UserID)
		if err != nil {
			return errors.New("检查邮箱失败")
		}
		if exists {
			return errors.New("邮箱已存在")
		}
	}

	// 检查手机号是否已存在（排除当前用户）
	if req.Phonenumber != "" {
		exists, err = s.userRepo.ExistsByPhoneNumber(ctx, req.Phonenumber, req.UserID)
		if err != nil {
			return errors.New("检查手机号失败")
		}
		if exists {
			return errors.New("手机号已存在")
		}
	}

	// 使用事务处理
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 构建更新的用户信息
		user := &model.SysUser{
			UserID:      req.UserID,
			UserDeptID:  req.DeptID,
			SuperiorID:  req.SuperiorID,
			LoginName:   req.LoginName,
			UserName:    req.UserName,
			UserType:    req.UserType,
			Email:       req.Email,
			Phonenumber: req.Phonenumber,
			Sex:         req.Sex,
			Avatar:      req.Avatar,
			Status:      req.Status,
			Remark:      req.Remark,
			BaseModel: model.BaseModel{
				UpdatedBy: ctx.Value("login_name").(string),
			},
		}

		// 如果提供了新密码，则更新密码
		if req.Password != "" {
			user.Password = s.userRepo.EncryptPassword(req.Password)
			now := time.Now()
			user.PwdUpdateDate = &now
		}

		if err := tx.Model(&model.SysUser{}).Where("user_id = ?", req.UserID).
			Omit("created_by", "created_at", "user_id").
			Updates(user).Error; err != nil {
			return err
		}

		// 删除原有的用户角色关联
		if err := tx.Where("user_id = ?", req.UserID).Unscoped().Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}

		// 重新创建用户角色关联
		if len(req.RoleIDs) > 0 {
			var userRoles []*model.SysUserRole
			for _, roleID := range req.RoleIDs {
				userRoles = append(userRoles, &model.SysUserRole{
					UserID: req.UserID,
					RoleID: roleID,
				})
			}
			if err := tx.Create(&userRoles).Error; err != nil {
				return err
			}
		}

		// 删除原有的用户岗位关联
		if err := tx.Where("user_id = ?", req.UserID).Unscoped().Delete(&model.SysUserPost{}).Error; err != nil {
			return err
		}

		// 重新创建用户岗位关联
		if len(req.PostIDs) > 0 {
			var userPosts []*model.SysUserPost
			for _, postID := range req.PostIDs {
				userPosts = append(userPosts, &model.SysUserPost{
					UserID: req.UserID,
					PostID: postID,
				})
			}
			if err := tx.Create(&userPosts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sysUserService) Delete(ctx context.Context, id uint) error {
	// 检查用户是否存在
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.UserType == "99" {
		return errors.New("非法用户类型")
	}

	// 使用事务删除
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除用户角色关联
		if err := tx.Where("user_id = ?", id).Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}

		// 删除用户岗位关联
		if err := tx.Where("user_id = ?", id).Delete(&model.SysUserPost{}).Error; err != nil {
			return err
		}

		// 删除用户
		if err := tx.Delete(&model.SysUser{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *sysUserService) FindByID(ctx context.Context, id uint) (*v1.SysUserDetailResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取用户角色ID
	roleIDs, err := s.userRoleRepo.FindRoleIDsByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取用户岗位ID
	postIDs, err := s.userPostRepo.FindPostIDsByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertToDetailResponse(user, roleIDs, postIDs), nil
}

func (s *sysUserService) FindWithPagination(ctx context.Context, query *v1.SysUserListQuery) ([]*v1.SysUserDetailResponse, int64, error) {

	// 查询数据
	users, total, err := s.userRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var records []*v1.SysUserDetailResponse
	for _, user := range users {
		// 获取用户角色ID
		roleIDs, err := s.userRoleRepo.FindRoleIDsByUserID(ctx, user.UserID)
		if err != nil {
			return nil, 0, err
		}

		// 获取用户岗位ID
		postIDs, err := s.userPostRepo.FindPostIDsByUserID(ctx, user.UserID)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, s.convertToDetailResponse(user, roleIDs, postIDs))
	}

	return records, total, nil
}

func (s *sysUserService) ResetPassword(ctx context.Context, req *v1.SysUserResetPasswordRequest) error {
	// 获取用户信息
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 生成新的加密密码
	encryptedPassword := s.userRepo.EncryptPassword(req.NewPassword)

	// 更新密码
	user.Password = encryptedPassword
	now := time.Now()
	user.PwdUpdateDate = &now

	return s.userRepo.Update(ctx, user)
}

func (s *sysUserService) ChangeStatus(ctx context.Context, req *v1.SysUserChangeStatusRequest) error {
	// 获取用户信息
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 更新状态
	user.Status = req.Status

	return s.userRepo.Update(ctx, user)
}

// 辅助方法
func (s *sysUserService) convertToDetailResponse(user *model.SysUser, roleIDs, postIDs []uint) *v1.SysUserDetailResponse {
	response := &v1.SysUserDetailResponse{
		UserID:        user.UserID,
		DeptID:        user.UserDeptID,
		SuperiorID:    user.SuperiorID,
		LoginName:     user.LoginName,
		UserName:      user.UserName,
		UserType:      user.UserType,
		Email:         user.Email,
		Phonenumber:   user.Phonenumber,
		Sex:           user.Sex,
		Avatar:        user.Avatar,
		Status:        user.Status,
		LoginIP:       user.LoginIP,
		LoginDate:     user.LoginDate,
		PwdUpdateDate: user.PwdUpdateDate,
		Remark:        user.Remark,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		RoleIDs:       roleIDs,
		PostIDs:       postIDs,
		Roles:         user.Roles,
		Posts:         user.Posts,
		Dept:          user.Dept,
		Superior:      user.Superior,
	}

	// 如果有部门ID，可以获取部门名称
	if user.UserDeptID != nil {
		if dept, err := s.deptRepo.FindByID(context.Background(), *user.UserDeptID); err == nil {
			response.DeptName = dept.DeptName
		}
	}

	// 如果有直属上级ID，可以获取上级姓名
	if user.SuperiorID != nil && *user.SuperiorID != 0 {
		if superior, err := s.userRepo.FindByID(context.Background(), *user.SuperiorID); err == nil {
			response.SuperiorName = superior.UserName
		}
	}

	return response
}

func (s *sysUserService) Enable(ctx context.Context, id uint) error {
	return s.userRepo.Enable(ctx, id)
}

func (s *sysUserService) Disable(ctx context.Context, id uint) error {
	return s.userRepo.Disable(ctx, id)
}

func (s *sysUserService) UpdatePassword(ctx context.Context, id uint, newPassword string) error {
	// 生成加密密码
	encryptedPassword := s.userRepo.EncryptPassword(newPassword)
	return s.userRepo.UpdatePassword(ctx, id, encryptedPassword)
}

func (s *sysUserService) UpdateCurrentPassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	// 获取用户信息（用户修改自己的密码，不需要数据权限过滤）
	user, err := s.userRepo.FindByIDWithoutScope(ctx, id)
	if err != nil {
		return errors.Wrap(err, "获取用户信息失败")
	}

	// 验证原密码
	if !s.userRepo.ComparePassword(user.Password, oldPassword) {
		return errors.New("原密码错误")
	}

	// 生成加密密码
	encryptedPassword := s.userRepo.EncryptPassword(newPassword)
	return s.userRepo.UpdatePassword(ctx, id, encryptedPassword)
}

func (s *sysUserService) Login(ctx *gin.Context, req *v1.SysUserLoginRequest) (string, error) {
	user, err := s.userRepo.FindByLoginName(ctx, req.LoginName)

	now := time.Now()
	ip, _ := GetIP(ctx.Request)
	userAgentStr := ctx.Request.UserAgent()

	ua := useragent.New(userAgentStr)

	name, version := ua.Browser()

	loginLocation, lerr := utils.GetCityByIp(ip)

	if lerr != nil {
		loginLocation = ""
		fmt.Println(lerr.Error())
	}

	loginInfo := &model.SysLogininfor{
		LoginName:     req.LoginName,
		IPAddr:        ip,
		LoginLocation: loginLocation,
		Browser:       name + " " + version,
		OS:            ua.OS(),
		Status:        "0",
		Msg:           "",
		LoginTime:     now,
	}

	if err != nil {
		loginInfo.Status = "1"
		loginInfo.Msg = "用户不存在"
		s.sysLogininforService.Create(ctx, loginInfo)
		return "", errors.New("用户不存在")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		loginInfo.Status = "1"
		loginInfo.Msg = "密码错误"
		s.sysLogininforService.Create(ctx, loginInfo)
		return "", errors.New("密码错误")
	}

	roleIDs := make([]uint, 0)
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.RoleID)
	}

	if user.Status == "1" && user.UserType != "99" {
		loginInfo.Status = "1"
		loginInfo.Msg = "用户已禁用"
		s.sysLogininforService.Create(ctx, loginInfo)
		return "", errors.New("用户已禁用")
	}

	dataScope, dataScopeDepts, err := s.GetDataScope(ctx, roleIDs, user.UserType)
	if err != nil {
		loginInfo.Status = "1"
		loginInfo.Msg = "获取数据权限失败"
		s.sysLogininforService.Create(ctx, loginInfo)
		return "", err
	}

	token, err := s.GrantToken(user.UserID, *user.UserDeptID, roleIDs, user.LoginName, dataScope, dataScopeDepts)
	if err != nil {
		loginInfo.Status = "1"
		loginInfo.Msg = "生成token失败"
		s.sysLogininforService.Create(ctx, loginInfo)
		return "", err
	}

	//更新登录时间、登录IP
	updateData := &model.SysUser{
		LoginDate: &now,
		LoginIP:   ip,
		UserID:    user.UserID,
	}

	loginInfo.Status = "0"
	loginInfo.Msg = ""
	s.sysLogininforService.Create(ctx, loginInfo)
	s.userRepo.UpdateLoginInfo(ctx, updateData)

	return token, nil
}

// GetIP returns request real ip.
func GetIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}

func (s *sysUserService) GetUserInfo(ctx context.Context, userID uint) (*model.SysUser, []*model.SysMenu, error) {
	// 获取当前登录用户信息，不需要数据权限过滤
	user, err := s.userRepo.FindByIDWithoutScope(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	roleIDs := make([]uint, 0)
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.RoleID)
	}

	//根据角色ID，查找关联的菜单
	menus, err := s.roleMenuRepo.FindMenusByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, nil, err
	}

	return user, menus, nil
}

func (s *sysUserService) GrantToken(userID uint, deptID uint, roleIDs []uint, loginName string, dataScope string, dataScopeDepts []uint) (string, error) {
	// 1. 构造包含关键信息的结构
	tokenData := v1.TokenData{
		Exp:            time.Now().Add(time.Hour * 24).Unix(),
		UserId:         userID,
		DeptId:         deptID,
		RoleIds:        roleIDs,
		LoginName:      loginName,
		DataScope:      dataScope,
		DataScopeDepts: dataScopeDepts,
	}

	// 2. 获取用户信息和菜单权限
	user, menus, err := s.GetUserInfo(context.Background(), userID)
	if err != nil {
		return "", fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 3. 获取用户的API权限列表

	apiPermissions := make([]string, 0)
	for _, menu := range menus {
		if menu.MenuType == "F" {
			apiPermissions = append(apiPermissions, menu.Perms)
		}
	}

	// 4. 获取用户的API路径映射
	apiPathMap, err := s.getUserApiPathMap(context.Background(), apiPermissions, user.UserType)
	if err != nil {
		return "", fmt.Errorf("获取API路径映射失败: %w", err)
	}

	// 5. 生成缓存使用的令牌

	// 构造包含时间戳的数据
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%d:%d", userID, timestamp)
	key := s.conf.GetString("security.jwt.key") // 从配置获取密钥
	token := cryptor.HmacSha256(data, key)

	// 6. 创建用户信息对象并存入缓存
	userInfo := cache.FromTokenDataWithApiInfo(&tokenData, user, menus, apiPermissions, apiPathMap)

	err = s.userCache.StoreUserToken(token, userInfo)
	if err != nil {
		return "", fmt.Errorf("存入缓存失败: %w", err)
	}

	return token, nil
}

// getUserApiPathMap 获取用户的API路径映射
func (s *sysUserService) getUserApiPathMap(ctx context.Context, apiPermissions []string, userType string) (map[string][]cache.ApiPathInfo, error) {
	// 超级管理员拥有所有API权限
	if userType == "99" {
		return map[string][]cache.ApiPathInfo{
			"*": {{Path: "*", Perms: "*"}}, // 通配符表示所有权限
		}, nil
	}

	// 如果没有权限，返回空映射
	if len(apiPermissions) == 0 {
		return map[string][]cache.ApiPathInfo{}, nil
	}

	// 查询用户权限对应的API路径信息
	var apis []model.SysApi
	if err := s.db.WithContext(ctx).Where("perms IN ?", apiPermissions).Find(&apis).Error; err != nil {
		return nil, fmt.Errorf("查询API路径信息失败: %w", err)
	}

	// 按HTTP方法分组API路径信息
	apiPathMap := make(map[string][]cache.ApiPathInfo)
	for _, api := range apis {
		pathInfo := cache.ApiPathInfo{
			Path:  api.Path,
			Perms: api.Perms,
		}
		apiPathMap[api.Method] = append(apiPathMap[api.Method], pathInfo)
	}

	return apiPathMap, nil
}

// 获取数据权限
func (s *sysUserService) GetDataScope(ctx context.Context, roleIDs []uint, userType string) (string, []uint, error) {
	if userType == "99" {
		return "1", []uint{}, nil
	}

	// 检查角色是否存在
	if len(roleIDs) == 0 {
		return "", nil, errors.New("没有角色")
	}

	// 初始化最终权限
	finalDataScope := "5" // 默认为最小权限：仅本人数据
	var allDeptIDs []uint

	// 批量获取所有角色，查找最大权限集
	roles, err := s.sysRoleService.FindByIDs(ctx, roleIDs)
	if err != nil {
		return "", nil, err
	}

	// 遍历所有角色，查找最大权限集
	for _, role := range roles {
		// 全部数据权限是最高级别，一旦有此权限直接返回
		if role.DataScope == "1" {
			return "1", []uint{}, nil
		}

		// 判断当前角色权限是否优于已有权限
		currentScopeLevel, _ := strconv.Atoi(role.DataScope)
		finalScopeLevel, _ := strconv.Atoi(finalDataScope)

		if currentScopeLevel < finalScopeLevel {
			// 当前角色权限更高，更新最终权限
			finalDataScope = role.DataScope
		}

		// 合并部门ID（针对自定义数据权限）
		if role.DataScope == "2" {
			// 获取角色关联的部门ID（这部分仍需单独查询）
			customDeptIDs, err := s.sysRoleDeptRepo.FindDeptIDsByRoleID(ctx, role.RoleID)
			if err == nil && len(customDeptIDs) > 0 {
				allDeptIDs = append(allDeptIDs, customDeptIDs...)
			}
		}
	}

	// 根据最终确定的数据范围处理部门ID列表
	var finalDeptIDs []uint

	if finalDataScope == "2" {
		// 自定义数据权限：使用合并后的部门ID
		finalDeptIDs = removeDuplicateUints(allDeptIDs)
	} else if finalDataScope == "3" {
		// 本部门数据权限：获取用户部门ID
		deptID, ok := ctx.Value("dept_id").(uint)
		if ok && deptID > 0 {
			finalDeptIDs = []uint{deptID}
		}
	} else if finalDataScope == "4" {
		// 本部门及以下数据权限：获取用户部门及其子部门
		deptID, ok := ctx.Value("dept_id").(uint)
		if ok && deptID > 0 {
			// 获取子部门ID
			childDeptIDs, err := s.getChildDeptIDs(ctx, deptID)
			if err == nil {
				finalDeptIDs = append([]uint{deptID}, childDeptIDs...)
			} else {
				finalDeptIDs = []uint{deptID}
			}
		}
	}
	// 角色为"5"时不需要设置部门ID，因为是按用户ID过滤

	return finalDataScope, finalDeptIDs, nil
}

// 辅助方法：获取子部门ID列表
func (s *sysUserService) getChildDeptIDs(ctx context.Context, parentDeptID uint) ([]uint, error) {
	// 构建查询条件：ancestors 包含当前部门ID
	allDepts, _, err := s.deptRepo.FindWithPagination(ctx, &v1.SysDeptListQuery{
		Status: "0", // 只查询启用的部门
	})
	if err != nil {
		return nil, err
	}

	var childDeptIDs []uint
	currentDeptIDStr := strconv.FormatUint(uint64(parentDeptID), 10)

	for _, dept := range allDepts {
		if dept.DeptID == parentDeptID || dept.Ancestors == "" {
			continue
		}
		for _, ancestor := range strings.Split(dept.Ancestors, ",") {
			if strings.TrimSpace(ancestor) == currentDeptIDStr {
				childDeptIDs = append(childDeptIDs, dept.DeptID)
				break
			}
		}
	}

	return childDeptIDs, nil
}

// removeDuplicateUints 去重uint切片
func removeDuplicateUints(slice []uint) []uint {
	keys := make(map[uint]bool)
	var result []uint

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Logout 退出登录
func (s *sysUserService) Logout(ctx *gin.Context, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}
	return s.userCache.RemoveToken(token)
}

// ImportUsers 导入用户（两轮处理：第一轮创建用户，第二轮更新直属上级）
func (s *sysUserService) ImportUsers(ctx context.Context, fileData []byte) (*v1.SysUserImportResponse, error) {
	// 解析Excel文件
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, errors.Wrap(err, "解析Excel文件失败")
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, errors.Wrap(err, "读取工作表失败")
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel文件中没有数据")
	}

	response := &v1.SysUserImportResponse{
		Errors: make([]v1.SysUserImportError, 0),
	}

	// 读取表头，判断直属上级匹配模式（默认按登录账号匹配）
	superiorMatchByLoginName := true
	headerRow := rows[0]
	if len(headerRow) > 3 {
		superiorHeader := strings.TrimSpace(headerRow[3])
		if strings.Contains(superiorHeader, "用户昵称") || strings.Contains(superiorHeader, "姓名") {
			superiorMatchByLoginName = false
		}
	}

	// 用于记录需要设置直属上级的用户
	type superiorRelation struct {
		LoginName    string
		SuperiorName string
	}
	var superiorRelations []superiorRelation

	// 收集本次导入的所有登录账号和用户昵称，用于验证直属上级是否在本次导入中
	importLoginNames := make(map[string]bool)
	importUserNames := make(map[string]bool)
	for i := 1; i < len(rows); i++ {
		userData := s.parseImportRow(rows[i])
		if userData.LoginName != "" {
			importLoginNames[userData.LoginName] = true
		}
		if userData.UserName != "" {
			importUserNames[userData.UserName] = true
		}
	}

	// 第一轮：创建用户（不设置直属上级）
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		lineNum := i + 1
		rowErrors := make([]string, 0)

		// 解析行数据
		userData := s.parseImportRow(row)

		// 验证必填字段
		if userData.LoginName == "" {
			rowErrors = append(rowErrors, "登录账号不能为空")
		}
		if userData.UserName == "" {
			rowErrors = append(rowErrors, "用户姓名不能为空")
		}

		// 检查登录名是否已存在
		if userData.LoginName != "" {
			exists, _ := s.userRepo.ExistsByLoginName(ctx, userData.LoginName, 0)
			if exists {
				rowErrors = append(rowErrors, "登录账号已存在")
			}
		}

		// 检查邮箱是否已存在
		if userData.Email != "" {
			exists, _ := s.userRepo.ExistsByEmail(ctx, userData.Email, 0)
			if exists {
				rowErrors = append(rowErrors, "邮箱已存在")
			}
		}

		// 检查手机号是否已存在
		if userData.Phonenumber != "" {
			exists, _ := s.userRepo.ExistsByPhoneNumber(ctx, userData.Phonenumber, 0)
			if exists {
				rowErrors = append(rowErrors, "手机号已存在")
			}
		}

		// 解析部门
		var deptID *uint
		if userData.DeptName != "" {
			dept, err := s.deptRepo.FindByDeptName(ctx, userData.DeptName)
			if err != nil {
				rowErrors = append(rowErrors, fmt.Sprintf("部门[%s]不存在", userData.DeptName))
			} else {
				deptID = &dept.DeptID
			}
		}

		// 验证直属上级（检查是否存在于系统中或本次导入中）
		if userData.SuperiorName != "" {
			var foundInImport bool
			if superiorMatchByLoginName {
				foundInImport = importLoginNames[userData.SuperiorName]
			} else {
				foundInImport = importUserNames[userData.SuperiorName]
			}

			if !foundInImport {
				// 不在本次导入中，检查是否已存在于系统中
				_, err := s.findSuperiorUser(ctx, userData.SuperiorName, superiorMatchByLoginName)
				if err != nil {
					rowErrors = append(rowErrors, fmt.Sprintf("直属上级[%s]不存在", userData.SuperiorName))
				}
			}
		}

		// 解析角色
		var roleIDs []uint
		if userData.RoleNames != "" {
			roleNames := strings.Split(userData.RoleNames, ",")
			for _, name := range roleNames {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				role, err := s.roleRepo.FindByRoleName(ctx, name)
				if err != nil {
					rowErrors = append(rowErrors, fmt.Sprintf("角色[%s]不存在", name))
				} else {
					roleIDs = append(roleIDs, role.RoleID)
				}
			}
		}

		// 解析岗位
		var postIDs []uint
		if userData.PostNames != "" {
			postNames := strings.Split(userData.PostNames, ",")
			for _, name := range postNames {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				post, err := s.postRepo.FindByPostName(ctx, name)
				if err != nil {
					rowErrors = append(rowErrors, fmt.Sprintf("岗位[%s]不存在", name))
				} else {
					postIDs = append(postIDs, post.PostID)
				}
			}
		}

		// 如果有错误，记录并跳过
		if len(rowErrors) > 0 {
			response.FailCount++
			response.Errors = append(response.Errors, v1.SysUserImportError{
				LineNum: lineNum,
				Errors:  rowErrors,
			})
			continue
		}

		// 解析性别
		sex := "0" // 默认男
		switch userData.Sex {
		case "女":
			sex = "1"
		case "未知":
			sex = "2"
		}

		// 解析状态
		status := "0" // 默认正常
		if userData.Status == "停用" {
			status = "1"
		}

		// 密码处理
		password := userData.Password
		if password == "" {
			password = "123456"
		}

		// 第一轮创建用户时不设置直属上级
		createReq := &v1.SysUserCreateRequest{
			BaseSysUserRequest: v1.BaseSysUserRequest{
				DeptID:      deptID,
				SuperiorID:  nil, // 第一轮不设置直属上级
				LoginName:   userData.LoginName,
				UserName:    userData.UserName,
				Email:       userData.Email,
				Phonenumber: userData.Phonenumber,
				Sex:         sex,
				Status:      status,
				RoleIDs:     roleIDs,
				PostIDs:     postIDs,
			},
			Password: cryptor.Sha256(password),
		}

		err = s.Create(ctx, createReq)
		if err != nil {
			response.FailCount++
			response.Errors = append(response.Errors, v1.SysUserImportError{
				LineNum: lineNum,
				Errors:  []string{err.Error()},
			})
			continue
		}

		// 记录需要设置直属上级的用户
		if userData.SuperiorName != "" {
			superiorRelations = append(superiorRelations, superiorRelation{
				LoginName:    userData.LoginName,
				SuperiorName: userData.SuperiorName,
			})
		}

		response.SuccessCount++
	}

	// 第二轮：更新直属上级关系
	for _, relation := range superiorRelations {
		// 查找直属上级（根据匹配模式）
		superior, err := s.findSuperiorUser(ctx, relation.SuperiorName, superiorMatchByLoginName)
		if err != nil {
			// 直属上级不存在，跳过（理论上不会发生，因为第一轮已经验证过）
			continue
		}

		// 查找当前用户
		user, err := s.userRepo.FindByLoginName(ctx, relation.LoginName)
		if err != nil {
			continue
		}

		// 更新直属上级
		user.SuperiorID = &superior.UserID
		s.db.WithContext(ctx).Model(&model.SysUser{}).
			Where("user_id = ?", user.UserID).
			Update("superior_id", superior.UserID)
	}

	return response, nil
}

// findSuperiorUser 根据登录账号或用户昵称查找直属上级
func (s *sysUserService) findSuperiorUser(ctx context.Context, nameOrLogin string, byLoginName bool) (*model.SysUser, error) {
	if byLoginName {
		// 按登录账号精确匹配
		return s.userRepo.FindByLoginName(ctx, nameOrLogin)
	}

	// 按用户昵称匹配
	var users []*model.SysUser
	err := s.db.WithContext(ctx).
		Where("user_name = ?", nameOrLogin).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("用户不存在")
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("存在%d个同名用户[%s]，请使用登录账号模式", len(users), nameOrLogin)
	}

	return users[0], nil
}

// parseImportRow 解析导入行数据
func (s *sysUserService) parseImportRow(row []string) *v1.SysUserImportRequest {
	data := &v1.SysUserImportRequest{}

	if len(row) > 0 {
		data.LoginName = strings.TrimSpace(row[0])
	}
	if len(row) > 1 {
		data.UserName = strings.TrimSpace(row[1])
	}
	if len(row) > 2 {
		data.DeptName = strings.TrimSpace(row[2])
	}
	if len(row) > 3 {
		data.SuperiorName = strings.TrimSpace(row[3])
	}
	if len(row) > 4 {
		data.Email = strings.TrimSpace(row[4])
	}
	if len(row) > 5 {
		data.Phonenumber = strings.TrimSpace(row[5])
	}
	if len(row) > 6 {
		data.Sex = strings.TrimSpace(row[6])
	}
	if len(row) > 7 {
		data.Status = strings.TrimSpace(row[7])
	}
	if len(row) > 8 {
		data.RoleNames = strings.TrimSpace(row[8])
	}
	if len(row) > 9 {
		data.PostNames = strings.TrimSpace(row[9])
	}
	if len(row) > 10 {
		data.Password = strings.TrimSpace(row[10])
	}

	return data
}

// GenerateImportTemplate 生成导入模板
func (s *sysUserService) GenerateImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "用户导入模板"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 设置表头（直属上级列默认使用登录账号模式）
	headers := []string{
		"登录账号*", "用户姓名*", "部门名称", "直属上级(登录账号)", "邮箱", "手机号码",
		"性别(男/女)", "状态(正常/停用)", "角色(多个用逗号分隔)", "岗位(多个用逗号分隔)", "密码(为空则使用默认密码)",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// 为"直属上级"列的表头单元格(D1)设置下拉框
	dvSuperiorHeader := excelize.NewDataValidation(true)
	dvSuperiorHeader.Sqref = "D1"
	dvSuperiorHeader.SetDropList([]string{"直属上级(登录账号)", "直属上级(用户昵称)"})
	f.AddDataValidation(sheetName, dvSuperiorHeader)

	// 设置列宽
	colWidths := []float64{15, 15, 20, 22, 25, 15, 12, 15, 25, 25, 25}
	for i, width := range colWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, width)
	}

	// 设置表头样式
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	f.SetRowStyle(sheetName, 1, 1, style)

	// 添加示例数据
	exampleData := [][]interface{}{
		{"zhangsan", "张三", "技术部", "", "zhangsan@example.com", "13800138001", "男", "正常", "普通用户", "开发工程师", ""},
		{"lisi", "李四", "产品部", "zhangsan", "lisi@example.com", "13800138002", "女", "正常", "普通用户,管理员", "产品经理", "123456"},
	}

	for i, rowData := range exampleData {
		for j, value := range rowData {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// 写入buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetUserOptions 获取用户选项列表（简化版，用于下拉选择）
func (s *sysUserService) GetUserOptions(ctx context.Context, query *v1.UserOptionQuery) ([]*v1.UserOptionResponse, error) {
	db := s.db.WithContext(ctx).Model(&model.SysUser{}).
		Select("sys_user.user_id, sys_user.user_name, sys_user.login_name").
		Where("sys_user.status = ?", "0") // 只查询启用状态的用户

	// 关键词搜索（用户名或登录名）
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("sys_user.user_name LIKE ? OR sys_user.login_name LIKE ?", keyword, keyword)
	}

	// 部门筛选
	if query.DeptID != nil && *query.DeptID > 0 {
		db = db.Where("sys_user.dept_id = ?", *query.DeptID)
	}

	var users []*v1.UserOptionResponse
	err := db.Order("sys_user.user_name ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}
