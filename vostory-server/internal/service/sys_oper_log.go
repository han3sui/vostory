package service

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysOperLogService interface {
	Create(ctx context.Context, operLog *model.SysOperLog) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysOperLogDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysOperLogListQuery) ([]*v1.SysOperLogDetailResponse, int64, error)
	Clean(ctx context.Context) error
}

func NewSysOperLogService(
	service *Service,
	sysOperLogRepository repository.SysOperLogRepository,
) SysOperLogService {
	return &sysOperLogService{
		Service:              service,
		sysOperLogRepository: sysOperLogRepository,
	}
}

type sysOperLogService struct {
	*Service
	sysOperLogRepository repository.SysOperLogRepository
}

func (s *sysOperLogService) Create(ctx context.Context, operLog *model.SysOperLog) error {
	return s.sysOperLogRepository.Create(ctx, operLog)
}

func (s *sysOperLogService) Delete(ctx context.Context, id uint) error {
	return s.sysOperLogRepository.Delete(ctx, id)
}

func (s *sysOperLogService) FindByID(ctx context.Context, id uint) (*v1.SysOperLogDetailResponse, error) {
	operLog, err := s.sysOperLogRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(operLog), nil
}

func (s *sysOperLogService) FindWithPagination(ctx context.Context, query *v1.SysOperLogListQuery) ([]*v1.SysOperLogDetailResponse, int64, error) {
	operLogs, total, err := s.sysOperLogRepository.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.SysOperLogDetailResponse
	for _, ol := range operLogs {
		responses = append(responses, s.convertToDetailResponse(ol))
	}

	return responses, total, nil
}

func (s *sysOperLogService) Clean(ctx context.Context) error {
	return s.sysOperLogRepository.Clean(ctx)
}

func (s *sysOperLogService) convertToDetailResponse(operLog *model.SysOperLog) *v1.SysOperLogDetailResponse {
	return &v1.SysOperLogDetailResponse{
		ID:            operLog.OperID,
		Title:         operLog.Title,
		BusinessType:  operLog.BusinessType,
		Method:        operLog.Method,
		RequestMethod: operLog.RequestMethod,
		OperatorType:  operLog.OperatorType,
		OperName:      operLog.OperName,
		DeptName:      operLog.DeptName,
		OperURL:       operLog.OperURL,
		OperIP:        operLog.OperIP,
		OperLocation:  operLog.OperLocation,
		OperParam:     operLog.OperParam,
		JSONResult:    operLog.JSONResult,
		Status:        operLog.Status,
		ErrorMsg:      operLog.ErrorMsg,
		OperTime:      operLog.OperTime,
		CostTime:      operLog.CostTime,
	}
}
