package model

// SysPost 岗位信息表
type SysPost struct {
	BaseModel
	PostID   uint   `json:"post_id" gorm:"primaryKey;autoIncrement;comment:岗位ID"`
	PostCode string `json:"post_code" gorm:"size:64;not null;comment:岗位编码"`
	PostName string `json:"post_name" gorm:"size:50;not null;comment:岗位名称"`
	PostSort int    `json:"post_sort" gorm:"default:0;comment:显示顺序"`
	Status   string `json:"status" gorm:"size:1;not null;comment:状态（0正常 1停用）"`
	Remark   string `json:"remark" gorm:"size:500;comment:备注"`
}

func (SysPost) TableName() string {
	return "sys_post"
}
