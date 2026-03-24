package service

import (
	"context"
	"fmt"
	"strconv"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysDeptService interface {
	Create(ctx context.Context, req *v1.SysDeptCreateRequest) error
	Update(ctx context.Context, req *v1.SysDeptUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysDeptDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysDeptListQuery) ([]*v1.SysDeptDetailResponse, int64, error)
	GetDeptTree(ctx context.Context, query *v1.SysDeptListQuery) ([]*v1.SysDeptTreeResponse, error)
	Enable(ctx context.Context, id uint) error
	Disable(ctx context.Context, id uint) error
}

type sysDeptService struct {
	deptRepo repository.SysDeptRepository
}

func NewSysDeptService(deptRepo repository.SysDeptRepository) SysDeptService {
	return &sysDeptService{
		deptRepo: deptRepo,
	}
}

func (s *sysDeptService) Create(ctx context.Context, req *v1.SysDeptCreateRequest) error {
	// 构建祖级列表
	ancestors, err := s.buildAncestors(ctx, *req.ParentID)
	if err != nil {
		return fmt.Errorf("构建祖级列表失败: %w", err)
	}

	dept := &model.SysDept{
		ParentID:  *req.ParentID,
		Ancestors: ancestors,
		DeptName:  req.DeptName,
		OrderNum:  req.OrderNum,
		LeaderID:  req.LeaderID,
		Leader:    req.Leader,
		Phone:     req.Phone,
		Email:     req.Email,
		Status:    req.Status,
		Remark:    req.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.deptRepo.Create(ctx, dept)
}

func (s *sysDeptService) Update(ctx context.Context, req *v1.SysDeptUpdateRequest) error {
	// 检查部门是否存在
	oldDept, err := s.deptRepo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("部门不存在: %w", err)
	}

	// 构建新的祖级列表
	ancestors, err := s.buildAncestors(ctx, *req.ParentID)
	if err != nil {
		return fmt.Errorf("构建祖级列表失败: %w", err)
	}

	// 如果父部门发生变化，需要更新所有子部门的祖级列表
	if oldDept.ParentID != *req.ParentID {
		if err := s.updateChildrenAncestors(ctx, req.ID, ancestors); err != nil {
			return fmt.Errorf("更新子部门祖级列表失败: %w", err)
		}
	}

	// 构建更新的部门信息
	var leaderID *uint
	var leader string
	if req.LeaderID != nil && *req.LeaderID != 0 {
		leaderID = req.LeaderID
		leader = req.Leader
	}

	dept := &model.SysDept{
		DeptID:    req.ID,
		ParentID:  *req.ParentID,
		Ancestors: ancestors,
		DeptName:  req.DeptName,
		OrderNum:  req.OrderNum,
		LeaderID:  leaderID,
		Leader:    leader,
		Phone:     req.Phone,
		Email:     req.Email,
		Status:    req.Status,
		Remark:    req.Remark,
		BaseModel: model.BaseModel{
			UpdatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.deptRepo.Update(ctx, dept)
}

func (s *sysDeptService) Delete(ctx context.Context, id uint) error {
	// 检查是否有子部门
	count, err := s.deptRepo.CountChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("检查子部门失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("存在子部门，不能删除")
	}

	return s.deptRepo.Delete(ctx, id)
}

func (s *sysDeptService) FindByID(ctx context.Context, id uint) (*v1.SysDeptDetailResponse, error) {
	dept, err := s.deptRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取子部门
	children, err := s.getDeptChildren(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertToDetailResponse(dept, children), nil
}

func (s *sysDeptService) FindWithPagination(ctx context.Context, query *v1.SysDeptListQuery) ([]*v1.SysDeptDetailResponse, int64, error) {
	// 构建查询参数

	// 查询数据
	depts, total, err := s.deptRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var records []*v1.SysDeptDetailResponse
	for _, dept := range depts {
		records = append(records, s.convertToDetailResponse(dept, nil))
	}

	return records, total, nil
}

func (s *sysDeptService) GetDeptTree(ctx context.Context, query *v1.SysDeptListQuery) ([]*v1.SysDeptTreeResponse, error) {
	// 获取所有部门
	depts, _, err := s.deptRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, err
	}

	// 如果有查询条件，找出结果中的所有parentID
	if query.DeptName != "" || query.Status != "" || query.ParentID != nil {
		// 找出所有可能的根节点
		parentIDs := make(map[uint]bool)
		deptIDs := make(map[uint]bool)

		for _, dept := range depts {
			parentIDs[dept.ParentID] = true
			deptIDs[dept.DeptID] = true
		}

		// 构建结果
		var result []*v1.SysDeptTreeResponse
		for parentID := range parentIDs {
			// 如果这个parentID不在菜单ID列表中，它就是树的顶层
			if !deptIDs[parentID] || parentID == 0 {
				result = append(result, s.buildDeptTree(depts, parentID)...)
			}
		}
		return result, nil
	}

	// 构建树形结构
	return s.buildDeptTree(depts, 0), nil
}

// 辅助方法
func (s *sysDeptService) buildAncestors(ctx context.Context, parentID uint) (string, error) {
	if parentID == 0 {
		return "0", nil
	}

	parent, err := s.deptRepo.FindByID(ctx, parentID)
	if err != nil {
		return "", err
	}

	return parent.Ancestors + "," + strconv.FormatUint(uint64(parentID), 10), nil
}

func (s *sysDeptService) updateChildrenAncestors(ctx context.Context, deptID uint, newAncestors string) error {
	// 获取所有子部门
	children, err := s.deptRepo.FindChildren(ctx, deptID)
	if err != nil {
		return err
	}

	for _, child := range children {
		// 更新子部门的祖级列表
		child.Ancestors = newAncestors + "," + strconv.FormatUint(uint64(deptID), 10)
		if err := s.deptRepo.Update(ctx, child); err != nil {
			return err
		}

		// 递归更新子部门的子部门
		if err := s.updateChildrenAncestors(ctx, child.DeptID, child.Ancestors); err != nil {
			return err
		}
	}

	return nil
}

func (s *sysDeptService) getDeptChildren(ctx context.Context, parentID uint) ([]*v1.SysDeptDetailResponse, error) {
	children, err := s.deptRepo.FindChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	var result []*v1.SysDeptDetailResponse
	for _, child := range children {
		// 递归获取子部门的子部门
		grandChildren, err := s.getDeptChildren(ctx, child.DeptID)
		if err != nil {
			return nil, err
		}
		result = append(result, s.convertToDetailResponse(child, grandChildren))
	}

	return result, nil
}

func (s *sysDeptService) buildDeptTree(depts []*model.SysDept, parentID uint) []*v1.SysDeptTreeResponse {
	var tree []*v1.SysDeptTreeResponse

	for _, dept := range depts {
		if dept.ParentID == parentID {
			node := &v1.SysDeptTreeResponse{
				ID:       dept.DeptID,
				ParentID: dept.ParentID,
				DeptName: dept.DeptName,
				OrderNum: dept.OrderNum,
				LeaderID: dept.LeaderID,
				Leader:   dept.Leader,
				Status:   dept.Status,
				Children: s.buildDeptTree(depts, dept.DeptID),
			}
			// 添加负责人详情
			if dept.LeaderUser != nil {
				node.LeaderUser = &v1.SysUserBriefResponse{
					UserID:   dept.LeaderUser.UserID,
					UserName: dept.LeaderUser.UserName,
					Avatar:   dept.LeaderUser.Avatar,
				}
			}
			tree = append(tree, node)
		}
	}

	return tree
}

func (s *sysDeptService) convertToDetailResponse(dept *model.SysDept, children []*v1.SysDeptDetailResponse) *v1.SysDeptDetailResponse {
	resp := &v1.SysDeptDetailResponse{
		ID:        dept.DeptID,
		ParentID:  dept.ParentID,
		Ancestors: dept.Ancestors,
		DeptName:  dept.DeptName,
		OrderNum:  dept.OrderNum,
		LeaderID:  dept.LeaderID,
		Leader:    dept.Leader,
		Phone:     dept.Phone,
		Email:     dept.Email,
		Status:    dept.Status,
		Remark:    dept.Remark,
		CreatedAt: dept.CreatedAt,
		UpdatedAt: dept.UpdatedAt,
		Children:  children,
	}
	// 添加负责人详情
	if dept.LeaderUser != nil {
		resp.LeaderUser = &v1.SysUserBriefResponse{
			UserID:   dept.LeaderUser.UserID,
			UserName: dept.LeaderUser.UserName,
			Avatar:   dept.LeaderUser.Avatar,
		}
	}
	return resp
}

func (s *sysDeptService) Enable(ctx context.Context, id uint) error {
	return s.deptRepo.Enable(ctx, id)
}

func (s *sysDeptService) Disable(ctx context.Context, id uint) error {
	return s.deptRepo.Disable(ctx, id)
}
