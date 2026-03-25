package model

// VsWorkspace 工作空间
type VsWorkspace struct {
	BaseModel
	WorkspaceID uint64   `json:"workspace_id" gorm:"primaryKey;comment:工作空间ID"`
	Name        string   `json:"name" gorm:"size:100;not null;comment:空间名称"`
	Description string   `json:"description" gorm:"size:500;comment:描述"`
	OwnerID     uint     `json:"owner_id" gorm:"comment:创建者用户ID"`
	Status      string   `json:"status" gorm:"size:1;default:'0';comment:状态（0正常 1停用）"`
	Owner       *SysUser `json:"owner" gorm:"foreignKey:OwnerID;references:UserID"`
}

func (VsWorkspace) TableName() string {
	return "vs_workspace"
}
