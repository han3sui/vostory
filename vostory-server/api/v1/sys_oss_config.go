package v1

// OSSConfigResponse 对象存储配置（SK 已脱敏）
type OSSConfigResponse struct {
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	AK       string `json:"ak"`
	SK       string `json:"sk"`
	Bucket   string `json:"bucket"`
	Prefix   string `json:"prefix"`
	UseSSL   string `json:"use_ssl"`
	CdnBase  string `json:"cdn_base"`
}

// OSSConfigUpdateRequest 保存 / 测试对象存储配置
type OSSConfigUpdateRequest struct {
	Type     string `json:"type" binding:"required"`
	Endpoint string `json:"endpoint" binding:"required"`
	AK       string `json:"ak" binding:"required"`
	SK       string `json:"sk" binding:"required"`
	Bucket   string `json:"bucket" binding:"required"`
	Prefix   string `json:"prefix"`
	UseSSL   string `json:"use_ssl"`
	CdnBase  string `json:"cdn_base"`
}
