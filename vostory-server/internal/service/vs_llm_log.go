package service

import (
	"context"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsLLMLogService interface {
	FindByID(ctx context.Context, id uint64) (*v1.VsLLMLogDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsLLMLogListQuery) ([]*v1.VsLLMLogDetailResponse, int64, error)
	Delete(ctx context.Context, id uint64) error
}

func NewVsLLMLogService(
	service *Service,
	repo repository.VsLLMLogRepository,
) VsLLMLogService {
	return &vsLLMLogService{Service: service, repo: repo}
}

type vsLLMLogService struct {
	*Service
	repo repository.VsLLMLogRepository
}

func (s *vsLLMLogService) FindByID(ctx context.Context, id uint64) (*v1.VsLLMLogDetailResponse, error) {
	log, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(log), nil
}

func (s *vsLLMLogService) FindWithPagination(ctx context.Context, query *v1.VsLLMLogListQuery) ([]*v1.VsLLMLogDetailResponse, int64, error) {
	logs, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsLLMLogDetailResponse
	for _, l := range logs {
		responses = append(responses, s.convertToDetailResponse(l))
	}
	return responses, total, nil
}

func (s *vsLLMLogService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsLLMLogService) convertToDetailResponse(l *model.VsLLMLog) *v1.VsLLMLogDetailResponse {
	resp := &v1.VsLLMLogDetailResponse{
		ID:            l.LogID,
		ProjectID:     l.ProjectID,
		ProviderID:    l.ProviderID,
		TemplateID:    l.TemplateID,
		ModelName:     l.ModelName,
		InputTokens:   l.InputTokens,
		OutputTokens:  l.OutputTokens,
		InputSummary:  l.InputSummary,
		OutputSummary: l.OutputSummary,
		CostTime:      l.CostTime,
		Status:        l.Status,
		ErrorMessage:  l.ErrorMessage,
		CreatedAt:     l.CreatedAt,
	}
	if l.Project != nil {
		resp.ProjectName = l.Project.Name
	}
	if l.LLMProvider != nil {
		resp.ProviderName = l.LLMProvider.Name
	}
	if l.PromptTemplate != nil {
		resp.TemplateName = l.PromptTemplate.Name
	}
	return resp
}
