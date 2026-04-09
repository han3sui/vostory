package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsTTSProviderService interface {
	Create(ctx context.Context, request *v1.VsTTSProviderCreateRequest) error
	Update(ctx context.Context, request *v1.VsTTSProviderUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsTTSProviderDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsTTSProviderListQuery) ([]*v1.VsTTSProviderDetailResponse, int64, error)
	FindAllEnabled(ctx context.Context) ([]*v1.VsTTSProviderOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
	TestConnection(ctx context.Context, request *v1.VsTTSProviderTestRequest) *v1.VsTTSProviderTestResponse
	GetStatus(ctx context.Context, id uint64) (*v1.VsTTSProviderStatusResponse, error)
}

func NewVsTTSProviderService(
	service *Service,
	repo repository.VsTTSProviderRepository,
) VsTTSProviderService {
	return &vsTTSProviderService{
		Service: service,
		repo:    repo,
	}
}

type vsTTSProviderService struct {
	*Service
	repo repository.VsTTSProviderRepository
}

func (s *vsTTSProviderService) Create(ctx context.Context, request *v1.VsTTSProviderCreateRequest) error {
	provider := &model.VsTTSProvider{
		Name:              request.Name,
		ProviderType:      request.ProviderType,
		APIBaseURL:        request.APIBaseURL,
		APIKey:            request.APIKey,
		SupportedFeatures: request.SupportedFeatures,
		CustomParams:      request.CustomParams,
		MaxConcurrency:    1,
		SortOrder:         request.SortOrder,
		Status:            request.Status,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.repo.Create(ctx, provider)
}

func (s *vsTTSProviderService) Update(ctx context.Context, request *v1.VsTTSProviderUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("提供商不存在")
	}

	existing.Name = request.Name
	existing.ProviderType = request.ProviderType
	existing.APIBaseURL = request.APIBaseURL
	existing.APIKey = request.APIKey
	existing.SupportedFeatures = request.SupportedFeatures
	existing.CustomParams = request.CustomParams
	existing.MaxConcurrency = 1
	existing.SortOrder = request.SortOrder
	existing.Status = request.Status
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsTTSProviderService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsTTSProviderService) FindByID(ctx context.Context, id uint64) (*v1.VsTTSProviderDetailResponse, error) {
	provider, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(provider), nil
}

func (s *vsTTSProviderService) FindWithPagination(ctx context.Context, query *v1.VsTTSProviderListQuery) ([]*v1.VsTTSProviderDetailResponse, int64, error) {
	providers, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsTTSProviderDetailResponse
	for _, p := range providers {
		responses = append(responses, s.convertToDetailResponse(p))
	}
	return responses, total, nil
}

func (s *vsTTSProviderService) FindAllEnabled(ctx context.Context) ([]*v1.VsTTSProviderOptionResponse, error) {
	providers, err := s.repo.FindAllEnabled(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsTTSProviderOptionResponse
	for _, p := range providers {
		responses = append(responses, &v1.VsTTSProviderOptionResponse{
			ID:                p.ProviderID,
			Name:              p.Name,
			ProviderType:      p.ProviderType,
			SupportedFeatures: p.SupportedFeatures,
		})
	}
	return responses, nil
}

func (s *vsTTSProviderService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsTTSProviderService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

func (s *vsTTSProviderService) TestConnection(_ context.Context, request *v1.VsTTSProviderTestRequest) *v1.VsTTSProviderTestResponse {
	start := time.Now()

	url := strings.TrimRight(request.APIBaseURL, "/")
	if !strings.HasSuffix(url, "/health") && !strings.HasSuffix(url, "/ping") {
		url += "/health"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &v1.VsTTSProviderTestResponse{
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
		return &v1.VsTTSProviderTestResponse{
			Success:  false,
			Message:  fmt.Sprintf("连接失败: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 500 {
		return &v1.VsTTSProviderTestResponse{
			Success:  true,
			Message:  "连接成功",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return &v1.VsTTSProviderTestResponse{
		Success:  false,
		Message:  fmt.Sprintf("服务端错误: HTTP %d", resp.StatusCode),
		Duration: time.Since(start).Milliseconds(),
	}
}

func (s *vsTTSProviderService) GetStatus(ctx context.Context, id uint64) (*v1.VsTTSProviderStatusResponse, error) {
	provider, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("提供商不存在")
	}

	url := strings.TrimRight(provider.APIBaseURL, "/") + "/status"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}

	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("远程服务返回错误: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var status v1.VsTTSProviderStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &status, nil
}

func (s *vsTTSProviderService) convertToDetailResponse(p *model.VsTTSProvider) *v1.VsTTSProviderDetailResponse {
	return &v1.VsTTSProviderDetailResponse{
		ID:                p.ProviderID,
		Name:              p.Name,
		ProviderType:      p.ProviderType,
		APIBaseURL:        p.APIBaseURL,
		APIKey:            p.APIKey,
		SupportedFeatures: p.SupportedFeatures,
		CustomParams:      p.CustomParams,
		MaxConcurrency:    p.MaxConcurrency,
		SortOrder:         p.SortOrder,
		Status:            p.Status,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}
