package service

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysApiService interface {
	Create(ctx context.Context, req *v1.SysApiCreateRequest) error
	Update(ctx context.Context, req *v1.SysApiUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindWithPagination(ctx context.Context, req *v1.SysApiListQuery) ([]*model.SysApi, int64, error)
	ListTag(ctx context.Context) ([]string, error)
}

func NewSysApiService(
	sysApiRepo repository.SysApiRepository,
) SysApiService {
	return &sysApiService{
		sysApiRepo: sysApiRepo,
	}
}

type sysApiService struct {
	sysApiRepo repository.SysApiRepository
}

func (s *sysApiService) Create(ctx context.Context, req *v1.SysApiCreateRequest) error {
	return s.sysApiRepo.Create(ctx, &model.SysApi{
		Method:      req.Method,
		Path:        req.Path,
		Name:        req.Name,
		Description: req.Desc,
	})
}

func (s *sysApiService) Update(ctx context.Context, req *v1.SysApiUpdateRequest) error {
	return s.sysApiRepo.Update(ctx, &model.SysApi{
		ID:          req.ID,
		Method:      req.Method,
		Path:        req.Path,
		Name:        req.Name,
		Description: req.Desc,
	})
}

func (s *sysApiService) Delete(ctx context.Context, id uint) error {
	return s.sysApiRepo.Delete(ctx, id)
}

func (s *sysApiService) FindWithPagination(ctx context.Context, req *v1.SysApiListQuery) ([]*model.SysApi, int64, error) {
	return s.sysApiRepo.FindWithPagination(ctx, req)
}

func (s *sysApiService) ListTag(ctx context.Context) ([]string, error) {
	return s.sysApiRepo.ListTag(ctx)
}
