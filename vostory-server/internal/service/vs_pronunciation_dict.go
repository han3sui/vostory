package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsPronunciationDictService interface {
	Create(ctx context.Context, request *v1.VsPronunciationDictCreateRequest) error
	Update(ctx context.Context, request *v1.VsPronunciationDictUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsPronunciationDictDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsPronunciationDictListQuery) ([]*v1.VsPronunciationDictDetailResponse, int64, error)
}

func NewVsPronunciationDictService(
	service *Service,
	repo repository.VsPronunciationDictRepository,
) VsPronunciationDictService {
	return &vsPronunciationDictService{Service: service, repo: repo}
}

type vsPronunciationDictService struct {
	*Service
	repo repository.VsPronunciationDictRepository
}

func (s *vsPronunciationDictService) Create(ctx context.Context, request *v1.VsPronunciationDictCreateRequest) error {
	dict := &model.VsPronunciationDict{
		ProjectID: request.ProjectID,
		Word:      request.Word,
		Phoneme:   request.Phoneme,
		Remark:    request.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, dict)
}

func (s *vsPronunciationDictService) Update(ctx context.Context, request *v1.VsPronunciationDictUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("词典条目不存在")
	}

	existing.Word = request.Word
	existing.Phoneme = request.Phoneme
	existing.Remark = request.Remark
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsPronunciationDictService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsPronunciationDictService) FindByID(ctx context.Context, id uint64) (*v1.VsPronunciationDictDetailResponse, error) {
	dict, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(dict), nil
}

func (s *vsPronunciationDictService) FindWithPagination(ctx context.Context, query *v1.VsPronunciationDictListQuery) ([]*v1.VsPronunciationDictDetailResponse, int64, error) {
	dicts, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsPronunciationDictDetailResponse
	for _, d := range dicts {
		responses = append(responses, s.convertToDetailResponse(d))
	}
	return responses, total, nil
}

func (s *vsPronunciationDictService) convertToDetailResponse(d *model.VsPronunciationDict) *v1.VsPronunciationDictDetailResponse {
	return &v1.VsPronunciationDictDetailResponse{
		ID:        d.DictID,
		ProjectID: d.ProjectID,
		Word:      d.Word,
		Phoneme:   d.Phoneme,
		Remark:    d.Remark,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
