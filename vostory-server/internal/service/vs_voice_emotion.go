package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsVoiceEmotionService interface {
	Create(ctx context.Context, request *v1.VsVoiceEmotionCreateRequest) error
	Update(ctx context.Context, request *v1.VsVoiceEmotionUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsVoiceEmotionDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsVoiceEmotionListQuery) ([]*v1.VsVoiceEmotionDetailResponse, int64, error)
	FindByVoiceProfileID(ctx context.Context, voiceProfileID uint64) ([]*v1.VsVoiceEmotionDetailResponse, error)
	FindByVoiceAssetID(ctx context.Context, voiceAssetID uint64) ([]*v1.VsVoiceEmotionDetailResponse, error)
}

func NewVsVoiceEmotionService(
	service *Service,
	repo repository.VsVoiceEmotionRepository,
) VsVoiceEmotionService {
	return &vsVoiceEmotionService{Service: service, repo: repo}
}

type vsVoiceEmotionService struct {
	*Service
	repo repository.VsVoiceEmotionRepository
}

func (s *vsVoiceEmotionService) Create(ctx context.Context, request *v1.VsVoiceEmotionCreateRequest) error {
	emotion := &model.VsVoiceEmotion{
		VoiceProfileID:    request.VoiceProfileID,
		VoiceAssetID:      request.VoiceAssetID,
		EmotionType:       request.EmotionType,
		EmotionStrength:   request.EmotionStrength,
		ReferenceAudioURL: request.ReferenceAudioURL,
		ReferenceText:     request.ReferenceText,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}
	return s.repo.Create(ctx, emotion)
}

func (s *vsVoiceEmotionService) Update(ctx context.Context, request *v1.VsVoiceEmotionUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("情绪音频不存在")
	}

	existing.EmotionType = request.EmotionType
	existing.EmotionStrength = request.EmotionStrength
	existing.ReferenceAudioURL = request.ReferenceAudioURL
	existing.ReferenceText = request.ReferenceText
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsVoiceEmotionService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsVoiceEmotionService) FindByID(ctx context.Context, id uint64) (*v1.VsVoiceEmotionDetailResponse, error) {
	emotion, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertEmotionToResponse(emotion), nil
}

func (s *vsVoiceEmotionService) FindWithPagination(ctx context.Context, query *v1.VsVoiceEmotionListQuery) ([]*v1.VsVoiceEmotionDetailResponse, int64, error) {
	emotions, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsVoiceEmotionDetailResponse
	for _, e := range emotions {
		responses = append(responses, convertEmotionToResponse(e))
	}
	return responses, total, nil
}

func (s *vsVoiceEmotionService) FindByVoiceProfileID(ctx context.Context, voiceProfileID uint64) ([]*v1.VsVoiceEmotionDetailResponse, error) {
	emotions, err := s.repo.FindByVoiceProfileID(ctx, voiceProfileID)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsVoiceEmotionDetailResponse
	for _, e := range emotions {
		responses = append(responses, convertEmotionToResponse(e))
	}
	return responses, nil
}

func (s *vsVoiceEmotionService) FindByVoiceAssetID(ctx context.Context, voiceAssetID uint64) ([]*v1.VsVoiceEmotionDetailResponse, error) {
	emotions, err := s.repo.FindByVoiceAssetID(ctx, voiceAssetID)
	if err != nil {
		return nil, err
	}

	var responses []*v1.VsVoiceEmotionDetailResponse
	for _, e := range emotions {
		responses = append(responses, convertEmotionToResponse(e))
	}
	return responses, nil
}

func convertEmotionToResponse(e *model.VsVoiceEmotion) *v1.VsVoiceEmotionDetailResponse {
	return &v1.VsVoiceEmotionDetailResponse{
		ID:                e.VoiceEmotionID,
		VoiceProfileID:    e.VoiceProfileID,
		VoiceAssetID:      e.VoiceAssetID,
		EmotionType:       e.EmotionType,
		EmotionStrength:   e.EmotionStrength,
		ReferenceAudioURL: e.ReferenceAudioURL,
		ReferenceText:     e.ReferenceText,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}
