package model

// SysUserPost 用户与岗位关联表
type SysUserPost struct {
	UserID uint `json:"user_id" gorm:"primaryKey;comment:用户ID"`
	PostID uint `json:"post_id" gorm:"primaryKey;comment:岗位ID"`
}

func (SysUserPost) TableName() string {
	return "sys_user_post"
}
