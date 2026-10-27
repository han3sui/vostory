package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"iot-alert-center/internal/filestorage"
)

// refAudioCache 参考音频本地缓存。参考音频 key 都带毫秒时间戳，换音色会写新 key、不会覆盖旧对象，
// 所以按 key 缓存不会读到旧内容；旁白这类角色几万条片段共用一份参考音频，只需下载一次。
// 缓存文件不随任务删除，重启后仍可复用
type refAudioCache struct {
	dir   string
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func newRefAudioCache(dir string) *refAudioCache {
	return &refAudioCache{dir: dir, locks: make(map[string]*sync.Mutex)}
}

func (c *refAudioCache) keyLock(key string) *sync.Mutex {
	c.mu.Lock()
	defer c.mu.Unlock()
	l, ok := c.locks[key]
	if !ok {
		l = &sync.Mutex{}
		c.locks[key] = l
	}
	return l
}

// Path 返回参考音频的本地路径，没有缓存时下载一次。同一 key 并发调用只会下载一次
func (c *refAudioCache) Path(ctx context.Context, fs filestorage.FileStorage, key string) (string, error) {
	sum := sha1.Sum([]byte(key))
	path := filepath.Join(c.dir, hex.EncodeToString(sum[:])+filepath.Ext(key))
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	l := c.keyLock(key)
	l.Lock()
	defer l.Unlock()
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return "", fmt.Errorf("创建参考音频缓存目录失败: %w", err)
	}
	r, err := fs.Open(ctx, key)
	if err != nil {
		return "", err
	}
	defer r.Close()

	// 先写临时文件再改名，下载中断不会留下半个文件被当成缓存
	tmp, err := os.CreateTemp(c.dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("下载参考音频失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return path, nil
}
