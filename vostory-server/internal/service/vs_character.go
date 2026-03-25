package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsCharacterService interface {
	Create(ctx context.Context, request *v1.VsCharacterCreateRequest) error
	Update(ctx context.Context, request *v1.VsCharacterUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsCharacterDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsCharacterListQuery) ([]*v1.VsCharacterDetailResponse, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*v1.VsCharacterOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsCharacterService(
	service *Service,
	repo repository.VsCharacterRepository,
) VsCharacterService {
	return &vsCharacterService{Service: service, repo: repo}
}

type vsCharacterService struct {
	*Service
	repo repository.VsCharacterRepository
}

func (s *vsCharacterService) Create(ctx context.Context, request *v1.VsCharacterCreateRequest) error {
	id, err := s.sid.GenUint64()
	if err != nil {
		return fmt.Errorf("生成ID失败: %w", err)
	}

	character := &model.VsCharacter{
		CharacterID:    id,
		ProjectID:      request.ProjectID,
		Name:           request.Name,
		Aliases:        model.StringList(request.Aliases),
		Gender:         request.Gender,
		Description:    request.Description,
		Level:          request.Level,
		VoiceProfileID: request.VoiceProfileID,
		SortOrder:      request.SortOrder,
		Status:         request.Status,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, character)
}

func (s *vsCharacterService) Update(ctx context.Context, request *v1.VsCharacterUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("角色不存在")
	}

	existing.Name = request.Name
	existing.Aliases = model.StringList(request.Aliases)
	existing.Gender = request.Gender
	existing.Description = request.Description
	existing.Level = request.Level
	existing.VoiceProfileID = request.VoiceProfileID
	existing.SortOrder = request.SortOrder
	existing.Status = request.Status
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsCharacterService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsCharacterService) FindByID(ctx context.Context, id uint64) (*v1.VsCharacterDetailResponse, error) {
	character, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(character), nil
}

func (s *vsCharacterService) FindWithPagination(ctx context.Context, query *v1.VsCharacterListQuery) ([]*v1.VsCharacterDetailResponse, int64, error) {
	characters, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsCharacterDetailResponse
	for _, c := range characters {
		responses = append(responses, s.convertToDetailResponse(c))
	}
	return responses, total, nil
}

func (s *vsCharacterService) FindByProjectID(ctx context.Context, projectID uint64) ([]*v1.VsCharacterOptionResponse, error) {
	characters, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsCharacterOptionResponse
	for _, c := range characters {
		responses = append(responses, &v1.VsCharacterOptionResponse{
			ID:   c.CharacterID,
			Name: c.Name,
		})
	}
	return responses, nil
}

func (s *vsCharacterService) Enable(ctx context.Context, id uint64) error {
	return s.repo.Enable(ctx, id)
}

func (s *vsCharacterService) Disable(ctx context.Context, id uint64) error {
	return s.repo.Disable(ctx, id)
}

func (s *vsCharacterService) convertToDetailResponse(c *model.VsCharacter) *v1.VsCharacterDetailResponse {
	return &v1.VsCharacterDetailResponse{
		ID:             c.CharacterID,
		ProjectID:      c.ProjectID,
		Name:           c.Name,
		Aliases:        []string(c.Aliases),
		Gender:         c.Gender,
		Description:    c.Description,
		Level:          c.Level,
		VoiceProfileID: c.VoiceProfileID,
		SortOrder:      c.SortOrder,
		Status:         c.Status,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
