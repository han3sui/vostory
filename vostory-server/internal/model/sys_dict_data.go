package model

// SysDictData 字典数据表
type SysDictData struct {
	BaseModel
	DictCode  uint   `json:"dict_code" gorm:"primaryKey;autoIncrement;comment:字典编码"`
	DictSort  int    `json:"dict_sort" gorm:"default:0;comment:字典排序"`
	DictLabel string `json:"dict_label" gorm:"size:100;default:'';comment:字典标签"`
	DictValue string `json:"dict_value" gorm:"size:100;default:'';comment:字典键值"`
	DictType  string `json:"dict_type" gorm:"size:100;default:'';comment:字典类型"`
	CSSClass  string `json:"css_class" gorm:"size:100;comment:样式属性（其他样式扩展）"`
	ListClass string `json:"list_class" gorm:"size:100;comment:表格回显样式"`
	IsDefault string `json:"is_default" gorm:"size:1;default:'N';comment:是否默认（Y是 N否）"`
	Status    string `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
	Remark    string `json:"remark" gorm:"size:500;comment:备注"`
}

func (SysDictData) TableName() string {
	return "sys_dict_data"
}
