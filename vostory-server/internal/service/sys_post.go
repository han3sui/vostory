package service

import (
	"context"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysPostService interface {
	Create(ctx context.Context, request *v1.SysPostCreateRequest) error
	Update(ctx context.Context, request *v1.SysPostUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysPostDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysPostListQuery) ([]*v1.SysPostDetailResponse, int64, error)

	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

func NewSysPostService(
	service *Service,
	sysPostRepository repository.SysPostRepository,
) SysPostService {
	return &sysPostService{
		Service:           service,
		sysPostRepository: sysPostRepository,
	}
}

type sysPostService struct {
	*Service
	sysPostRepository repository.SysPostRepository
}

// Create 创建岗位
func (s *sysPostService) Create(ctx context.Context, request *v1.SysPostCreateRequest) error {
	post := &model.SysPost{
		PostCode: request.PostCode,
		PostName: request.PostName,
		PostSort: request.PostSort,
		Status:   request.Status,
		Remark:   request.Remark,
		BaseModel: model.BaseModel{
			DeptID:    ctx.Value("dept_id").(uint),
			CreatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.sysPostRepository.Create(ctx, post)
}

// Update 更新岗位
func (s *sysPostService) Update(ctx context.Context, request *v1.SysPostUpdateRequest) error {
	post := &model.SysPost{
		PostID:   request.ID,
		PostCode: request.PostCode,
		PostName: request.PostName,
		PostSort: request.PostSort,
		Status:   request.Status,
		Remark:   request.Remark,
		BaseModel: model.BaseModel{
			UpdatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.sysPostRepository.Update(ctx, post)
}

// Delete 删除岗位
func (s *sysPostService) Delete(ctx context.Context, id uint) error {
	return s.sysPostRepository.Delete(ctx, id)
}

// FindByID 根据ID查找岗位
func (s *sysPostService) FindByID(ctx context.Context, id uint) (*v1.SysPostDetailResponse, error) {
	post, err := s.sysPostRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertToDetailResponse(post), nil
}

// FindWithPagination 分页查询岗位
func (s *sysPostService) FindWithPagination(ctx context.Context, query *v1.SysPostListQuery) ([]*v1.SysPostDetailResponse, int64, error) {
	posts, total, err := s.sysPostRepository.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var responses []*v1.SysPostDetailResponse
	for _, post := range posts {
		responses = append(responses, s.convertToDetailResponse(post))
	}

	return responses, total, nil
}

// convertToDetailResponse 转换为详情响应
func (s *sysPostService) convertToDetailResponse(post *model.SysPost) *v1.SysPostDetailResponse {
	return &v1.SysPostDetailResponse{
		ID:        post.PostID,
		PostCode:  post.PostCode,
		PostName:  post.PostName,
		PostSort:  post.PostSort,
		Status:    post.Status,
		Remark:    post.Remark,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}

func (s *sysPostService) Enable(ctx context.Context, id uint) error {
	return s.sysPostRepository.Enable(ctx, id)
}

func (s *sysPostService) Disable(ctx context.Context, id uint) error {
	return s.sysPostRepository.Disable(ctx, id)
}
