package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsVoiceAssetService interface {
	Create(ctx context.Context, request *v1.VsVoiceAssetCreateRequest) error
	Update(ctx context.Context, request *v1.VsVoiceAssetUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsVoiceAssetDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsVoiceAssetListQuery) ([]*v1.VsVoiceAssetDetailResponse, int64, error)
	FindAllEnabled(ctx context.Context) ([]*v1.VsVoiceAssetOptionResponse, error)
	Enable(ctx context.Context, id uint64) error
	Disable(ctx context.Context, id uint64) error
}

func NewVsVoiceAssetService(
	service *Service,
	repo repository.VsVoiceAssetRepository,
	voiceEmotionRepo repository.VsVoiceEmotionRepository,
) VsVoiceAssetService {
	return &vsVoiceAssetService{Service: service, repo: repo, voiceEmotionRepo: voiceEmotionRepo}
}

type vsVoiceAssetService struct {
	*Service
	repo             repository.VsVoiceAssetRepository
	voiceEmotionRepo repository.VsVoiceEmotionRepository
}

func (s *vsVoiceAssetService) Create(ctx context.Context, request *v1.VsVoiceAssetCreateRequest) error {
	asset := &model.VsVoiceAsset{
		WorkspaceID:       1,
		Name:              request.Name,
		Gender:            request.Gender,
		Description:       request.Description,
		ReferenceAudioURL: request.ReferenceAudioURL,
		ReferenceText:     request.ReferenceText,
		TTSProviderID:     request.TTSProviderID,
		Tags:              model.StringList(request.Tags),
		Status:            "0",
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}
	return s.repo.Create(ctx, asset)
}

func (s *vsVoiceAssetService) Update(ctx context.Context, request *v1.VsVoiceAssetUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("声音资产不存在")
	}

	existing.Name = request.Name
	existing.Gender = request.Gender
	existing.Description = request.Description
	existing.ReferenceAudioURL = request.ReferenceAudioURL
	existing.ReferenceText = request.ReferenceText
	existing.TTSProviderID = request.TTSProviderID
	existing.Tags = model.StringList(request.Tags)
	if request.Status != "" {
		existing.Status = request.Status
	}
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsVoiceAssetService) Delete(ctx context.Context, id uint64) error {
	_ = s.voiceEmotionRepo.DeleteByVoiceAssetID(ctx, id)
	return s.repo.Delete(ctx, id)
}

func (s *vsVoiceAssetService) FindByID(ctx context.Context, id uint64) (*v1.VsVoiceAssetDetailResponse, error) {
	asset, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toDetail(asset), nil
}

func (s *vsVoiceAssetService) FindWithPagination(ctx context.Context, query *v1.VsVoiceAssetListQuery) ([]*v1.VsVoiceAssetDetailResponse, int64, error) {
	assets, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	var responses []*v1.VsVoiceAssetDetailResponse
	for _, a := range assets {
		responses = append(responses, s.toDetail(a))
	}
	return responses, total, nil
}

func (s *vsVoiceAssetService) FindAllEnabled(ctx context.Context) ([]*v1.VsVoiceAssetOptionResponse, error) {
	assets, err := s.repo.FindAllEnabled(ctx)
	if err != nil {
		return nil, err
	}
	var options []*v1.VsVoiceAssetOptionResponse
	for _, a := range assets {
		options = append(options, &v1.VsVoiceAssetOptionResponse{
			ID:                a.VoiceAssetID,
			Name:              a.Name,
			Gender:            a.Gender,
			Tags:              []string(a.Tags),
			ReferenceAudioURL: a.ReferenceAudioURL,
		})
	}
	return options, nil
}

func (s *vsVoiceAssetService) Enable(ctx context.Context, id uint64) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("声音资产不存在")
	}
	existing.Status = "0"
	existing.UpdatedBy = ctx.Value("login_name").(string)
	return s.repo.Update(ctx, existing)
}

func (s *vsVoiceAssetService) Disable(ctx context.Context, id uint64) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("声音资产不存在")
	}
	existing.Status = "1"
	existing.UpdatedBy = ctx.Value("login_name").(string)
	return s.repo.Update(ctx, existing)
}

func (s *vsVoiceAssetService) toDetail(a *model.VsVoiceAsset) *v1.VsVoiceAssetDetailResponse {
	resp := &v1.VsVoiceAssetDetailResponse{
		ID:                a.VoiceAssetID,
		Name:              a.Name,
		Gender:            a.Gender,
		Description:       a.Description,
		ReferenceAudioURL: a.ReferenceAudioURL,
		ReferenceText:     a.ReferenceText,
		TTSProviderID:     a.TTSProviderID,
		Tags:              []string(a.Tags),
		Status:            a.Status,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
	if a.TTSProvider != nil {
		resp.TTSProviderName = a.TTSProvider.Name
	}
	if resp.Tags == nil {
		resp.Tags = []string{}
	}
	return resp
}
