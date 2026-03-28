package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type VsScriptSegmentService interface {
	Create(ctx context.Context, request *v1.VsScriptSegmentCreateRequest) error
	Update(ctx context.Context, request *v1.VsScriptSegmentUpdateRequest) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*v1.VsScriptSegmentDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.VsScriptSegmentListQuery) ([]*v1.VsScriptSegmentDetailResponse, int64, error)
	FindByChapterID(ctx context.Context, chapterID uint64) ([]*v1.VsScriptSegmentDetailResponse, error)
}

func NewVsScriptSegmentService(
	service *Service,
	repo repository.VsScriptSegmentRepository,
	audioClipRepo repository.VsAudioClipRepository,
) VsScriptSegmentService {
	return &vsScriptSegmentService{Service: service, repo: repo, audioClipRepo: audioClipRepo}
}

type vsScriptSegmentService struct {
	*Service
	repo          repository.VsScriptSegmentRepository
	audioClipRepo repository.VsAudioClipRepository
}

func (s *vsScriptSegmentService) Create(ctx context.Context, request *v1.VsScriptSegmentCreateRequest) error {
	segment := &model.VsScriptSegment{
		ProjectID:       request.ProjectID,
		SceneID:         request.SceneID,
		ChapterID:       request.ChapterID,
		SegmentNum:      request.SegmentNum,
		SegmentType:     request.SegmentType,
		Content:         request.Content,
		OriginalContent: request.OriginalContent,
		CharacterID:     request.CharacterID,
		EmotionType:     request.EmotionType,
		EmotionStrength: request.EmotionStrength,
		Status:          "raw",
		Version:         1,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.repo.Create(ctx, segment)
}

func (s *vsScriptSegmentService) Update(ctx context.Context, request *v1.VsScriptSegmentUpdateRequest) error {
	existing, err := s.repo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("片段不存在")
	}

	if request.SegmentNum > 0 {
		existing.SegmentNum = request.SegmentNum
	}
	if request.SegmentType != "" {
		existing.SegmentType = request.SegmentType
	}
	if request.Content != "" {
		existing.Content = request.Content
		existing.Version++
	}
	existing.CharacterID = request.CharacterID
	if request.EmotionType != "" {
		existing.EmotionType = request.EmotionType
	}
	if request.EmotionStrength != "" {
		existing.EmotionStrength = request.EmotionStrength
	}
	if request.Status != "" {
		existing.Status = request.Status
	}
	existing.UpdatedBy = ctx.Value("login_name").(string)

	return s.repo.Update(ctx, existing)
}

func (s *vsScriptSegmentService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *vsScriptSegmentService) FindByID(ctx context.Context, id uint64) (*v1.VsScriptSegmentDetailResponse, error) {
	segment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(segment), nil
}

func (s *vsScriptSegmentService) FindWithPagination(ctx context.Context, query *v1.VsScriptSegmentListQuery) ([]*v1.VsScriptSegmentDetailResponse, int64, error) {
	segments, total, err := s.repo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.VsScriptSegmentDetailResponse
	for _, seg := range segments {
		responses = append(responses, s.convertToDetailResponse(seg))
	}
	return responses, total, nil
}

func (s *vsScriptSegmentService) FindByChapterID(ctx context.Context, chapterID uint64) ([]*v1.VsScriptSegmentDetailResponse, error) {
	segments, err := s.repo.FindByChapterID(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	segmentIDs := make([]uint64, 0, len(segments))
	for _, seg := range segments {
		segmentIDs = append(segmentIDs, seg.SegmentID)
	}

	audioMap, _ := s.audioClipRepo.FindCurrentBySegmentIDs(ctx, segmentIDs)

	var responses []*v1.VsScriptSegmentDetailResponse
	for _, seg := range segments {
		resp := s.convertToDetailResponse(seg)
		if clip, ok := audioMap[seg.SegmentID]; ok {
			resp.HasAudio = true
			resp.AudioURL = clip.AudioURL
			resp.ClipID = &clip.ClipID
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *vsScriptSegmentService) convertToDetailResponse(seg *model.VsScriptSegment) *v1.VsScriptSegmentDetailResponse {
	resp := &v1.VsScriptSegmentDetailResponse{
		ID:              seg.SegmentID,
		ProjectID:       seg.ProjectID,
		SceneID:         seg.SceneID,
		ChapterID:       seg.ChapterID,
		SegmentNum:      seg.SegmentNum,
		SegmentType:     seg.SegmentType,
		Content:         seg.Content,
		OriginalContent: seg.OriginalContent,
		CharacterID:     seg.CharacterID,
		EmotionType:     seg.EmotionType,
		EmotionStrength: seg.EmotionStrength,
		Status:          seg.Status,
		ErrorMessage:    seg.ErrorMessage,
		Version:         seg.Version,
		CreatedAt:       seg.CreatedAt,
		UpdatedAt:       seg.UpdatedAt,
	}
	if seg.Character != nil {
		resp.CharacterName = seg.Character.Name
	}
	return resp
}
