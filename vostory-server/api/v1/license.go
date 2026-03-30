package v1

import "time"

// LicenseActivateOnlineRequest 在线激活请求
type LicenseActivateOnlineRequest struct {
	LicenseCode string `json:"license_code" binding:"required"`
}

// LicenseActivateOfflineRequest 离线激活请求（前端上传 License 文件内容）
type LicenseActivateOfflineRequest struct {
	LicenseFileContent string `json:"license_file_content" binding:"required"`
	PublicKey          string `json:"public_key" binding:"required"`
}

// LicenseStatusResponse 授权状态响应
type LicenseStatusResponse struct {
	Activated   bool       `json:"activated"`
	LicenseCode string     `json:"license_code,omitempty"`
	ProductCode string     `json:"product_code,omitempty"`
	LicenseType string     `json:"license_type,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Features    string     `json:"features,omitempty"`
	Mode        string     `json:"mode,omitempty"`
	Fingerprint string     `json:"fingerprint,omitempty"`
	Hostname    string     `json:"hostname,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}
