package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"iot-alert-center/internal/filestorage"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsFileImportService interface {
	UploadFile(ctx context.Context, projectID uint64, file *multipart.FileHeader) (string, string, error)
	// RecoverParsing 服务启动时重新解析上次中断、停在「解析中」的项目
	RecoverParsing(ctx context.Context)
}

func NewVsFileImportService(
	service *Service,
	projectRepo repository.VsProjectRepository,
	chapterRepo repository.VsChapterRepository,
	fileStorage filestorage.FileStorage,
) VsFileImportService {
	return &vsFileImportService{
		Service:     service,
		projectRepo: projectRepo,
		chapterRepo: chapterRepo,
		fileStorage: fileStorage,
	}
}

type vsFileImportService struct {
	*Service
	projectRepo repository.VsProjectRepository
	chapterRepo repository.VsChapterRepository
	fileStorage filestorage.FileStorage
}

func (s *vsFileImportService) UploadFile(ctx context.Context, projectID uint64, file *multipart.FileHeader) (string, string, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return "", "", fmt.Errorf("项目不存在")
	}
	if project.Status == projectStatusParsing && isAutoParsedSource(project.SourceType) {
		return "", "", fmt.Errorf("项目正在解析上一次上传的文件，请等待完成后再上传")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var sourceType string
	switch ext {
	case ".txt":
		sourceType = "txt"
	case ".docx":
		sourceType = "docx"
	case ".epub":
		sourceType = "epub"
	default:
		return "", "", fmt.Errorf("不支持的文件格式: %s，仅支持 txt/docx/epub", ext)
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UnixMilli(), ext)
	filePath := fmt.Sprintf("uploads/%d/%s", projectID, fileName)

	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	if err := s.fileStorage.Put(ctx, filePath, src, file.Size, "application/octet-stream"); err != nil {
		return "", "", fmt.Errorf("保存文件失败: %w", err)
	}

	loginName := ctx.Value("login_name").(string)
	deptID := ctx.Value("dept_id").(uint)

	if err := s.projectRepo.UpdateFields(ctx, project.ProjectID, map[string]interface{}{
		"source_type":     sourceType,
		"source_file_url": filePath,
		"status":          projectStatusParsing,
		"updated_by":      loginName,
	}); err != nil {
		return "", "", fmt.Errorf("更新项目失败: %w", err)
	}

	if isAutoParsedSource(sourceType) {
		go s.parseInBackground(projectID, loginName, deptID)
	}

	return sourceType, filePath, nil
}

const (
	projectStatusParsing     = "parsing"
	projectStatusParseFailed = "parse_failed"
)

// isAutoParsedSource 上传后会在后台自动解析章节的文件类型
func isAutoParsedSource(sourceType string) bool {
	return sourceType == "txt" || sourceType == "epub"
}

// parseInBackground 后台解析源文件。解析代码 panic（例如格式异常的 epub）时只把项目标为解析失败，不会带崩整个服务
func (s *vsFileImportService) parseInBackground(projectID uint64, loginName string, deptID uint) {
	ctx := context.WithValue(context.Background(), "login_name", loginName)
	ctx = context.WithValue(ctx, "dept_id", deptID)
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(fmt.Sprintf("解析项目 %d 源文件时发生 panic: %v", projectID, r))
			_ = s.projectRepo.UpdateFields(ctx, projectID, map[string]interface{}{
				"status":     projectStatusParseFailed,
				"updated_by": loginName,
			})
		}
	}()
	s.parseSourceFile(ctx, projectID)
}

func (s *vsFileImportService) RecoverParsing(ctx context.Context) {
	projects, err := s.projectRepo.FindByStatus(ctx, projectStatusParsing)
	if err != nil {
		s.logger.Error(fmt.Sprintf("查询解析中的项目失败: %v", err))
		return
	}
	for _, p := range projects {
		if ctx.Err() != nil {
			return
		}
		if !isAutoParsedSource(p.SourceType) {
			continue
		}
		operator := p.UpdatedBy
		if operator == "" {
			operator = p.CreatedBy
		}
		// 章节是一次性事务写入的，中断时不会留下部分章节，可以直接重新解析
		s.logger.Info(fmt.Sprintf("恢复解析项目 %d（%s）", p.ProjectID, p.Name))
		s.parseInBackground(p.ProjectID, operator, p.DeptID)
	}
}

var chapterPattern = regexp.MustCompile(`^第[零一二三四五六七八九十百千万\d]+[章节回卷集篇]`)

func (s *vsFileImportService) parseSourceFile(ctx context.Context, projectID uint64) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return
	}

	var parseErr error
	switch project.SourceType {
	case "txt":
		_, _, parseErr = s.parseTxtFile(ctx, project)
	case "epub":
		_, _, parseErr = s.parseEpubFile(ctx, project)
	default:
		parseErr = fmt.Errorf("不支持 %s 格式的自动解析", project.SourceType)
	}

	if parseErr != nil {
		s.logger.Warn(fmt.Sprintf("解析项目 %d 源文件失败: %v", project.ProjectID, parseErr))
		_ = s.projectRepo.UpdateFields(ctx, project.ProjectID, map[string]interface{}{
			"status":     projectStatusParseFailed,
			"updated_by": ctx.Value("login_name").(string),
		})
	}
}

func (s *vsFileImportService) parseTxtFile(ctx context.Context, project *model.VsProject) (int, int, error) {
	f, cleanup, err := s.openSource(ctx, project.SourceFileURL)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源文件失败: %w", err)
	}
	defer cleanup()
	defer f.Close()

	type chapterData struct {
		title   string
		content strings.Builder
	}

	var chapters []chapterData
	var current *chapterData
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if current != nil {
				current.content.WriteString("\n")
			}
			continue
		}

		if chapterPattern.MatchString(line) {
			chapters = append(chapters, chapterData{title: line})
			current = &chapters[len(chapters)-1]
		} else if current != nil {
			if current.content.Len() > 0 {
				current.content.WriteString("\n")
			}
			current.content.WriteString(line)
		} else {
			chapters = append(chapters, chapterData{title: "序章"})
			current = &chapters[len(chapters)-1]
			current.content.WriteString(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("读取文件失败: %w", err)
	}

	if len(chapters) == 0 {
		return 0, 0, fmt.Errorf("未识别到任何章节")
	}

	parsed := make([]parsedChapter, 0, len(chapters))
	for _, ch := range chapters {
		parsed = append(parsed, parsedChapter{Title: ch.title, Content: ch.content.String()})
	}
	return s.saveChapters(ctx, project, parsed)
}

func (s *vsFileImportService) parseEpubFile(ctx context.Context, project *model.VsProject) (int, int, error) {
	path, cleanup, err := s.sourceLocalPath(ctx, project.SourceFileURL)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源文件失败: %w", err)
	}
	defer cleanup()
	epubChapters, err := ParseEpubFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("解析 epub 失败: %w", err)
	}

	parsed := make([]parsedChapter, 0, len(epubChapters))
	for _, ch := range epubChapters {
		parsed = append(parsed, parsedChapter{Title: ch.Title, Content: ch.Content})
	}
	return s.saveChapters(ctx, project, parsed)
}

type parsedChapter struct {
	Title   string
	Content string
}

func (s *vsFileImportService) saveChapters(ctx context.Context, project *model.VsProject, chapters []parsedChapter) (int, int, error) {
	if len(chapters) == 0 {
		return 0, 0, fmt.Errorf("未识别到任何章节")
	}

	loginName, _ := ctx.Value("login_name").(string)
	deptID, _ := ctx.Value("dept_id").(uint)
	totalWords := 0

	records := make([]*model.VsChapter, 0, len(chapters))
	for i, ch := range chapters {
		wordCount := utf8.RuneCountInString(ch.Content)
		totalWords += wordCount

		chapter := &model.VsChapter{
			ProjectID:  project.ProjectID,
			Title:      ch.Title,
			ChapterNum: i + 1,
			Content:    ch.Content,
			WordCount:  wordCount,
			Status:     "raw",
			BaseModel: model.BaseModel{
				CreatedBy: loginName,
				DeptID:    deptID,
			},
		}

		records = append(records, chapter)
	}
	if err := s.chapterRepo.CreateBatch(ctx, records); err != nil {
		return 0, 0, fmt.Errorf("创建章节失败: %w", err)
	}

	if err := s.projectRepo.UpdateFields(ctx, project.ProjectID, map[string]interface{}{
		"status":         "parsed",
		"total_chapters": len(chapters),
		"updated_by":     loginName,
	}); err != nil {
		return len(chapters), totalWords, fmt.Errorf("更新项目失败: %w", err)
	}

	return len(chapters), totalWords, nil
}

// openSource 优先按存储 Key 打开，兼容历史本地绝对路径
func (s *vsFileImportService) openSource(ctx context.Context, stored string) (io.ReadCloser, func(), error) {
	if r, err := s.fileStorage.Open(ctx, filestorage.NormalizeKey(stored)); err == nil {
		return r, func() {}, nil
	}
	f, err := os.Open(stored)
	if err != nil {
		return nil, nil, err
	}
	return f, func() {}, nil
}

// sourceLocalPath 给依赖本地路径的解析器（如 epub）用
func (s *vsFileImportService) sourceLocalPath(ctx context.Context, stored string) (string, func(), error) {
	if path, cleanup, err := s.fileStorage.DownloadToTemp(ctx, filestorage.NormalizeKey(stored)); err == nil {
		return path, cleanup, nil
	}
	if _, err := os.Stat(stored); err == nil {
		return stored, func() {}, nil
	}
	return "", nil, fmt.Errorf("源文件不存在")
}
