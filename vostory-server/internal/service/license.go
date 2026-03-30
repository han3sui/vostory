package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	v1 "iot-alert-center/api/v1"
	"iot-alert-center/pkg/log"
	"os"
	"path/filepath"
	"sync"
	"time"

	license "github.com/license-core/license-sdk"
	"github.com/spf13/viper"
)

type LicenseService interface {
	GetStatus() *v1.LicenseStatusResponse
	ActivateOnline(licenseCode string) error
	ActivateOffline(fileContent string) error
	Deactivate() error
	IsActivated() bool
}

var defaultLicenseFile = filepath.Join("storage", "license.json")

type licenseService struct {
	logger      *log.Logger
	conf        *viper.Viper
	client      *license.Client
	mu          sync.RWMutex
	activated   bool
	lastResult  *license.VerifyResult
	licenseCode string
}

func NewLicenseService(logger *log.Logger, conf *viper.Viper) LicenseService {
	svc := &licenseService{
		logger: logger,
		conf:   conf,
	}

	svc.tryAutoRestore()

	return svc
}

// tryAutoRestore 尝试从本地存储的 License 文件自动恢复激活状态
func (s *licenseService) tryAutoRestore() {
	data, err := os.ReadFile(defaultLicenseFile)
	if err != nil {
		return
	}

	var stored struct {
		LicenseCode string `json:"license_code"`
		ServerURL   string `json:"server_url"`
		PublicKey   string `json:"public_key"`
		Mode        string `json:"mode"`
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		s.tryRestoreOffline(defaultLicenseFile, data)
		return
	}

	if stored.LicenseCode != "" && stored.Mode == "online" {
		s.licenseCode = stored.LicenseCode
		serverURL := stored.ServerURL
		if serverURL == "" {
			serverURL = s.conf.GetString("license.server_url")
		}
		client := license.NewClient(license.Config{
			LicenseCode: stored.LicenseCode,
			ServerURL:   serverURL,
		})
		result, err := client.VerifyOnline()
		if err == nil && result.Valid {
			s.mu.Lock()
			s.client = client
			s.activated = true
			s.lastResult = result
			s.mu.Unlock()
			client.StartHeartbeat(5 * time.Minute)
			s.logger.Info("License auto-restored (online)")
			return
		}
	}

	if stored.PublicKey != "" {
		s.tryRestoreOffline(defaultLicenseFile, data)
	}
}

func (s *licenseService) tryRestoreOffline(filePath string, data []byte) {
	var stored struct {
		PublicKey string `json:"public_key"`
	}
	_ = json.Unmarshal(data, &stored)
	if stored.PublicKey == "" {
		return
	}

	client := license.NewClient(license.Config{
		LicenseFile: filePath,
		PublicKey:   stored.PublicKey,
		ServerURL:   s.conf.GetString("license.server_url"),
	})
	result, err := client.VerifyOffline()
	if err == nil && result.Valid {
		s.mu.Lock()
		s.client = client
		s.activated = true
		s.lastResult = result
		s.licenseCode = result.LicenseCode
		s.mu.Unlock()
		s.logger.Info("License auto-restored (offline)")
	}
}

func (s *licenseService) GetStatus() *v1.LicenseStatusResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := &v1.LicenseStatusResponse{
		Activated: s.activated,
	}

	if s.client != nil {
		resp.Fingerprint = s.client.GetFingerprint()
		resp.Hostname = s.client.GetHostname()
	} else {
		resp.Fingerprint = license.CollectFingerprint()
		hostname, _ := os.Hostname()
		resp.Hostname = hostname
	}

	if s.lastResult != nil {
		resp.LicenseCode = s.lastResult.LicenseCode
		resp.ProductCode = s.lastResult.ProductCode
		resp.LicenseType = s.lastResult.LicenseType
		resp.ExpiresAt = s.lastResult.ExpiresAt
		resp.Features = s.lastResult.Features
		resp.Mode = s.lastResult.Mode
	}

	return resp
}

func (s *licenseService) ActivateOnline(licenseCode string) error {
	serverURL := s.conf.GetString("license.server_url")
	if serverURL == "" {
		return fmt.Errorf("未配置授权服务地址")
	}

	client := license.NewClient(license.Config{
		LicenseCode: licenseCode,
		ServerURL:   serverURL,
	})

	if err := client.Activate(); err != nil {
		return fmt.Errorf("激活失败: %w", err)
	}

	result, err := client.VerifyOnline()
	if err != nil {
		return fmt.Errorf("验证失败: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("授权无效: %s", result.Reason)
	}

	s.mu.Lock()
	if s.client != nil {
		s.client.StopHeartbeat()
	}
	s.client = client
	s.activated = true
	s.lastResult = result
	s.licenseCode = licenseCode
	s.mu.Unlock()

	client.StartHeartbeat(5 * time.Minute)

	s.persistState("online", licenseCode, "", "")

	return nil
}

func (s *licenseService) ActivateOffline(fileContent string) error {
	if err := os.MkdirAll("./storage", 0755); err != nil {
		return fmt.Errorf("创建存储目录失败: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		return fmt.Errorf("License 文件解码失败，请确认粘贴的是完整的 License 文件内容: %w", err)
	}

	var fullFile struct {
		License   json.RawMessage `json:"license"`
		Signature string          `json:"signature"`
		PublicKey string          `json:"public_key"`
	}
	if err := json.Unmarshal(decoded, &fullFile); err != nil {
		return fmt.Errorf("License 文件格式错误: %w", err)
	}
	if fullFile.PublicKey == "" {
		return fmt.Errorf("License 文件中缺少公钥信息")
	}

	persistData := struct {
		License   json.RawMessage `json:"license"`
		Signature string          `json:"signature"`
		PublicKey string          `json:"public_key"`
		Mode      string          `json:"mode"`
	}{
		License:   fullFile.License,
		Signature: fullFile.Signature,
		PublicKey: fullFile.PublicKey,
		Mode:      "offline",
	}

	persistBytes, _ := json.MarshalIndent(persistData, "", "  ")
	if err := os.WriteFile(defaultLicenseFile, persistBytes, 0644); err != nil {
		return fmt.Errorf("保存 License 文件失败: %w", err)
	}

	pureLicBytes, _ := json.Marshal(struct {
		License   json.RawMessage `json:"license"`
		Signature string          `json:"signature"`
	}{
		License:   fullFile.License,
		Signature: fullFile.Signature,
	})
	pureLicFile := defaultLicenseFile + ".lic"
	if err := os.WriteFile(pureLicFile, pureLicBytes, 0644); err != nil {
		return fmt.Errorf("保存 License 文件失败: %w", err)
	}

	client := license.NewClient(license.Config{
		LicenseFile: pureLicFile,
		PublicKey:   fullFile.PublicKey,
		ServerURL:   s.conf.GetString("license.server_url"),
	})

	result, err := client.VerifyOffline()
	if err != nil {
		return fmt.Errorf("离线验证失败: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("授权无效: %s", result.Reason)
	}

	s.mu.Lock()
	if s.client != nil {
		s.client.StopHeartbeat()
	}
	s.client = client
	s.activated = true
	s.lastResult = result
	s.licenseCode = result.LicenseCode
	s.mu.Unlock()

	return nil
}

func (s *licenseService) Deactivate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.StopHeartbeat()
		_ = s.client.Deactivate()
	}

	s.activated = false
	s.lastResult = nil
	s.client = nil
	s.licenseCode = ""

	_ = os.Remove(defaultLicenseFile)
	_ = os.Remove(defaultLicenseFile + ".lic")

	return nil
}

func (s *licenseService) IsActivated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activated
}

func (s *licenseService) persistState(mode, licenseCode, publicKey, fileContent string) {
	_ = os.MkdirAll("./storage", 0755)

	data := map[string]string{
		"license_code": licenseCode,
		"server_url":   s.conf.GetString("license.server_url"),
		"public_key":   publicKey,
		"mode":         mode,
	}
	bytes, _ := json.MarshalIndent(data, "", "  ")
	_ = os.WriteFile(defaultLicenseFile, bytes, 0644)
}
