package service

import (
	"context"
	"fmt"
	"strings"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsLLMModelService interface {
	Create(ctx context.Context, request *v1.VsLLMModelCreateRequest) error
	Update(ctx context.Context, request *v1.VsLLMModelUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsLLMModelDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMModelListQuery) ([]*v1.VsLLMModelDetailResponse, int64, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsLLMModelService(
	service *Service,
	modelRepo repository.VsLLMModelRepository,
	providerRepo repository.VsLLMProviderRepository,
	projectRepo repository.VsProjectRepository,
) VsLLMModelService {
	return &vsLLMModelService{
		Service:      service,
		modelRepo:    modelRepo,
		providerRepo: providerRepo,
		projectRepo:  projectRepo,
	}
}

type vsLLMModelService struct {
	*Service
	modelRepo    repository.VsLLMModelRepository
	providerRepo repository.VsLLMProviderRepository
	projectRepo  repository.VsProjectRepository
}

func (s *vsLLMModelService) Create(ctx context.Context, request *v1.VsLLMModelCreateRequest) error {
	if _, err := s.providerRepo.FindByID(ctx, request.ProviderID); err != nil {
		return fmt.Errorf("所属厂家不存在")
	}
	createdBy, _ := ctx.Value("login_name").(string)
	m := &model.VsLLMModel{
		ProviderID: request.ProviderID,
		ModelName:  strings.TrimSpace(request.ModelName),
		SortOrder:  request.SortOrder,
		Status:     request.Status,
		Remark:     request.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: createdBy,
		},
	}
	return s.modelRepo.Create(ctx, m)
}

func (s *vsLLMModelService) Update(ctx context.Context, request *v1.VsLLMModelUpdateRequest) error {
	if _, err := s.providerRepo.FindByID(ctx, request.ProviderID); err != nil {
		return fmt.Errorf("所属厂家不存在")
	}
	updatedBy, _ := ctx.Value("login_name").(string)
	return s.modelRepo.UpdateFields(ctx, request.ID, map[string]interface{}{
		"provider_id": request.ProviderID,
		"model_name":  strings.TrimSpace(request.ModelName),
		"sort_order":  request.SortOrder,
		"status":      request.Status,
		"remark":      request.Remark,
		"updated_by":  updatedBy,
	})
}

func (s *vsLLMModelService) Delete(ctx context.Context, id uint64) error {
	count, err := s.projectRepo.CountByLLMModelID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该模型仍被 %d 个项目引用，请先在项目中改绑其它模型", count)
	}
	return s.modelRepo.Delete(ctx, id)
}

func (s *vsLLMModelService) DeleteByIDs(ctx context.Context, ids []uint64) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *vsLLMModelService) FindByID(ctx context.Context, id uint64) (*v1.VsLLMModelDetailResponse, error) {
	m, err := s.modelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(m, s.providerMap(ctx)), nil
}

func (s *vsLLMModelService) FindWithPagination(ctx context.Context, query *v1.VsLLMModelListQuery) ([]*v1.VsLLMModelDetailResponse, int64, error) {
	models, total, err := s.modelRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	providers := s.providerMap(ctx)
	responses := make([]*v1.VsLLMModelDetailResponse, 0, len(models))
	for _, m := range models {
		responses = append(responses, s.convertToDetailResponse(m, providers))
	}
	return responses, total, nil
}

func (s *vsLLMModelService) Enable(ctx context.Context, id uint64) error {
	return s.modelRepo.Enable(ctx, id)
}

func (s *vsLLMModelService) Disable(ctx context.Context, id uint64) error {
	return s.modelRepo.Disable(ctx, id)
}

func (s *vsLLMModelService) providerMap(ctx context.Context) map[uint64]*model.VsLLMProvider {
	providers, err := s.providerRepo.FindAll(ctx)
	if err != nil {
		return map[uint64]*model.VsLLMProvider{}
	}
	m := make(map[uint64]*model.VsLLMProvider, len(providers))
	for _, p := range providers {
		m[p.ProviderID] = p
	}
	return m
}

func (s *vsLLMModelService) convertToDetailResponse(m *model.VsLLMModel, providers map[uint64]*model.VsLLMProvider) *v1.VsLLMModelDetailResponse {
	resp := &v1.VsLLMModelDetailResponse{
		ID:         m.ModelID,
		ProviderID: m.ProviderID,
		ModelName:  m.ModelName,
		SortOrder:  m.SortOrder,
		Status:     m.Status,
		Remark:     m.Remark,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if p, ok := providers[m.ProviderID]; ok {
		resp.ProviderName = p.Name
	}
	return resp
}
