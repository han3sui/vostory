package model

// SysDictType 字典类型表
type SysDictType struct {
	BaseModel
	DictID   uint   `json:"dict_id" gorm:"primaryKey;autoIncrement;comment:字典主键"`
	DictName string `json:"dict_name" gorm:"size:100;default:'';comment:字典名称"`
	DictType string `json:"dict_type" gorm:"size:100;default:'';uniqueIndex;comment:字典类型"`
	Status   string `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
	Remark   string `json:"remark" gorm:"size:500;comment:备注"`
}

func (SysDictType) TableName() string {
	return "sys_dict_type"
}
