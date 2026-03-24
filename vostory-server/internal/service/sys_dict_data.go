package service

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysDictDataService interface {
	Create(ctx context.Context, request *v1.SysDictDataCreateRequest) error
	Update(ctx context.Context, request *v1.SysDictDataUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysDictDataDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysDictDataListQuery) ([]*v1.SysDictDataDetailResponse, int64, error)
	FindByDictType(ctx context.Context, dictType string) ([]*v1.SysDictDataDetailResponse, error)
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

func NewSysDictDataService(
	service *Service,
	sysDictDataRepository repository.SysDictDataRepository,
) SysDictDataService {
	return &sysDictDataService{
		Service:               service,
		sysDictDataRepository: sysDictDataRepository,
	}
}

type sysDictDataService struct {
	*Service
	sysDictDataRepository repository.SysDictDataRepository
}

func (s *sysDictDataService) Create(ctx context.Context, request *v1.SysDictDataCreateRequest) error {
	dictData := &model.SysDictData{
		DictSort:  request.DictSort,
		DictLabel: request.DictLabel,
		DictValue: request.DictValue,
		DictType:  request.DictType,
		CSSClass:  request.CSSClass,
		ListClass: request.ListClass,
		IsDefault: request.IsDefault,
		Status:    request.Status,
		Remark:    request.Remark,
		BaseModel: model.BaseModel{
			DeptID:    ctx.Value("dept_id").(uint),
			CreatedBy: ctx.Value("login_name").(string),
		},
	}
	return s.sysDictDataRepository.Create(ctx, dictData)
}

func (s *sysDictDataService) Update(ctx context.Context, request *v1.SysDictDataUpdateRequest) error {
	dictData := &model.SysDictData{
		DictCode:  request.ID,
		DictSort:  request.DictSort,
		DictLabel: request.DictLabel,
		DictValue: request.DictValue,
		DictType:  request.DictType,
		CSSClass:  request.CSSClass,
		ListClass: request.ListClass,
		IsDefault: request.IsDefault,
		Status:    request.Status,
		Remark:    request.Remark,
		BaseModel: model.BaseModel{
			UpdatedBy: ctx.Value("login_name").(string),
		},
	}
	return s.sysDictDataRepository.Update(ctx, dictData)
}

func (s *sysDictDataService) Delete(ctx context.Context, id uint) error {
	return s.sysDictDataRepository.Delete(ctx, id)
}

func (s *sysDictDataService) FindByID(ctx context.Context, id uint) (*v1.SysDictDataDetailResponse, error) {
	dictData, err := s.sysDictDataRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(dictData), nil
}

func (s *sysDictDataService) FindWithPagination(ctx context.Context, query *v1.SysDictDataListQuery) ([]*v1.SysDictDataDetailResponse, int64, error) {
	dictDataList, total, err := s.sysDictDataRepository.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.SysDictDataDetailResponse
	for _, dd := range dictDataList {
		responses = append(responses, s.convertToDetailResponse(dd))
	}

	return responses, total, nil
}

func (s *sysDictDataService) FindByDictType(ctx context.Context, dictType string) ([]*v1.SysDictDataDetailResponse, error) {
	dictDataList, err := s.sysDictDataRepository.FindByDictType(ctx, dictType)
	if err != nil {
		return nil, err
	}

	var responses []*v1.SysDictDataDetailResponse
	for _, dd := range dictDataList {
		responses = append(responses, s.convertToDetailResponse(dd))
	}

	return responses, nil
}

func (s *sysDictDataService) Enable(ctx context.Context, id uint) error {
	return s.sysDictDataRepository.Enable(ctx, id)
}

func (s *sysDictDataService) Disable(ctx context.Context, id uint) error {
	return s.sysDictDataRepository.Disable(ctx, id)
}

func (s *sysDictDataService) convertToDetailResponse(dictData *model.SysDictData) *v1.SysDictDataDetailResponse {
	return &v1.SysDictDataDetailResponse{
		ID:        dictData.DictCode,
		DictSort:  dictData.DictSort,
		DictLabel: dictData.DictLabel,
		DictValue: dictData.DictValue,
		DictType:  dictData.DictType,
		CSSClass:  dictData.CSSClass,
		ListClass: dictData.ListClass,
		IsDefault: dictData.IsDefault,
		Status:    dictData.Status,
		Remark:    dictData.Remark,
		CreatedAt: dictData.CreatedAt,
		UpdatedAt: dictData.UpdatedAt,
	}
}
