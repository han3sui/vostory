package service

import (
	"context"
	"fmt"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/repository"
)

type SysMenuService interface {
	Create(ctx context.Context, req *v1.SysMenuCreateRequest) error
	Update(ctx context.Context, req *v1.SysMenuUpdateRequest) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*v1.SysMenuDetailResponse, error)
	FindWithPagination(ctx context.Context, query *v1.SysMenuListQuery) ([]*v1.SysMenuDetailResponse, int64, error)
	GetMenuTree(ctx context.Context, query *v1.SysMenuListQuery) ([]*v1.SysMenuTreeResponse, error)
	GetMenusByType(ctx context.Context, menuType string) ([]*v1.SysMenuDetailResponse, error)
	CreateMutiByPerms(ctx context.Context, req *v1.SysPermsMenuMutiCreateRequest) error
}

type sysMenuService struct {
	menuRepo repository.SysMenuRepository
	apiRepo  repository.SysApiRepository
}

func NewSysMenuService(menuRepo repository.SysMenuRepository, apiRepo repository.SysApiRepository) SysMenuService {
	return &sysMenuService{
		menuRepo: menuRepo,
		apiRepo:  apiRepo,
	}
}

func (s *sysMenuService) Create(ctx context.Context, req *v1.SysMenuCreateRequest) error {
	menu := &model.SysMenu{
		ParentID:  req.ParentID,
		MenuName:  req.MenuName,
		OrderNum:  req.OrderNum,
		URL:       req.URL,
		Target:    req.Target,
		MenuType:  req.MenuType,
		Visible:   req.Visible,
		IsRefresh: req.IsRefresh,
		Perms:     req.Perms,
		Icon:      req.Icon,
		Remark:    req.Remark,
		BaseModel: model.BaseModel{
			CreatedBy: ctx.Value("login_name").(string),
			DeptID:    ctx.Value("dept_id").(uint),
		},
	}

	return s.menuRepo.Create(ctx, menu)
}

func (s *sysMenuService) Update(ctx context.Context, req *v1.SysMenuUpdateRequest) error {
	// 检查菜单是否存在
	_, err := s.menuRepo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("菜单不存在: %w", err)
	}

	// 构建更新的菜单信息
	menu := &model.SysMenu{
		MenuID:    req.ID,
		ParentID:  req.ParentID,
		MenuName:  req.MenuName,
		OrderNum:  req.OrderNum,
		URL:       req.URL,
		Target:    req.Target,
		MenuType:  req.MenuType,
		Visible:   req.Visible,
		IsRefresh: req.IsRefresh,
		Perms:     req.Perms,
		Icon:      req.Icon,
		Remark:    req.Remark,
		BaseModel: model.BaseModel{
			UpdatedBy: ctx.Value("login_name").(string),
		},
	}

	return s.menuRepo.Update(ctx, menu)
}

func (s *sysMenuService) Delete(ctx context.Context, id uint) error {
	// 检查是否有子菜单
	count, err := s.menuRepo.CountChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("检查子菜单失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("存在子菜单，不能删除")
	}

	return s.menuRepo.Delete(ctx, id)
}

func (s *sysMenuService) FindByID(ctx context.Context, id uint) (*v1.SysMenuDetailResponse, error) {
	menu, err := s.menuRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取子菜单
	children, err := s.getMenuChildren(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.convertToDetailResponse(menu, children), nil
}

func (s *sysMenuService) FindWithPagination(ctx context.Context, query *v1.SysMenuListQuery) ([]*v1.SysMenuDetailResponse, int64, error) {

	// 查询数据
	menus, total, err := s.menuRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	var records []*v1.SysMenuDetailResponse
	for _, menu := range menus {
		records = append(records, s.convertToDetailResponse(menu, nil))
	}

	return records, total, nil
}

func (s *sysMenuService) GetMenuTree(ctx context.Context, query *v1.SysMenuListQuery) ([]*v1.SysMenuTreeResponse, error) {
	// 获取所有显示的菜单
	menus, _, err := s.menuRepo.FindWithPagination(ctx, query)
	if err != nil {
		return nil, err
	}

	// 收集所有按钮类型菜单的权限标识，用于查询对应的API方法
	var buttonPerms []string
	for _, menu := range menus {
		if menu.MenuType == "F" && menu.Perms != "" {
			buttonPerms = append(buttonPerms, menu.Perms)
		}
	}

	// 查询按钮对应的API方法
	permsMethodMap := make(map[string]string)
	if len(buttonPerms) > 0 {
		apiMap, err := s.apiRepo.FindByPerms(ctx, buttonPerms)
		if err == nil {
			for perms, api := range apiMap {
				permsMethodMap[perms] = api.Method
			}
		}
	}

	// 如果有查询条件，找出结果中的所有parentID
	if query.MenuName != "" || query.Visible != "" || query.MenuType != "" {
		// 找出所有可能的根节点
		parentIDs := make(map[uint]bool)
		menuIDs := make(map[uint]bool)

		for _, menu := range menus {
			parentIDs[menu.ParentID] = true
			menuIDs[menu.MenuID] = true
		}

		// 构建结果
		var result []*v1.SysMenuTreeResponse
		for parentID := range parentIDs {
			// 如果这个parentID不在菜单ID列表中，它就是树的顶层
			if !menuIDs[parentID] || parentID == 0 {
				result = append(result, s.buildMenuTree(menus, parentID, permsMethodMap)...)
			}
		}
		return result, nil
	}

	// 构建树形结构
	return s.buildMenuTree(menus, 0, permsMethodMap), nil
}

func (s *sysMenuService) GetMenusByType(ctx context.Context, menuType string) ([]*v1.SysMenuDetailResponse, error) {
	menus, err := s.menuRepo.FindByType(ctx, menuType)
	if err != nil {
		return nil, err
	}

	var records []*v1.SysMenuDetailResponse
	for _, menu := range menus {
		records = append(records, s.convertToDetailResponse(menu, nil))
	}

	return records, nil
}

// 辅助方法
func (s *sysMenuService) getMenuChildren(ctx context.Context, parentID uint) ([]*v1.SysMenuDetailResponse, error) {
	children, err := s.menuRepo.FindChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	var result []*v1.SysMenuDetailResponse
	for _, child := range children {
		// 递归获取子菜单的子菜单
		grandChildren, err := s.getMenuChildren(ctx, child.MenuID)
		if err != nil {
			return nil, err
		}
		result = append(result, s.convertToDetailResponse(child, grandChildren))
	}

	return result, nil
}

func (s *sysMenuService) buildMenuTree(menus []*model.SysMenu, parentID uint, permsMethodMap map[string]string) []*v1.SysMenuTreeResponse {
	var tree []*v1.SysMenuTreeResponse

	for _, menu := range menus {
		if menu.ParentID == parentID {
			node := &v1.SysMenuTreeResponse{
				ID:       menu.MenuID,
				ParentID: menu.ParentID,
				MenuName: menu.MenuName,
				OrderNum: menu.OrderNum,
				URL:      menu.URL,
				MenuType: menu.MenuType,
				Visible:  menu.Visible,
				Perms:    menu.Perms,
				Icon:     menu.Icon,
				Children: s.buildMenuTree(menus, menu.MenuID, permsMethodMap),
			}
			// 如果是按钮类型，添加对应的请求方法
			if menu.MenuType == "F" && menu.Perms != "" {
				if method, ok := permsMethodMap[menu.Perms]; ok {
					node.Method = method
				}
			}
			tree = append(tree, node)
		}
	}

	return tree
}

func (s *sysMenuService) convertToDetailResponse(menu *model.SysMenu, children []*v1.SysMenuDetailResponse) *v1.SysMenuDetailResponse {
	return &v1.SysMenuDetailResponse{
		ID:        menu.MenuID,
		ParentID:  menu.ParentID,
		MenuName:  menu.MenuName,
		OrderNum:  menu.OrderNum,
		URL:       menu.URL,
		Target:    menu.Target,
		MenuType:  menu.MenuType,
		Visible:   menu.Visible,
		IsRefresh: menu.IsRefresh,
		Perms:     menu.Perms,
		Icon:      menu.Icon,
		Remark:    menu.Remark,
		CreatedAt: menu.CreatedAt,
		UpdatedAt: menu.UpdatedAt,
		Children:  children,
	}
}

func (s *sysMenuService) CreateMutiByPerms(ctx context.Context, req *v1.SysPermsMenuMutiCreateRequest) error {
	menus := make([]*model.SysMenu, len(*req))
	for i, menu := range *req {
		menus[i] = &model.SysMenu{
			ParentID:  menu.ParentID,
			MenuName:  menu.MenuName,
			Perms:     menu.Perms,
			MenuType:  "F",
			Visible:   "0",
			IsRefresh: "0",
			Icon:      "",
			Remark:    "",
		}
	}
	return s.menuRepo.CreateMutiByPerms(ctx, menus)
}
