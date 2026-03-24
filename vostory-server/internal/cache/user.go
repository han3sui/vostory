package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"

	"github.com/redis/go-redis/v9"
)

// UserCache 定义用户缓存接口
type UserCache interface {
	// StoreUserToken 存储用户令牌和用户信息
	StoreUserToken(token string, userInfo *UserInfo) error

	// GetUserInfo 根据令牌获取用户信息
	GetUserInfo(token string) (*UserInfo, error)

	// RemoveToken 移除令牌
	RemoveToken(token string) error

	// RefreshTokenExpiration 刷新令牌过期时间
	RefreshTokenExpiration(token string, newExpiredAt int64) error

	// Close 关闭缓存
	Close() error
}

// UserInfo 用户信息结构
type UserInfo struct {
	UserID         uint                     `json:"user_id"`
	DeptID         uint                     `json:"dept_id"`
	RoleIDs        []uint                   `json:"role_ids"`
	LoginName      string                   `json:"login_name"`
	DataScope      string                   `json:"data_scope"`
	DataScopeDepts []uint                   `json:"data_scope_depts"`
	User           *model.SysUser           `json:"user,omitempty"`
	Menus          []*model.SysMenu         `json:"menus,omitempty"`
	ExpiredAt      int64                    `json:"expired_at"`
	ApiPermissions []string                 `json:"api_permissions,omitempty"`
	ApiPathMap     map[string][]ApiPathInfo `json:"api_path_map,omitempty"` // 新增：API路径映射，key为HTTP方法，value为该方法下的API路径信息
}

// ApiPathInfo API路径信息
type ApiPathInfo struct {
	Path  string `json:"path"`  // API路径模式
	Perms string `json:"perms"` // 权限标识
}

// redisCache 使用Redis实现的用户缓存
type redisCache struct {
	rdb *redis.Client
	ctx context.Context
}

var userTokenKey = "user:token:%s"

// NewUserCache 创建基于Redis的用户缓存
func NewUserCache(rdb *redis.Client) UserCache {
	return &redisCache{
		rdb: rdb,
		ctx: context.Background(),
	}
}

// StoreUserToken 存储用户令牌和用户信息
func (c *redisCache) StoreUserToken(token string, userInfo *UserInfo) error {
	// 序列化用户信息
	data, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("序列化用户信息失败: %w", err)
	}

	// 计算过期时间
	expiration := time.Until(time.Unix(userInfo.ExpiredAt, 0))
	if expiration <= 0 {
		return fmt.Errorf("令牌已过期")
	}

	// 存入Redis
	key := fmt.Sprintf(userTokenKey, token)
	return c.rdb.Set(c.ctx, key, data, expiration).Err()
}

// GetUserInfo 根据令牌获取用户信息
func (c *redisCache) GetUserInfo(token string) (*UserInfo, error) {
	// 从Redis获取
	key := fmt.Sprintf(userTokenKey, token)
	data, err := c.rdb.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("令牌不存在或已过期")
		}
		return nil, fmt.Errorf("获取缓存数据失败: %w", err)
	}

	// 反序列化
	var userInfo UserInfo
	if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
		return nil, fmt.Errorf("反序列化用户信息失败: %w", err)
	}

	// 检查是否过期
	if userInfo.ExpiredAt < time.Now().Unix() {
		c.RemoveToken(token) // 移除过期令牌
		return nil, fmt.Errorf("令牌已过期")
	}

	return &userInfo, nil
}

// RefreshTokenExpiration 刷新令牌过期时间
func (c *redisCache) RefreshTokenExpiration(token string, newExpiredAt int64) error {
	// 从Redis获取当前用户信息
	key := fmt.Sprintf(userTokenKey, token)
	data, err := c.rdb.Get(c.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("令牌不存在或已过期")
		}
		return fmt.Errorf("获取缓存数据失败: %w", err)
	}

	// 反序列化
	var userInfo UserInfo
	if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
		return fmt.Errorf("反序列化用户信息失败: %w", err)
	}

	// 更新过期时间
	userInfo.ExpiredAt = newExpiredAt

	// 重新序列化
	newData, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("序列化用户信息失败: %w", err)
	}

	// 计算新的过期时间
	expiration := time.Until(time.Unix(newExpiredAt, 0))
	if expiration <= 0 {
		return fmt.Errorf("新的过期时间无效")
	}

	// 更新Redis中的数据和过期时间
	return c.rdb.Set(c.ctx, key, newData, expiration).Err()
}

// RemoveToken 移除令牌
func (c *redisCache) RemoveToken(token string) error {
	key := fmt.Sprintf(userTokenKey, token)
	return c.rdb.Del(c.ctx, key).Err()
}

// Close 关闭缓存
func (c *redisCache) Close() error {
	return c.rdb.Close()
}

// FromTokenData 从TokenData创建UserInfo
func FromTokenData(tokenData *v1.TokenData, user *model.SysUser, menus []*model.SysMenu) *UserInfo {
	apiPermissions := make([]string, 0)
	for _, menu := range menus {
		if menu.MenuType == "F" {
			apiPermissions = append(apiPermissions, menu.Perms)
		}
	}
	return &UserInfo{
		UserID:         tokenData.UserId,
		DeptID:         tokenData.DeptId,
		RoleIDs:        tokenData.RoleIds,
		LoginName:      tokenData.LoginName,
		DataScope:      tokenData.DataScope,
		DataScopeDepts: tokenData.DataScopeDepts,
		User:           user,
		Menus:          menus,
		ExpiredAt:      tokenData.Exp,
		ApiPermissions: apiPermissions,
	}
}

// FromTokenDataWithApiInfo 从TokenData创建UserInfo，传入API权限和路径映射
func FromTokenDataWithApiInfo(tokenData *v1.TokenData, user *model.SysUser, menus []*model.SysMenu, apiPermissions []string, apiPathMap map[string][]ApiPathInfo) *UserInfo {
	return &UserInfo{
		UserID:         tokenData.UserId,
		DeptID:         tokenData.DeptId,
		RoleIDs:        tokenData.RoleIds,
		LoginName:      tokenData.LoginName,
		DataScope:      tokenData.DataScope,
		DataScopeDepts: tokenData.DataScopeDepts,
		User:           user,
		Menus:          menus,
		ExpiredAt:      tokenData.Exp,
		ApiPermissions: apiPermissions,
		ApiPathMap:     apiPathMap,
	}
}

// ToTokenData 转换为TokenData
func (u *UserInfo) ToTokenData() *v1.TokenData {
	return &v1.TokenData{
		Exp:            u.ExpiredAt,
		UserId:         u.UserID,
		DeptId:         u.DeptID,
		RoleIds:        u.RoleIDs,
		LoginName:      u.LoginName,
		DataScope:      u.DataScope,
		DataScopeDepts: u.DataScopeDepts,
		ApiPermissions: u.ApiPermissions,
	}
}
