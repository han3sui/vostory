package filestorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	s3client "iot-alert-center/internal/client/s3"
	"iot-alert-center/internal/repository"
)

// s3Storage OSS 对象存储实现。
// 每次使用时从数据库读取 OSS 配置构建 client，管理页修改配置即时生效。
type s3Storage struct {
	ossConfigRepo repository.SysOSSConfigRepository
}

func newS3Storage(ossConfigRepo repository.SysOSSConfigRepository) FileStorage {
	return &s3Storage{ossConfigRepo: ossConfigRepo}
}

// getClient 从数据库配置构建 S3 客户端
func (s *s3Storage) getClient(ctx context.Context) (*s3client.Client, error) {
	configs, err := s.ossConfigRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oss config: %w", err)
	}
	endpoint := configs["oss.endpoint"]
	ak := configs["oss.ak"]
	sk := configs["oss.sk"]
	bucket := configs["oss.bucket"]
	if endpoint == "" || ak == "" || sk == "" || bucket == "" {
		return nil, fmt.Errorf("对象存储未配置，请先到 系统管理 > 对象存储 中完成配置")
	}
	useSSL := configs["oss.use_ssl"] != "false"
	return s3client.NewClient(s3client.Config{
		Type:     configs["oss.type"],
		Endpoint: endpoint,
		AK:       ak,
		SK:       sk,
		Bucket:   bucket,
		Prefix:   configs["oss.prefix"],
		UseSSL:   useSSL,
	})
}

func (s *s3Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}
	return client.PutObject(ctx, key, r, size, contentType)
}

func (s *s3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.GetObject(ctx, key)
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}
	return client.RemoveObject(ctx, key)
}

// DownloadToTemp 下载对象到临时文件，cleanup 负责删除
func (s *s3Storage) DownloadToTemp(ctx context.Context, key string) (string, func(), error) {
	reader, err := s.Open(ctx, key)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	// 保留原扩展名，部分解析库依赖扩展名判断格式
	tmp, err := os.CreateTemp("", "filestorage_*"+filepath.Ext(key))
	if err != nil {
		return "", nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	if _, err := io.Copy(tmp, reader); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("下载文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", nil, err
	}
	path := tmp.Name()
	return path, func() { os.Remove(path) }, nil
}

func (s *s3Storage) PublicURLPrefix(ctx context.Context) string {
	configs, err := s.ossConfigRepo.GetAll(ctx)
	if err != nil {
		return ""
	}
	cdn := strings.TrimRight(strings.TrimSpace(configs["oss.cdn_base"]), "/")
	if cdn == "" {
		return ""
	}
	return cdn + "/" + strings.TrimLeft(configs["oss.prefix"], "/")
}
