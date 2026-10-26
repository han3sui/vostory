package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Type     string // huawei / aliyun / tencent / minio
	Endpoint string
	AK       string
	SK       string
	Bucket   string
	Prefix   string
	UseSSL   bool
}

type Client struct {
	mc     *minio.Client
	bucket string
	prefix string
}

func NewClient(cfg Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AK, cfg.SK, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}
	return &Client{mc: mc, bucket: cfg.Bucket, prefix: cfg.Prefix}, nil
}

func (c *Client) PresignedPutURL(ctx context.Context, objectKey string, expires time.Duration) (*url.URL, error) {
	return c.mc.PresignedPutObject(ctx, c.bucket, c.prefix+objectKey, expires)
}

func (c *Client) PresignedGetURL(ctx context.Context, objectKey string, expires time.Duration) (*url.URL, error) {
	return c.mc.PresignedGetObject(ctx, c.bucket, c.prefix+objectKey, expires, nil)
}

// PutObject 服务端直传对象。size 传 -1 时由 SDK 走分片上传。
func (c *Client) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := c.mc.PutObject(ctx, c.bucket, c.prefix+objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// GetObject 获取对象读取流
func (c *Client) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, c.prefix+objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// GetObject 是惰性请求，Stat 触发一次请求以便尽早发现对象不存在等错误
	if _, err := obj.Stat(); err != nil {
		obj.Close()
		return nil, err
	}
	return obj, nil
}

// RemoveObject 删除对象
func (c *Client) RemoveObject(ctx context.Context, objectKey string) error {
	return c.mc.RemoveObject(ctx, c.bucket, c.prefix+objectKey, minio.RemoveObjectOptions{})
}

func (c *Client) TestConnection(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("test connection: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket %q does not exist", c.bucket)
	}
	return nil
}
