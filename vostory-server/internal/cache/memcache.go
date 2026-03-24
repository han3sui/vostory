package cache

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	v1 "iot-alert-center/api/v1"
// 	"iot-alert-center/internal/model"

// 	"github.com/allegro/bigcache/v3"
// )

// // TokenCache 定义令牌缓存接口
// type BigCache interface {
// 	// StoreUserToken 存储用户令牌和用户信息
// 	StoreUserToken(token string, userInfo *UserInfo) error

// 	// GetUserInfo 根据令牌获取用户信息
// 	GetUserInfo(token string) (*UserInfo, error)

// 	// RemoveToken 移除令牌
// 	RemoveToken(token string) error

// 	// Close 关闭缓存
// 	Close() error
// }

// // UserInfo 用户信息结构
// type UserInfo struct {
// 	UserID         uint                     `json:"user_id"`
// 	DeptID         uint                     `json:"dept_id"`
// 	RoleIDs        []uint                   `json:"role_ids"`
// 	LoginName      string                   `json:"login_name"`
// 	DataScope      string                   `json:"data_scope"`
// 	DataScopeDepts []uint                   `json:"data_scope_depts"`
// 	User           *model.SysUser           `json:"user,omitempty"`
// 	Menus          []*model.SysMenu         `json:"menus,omitempty"`
// 	ExpiredAt      int64                    `json:"expired_at"`
// 	ApiPermissions []string                 `json:"api_permissions,omitempty"`
// 	ApiPathMap     map[string][]ApiPathInfo `json:"api_path_map,omitempty"` // 新增：API路径映射，key为HTTP方法，value为该方法下的API路径信息
// }

// // ApiPathInfo API路径信息
// type ApiPathInfo struct {
// 	Path  string `json:"path"`  // API路径模式
// 	Perms string `json:"perms"` // 权限标识
// }

// // bigCache 使用bigcache实现的令牌缓存
// type bigCache struct {
// 	cache *bigcache.BigCache
// }

// // NewBigCache 创建基于BigCache的令牌缓存
// func NewBigCache() (BigCache, error) {
// 	// 默认配置，1小时过期
// 	config := bigcache.DefaultConfig(time.Hour)
// 	config.CleanWindow = 5 * time.Minute // 设置清理窗口
// 	config.MaxEntriesInWindow = 10000    // 设置最大条目数

// 	cache, err := bigcache.New(context.Background(), config)
// 	if err != nil {
// 		return nil, fmt.Errorf("创建BigCache失败: %w", err)
// 	}

// 	return &bigCache{
// 		cache: cache,
// 	}, nil
// }

// // StoreUserToken 存储用户令牌和用户信息
// func (c *bigCache) StoreUserToken(token string, userInfo *UserInfo) error {
// 	// 序列化用户信息
// 	data, err := json.Marshal(userInfo)
// 	if err != nil {
// 		return fmt.Errorf("序列化用户信息失败: %w", err)
// 	}

// 	// 存入缓存
// 	return c.cache.Set(token, data)
// }

// // GetUserInfo 根据令牌获取用户信息
// func (c *bigCache) GetUserInfo(token string) (*UserInfo, error) {
// 	// 从缓存获取
// 	data, err := c.cache.Get(token)
// 	if err != nil {
// 		if err == bigcache.ErrEntryNotFound {
// 			return nil, fmt.Errorf("令牌不存在或已过期")
// 		}
// 		return nil, fmt.Errorf("获取缓存数据失败: %w", err)
// 	}

// 	// 反序列化
// 	var userInfo UserInfo
// 	if err := json.Unmarshal(data, &userInfo); err != nil {
// 		return nil, fmt.Errorf("反序列化用户信息失败: %w", err)
// 	}

// 	// 检查是否过期
// 	if userInfo.ExpiredAt < time.Now().Unix() {
// 		c.RemoveToken(token) // 移除过期令牌
// 		return nil, fmt.Errorf("令牌已过期")
// 	}

// 	return &userInfo, nil
// }

// // RemoveToken 移除令牌
// func (c *bigCache) RemoveToken(token string) error {
// 	return c.cache.Delete(token)
// }

// // Close 关闭缓存
// func (c *bigCache) Close() error {
// 	return c.cache.Close()
// }

// // FromTokenData 从TokenData创建UserInfo
// func FromTokenData(tokenData *v1.TokenData, user *model.SysUser, menus []*model.SysMenu) *UserInfo {
// 	apiPermissions := make([]string, 0)
// 	for _, menu := range menus {
// 		if menu.MenuType == "F" {
// 			apiPermissions = append(apiPermissions, menu.Perms)
// 		}
// 	}
// 	return &UserInfo{
// 		UserID:         tokenData.UserId,
// 		DeptID:         tokenData.DeptId,
// 		RoleIDs:        tokenData.RoleIds,
// 		LoginName:      tokenData.LoginName,
// 		DataScope:      tokenData.DataScope,
// 		DataScopeDepts: tokenData.DataScopeDepts,
// 		User:           user,
// 		Menus:          menus,
// 		ExpiredAt:      tokenData.Exp,
// 		ApiPermissions: apiPermissions,
// 	}
// }

// // FromTokenDataWithApiInfo 从TokenData创建UserInfo，传入API权限和路径映射
// func FromTokenDataWithApiInfo(tokenData *v1.TokenData, user *model.SysUser, menus []*model.SysMenu, apiPermissions []string, apiPathMap map[string][]ApiPathInfo) *UserInfo {
// 	return &UserInfo{
// 		UserID:         tokenData.UserId,
// 		DeptID:         tokenData.DeptId,
// 		RoleIDs:        tokenData.RoleIds,
// 		LoginName:      tokenData.LoginName,
// 		DataScope:      tokenData.DataScope,
// 		DataScopeDepts: tokenData.DataScopeDepts,
// 		User:           user,
// 		Menus:          menus,
// 		ExpiredAt:      tokenData.Exp,
// 		ApiPermissions: apiPermissions,
// 		ApiPathMap:     apiPathMap,
// 	}
// }

// // ToTokenData 转换为TokenData
// func (u *UserInfo) ToTokenData() *v1.TokenData {
// 	return &v1.TokenData{
// 		Exp:            u.ExpiredAt,
// 		UserId:         u.UserID,
// 		DeptId:         u.DeptID,
// 		RoleIds:        u.RoleIDs,
// 		LoginName:      u.LoginName,
// 		DataScope:      u.DataScope,
// 		DataScopeDepts: u.DataScopeDepts,
// 		ApiPermissions: u.ApiPermissions,
// 	}
// }
