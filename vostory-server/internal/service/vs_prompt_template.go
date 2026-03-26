package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsPromptTemplateService interface {
	Create(ctx context.Context, request *v1.VsPromptTemplateCreateRequest) error
	Update(ctx context.Context, request *v1.VsPromptTemplateUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsPromptTemplateDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsPromptTemplateListQuery) ([]*v1.VsPromptTemplateDetailResponse, int64, error)
	FindByType(ctx context.Context, templateType string) ([]*v1.VsPromptTemplateOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
	SeedDefaults(ctx context.Context) error
}

func NewVsPromptTemplateService(
	service *Service,
	repo repository.VsPromptTemplateRepository,
) VsPromptTemplateService {
	return &vsPromptTemplateService{
		Service: service,
		repo:    repo,
	}
}

type vsPromptTemplateService struct {
	*Service
	repo repository.VsPromptTemplateRepository
}

func (s *vsPromptTemplateService) Create(ctx context.Context, request *v1.VsPromptTemplateCreateRequest) error {
	template := &model.VsPromptTemplate{
		Name:         request.Name,
		TemplateType: request.TemplateType,
		Content:      request.Content,
		Description:  request.Description,
		IsSystem:     "0",
		Version:      1,
		SortOrder:    request.SortOrder,
		Status:       request.Status,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.repo.Create(ctx, template)
}

func (s *vsPromptTemplateService) Update(ctx context.Context, request *v1.VsPromptTemplateUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("模板不存在")
	}

	existing.Name = request.Name
	existing.TemplateType = request.TemplateType
	existing.Content = request.Content
	existing.Description = request.Description
	existing.SortOrder = request.SortOrder
	existing.Status = request.Status
	existing.Version = existing.Version + 1
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsPromptTemplateService) Delete(ctx context.Context, id uint64) error {
	template, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("模板不存在")
	}
	if template.IsSystem == "1" {
		return fmt.Errorf("系统内置模板不允许删除")
	}
	return s.repo.Delete(ctx, id)
}

func (s *vsPromptTemplateService) FindByID(ctx context.Context, id uint64) (*v1.VsPromptTemplateDetailResponse, error) {
	template, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(template), nil
}

func (s *vsPromptTemplateService) FindWithPagination(ctx context.Context, query *v1.VsPromptTemplateListQuery) ([]*v1.VsPromptTemplateDetailResponse, int64, error) {
	templates, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsPromptTemplateDetailResponse
	for _, t := range templates {
		responses = append(responses, s.convertToDetailResponse(t))
	}
	return responses, total, nil
}

func (s *vsPromptTemplateService) FindByType(ctx context.Context, templateType string) ([]*v1.VsPromptTemplateOptionResponse, error) {
	templates, err := s.repo.FindByType(ctx, templateType)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsPromptTemplateOptionResponse
	for _, t := range templates {
		responses = append(responses, &v1.VsPromptTemplateOptionResponse{
			ID:           t.TemplateID,
			Name:         t.Name,
			TemplateType: t.TemplateType,
		})
	}
	return responses, nil
}

func (s *vsPromptTemplateService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsPromptTemplateService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

func (s *vsPromptTemplateService) SeedDefaults(ctx context.Context) error {
	for _, d := range model.DefaultPromptTemplateSeeds {
		count, err := s.repo.CountByType(ctx, d.TemplateType)
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		template := &model.VsPromptTemplate{
			Name:         d.Name,
			TemplateType: d.TemplateType,
			Content:      d.Content,
			Description:  d.Description,
			IsSystem:     "1",
			Version:      1,
			SortOrder:    0,
			Status:       "0",
			BaseModel: model.BaseModel{
				CreatedBy: "system",
			},
		}

		if err := s.repo.Create(ctx, template); err != nil {
			return fmt.Errorf("创建默认模板[%s]失败: %w", d.Name, err)
		}
	}

	return nil
}

func (s *vsPromptTemplateService) convertToDetailResponse(t *model.VsPromptTemplate) *v1.VsPromptTemplateDetailResponse {
	return &v1.VsPromptTemplateDetailResponse{
		ID:           t.TemplateID,
		Name:         t.Name,
		TemplateType: t.TemplateType,
		Content:      t.Content,
		Description:  t.Description,
		IsSystem:     t.IsSystem,
		Version:      t.Version,
		SortOrder:    t.SortOrder,
		Status:       t.Status,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
