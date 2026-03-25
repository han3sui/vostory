package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsLLMProviderService interface {
	Create(ctx context.Context, request *v1.VsLLMProviderCreateRequest) error
	Update(ctx context.Context, request *v1.VsLLMProviderUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsLLMProviderDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMProviderListQuery) ([]*v1.VsLLMProviderDetailResponse, int64, error)
	FindAllEnabled(ctx context.Context) ([]*v1.VsLLMProviderOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
	TestConnection(ctx context.Context, request *v1.VsLLMProviderTestRequest) *v1.VsLLMProviderTestResponse
}

func NewVsLLMProviderService(
	service *Service,
	repo repository.VsLLMProviderRepository,
) VsLLMProviderService {
	return &vsLLMProviderService{
		Service: service,
		repo:    repo,
	}
}

type vsLLMProviderService struct {
	*Service
	repo repository.VsLLMProviderRepository
}

func (s *vsLLMProviderService) Create(ctx context.Context, request *v1.VsLLMProviderCreateRequest) error {
	id, err := s.sid.GenUint64()
	if err != nil {
		return fmt.Errorf("生成ID失败: %w", err)
	}

	provider := &model.VsLLMProvider{
		ProviderID:   id,
		Name:         request.Name,
		ProviderType: request.ProviderType,
		APIBaseURL:   request.APIBaseURL,
		APIKey:       request.APIKey,
		ModelList:    request.ModelList,
		DefaultModel: request.DefaultModel,
		CustomParams: request.CustomParams,
		SortOrder:    request.SortOrder,
		Status:       request.Status,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.repo.Create(ctx, provider)
}

func (s *vsLLMProviderService) Update(ctx context.Context, request *v1.VsLLMProviderUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("提供商不存在")
	}

	existing.Name = request.Name
	existing.ProviderType = request.ProviderType
	existing.APIBaseURL = request.APIBaseURL
	existing.APIKey = request.APIKey
	existing.ModelList = request.ModelList
	existing.DefaultModel = request.DefaultModel
	existing.CustomParams = request.CustomParams
	existing.SortOrder = request.SortOrder
	existing.Status = request.Status
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsLLMProviderService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsLLMProviderService) FindByID(ctx context.Context, id uint64) (*v1.VsLLMProviderDetailResponse, error) {
	provider, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(provider), nil
}

func (s *vsLLMProviderService) FindWithPagination(ctx context.Context, query *v1.VsLLMProviderListQuery) ([]*v1.VsLLMProviderDetailResponse, int64, error) {
	providers, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsLLMProviderDetailResponse
	for _, p := range providers {
		responses = append(responses, s.convertToDetailResponse(p))
	}
	return responses, total, nil
}

func (s *vsLLMProviderService) FindAllEnabled(ctx context.Context) ([]*v1.VsLLMProviderOptionResponse, error) {
	providers, err := s.repo.FindAllEnabled(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsLLMProviderOptionResponse
	for _, p := range providers {
		responses = append(responses, &v1.VsLLMProviderOptionResponse{
			ID:           p.ProviderID,
			Name:         p.Name,
			ProviderType: p.ProviderType,
			ModelList:    p.ModelList,
			DefaultModel: p.DefaultModel,
		})
	}
	return responses, nil
}

func (s *vsLLMProviderService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsLLMProviderService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

func (s *vsLLMProviderService) TestConnection(_ context.Context, request *v1.VsLLMProviderTestRequest) *v1.VsLLMProviderTestResponse {
	start := time.Now()

	url := strings.TrimRight(request.APIBaseURL, "/") + "/v1/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &v1.VsLLMProviderTestResponse{
			Success:  false,
			Message:  fmt.Sprintf("构建请求失败: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if request.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+request.APIKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &v1.VsLLMProviderTestResponse{
			Success:  false,
			Message:  fmt.Sprintf("连接失败: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode == http.StatusOK {
		return &v1.VsLLMProviderTestResponse{
			Success:  true,
			Message:  "连接成功",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return &v1.VsLLMProviderTestResponse{
		Success:  false,
		Message:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		Duration: time.Since(start).Milliseconds(),
	}
}

func (s *vsLLMProviderService) convertToDetailResponse(p *model.VsLLMProvider) *v1.VsLLMProviderDetailResponse {
	return &v1.VsLLMProviderDetailResponse{
		ID:           p.ProviderID,
		Name:         p.Name,
		ProviderType: p.ProviderType,
		APIBaseURL:   p.APIBaseURL,
		APIKey:       p.APIKey,
		ModelList:    p.ModelList,
		DefaultModel: p.DefaultModel,
		CustomParams: p.CustomParams,
		SortOrder:    p.SortOrder,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}
