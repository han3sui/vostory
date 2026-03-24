package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/config"
	"iot-alert-center/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	configPath = flag.String("conf", "config/dev.yml", "config path, eg: -conf ./config/dev.yml")
)

// API信息结构
type APIInfo struct {
	Name        string // Summary
	Description string // Description
	Method      string // HTTP方法
	Path        string // 路径
	PathRegex   string // 路径正则表达式，用于匹配
	Perms       string // 权限标识（来自operationId）
	Tag         string // 标签
}

// Swagger文档结构
type SwaggerDoc struct {
	Paths map[string]map[string]SwaggerOperation `json:"paths"`
}

// Swagger操作结构
type SwaggerOperation struct {
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	OperationID string   `json:"operationId"`
	Tags        []string `json:"tags"`
}

func main() {
	flag.Parse()

	// 初始化配置
	conf := config.NewConfig(*configPath)

	// 初始化日志
	logger := log.NewLog(conf)
	defer logger.Sync()

	// 初始化数据库
	db := repository.NewDB(conf, logger)

	// 从swagger.json文件中读取API信息
	swaggerFile := "docs/swagger.json"
	apis, err := parseSwaggerAPIs(swaggerFile, logger)
	if err != nil {
		logger.Fatal("解析swagger.json失败", zap.Error(err))
	}

	// 显示解析结果的详细信息
	logger.Info("解析结果统计")
	validCount := 0
	invalidCount := 0
	for _, api := range apis {
		if api.Perms != "" {
			validCount++
			logger.Debug("解析到有效API",
				zap.String("operationId", api.Perms),
				zap.String("名称", api.Name),
				zap.String("路径", api.Path),
				zap.String("方法", api.Method),
				zap.String("标签", api.Tag))
		} else {
			invalidCount++
			logger.Debug("解析到无operationId的API",
				zap.String("名称", api.Name),
				zap.String("路径", api.Path),
				zap.String("方法", api.Method),
				zap.String("标签", api.Tag))
		}
	}

	logger.Info("解析统计",
		zap.Int("总API数", len(apis)),
		zap.Int("有operationId的API", validCount),
		zap.Int("无operationId的API", invalidCount))

	// 同步API到数据库
	err = syncApisWithDB(context.Background(), db, apis, logger)
	if err != nil {
		logger.Fatal("同步API失败", zap.Error(err))
	}

	logger.Info("API同步完成", zap.Int("总数", len(apis)))
}

// 从swagger.json文件中解析API信息
func parseSwaggerAPIs(swaggerFile string, logger *log.Logger) ([]APIInfo, error) {
	logger.Info("开始解析swagger.json文件", zap.String("文件", swaggerFile))

	// 读取swagger.json文件
	content, err := ioutil.ReadFile(swaggerFile)
	if err != nil {
		return nil, fmt.Errorf("读取swagger.json文件失败: %w", err)
	}

	// 解析JSON
	var swaggerDoc SwaggerDoc
	if err := json.Unmarshal(content, &swaggerDoc); err != nil {
		return nil, fmt.Errorf("解析swagger.json失败: %w", err)
	}

	var apis []APIInfo

	// 遍历所有路径和方法
	for path, methods := range swaggerDoc.Paths {
		for method, operation := range methods {
			api := APIInfo{
				Name:        operation.Summary,
				Description: operation.Description,
				Method:      strings.ToUpper(method),
				Path:        path,
				PathRegex:   generatePathRegex(path),
				Perms:       operation.OperationID,
			}

			// 处理标签
			if len(operation.Tags) > 0 {
				api.Tag = operation.Tags[0] // 取第一个标签
			}

			// 如果没有Summary，使用Description
			if api.Name == "" {
				api.Name = api.Description
			}

			// 如果没有Description，使用Summary
			if api.Description == "" {
				api.Description = api.Name
			}

			// 如果没有operationId，记录警告但仍然保存API
			if api.Perms == "" {
				logger.Warn("API缺少operationId",
					zap.String("路径", api.Path),
					zap.String("方法", api.Method))
			}

			apis = append(apis, api)
		}
	}

	logger.Info("解析完成", zap.Int("API总数", len(apis)))
	return apis, nil
}

// 生成路径正则表达式，用于匹配gin路由
func generatePathRegex(path string) string {
	// 将路径参数 {id} 转换为正则表达式 ([^/]+)
	// 例如: /api/v1/alert-event/{id} -> ^/api/v1/alert-event/([^/]+)$

	// 转义特殊字符
	regexPath := regexp.QuoteMeta(path)

	// 替换路径参数
	paramPattern := regexp.MustCompile(`\\\{[^}]+\\\}`)
	regexPath = paramPattern.ReplaceAllString(regexPath, `([^/]+)`)

	// 添加开始和结束锚点
	// regexPath = "^" + regexPath + "$"

	return regexPath
}

// 同步API到数据库
func syncApisWithDB(ctx context.Context, db *gorm.DB, apis []APIInfo, logger *log.Logger) error {
	logger.Info("开始同步API到数据库")

	// 获取所有现有API
	var existingAPIs []*model.SysApi
	if err := db.WithContext(ctx).Find(&existingAPIs).Error; err != nil {
		return fmt.Errorf("获取API列表失败: %w", err)
	}

	// 构建本地API集合（只包含有operationId的）
	localPermsSet := make(map[string]APIInfo)
	for _, api := range apis {
		if api.Perms != "" {
			localPermsSet[api.Perms] = api
		}
	}

	// 构建远程API集合（只包含有operationId的）
	remotePermsSet := make(map[string]*model.SysApi)
	for _, api := range existingAPIs {
		if api.Perms != "" {
			remotePermsSet[api.Perms] = api
		}
	}

	logger.Info("比对统计",
		zap.Int("本地API数", len(localPermsSet)),
		zap.Int("远程API数", len(remotePermsSet)))

	updateCount := 0
	addCount := 0
	deleteCount := 0

	// 遍历本地API
	for perms, localAPI := range localPermsSet {
		if remoteAPI, exists := remotePermsSet[perms]; exists {
			// 本地有，远程有：更新
			remoteAPI.Name = localAPI.Name
			remoteAPI.Description = localAPI.Description
			remoteAPI.Path = localAPI.Path
			remoteAPI.Method = localAPI.Method
			remoteAPI.Tag = localAPI.Tag

			if err := db.WithContext(ctx).Save(remoteAPI).Error; err != nil {
				logger.Error("更新API失败",
					zap.String("operationId", perms),
					zap.Error(err))
				continue
			}

			updateCount++
			logger.Debug("更新API",
				zap.String("operationId", perms),
				zap.String("名称", localAPI.Name),
				zap.String("标签", localAPI.Tag))
		} else {
			// 本地有，远程没有：添加
			newAPI := &model.SysApi{
				Name:        localAPI.Name,
				Path:        localAPI.Path,
				Method:      localAPI.Method,
				Description: localAPI.Description,
				Perms:       localAPI.Perms,
				Tag:         localAPI.Tag,
			}

			if err := db.WithContext(ctx).Create(newAPI).Error; err != nil {
				logger.Error("添加API失败",
					zap.String("operationId", perms),
					zap.Error(err))
				continue
			}

			addCount++
			logger.Debug("添加API",
				zap.String("operationId", perms),
				zap.String("名称", localAPI.Name),
				zap.String("标签", localAPI.Tag))
		}
	}

	// 遍历远程API，查找本地没有的
	for perms, remoteAPI := range remotePermsSet {
		if _, exists := localPermsSet[perms]; !exists {
			// 本地没有，远程有：删除
			if err := db.WithContext(ctx).Where("perms = ?", perms).Delete(&model.SysApi{}).Error; err != nil {
				logger.Error("删除API失败",
					zap.String("operationId", perms),
					zap.Error(err))
				continue
			}

			deleteCount++
			logger.Debug("删除API",
				zap.String("operationId", perms),
				zap.String("名称", remoteAPI.Name))
		}
	}

	logger.Info("API同步完成",
		zap.Int("更新", updateCount),
		zap.Int("添加", addCount),
		zap.Int("删除", deleteCount))
	return nil
}
