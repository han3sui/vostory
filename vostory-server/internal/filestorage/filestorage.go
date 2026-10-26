package filestorage

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"iot-alert-center/internal/repository"
	"iot-alert-center/pkg/log"

	"github.com/spf13/viper"
)

// FileStorage 统一文件存储抽象（local / oss 可插拔）。
// 业务代码只依赖本接口，不直接操作本地磁盘或 OSS。
// Key 统一使用正斜杠的相对路径，如 audio/{chapterID}/{segmentID}_v{version}.wav。
type FileStorage interface {
	// Put 保存文件流到指定 Key
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	// Open 打开文件读取流（下载代理、导出读取用），调用方负责 Close
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete 删除文件
	Delete(ctx context.Context, key string) error
	// DownloadToTemp 供需要本地文件路径的处理逻辑使用。
	// local 实现直接返回原路径 + 空 cleanup；oss 实现下载到临时文件，cleanup 负责删除。
	DownloadToTemp(ctx context.Context, key string) (path string, cleanup func(), err error)
	// PublicURLPrefix CDN 前缀（含 oss.prefix，可直接拼 key）；未配置返回空
	PublicURLPrefix(ctx context.Context) string
}

// NewFileStorage 按配置 storage.type 构建存储实现（默认 local）
func NewFileStorage(conf *viper.Viper, ossConfigRepo repository.SysOSSConfigRepository, logger *log.Logger) FileStorage {
	storageType := conf.GetString("storage.type")
	if storageType == "oss" {
		logger.Info("文件存储使用 oss 模式")
		return newS3Storage(ossConfigRepo)
	}
	root := conf.GetString("storage.local.root")
	if root == "" {
		root = "storage"
	}
	logger.Info("文件存储使用 local 模式: " + root)
	return newLocalStorage(root)
}

// NewLocal 仅用于测试
func NewLocal(root string) FileStorage {
	if root == "" {
		root = "storage"
	}
	return newLocalStorage(root)
}

// NormalizeKey 把库里存的历史本地路径收成统一 Key。
// 兼容 storage/audio/...、storage\audio\...，新写入本身就是 audio/...。
func NormalizeKey(stored string) string {
	p := strings.TrimSpace(stored)
	if p == "" {
		return ""
	}
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	for strings.HasPrefix(p, "storage/") {
		p = strings.TrimPrefix(p, "storage/")
	}
	return strings.TrimPrefix(p, "/")
}

// Copy 把 src 对象复制到 dst（Open + Put）
func Copy(ctx context.Context, fs FileStorage, src, dst, contentType string) error {
	r, err := fs.Open(ctx, NormalizeKey(src))
	if err != nil {
		return err
	}
	defer r.Close()
	return fs.Put(ctx, dst, r, -1, contentType)
}

// Exists 判断对象是否存在
func Exists(ctx context.Context, fs FileStorage, key string) bool {
	r, err := fs.Open(ctx, NormalizeKey(key))
	if err != nil {
		return false
	}
	_ = r.Close()
	return true
}
