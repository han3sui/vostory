package handler

import (
	"net/http"
	"strconv"

	v1 "iot-alert-center/api/v1"
	"iot-alert-center/internal/model"
	"iot-alert-center/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type SysDeptHandler struct {
	deptService service.SysDeptService
	db          *gorm.DB
}

func NewSysDeptHandler(deptService service.SysDeptService, db *gorm.DB) *SysDeptHandler {
	return &SysDeptHandler{
		deptService: deptService,
		db:          db,
	}
}

// CreateDept godoc
// @Summary 创建部门
// @Description 创建新的部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param request body v1.SysDeptCreateRequest true "创建部门请求"
// @Success 200 {object} v1.Response[any]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept [post]
// @Id system:dept:add
func (h *SysDeptHandler) CreateDept(c *gin.Context) {
	var req v1.SysDeptCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, err.Error())
		return
	}

	if err := h.deptService.Create(c, &req); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// UpdateDept godoc
// @Summary 更新部门
// @Description 更新部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param id path int true "部门ID"
// @Param request body v1.SysDeptUpdateRequest true "更新部门请求"
// @Success 200 {object} v1.Response[any]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/{id} [put]
// @Id system:dept:edit
// UpdateDept 更新部门
func (h *SysDeptHandler) UpdateDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid department ID")
		return
	}

	var req v1.SysDeptUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, err.Error())
		return
	}

	req.ID = uint(id)

	if err := h.deptService.Update(c, &req); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// DeleteDept godoc
// @Summary 删除部门
// @Description 删除指定部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} v1.Response[any]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/{id} [delete]
// @Id system:dept:remove
// DeleteDept 删除部门
func (h *SysDeptHandler) DeleteDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid department ID")
		return
	}

	if err := h.deptService.Delete(c, uint(id)); err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// GetDept godoc
// @Summary 获取部门详情
// @Description 根据ID获取部门详情
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} v1.Response[v1.SysDeptDetailResponse]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/{id} [get]
// @Id system:dept:detail
// GetDept 获取部门详情
func (h *SysDeptHandler) GetDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid department ID")
		return
	}

	dept, err := h.deptService.FindByID(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, dept)
}

// ListDepts godoc
// @Summary 获取部门列表
// @Description 分页获取部门列表
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param page query int false "当前页"
// @Param size query int false "每页数量"
// @Param dept_name query string false "部门名称"
// @Param status query string false "状态"
// @Success 200 {object} v1.Response[v1.PageResponse[v1.SysDeptDetailResponse]]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/list [get]
// @Id system:dept:list
// ListDepts 获取部门列表（分页）
func (h *SysDeptHandler) ListDepts(c *gin.Context) {

	query := &v1.SysDeptListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}

	page := c.Query("page")
	size := c.Query("size")

	query.Page = cast.ToInt(page)
	query.Size = cast.ToInt(size)

	query.DeptName = c.Query("dept_name")

	result, total, err := h.deptService.FindWithPagination(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, v1.NewPageResponse(query.Page, query.Size, total, result))
}

// GetDeptTree godoc
// @Summary 获取部门树
// @Description 获取部门树形结构
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param dept_name query string false "部门名称"
// @Param status query string false "状态"
// @Success 200 {object} v1.Response[[]v1.SysDeptTreeResponse]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/tree [get]
// @Id system:dept:tree
// GetDeptTree 获取部门树
func (h *SysDeptHandler) GetDeptTree(c *gin.Context) {
	query := &v1.SysDeptListQuery{}
	query.BasePageQuery = &v1.BasePageQuery{}
	query.DeptName = c.Query("dept_name")
	query.Status = c.Query("status")

	tree, err := h.deptService.GetDeptTree(c, query)
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	if len(tree) > 0 {
		v1.HandleSuccess(c, tree)
	} else {
		v1.HandleSuccess(c, []map[string]interface{}{})
	}
}

// Enable godoc
// @Summary 启用部门
// @Description 启用指定部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} v1.Response[any]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/{id}/enable [put]
// @Id system:dept:enable
// Enable 启用部门
func (h *SysDeptHandler) Enable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid department ID")
		return
	}

	err = h.deptService.Enable(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// Disable godoc
// @Summary 禁用部门
// @Description 禁用指定部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} v1.Response[any]
// @Failure 400 {object} v1.Response[any]
// @Failure 500 {object} v1.Response[any]
// @Router /api/v1/system/dept/{id}/disable [put]
// @Id system:dept:disable
// Disable 禁用部门
func (h *SysDeptHandler) Disable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		v1.HandleError(c, http.StatusBadRequest, v1.ErrInvalidParams, "Invalid department ID")
		return
	}

	err = h.deptService.Disable(c, uint(id))
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	v1.HandleSuccess(c, nil)
}

// GetDeptOptionsTree godoc
// @Summary      获取部门选项树
// @Description  获取简化的部门树形结构，用于下拉选择（指定部门等场景）
// @Tags         通用接口
// @Accept       json
// @Produce      json
// @Success      200      {object}  v1.Response[[]v1.DeptOptionTreeResponse]
// @Failure      500      {object}  v1.Response[any]
// @Router       /api/v1/common/dept/options [get]
// @Whitelist    true
func (h *SysDeptHandler) GetDeptOptionsTree(c *gin.Context) {
	// 查询所有启用的部门
	var depts []model.SysDept
	err := h.db.WithContext(c).
		Select("dept_id, dept_name, parent_id").
		Where("status = ?", "0").
		Order("order_num ASC").
		Find(&depts).Error
	if err != nil {
		v1.HandleError(c, http.StatusInternalServerError, v1.NewError(500, "获取部门列表失败"), err)
		return
	}

	// 构建树形结构
	tree := h.buildDeptOptionTree(depts, 0)
	v1.HandleSuccess(c, tree)
}

// buildDeptOptionTree 构建部门选项树
func (h *SysDeptHandler) buildDeptOptionTree(depts []model.SysDept, parentID uint) []*v1.DeptOptionTreeResponse {
	var tree []*v1.DeptOptionTreeResponse
	for _, dept := range depts {
		if dept.ParentID == parentID {
			node := &v1.DeptOptionTreeResponse{
				DeptID:   dept.DeptID,
				DeptName: dept.DeptName,
				Children: h.buildDeptOptionTree(depts, dept.DeptID),
			}
			if len(node.Children) == 0 {
				node.Children = nil
			}
			tree = append(tree, node)
		}
	}
	return tree
}
