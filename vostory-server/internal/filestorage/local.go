package filestorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// localStorage 本地磁盘实现，文件落盘于 {root}/{key}
type localStorage struct {
	root string
}

func newLocalStorage(root string) FileStorage {
	return &localStorage{root: root}
}

// resolvePath 将 Key 转换为本地路径，清洗防止路径穿越
func (s *localStorage) resolvePath(key string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if cleaned == "." || strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("非法的文件 key: %s", key)
	}
	return filepath.Join(s.root, cleaned), nil
}

func (s *localStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	path, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建存储目录失败: %w", err)
	}
	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, r); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

func (s *localStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	return f, nil
}

func (s *localStorage) Delete(ctx context.Context, key string) error {
	path, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DownloadToTemp 本地模式直接返回真实路径，零复制开销
func (s *localStorage) DownloadToTemp(ctx context.Context, key string) (string, func(), error) {
	path, err := s.resolvePath(key)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return "", nil, fmt.Errorf("文件不存在: %w", err)
	}
	return path, func() {}, nil
}

func (s *localStorage) PublicURLPrefix(_ context.Context) string {
	return ""
}
