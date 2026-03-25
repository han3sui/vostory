package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsProjectService interface {
	Create(ctx context.Context, request *v1.VsProjectCreateRequest) error
	Update(ctx context.Context, request *v1.VsProjectUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsProjectDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsProjectListQuery) ([]*v1.VsProjectDetailResponse, int64, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*v1.VsProjectOptionResponse, error)
}

func NewVsProjectService(
	service *Service,
	repo repository.VsProjectRepository,
) VsProjectService {
	return &vsProjectService{
		Service: service,
		repo:    repo,
	}
}

type vsProjectService struct {
	*Service
	repo repository.VsProjectRepository
}

func (s *vsProjectService) Create(ctx context.Context, request *v1.VsProjectCreateRequest) error {
	id, err := s.sid.GenUint64()
	if err != nil {
		return fmt.Errorf("生成ID失败: %w", err)
	}

	project := &model.VsProject{
		ProjectID:         id,
		WorkspaceID:       request.WorkspaceID,
		Name:              request.Name,
		Description:       request.Description,
		CoverURL:          request.CoverURL,
		Status:            "draft",
		LLMProviderID:     request.LLMProviderID,
		TTSProviderID:     request.TTSProviderID,
		PromptTemplateIDs: model.PromptTemplateIDMap(request.PromptTemplateIDs),
		Remark:            request.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, project)
}

func (s *vsProjectService) Update(ctx context.Context, request *v1.VsProjectUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("项目不存在")
	}

	existing.Name = request.Name
	existing.Description = request.Description
	existing.CoverURL = request.CoverURL
	existing.LLMProviderID = request.LLMProviderID
	existing.TTSProviderID = request.TTSProviderID
	existing.PromptTemplateIDs = model.PromptTemplateIDMap(request.PromptTemplateIDs)
	existing.Remark = request.Remark
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsProjectService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsProjectService) FindByID(ctx context.Context, id uint64) (*v1.VsProjectDetailResponse, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(project), nil
}

func (s *vsProjectService) FindWithPagination(ctx context.Context, query *v1.VsProjectListQuery) ([]*v1.VsProjectDetailResponse, int64, error) {
	projects, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsProjectDetailResponse
	for _, p := range projects {
		responses = append(responses, s.convertToDetailResponse(p))
	}
	return responses, total, nil
}

func (s *vsProjectService) FindByWorkspaceID(ctx context.Context, workspaceID uint64) ([]*v1.VsProjectOptionResponse, error) {
	projects, err := s.repo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsProjectOptionResponse
	for _, p := range projects {
		responses = append(responses, &v1.VsProjectOptionResponse{
			ID:   p.ProjectID,
			Name: p.Name,
		})
	}
	return responses, nil
}

func (s *vsProjectService) convertToDetailResponse(p *model.VsProject) *v1.VsProjectDetailResponse {
	resp := &v1.VsProjectDetailResponse{
		ID:                p.ProjectID,
		WorkspaceID:       p.WorkspaceID,
		Name:              p.Name,
		Description:       p.Description,
		CoverURL:          p.CoverURL,
		SourceType:        p.SourceType,
		SourceFileURL:     p.SourceFileURL,
		Status:            p.Status,
		LLMProviderID:     p.LLMProviderID,
		TTSProviderID:     p.TTSProviderID,
		PromptTemplateIDs: map[string]uint64(p.PromptTemplateIDs),
		TotalChapters:     p.TotalChapters,
		TotalCharacters:   p.TotalCharacters,
		Remark:            p.Remark,
		CreatedBy:         p.CreatedBy,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	if p.Workspace != nil {
		resp.WorkspaceName = p.Workspace.Name
	}
	if p.LLMProvider != nil {
		resp.LLMProviderName = p.LLMProvider.Name
	}
	if p.TTSProvider != nil {
		resp.TTSProviderName = p.TTSProvider.Name
	}
	return resp
}
