package service

import (
	"context"
	"fmt"
	v1 "iot-alert-center/api/v1"
	s3client "iot-alert-center/internal/client/s3"
	"iot-alert-center/internal/repository"
	"strings"
)

type SysOSSConfigService interface {
	GetConfig(ctx context.Context) (*v1.OSSConfigResponse, error)
	UpdateConfig(ctx context.Context, req *v1.OSSConfigUpdateRequest) error
	TestConnection(ctx context.Context, req *v1.OSSConfigUpdateRequest) error
}

func NewSysOSSConfigService(
	service *Service,
	ossConfigRepo repository.SysOSSConfigRepository,
) SysOSSConfigService {
	return &sysOSSConfigService{
		Service:       service,
		ossConfigRepo: ossConfigRepo,
	}
}

type sysOSSConfigService struct {
	*Service
	ossConfigRepo repository.SysOSSConfigRepository
}

func (s *sysOSSConfigService) GetConfig(ctx context.Context) (*v1.OSSConfigResponse, error) {
	configs, err := s.ossConfigRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.OSSConfigResponse{
		Type:     configs["oss.type"],
		Endpoint: configs["oss.endpoint"],
		AK:       configs["oss.ak"],
		SK:       maskSK(configs["oss.sk"]),
		Bucket:   configs["oss.bucket"],
		Prefix:   configs["oss.prefix"],
		UseSSL:   configs["oss.use_ssl"],
		CdnBase:  configs["oss.cdn_base"],
	}, nil
}

func (s *sysOSSConfigService) UpdateConfig(ctx context.Context, req *v1.OSSConfigUpdateRequest) error {
	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.AK = strings.TrimSpace(req.AK)
	req.SK = strings.TrimSpace(req.SK)
	req.Bucket = strings.TrimSpace(req.Bucket)
	req.Prefix = strings.TrimSpace(req.Prefix)

	items := map[string]struct {
		name  string
		value string
	}{
		"oss.type":     {name: "存储类型", value: req.Type},
		"oss.endpoint": {name: "服务端点", value: req.Endpoint},
		"oss.ak":       {name: "AccessKey", value: req.AK},
		"oss.bucket":   {name: "桶名称", value: req.Bucket},
		"oss.prefix":   {name: "路径前缀", value: req.Prefix},
		"oss.use_ssl":  {name: "启用SSL", value: req.UseSSL},
		"oss.cdn_base": {name: "CDN访问域名", value: strings.TrimRight(strings.TrimSpace(req.CdnBase), "/")},
	}
	for key, item := range items {
		if err := s.ossConfigRepo.SetByKey(ctx, key, item.name, item.value); err != nil {
			return fmt.Errorf("save %s: %w", key, err)
		}
	}
	// SK 只在非脱敏值时更新
	if req.SK != "" && !strings.Contains(req.SK, "****") {
		if err := s.ossConfigRepo.SetByKey(ctx, "oss.sk", "SecretKey", req.SK); err != nil {
			return fmt.Errorf("save oss.sk: %w", err)
		}
	}
	return nil
}

func (s *sysOSSConfigService) TestConnection(ctx context.Context, req *v1.OSSConfigUpdateRequest) error {
	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.AK = strings.TrimSpace(req.AK)
	req.Bucket = strings.TrimSpace(req.Bucket)

	sk := strings.TrimSpace(req.SK)
	if strings.Contains(sk, "****") {
		existing, err := s.ossConfigRepo.GetByKey(ctx, "oss.sk")
		if err != nil {
			return fmt.Errorf("获取已保存的 SK 失败: %w", err)
		}
		sk = existing
	}
	useSSL := req.UseSSL != "false"
	client, err := s3client.NewClient(s3client.Config{
		Type:     req.Type,
		Endpoint: req.Endpoint,
		AK:       req.AK,
		SK:       sk,
		Bucket:   req.Bucket,
		Prefix:   req.Prefix,
		UseSSL:   useSSL,
	})
	if err != nil {
		return fmt.Errorf("创建客户端失败: %w", err)
	}
	return client.TestConnection(ctx)
}

func maskSK(sk string) string {
	if len(sk) <= 8 {
		return "****"
	}
	return sk[:4] + "****" + sk[len(sk)-4:]
}
