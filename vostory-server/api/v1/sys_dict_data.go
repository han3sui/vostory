package v1

import "time"

// SysDictDataListQuery 字典数据列表查询参数
type SysDictDataListQuery struct {
	*BasePageQuery
	DictType  string `json:"dict_type"`  // 字典类型
	DictLabel string `json:"dict_label"` // 字典标签
	Status    string `json:"status"`     // 状态（0正常 1停用）
}

// SysDictDataCreateRequest 创建字典数据请求
type SysDictDataCreateRequest struct {
	DictSort  int    `json:"dict_sort"`                      // 字典排序
	DictLabel string `json:"dict_label" binding:"required"`  // 字典标签
	DictValue string `json:"dict_value" binding:"required"`  // 字典键值
	DictType  string `json:"dict_type" binding:"required"`   // 字典类型
	CSSClass  string `json:"css_class"`                      // 样式属性
	ListClass string `json:"list_class"`                     // 表格回显样式
	IsDefault string `json:"is_default"`                     // 是否默认（Y是 N否）
	Status    string `json:"status" binding:"required"`      // 状态（0正常 1停用）
	Remark    string `json:"remark"`                         // 备注
}

// SysDictDataUpdateRequest 更新字典数据请求
type SysDictDataUpdateRequest struct {
	SysDictDataCreateRequest
	ID uint `json:"id" binding:"required"`
}

// SysDictDataDetailResponse 字典数据详情响应
type SysDictDataDetailResponse struct {
	ID        uint      `json:"id"`
	DictSort  int       `json:"dict_sort"`
	DictLabel string    `json:"dict_label"`
	DictValue string    `json:"dict_value"`
	DictType  string    `json:"dict_type"`
	CSSClass  string    `json:"css_class"`
	ListClass string    `json:"list_class"`
	IsDefault string    `json:"is_default"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
