package middleware

import (
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/cache"
	"iot-alert-center/pkg/jwt"
	"iot-alert-center/pkg/log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func StrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			logger.WithContext(ctx).Warn("No token", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logger.WithContext(ctx).Error("token error", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}), zap.Error(err))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

func NoStrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			tokenString, _ = ctx.Cookie("accessToken")
		}
		if tokenString == "" {
			tokenString = ctx.Query("accessToken")
		}
		if tokenString == "" {
			ctx.Next()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

func recoveryLoggerFunc(ctx *gin.Context, logger *log.Logger) {
	if userInfo, ok := ctx.MustGet("claims").(*jwt.MyCustomClaims); ok {
		logger.WithValue(ctx, zap.String("UserId", userInfo.UserId))
	}
}

// const (
// 	UserIdKey         = "user_id"
// 	DeptIdKey         = "dept_id"
// 	RoleIdsKey        = "role_ids"
// 	LoginNameKey      = "login_name"
// 	DataScopeKey      = "data_scope"
// 	DataScopeDeptsKey = "data_scope_depts"
// )

// 使用缓存的TokenAuth中间件
func TokenCacheAuthMiddleware(conf *viper.Viper, logger *log.Logger, tokenCache cache.UserCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			v1.HandleError(c, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		// 去除Bearer前缀
		token = strings.Replace(token, "Bearer ", "", 1)

		// 从缓存中获取用户信息
		userInfo, err := tokenCache.GetUserInfo(token)
		if err != nil {
			logger.WithContext(c).Warn("Token验证失败", zap.Error(err))
			v1.HandleError(c, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		// 验证过期时间
		if userInfo.ExpiredAt < time.Now().Unix() {
			tokenCache.RemoveToken(token) // 移除过期令牌
			logger.WithContext(c).Warn("令牌已过期")
			v1.HandleError(c, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		// 检查是否需要刷新token（滑动过期机制）
		refreshTokenIfNeeded(c, token, userInfo, tokenCache, logger)

		// 将用户信息保存到上下文中
		c.Set("user_id", userInfo.UserID)
		c.Set("dept_id", userInfo.DeptID)
		c.Set("role_ids", userInfo.RoleIDs)
		c.Set("login_name", userInfo.LoginName)
		c.Set("data_scope", userInfo.DataScope)
		c.Set("data_scope_depts", userInfo.DataScopeDepts)

		// 如果需要，也可以将完整的用户信息设置到上下文
		c.Set("user_info", userInfo)

		c.Next()
	}
}

// API权限验证中间件
func APIPermissionMiddleware(logger *log.Logger, tokenCache cache.UserCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			v1.HandleError(c, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		// 去除Bearer前缀
		token = strings.Replace(token, "Bearer ", "", 1)

		// 从缓存中获取用户信息
		userInfo, err := tokenCache.GetUserInfo(token)
		if err != nil {
			logger.WithContext(c).Warn("Token验证失败", zap.Error(err))
			v1.HandleError(c, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		// 验证过期时间
		if userInfo.ExpiredAt < time.Now().Unix() {
			tokenCache.RemoveToken(token) // 移除过期令牌
			logger.WithContext(c).Warn("令牌已过期")
			v1.HandleError(c, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		// 获取请求路径和方法
		requestPath := c.Request.URL.Path
		requestMethod := strings.ToUpper(c.Request.Method)

		// 验证API权限（完全使用缓存，无数据库查询）
		hasPermission := checkAPIPermissionFromCacheOnly(userInfo, requestPath, requestMethod, logger)
		if !hasPermission {
			logger.WithContext(c).Warn("API权限不足",
				zap.Uint("user_id", userInfo.UserID),
				zap.String("path", requestPath),
				zap.String("method", requestMethod),
				zap.Strings("user_permissions", userInfo.ApiPermissions))
			v1.HandleError(c, http.StatusForbidden, v1.NewError(403, "权限不足"), nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// 完全基于缓存的API权限验证（无数据库查询）
func checkAPIPermissionFromCacheOnly(userInfo *cache.UserInfo, requestPath, requestMethod string, logger *log.Logger) bool {
	// 超级管理员直接通过
	if userInfo.User.UserType == "99" {
		return true
	}

	// 检查是否有通配符权限
	if apiPaths, exists := userInfo.ApiPathMap["*"]; exists {
		for _, pathInfo := range apiPaths {
			if pathInfo.Perms == "*" {
				return true
			}
		}
	}

	// 获取对应HTTP方法的API路径列表
	apiPaths, exists := userInfo.ApiPathMap[requestMethod]
	if !exists {
		return false
	}

	// 遍历API路径，检查是否匹配
	for _, pathInfo := range apiPaths {
		// 将API路径模式转换为正则表达式
		pathRegex, err := convertPathToRegex(pathInfo.Path)
		if err != nil {
			logger.Warn("路径正则转换失败",
				zap.String("api_path", pathInfo.Path),
				zap.String("perms", pathInfo.Perms),
				zap.Error(err))
			continue
		}

		// 使用正则表达式匹配请求路径
		matched, err := regexp.MatchString(pathRegex, requestPath)
		if err != nil {
			logger.Warn("正则匹配失败",
				zap.String("regex", pathRegex),
				zap.String("request_path", requestPath),
				zap.Error(err))
			continue
		}

		if matched {
			logger.Debug("API权限匹配成功",
				zap.String("perms", pathInfo.Perms),
				zap.String("api_path", pathInfo.Path),
				zap.String("request_path", requestPath),
				zap.String("method", requestMethod))
			return true
		}
	}

	return false
}

// 将API路径模式转换为正则表达式
func convertPathToRegex(apiPath string) (string, error) {
	// 转义特殊字符
	regexPath := regexp.QuoteMeta(apiPath)

	// 替换路径参数 {id} -> ([^/]+)
	paramPattern := regexp.MustCompile(`\\\{[^}]+\\\}`)
	regexPath = paramPattern.ReplaceAllString(regexPath, `([^/]+)`)

	// 添加开始和结束锚点
	regexPath = "^" + regexPath + "$"

	return regexPath, nil
}

// refreshTokenIfNeeded 检查并刷新token（滑动过期机制）
// 默认：token有效期24小时，剩余时间不足12小时时自动刷新
func refreshTokenIfNeeded(c *gin.Context, token string, userInfo *cache.UserInfo, tokenCache cache.UserCache, logger *log.Logger) {
	now := time.Now()
	expiredAt := time.Unix(userInfo.ExpiredAt, 0)
	remainingTime := expiredAt.Sub(now)
	refreshThreshold := 12 * time.Hour // 刷新阈值：12小时
	tokenExpiration := 24 * time.Hour  // token有效期：24小时

	if remainingTime < refreshThreshold {
		// 剩余时间不足阈值，刷新token过期时间
		newExpiredAt := now.Add(tokenExpiration).Unix()
		err := tokenCache.RefreshTokenExpiration(token, newExpiredAt)
		if err != nil {
			// 刷新失败只记录日志，不影响本次请求
			logger.WithContext(c).Warn("刷新token过期时间失败",
				zap.Error(err),
				zap.Uint("user_id", userInfo.UserID))
		} else {
			logger.WithContext(c).Debug("Token过期时间已刷新",
				zap.Uint("user_id", userInfo.UserID),
				zap.Time("old_expired_at", expiredAt),
				zap.Time("new_expired_at", time.Unix(newExpiredAt, 0)),
				zap.Duration("remaining_time", remainingTime))

			// 更新userInfo中的过期时间，以便后续使用
			userInfo.ExpiredAt = newExpiredAt
		}
	}
}

/*
使用示例：

在路由中使用API权限验证中间件：

```go
// 在路由组中使用
apiGroup := r.Group("/api/v1")
apiGroup.Use(middleware.APIPermissionMiddleware(db, logger, tokenCache))
{
    apiGroup.GET("/user/:id", userHandler.GetUser)
    apiGroup.POST("/user", userHandler.CreateUser)
    apiGroup.PUT("/user/:id/disabled", userHandler.DisableUser)
}
```

工作流程：
1. 用户请求 GET /api/v1/user/12
2. 中间件从token获取用户的ApiPermissions，如：["sys:user:query", "sys:user:edit"]
3. 查询sys_api表，找到perms="sys:user:query"且method="GET"的记录
4. 假设找到记录：path="/api/v1/user/{id}", method="GET", perms="sys:user:query"
5. 将"/api/v1/user/{id}"转换为正则："^/api/v1/user/([^/]+)$"
6. 用正则匹配请求路径"/api/v1/user/12"，匹配成功，允许访问
7. 如果匹配失败或用户没有对应权限，返回403错误
*/
