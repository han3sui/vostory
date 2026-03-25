package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsVoiceProfileService interface {
	Create(ctx context.Context, request *v1.VsVoiceProfileCreateRequest) error
	Update(ctx context.Context, request *v1.VsVoiceProfileUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsVoiceProfileDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsVoiceProfileListQuery) ([]*v1.VsVoiceProfileDetailResponse, int64, error)
	FindByProjectID(ctx context.Context, projectID uint64) ([]*v1.VsVoiceProfileOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsVoiceProfileService(
	service *Service,
	repo repository.VsVoiceProfileRepository,
) VsVoiceProfileService {
	return &vsVoiceProfileService{Service: service, repo: repo}
}

type vsVoiceProfileService struct {
	*Service
	repo repository.VsVoiceProfileRepository
}

func (s *vsVoiceProfileService) Create(ctx context.Context, request *v1.VsVoiceProfileCreateRequest) error {
	id, err := s.sid.GenUint64()
	if err != nil {
		return fmt.Errorf("生成ID失败: %w", err)
	}

	profile := &model.VsVoiceProfile{
		VoiceProfileID:    id,
		ProjectID:         request.ProjectID,
		Name:              request.Name,
		VoiceAssetID:      request.VoiceAssetID,
		ReferenceAudioURL: request.ReferenceAudioURL,
		ReferenceText:     request.ReferenceText,
		TTSProviderID:     request.TTSProviderID,
		TTSParams:         model.TTSParamsMap(request.TTSParams),
		Status:            "0",
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, profile)
}

func (s *vsVoiceProfileService) Update(ctx context.Context, request *v1.VsVoiceProfileUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("声音配置不存在")
	}

	existing.Name = request.Name
	existing.VoiceAssetID = request.VoiceAssetID
	existing.ReferenceAudioURL = request.ReferenceAudioURL
	existing.ReferenceText = request.ReferenceText
	existing.TTSProviderID = request.TTSProviderID
	existing.TTSParams = model.TTSParamsMap(request.TTSParams)
	if request.Status != "" {
		existing.Status = request.Status
	}
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsVoiceProfileService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsVoiceProfileService) FindByID(ctx context.Context, id uint64) (*v1.VsVoiceProfileDetailResponse, error) {
	profile, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(profile), nil
}

func (s *vsVoiceProfileService) FindWithPagination(ctx context.Context, query *v1.VsVoiceProfileListQuery) ([]*v1.VsVoiceProfileDetailResponse, int64, error) {
	profiles, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsVoiceProfileDetailResponse
	for _, p := range profiles {
		responses = append(responses, s.convertToDetailResponse(p))
	}
	return responses, total, nil
}

func (s *vsVoiceProfileService) FindByProjectID(ctx context.Context, projectID uint64) ([]*v1.VsVoiceProfileOptionResponse, error) {
	profiles, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var options []*v1.VsVoiceProfileOptionResponse
	for _, p := range profiles {
		options = append(options, &v1.VsVoiceProfileOptionResponse{
			ID:   p.VoiceProfileID,
			Name: p.Name,
		})
	}
	return options, nil
}

func (s *vsVoiceProfileService) Enable(ctx context.Context, id uint64) error {
	profile, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("声音配置不存在")
	}
	profile.Status = "0"
	profile.UpdatedBy = ctx.Value("login_name").(string)
	return s.repo.Update(ctx, profile)
}

func (s *vsVoiceProfileService) Disable(ctx context.Context, id uint64) error {
	profile, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("声音配置不存在")
	}
	profile.Status = "1"
	profile.UpdatedBy = ctx.Value("login_name").(string)
	return s.repo.Update(ctx, profile)
}

func (s *vsVoiceProfileService) convertToDetailResponse(p *model.VsVoiceProfile) *v1.VsVoiceProfileDetailResponse {
	resp := &v1.VsVoiceProfileDetailResponse{
		ID:                p.VoiceProfileID,
		ProjectID:         p.ProjectID,
		Name:              p.Name,
		VoiceAssetID:      p.VoiceAssetID,
		ReferenceAudioURL: p.ReferenceAudioURL,
		ReferenceText:     p.ReferenceText,
		TTSProviderID:     p.TTSProviderID,
		TTSParams:         map[string]interface{}(p.TTSParams),
		Status:            p.Status,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	if p.TTSProvider != nil {
		resp.TTSProviderName = p.TTSProvider.Name
	}
	return resp
}
