package model

// SysConfig 参数配置表
type SysConfig struct {
	BaseModel
	ConfigID    int    `json:"config_id" gorm:"primaryKey;autoIncrement;comment:参数主键"`
	ConfigName  string `json:"config_name" gorm:"size:100;default:'';comment:参数名称"`
	ConfigKey   string `json:"config_key" gorm:"size:100;default:'';comment:参数键名"`
	ConfigValue string `json:"config_value" gorm:"size:500;default:'';comment:参数键值"`
	ConfigType  string `json:"config_type" gorm:"size:1;default:'N';comment:系统内置（Y是 N否）"`
	Remark      string `json:"remark" gorm:"size:500;comment:备注"`
}

func (SysConfig) TableName() string {
	return "sys_config"
}
