package model

// SysNotice 通知公告表
type SysNotice struct {
	BaseModel
	NoticeID      int    `json:"notice_id" gorm:"primaryKey;autoIncrement;comment:公告ID"`
	NoticeTitle   string `json:"notice_title" gorm:"size:50;not null;comment:公告标题"`
	NoticeType    string `json:"notice_type" gorm:"size:1;not null;comment:公告类型（1通知 2公告）"`
	NoticeContent []byte `json:"notice_content" gorm:"type:longblob;comment:公告内容"`
	Status        string `json:"status" gorm:"size:1;default:'0';comment:公告状态（0正常 1关闭）"`
	Remark        string `json:"remark" gorm:"size:255;comment:备注"`
}

func (SysNotice) TableName() string {
	return "sys_notice"
}
