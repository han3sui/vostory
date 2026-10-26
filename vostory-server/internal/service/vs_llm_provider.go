package service

import (
	"bytes"
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

type VsLLMProviderService interface {
	Create(ctx context.Context, request *v1.VsLLMProviderCreateRequest) error
	Update(ctx context.Context, request *v1.VsLLMProviderUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsLLMProviderDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMProviderListQuery) ([]*v1.VsLLMProviderDetailResponse, int64, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
	TestConnection(ctx context.Context, request *v1.VsLLMProviderTestRequest) *v1.VsLLMProviderTestResponse
}

func NewVsLLMProviderService(
	service *Service,
	repo repository.VsLLMProviderRepository,
	modelRepo repository.VsLLMModelRepository,
) VsLLMProviderService {
	return &vsLLMProviderService{
		Service:   service,
		repo:      repo,
		modelRepo: modelRepo,
	}
}

type vsLLMProviderService struct {
	*Service
	repo      repository.VsLLMProviderRepository
	modelRepo repository.VsLLMModelRepository
}

func (s *vsLLMProviderService) Create(ctx context.Context, request *v1.VsLLMProviderCreateRequest) error {
	createdBy, _ := ctx.Value("login_name").(string)
	provider := &model.VsLLMProvider{
		Name:           strings.TrimSpace(request.Name),
		APIBaseURL:     strings.TrimSpace(request.APIBaseURL),
		APIKey:         strings.TrimSpace(request.APIKey),
		MaxConcurrency: request.MaxConcurrency,
		SortOrder:      request.SortOrder,
		Status:         request.Status,
		Remark:         request.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: createdBy,
		},
	}
	if provider.MaxConcurrency <= 0 {
		provider.MaxConcurrency = 1
	}
	return s.repo.Create(ctx, provider)
}

func (s *vsLLMProviderService) Update(ctx context.Context, request *v1.VsLLMProviderUpdateRequest) error {
	updatedBy, _ := ctx.Value("login_name").(string)
	fields := map[string]interface{}{
		"name":            strings.TrimSpace(request.Name),
		"api_base_url":    strings.TrimSpace(request.APIBaseURL),
		"max_concurrency": request.MaxConcurrency,
		"sort_order":      request.SortOrder,
		"status":          request.Status,
		"remark":          request.Remark,
		"updated_by":      updatedBy,
	}
	if request.MaxConcurrency <= 0 {
		fields["max_concurrency"] = 1
	}
	// API Key 仅在传入非脱敏值时更新
	apiKey := strings.TrimSpace(request.APIKey)
	if apiKey != "" && !strings.Contains(apiKey, "****") {
		fields["api_key"] = apiKey
	}
	return s.repo.UpdateFields(ctx, request.ID, fields)
}

func (s *vsLLMProviderService) Delete(ctx context.Context, id uint64) error {
	count, err := s.modelRepo.CountByProviderID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该厂家下仍有 %d 个模型，请先删除模型", count)
	}
	return s.repo.Delete(ctx, id)
}

func (s *vsLLMProviderService) DeleteByIDs(ctx context.Context, ids []uint64) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
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

	responses := make([]*v1.VsLLMProviderDetailResponse, 0, len(providers))
	for _, p := range providers {
		responses = append(responses, s.convertToDetailResponse(p))
	}
	return responses, total, nil
}

func (s *vsLLMProviderService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsLLMProviderService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

// TestConnection 通过一次最小的 Chat Completions 请求测试厂家连通性
func (s *vsLLMProviderService) TestConnection(ctx context.Context, request *v1.VsLLMProviderTestRequest) *v1.VsLLMProviderTestResponse {
	start := time.Now()

	baseURL := strings.TrimSpace(request.APIBaseURL)
	apiKey := strings.TrimSpace(request.APIKey)
	modelName := strings.TrimSpace(request.Model)
	if baseURL == "" || modelName == "" {
		return &v1.VsLLMProviderTestResponse{
			Success:  false,
			Message:  "API地址和测试模型不能为空",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// 脱敏或留空时回查已保存的原始 Key
	if request.ID > 0 && (apiKey == "" || strings.Contains(apiKey, "****")) {
		existing, err := s.repo.FindByID(ctx, request.ID)
		if err != nil {
			return &v1.VsLLMProviderTestResponse{
				Success:  false,
				Message:  fmt.Sprintf("获取已保存的 API Key 失败: %v", err),
				Duration: time.Since(start).Milliseconds(),
			}
		}
		apiKey = existing.APIKey
	}

	baseURL = strings.TrimRight(baseURL, "/")
	chatURL := baseURL + "/chat/completions"
	if !strings.Contains(baseURL, "/v1") {
		chatURL = baseURL + "/v1/chat/completions"
	}

	reqBody := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
		"max_tokens": 1,
		"stream":     false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return &v1.VsLLMProviderTestResponse{
			Success:  false,
			Message:  fmt.Sprintf("构建请求体失败: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	req, err := http.NewRequest("POST", chatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return &v1.VsLLMProviderTestResponse{
			Success:  false,
			Message:  fmt.Sprintf("构建请求失败: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &v1.VsLLMProviderTestResponse{
			Success:  false,
			Message:  fmt.Sprintf("连接失败: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err == nil {
			if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
				return &v1.VsLLMProviderTestResponse{
					Success:  true,
					Message:  fmt.Sprintf("连接成功，模型 %s 响应正常", modelName),
					Duration: time.Since(start).Milliseconds(),
				}
			}
		}
		return &v1.VsLLMProviderTestResponse{
			Success:  true,
			Message:  "连接成功（HTTP 200）",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return &v1.VsLLMProviderTestResponse{
		Success:  false,
		Message:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		Duration: time.Since(start).Milliseconds(),
	}
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func (s *vsLLMProviderService) convertToDetailResponse(p *model.VsLLMProvider) *v1.VsLLMProviderDetailResponse {
	return &v1.VsLLMProviderDetailResponse{
		ID:             p.ProviderID,
		Name:           p.Name,
		APIBaseURL:     p.APIBaseURL,
		APIKey:         maskAPIKey(p.APIKey),
		MaxConcurrency: p.MaxConcurrency,
		SortOrder:      p.SortOrder,
		Status:         p.Status,
		Remark:         p.Remark,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
