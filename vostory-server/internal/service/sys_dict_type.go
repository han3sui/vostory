package service

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysDictTypeService interface {
	Create(ctx context.Context, request *v1.SysDictTypeCreateRequest) error
	Update(ctx context.Context, request *v1.SysDictTypeUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysDictTypeDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysDictTypeListQuery) ([]*v1.SysDictTypeDetailResponse, int64, error)
	FindAll(ctx context.Context) ([]*v1.SysDictTypeDetailResponse, error)
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

func NewSysDictTypeService(
	service *Service,
	sysDictTypeRepository repository.SysDictTypeRepository,
) SysDictTypeService {
	return &sysDictTypeService{
		Service:               service,
		sysDictTypeRepository: sysDictTypeRepository,
	}
}

type sysDictTypeService struct {
	*Service
	sysDictTypeRepository repository.SysDictTypeRepository
}

func (s *sysDictTypeService) Create(ctx context.Context, request *v1.SysDictTypeCreateRequest) error {
	dictType := &model.SysDictType{
		DictName: request.DictName,
		DictType: request.DictType,
		Status:   request.Status,
		Remark:   request.Remark,
		BaseModel: model.BaseModel{
			DeptID:    ctx.Value("dept_id").(uint),
			CreatedBy: ctx.Value("login_name").(string),
		},
	}
	return s.sysDictTypeRepository.Create(ctx, dictType)
}

func (s *sysDictTypeService) Update(ctx context.Context, request *v1.SysDictTypeUpdateRequest) error {
	dictType := &model.SysDictType{
		DictID:   request.ID,
		DictName: request.DictName,
		DictType: request.DictType,
		Status:   request.Status,
		Remark:   request.Remark,
		BaseModel: model.BaseModel{
			UpdatedBy: ctx.Value("login_name").(string),
		},
	}
	return s.sysDictTypeRepository.Update(ctx, dictType)
}

func (s *sysDictTypeService) Delete(ctx context.Context, id uint) error {
	return s.sysDictTypeRepository.Delete(ctx, id)
}

func (s *sysDictTypeService) FindByID(ctx context.Context, id uint) (*v1.SysDictTypeDetailResponse, error) {
	dictType, err := s.sysDictTypeRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(dictType), nil
}

func (s *sysDictTypeService) FindWithPagination(ctx context.Context, query *v1.SysDictTypeListQuery) ([]*v1.SysDictTypeDetailResponse, int64, error) {
	dictTypes, total, err := s.sysDictTypeRepository.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.SysDictTypeDetailResponse
	for _, dt := range dictTypes {
		responses = append(responses, s.convertToDetailResponse(dt))
	}

	return responses, total, nil
}

func (s *sysDictTypeService) FindAll(ctx context.Context) ([]*v1.SysDictTypeDetailResponse, error) {
	dictTypes, err := s.sysDictTypeRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []*v1.SysDictTypeDetailResponse
	for _, dt := range dictTypes {
		responses = append(responses, s.convertToDetailResponse(dt))
	}

	return responses, nil
}

func (s *sysDictTypeService) Enable(ctx context.Context, id uint) error {
	return s.sysDictTypeRepository.Enable(ctx, id)
}

func (s *sysDictTypeService) Disable(ctx context.Context, id uint) error {
	return s.sysDictTypeRepository.Disable(ctx, id)
}

func (s *sysDictTypeService) convertToDetailResponse(dictType *model.SysDictType) *v1.SysDictTypeDetailResponse {
	return &v1.SysDictTypeDetailResponse{
		ID:        dictType.DictID,
		DictName:  dictType.DictName,
		DictType:  dictType.DictType,
		Status:    dictType.Status,
		Remark:    dictType.Remark,
		CreatedAt: dictType.CreatedAt,
		UpdatedAt: dictType.UpdatedAt,
	}
}
