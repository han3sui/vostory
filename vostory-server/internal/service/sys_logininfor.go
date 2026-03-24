package service

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysLogininforService interface {
	Create(ctx context.Context, data *model.SysLogininfor) error
	FindByID(ctx context.Context, id uint) (*model.SysLogininfor, error)
	FindWithPagination(ctx context.Context, query *v1.SysLogininforQueryParams) ([]*model.SysLogininfor, int64, error)
}

type sysLogininforService struct {
	repo repository.SysLogininforRepository
}

func NewSysLogininforService(repo repository.SysLogininforRepository) SysLogininforService {
	return &sysLogininforService{repo: repo}
}

func (s *sysLogininforService) Create(ctx context.Context, data *model.SysLogininfor) error {
	return s.repo.Create(ctx, data)
}

func (s *sysLogininforService) FindByID(ctx context.Context, id uint) (*model.SysLogininfor, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *sysLogininforService) FindWithPagination(ctx context.Context, query *v1.SysLogininforQueryParams) ([]*model.SysLogininfor, int64, error) {
	return s.repo.FindWithPagination(ctx, query)
}
