package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsWorkspaceService interface {
	Create(ctx context.Context, request *v1.VsWorkspaceCreateRequest) error
	Update(ctx context.Context, request *v1.VsWorkspaceUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsWorkspaceDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsWorkspaceListQuery) ([]*v1.VsWorkspaceDetailResponse, int64, error)
	FindAllEnabled(ctx context.Context) ([]*v1.VsWorkspaceOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsWorkspaceService(
	service *Service,
	repo repository.VsWorkspaceRepository,
) VsWorkspaceService {
	return &vsWorkspaceService{
		Service: service,
		repo:    repo,
	}
}

type vsWorkspaceService struct {
	*Service
	repo repository.VsWorkspaceRepository
}

func (s *vsWorkspaceService) Create(ctx context.Context, request *v1.VsWorkspaceCreateRequest) error {
	workspace := &model.VsWorkspace{
		Name:        request.Name,
		Description: request.Description,
		OwnerID:     ctx.Value("user_id").(uint),
		Status:      request.Status,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, workspace)
}

func (s *vsWorkspaceService) Update(ctx context.Context, request *v1.VsWorkspaceUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("工作空间不存在")
	}

	existing.Name = request.Name
	existing.Description = request.Description
	existing.Status = request.Status
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsWorkspaceService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsWorkspaceService) FindByID(ctx context.Context, id uint64) (*v1.VsWorkspaceDetailResponse, error) {
	workspace, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(workspace), nil
}

func (s *vsWorkspaceService) FindWithPagination(ctx context.Context, query *v1.VsWorkspaceListQuery) ([]*v1.VsWorkspaceDetailResponse, int64, error) {
	workspaces, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsWorkspaceDetailResponse
	for _, w := range workspaces {
		responses = append(responses, s.convertToDetailResponse(w))
	}
	return responses, total, nil
}

func (s *vsWorkspaceService) FindAllEnabled(ctx context.Context) ([]*v1.VsWorkspaceOptionResponse, error) {
	workspaces, err := s.repo.FindAllEnabled(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsWorkspaceOptionResponse
	for _, w := range workspaces {
		responses = append(responses, &v1.VsWorkspaceOptionResponse{
			ID:   w.WorkspaceID,
			Name: w.Name,
		})
	}
	return responses, nil
}

func (s *vsWorkspaceService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsWorkspaceService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

func (s *vsWorkspaceService) convertToDetailResponse(w *model.VsWorkspace) *v1.VsWorkspaceDetailResponse {
	resp := &v1.VsWorkspaceDetailResponse{
		ID:          w.WorkspaceID,
		Name:        w.Name,
		Description: w.Description,
		OwnerID:     w.OwnerID,
		Status:      w.Status,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
	if w.Owner != nil {
		resp.OwnerName = w.Owner.UserName
	}
	return resp
}
