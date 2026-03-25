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

	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsFileImportService interface {
	UploadFile(ctx context.Context, projectID uint64, file *multipart.FileHeader) (string, string, error)
	ParseTxt(ctx context.Context, projectID uint64) (int, int, error)
}

func NewVsFileImportService(
	service *Service,
	projectRepo repository.VsProjectRepository,
	chapterRepo repository.VsChapterRepository,
) VsFileImportService {
	return &vsFileImportService{
		Service:     service,
		projectRepo: projectRepo,
		chapterRepo: chapterRepo,
	}
}

type vsFileImportService struct {
	*Service
	projectRepo repository.VsProjectRepository
	chapterRepo repository.VsChapterRepository
}

func (s *vsFileImportService) UploadFile(ctx context.Context, projectID uint64, file *multipart.FileHeader) (string, string, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return "", "", fmt.Errorf("项目不存在")
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

	uploadDir := filepath.Join("storage", "uploads", fmt.Sprintf("%d", projectID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UnixMilli(), ext)
	filePath := filepath.Join(uploadDir, fileName)

	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", "", fmt.Errorf("保存文件失败: %w", err)
	}

	project.SourceType = sourceType
	project.SourceFileURL = filePath
	project.Status = "parsing"
	project.UpdatedBy = ctx.Value("login_name").(string)

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return "", "", fmt.Errorf("更新项目失败: %w", err)
	}

	return sourceType, filePath, nil
}

var chapterPattern = regexp.MustCompile(`^第[零一二三四五六七八九十百千万\d]+[章节回卷集篇]`)

func (s *vsFileImportService) ParseTxt(ctx context.Context, projectID uint64) (int, int, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return 0, 0, fmt.Errorf("项目不存在")
	}

	if project.SourceType != "txt" {
		return 0, 0, fmt.Errorf("仅支持 txt 格式的自动解析，当前格式: %s", project.SourceType)
	}

	f, err := os.Open(project.SourceFileURL)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源文件失败: %w", err)
	}
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

	loginName := ctx.Value("login_name").(string)
	deptID := ctx.Value("dept_id").(uint)
	totalWords := 0

	for i, ch := range chapters {
		content := ch.content.String()
		wordCount := utf8.RuneCountInString(content)
		totalWords += wordCount

		chapter := &model.VsChapter{
			ProjectID:  projectID,
			Title:      ch.title,
			ChapterNum: i + 1,
			Content:    content,
			WordCount:  wordCount,
			Status:     "raw",
			BaseModel: model.BaseModel{
				CreatedBy: loginName,
				DeptID:    deptID,
			},
		}

		if err := s.chapterRepo.Create(ctx, chapter); err != nil {
			return 0, 0, fmt.Errorf("创建章节失败: %w", err)
		}
	}

	project.Status = "parsed"
	project.TotalChapters = len(chapters)
	project.UpdatedBy = loginName
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return len(chapters), totalWords, fmt.Errorf("更新项目失败: %w", err)
	}

	return len(chapters), totalWords, nil
}
