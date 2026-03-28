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
	ImportFromAssets(ctx context.Context, request *v1.VsVoiceProfileImportRequest) (int, error)
}

func NewVsVoiceProfileService(
	service *Service,
	repo repository.VsVoiceProfileRepository,
	voiceAssetRepo repository.VsVoiceAssetRepository,
	voiceEmotionRepo repository.VsVoiceEmotionRepository,
) VsVoiceProfileService {
	return &vsVoiceProfileService{Service: service, repo: repo, voiceAssetRepo: voiceAssetRepo, voiceEmotionRepo: voiceEmotionRepo}
}

type vsVoiceProfileService struct {
	*Service
	repo             repository.VsVoiceProfileRepository
	voiceAssetRepo   repository.VsVoiceAssetRepository
	voiceEmotionRepo repository.VsVoiceEmotionRepository
}

func (s *vsVoiceProfileService) Create(ctx context.Context, request *v1.VsVoiceProfileCreateRequest) error {
	profile := &model.VsVoiceProfile{
		ProjectID:         request.ProjectID,
		Name:              request.Name,
		Gender:            request.Gender,
		Description:       request.Description,
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
	existing.Gender = request.Gender
	existing.Description = request.Description
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
	_ = s.voiceEmotionRepo.DeleteByVoiceProfileID(ctx, id)
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

func (s *vsVoiceProfileService) ImportFromAssets(ctx context.Context, request *v1.VsVoiceProfileImportRequest) (int, error) {
	assets, err := s.voiceAssetRepo.FindByIDs(ctx, request.VoiceAssetIDs)
	if err != nil {
		return 0, fmt.Errorf("查询音色资产失败: %w", err)
	}
	if len(assets) == 0 {
		return 0, nil
	}

	loginName := ctx.Value("login_name").(string)
	deptID := ctx.Value("dept_id").(uint)

	profiles := make([]*model.VsVoiceProfile, 0, len(assets))
	for _, asset := range assets {
		profiles = append(profiles, &model.VsVoiceProfile{
			ProjectID:         request.ProjectID,
			Name:              asset.Name,
			Gender:            asset.Gender,
			Description:       asset.Description,
			VoiceAssetID:      &asset.VoiceAssetID,
			ReferenceAudioURL: asset.ReferenceAudioURL,
			ReferenceText:     asset.ReferenceText,
			Status:            "0",
			BaseModel: model.BaseModel{
				CreatedBy: loginName,
				DeptID:    deptID,
			},
		})
	}

	if err := s.repo.BatchCreate(ctx, profiles); err != nil {
		return 0, fmt.Errorf("批量创建声音配置失败: %w", err)
	}

	// 复制音色资产的情绪音频到新建的声音配置
	assetEmotions, _ := s.voiceEmotionRepo.FindByVoiceAssetIDs(ctx, request.VoiceAssetIDs)
	if len(assetEmotions) > 0 {
		assetToProfile := make(map[uint64]uint64)
		for _, p := range profiles {
			if p.VoiceAssetID != nil {
				assetToProfile[*p.VoiceAssetID] = p.VoiceProfileID
			}
		}

		var emotionCopies []*model.VsVoiceEmotion
		for _, e := range assetEmotions {
			if e.VoiceAssetID == nil {
				continue
			}
			profileID, ok := assetToProfile[*e.VoiceAssetID]
			if !ok {
				continue
			}
			emotionCopies = append(emotionCopies, &model.VsVoiceEmotion{
				VoiceProfileID:    &profileID,
				EmotionType:       e.EmotionType,
				EmotionStrength:   e.EmotionStrength,
				ReferenceAudioURL: e.ReferenceAudioURL,
				ReferenceText:     e.ReferenceText,
				BaseModel: model.BaseModel{
					CreatedBy: loginName,
					DeptID:    deptID,
				},
			})
		}
		if len(emotionCopies) > 0 {
			_ = s.voiceEmotionRepo.BatchCreate(ctx, emotionCopies)
		}
	}

	return len(profiles), nil
}

func (s *vsVoiceProfileService) convertToDetailResponse(p *model.VsVoiceProfile) *v1.VsVoiceProfileDetailResponse {
	resp := &v1.VsVoiceProfileDetailResponse{
		ID:                p.VoiceProfileID,
		ProjectID:         p.ProjectID,
		Name:              p.Name,
		Gender:            p.Gender,
		Description:       p.Description,
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
