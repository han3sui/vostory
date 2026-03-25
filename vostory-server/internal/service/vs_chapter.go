package service

import (
	"context"
	"fmt"
	"unicode/utf8"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsChapterService interface {
	Create(ctx context.Context, request *v1.VsChapterCreateRequest) error
	Update(ctx context.Context, request *v1.VsChapterUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsChapterDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsChapterListQuery) ([]*v1.VsChapterDetailResponse, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*v1.VsChapterDetailResponse, error)
}

func NewVsChapterService(
	service *Service,
	repo repository.VsChapterRepository,
) VsChapterService {
	return &vsChapterService{Service: service, repo: repo}
}

type vsChapterService struct {
	*Service
	repo repository.VsChapterRepository
}

func (s *vsChapterService) Create(ctx context.Context, request *v1.VsChapterCreateRequest) error {
	chapter := &model.VsChapter{
		ProjectID:  request.ProjectID,
		Title:      request.Title,
		ChapterNum: request.ChapterNum,
		Content:    request.Content,
		WordCount:  utf8.RuneCountInString(request.Content),
		Status:     "raw",
		Remark:     request.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, chapter)
}

func (s *vsChapterService) Update(ctx context.Context, request *v1.VsChapterUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("章节不存在")
	}

	existing.Title = request.Title
	if request.ChapterNum > 0 {
		existing.ChapterNum = request.ChapterNum
	}
	if request.Content != "" {
		existing.Content = request.Content
		existing.WordCount = utf8.RuneCountInString(request.Content)
	}
	if request.Status != "" {
		existing.Status = request.Status
	}
	existing.Remark = request.Remark
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsChapterService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsChapterService) FindByID(ctx context.Context, id uint64) (*v1.VsChapterDetailResponse, error) {
	chapter, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(chapter), nil
}

func (s *vsChapterService) FindWithPagination(ctx context.Context, query *v1.VsChapterListQuery) ([]*v1.VsChapterDetailResponse, int64, error) {
	chapters, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsChapterDetailResponse
	for _, c := range chapters {
		responses = append(responses, s.convertToDetailResponse(c))
	}
	return responses, total, nil
}

func (s *vsChapterService) FindByProjectID(ctx context.Context, projectID uint64) ([]*v1.VsChapterDetailResponse, error) {
	chapters, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsChapterDetailResponse
	for _, c := range chapters {
		responses = append(responses, s.convertToDetailResponse(c))
	}
	return responses, nil
}

func (s *vsChapterService) convertToDetailResponse(c *model.VsChapter) *v1.VsChapterDetailResponse {
	return &v1.VsChapterDetailResponse{
		ID:         c.ChapterID,
		ProjectID:  c.ProjectID,
		Title:      c.Title,
		ChapterNum: c.ChapterNum,
		Content:    c.Content,
		WordCount:  c.WordCount,
		Status:     c.Status,
		Remark:     c.Remark,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}
