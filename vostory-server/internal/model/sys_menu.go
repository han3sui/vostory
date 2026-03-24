package model

// SysMenu 菜单权限表
type SysMenu struct {
	BaseModel
	MenuID    uint   `json:"menu_id" gorm:"primaryKey;autoIncrement;comment:菜单ID"`
	MenuName  string `json:"menu_name" gorm:"size:50;not null;comment:菜单名称"`
	ParentID  uint   `json:"parent_id" gorm:"default:0;comment:父菜单ID"`
	OrderNum  int    `json:"order_num" gorm:"default:0;comment:显示顺序"`
	URL       string `json:"url" gorm:"size:200;default:'';comment:请求地址"`
	Target    string `json:"target" gorm:"size:20;default:'';comment:打开方式（menuItem页签 menuBlank新窗口）"`
	MenuType  string `json:"menu_type" gorm:"size:1;default:'';comment:菜单类型（M目录 C菜单 F按钮）"`
	Visible   string `json:"visible" gorm:"size:1;default:'0';comment:菜单状态（0显示 1隐藏）"`
	IsRefresh string `json:"is_refresh" gorm:"size:1;default:'1';comment:是否刷新（0刷新 1不刷新）"`
	Perms     string `json:"perms" gorm:"size:100;comment:权限标识"`
	Icon      string `json:"icon" gorm:"size:100;default:'';comment:菜单图标"`
	Remark    string `json:"remark" gorm:"size:500;default:'';comment:备注"`
}

func (SysMenu) TableName() string {
	return "sys_menu"
}
