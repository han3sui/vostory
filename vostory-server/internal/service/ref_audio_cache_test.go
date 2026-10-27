package service

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"iot-alert-center/internal/filestorage"
)

// countingStorage 统计 Open 次数，其余方法沿用本地存储
type countingStorage struct {
	filestorage.FileStorage
	opens int32
}

func (c *countingStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	atomic.AddInt32(&c.opens, 1)
	return c.FileStorage.Open(ctx, key)
}

func TestRefAudioCacheDownloadsOncePerKey(t *testing.T) {
	ctx := context.Background()
	fs := &countingStorage{FileStorage: filestorage.NewLocal(t.TempDir())}
	data := []byte("RIFF-fake-wav")
	if err := fs.Put(ctx, "reference-audio/1_a.wav", bytes.NewReader(data), int64(len(data)), "audio/wav"); err != nil {
		t.Fatal(err)
	}
	cache := newRefAudioCache(t.TempDir())

	var wg sync.WaitGroup
	paths := make([]string, 8)
	for i := range paths {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p, err := cache.Path(ctx, fs, "reference-audio/1_a.wav")
			if err != nil {
				t.Error(err)
			}
			paths[i] = p
		}(i)
	}
	wg.Wait()

	if n := atomic.LoadInt32(&fs.opens); n != 1 {
		t.Fatalf("opens=%d, want 1", n)
	}
	for _, p := range paths {
		if p != paths[0] {
			t.Fatalf("paths differ: %v", paths)
		}
	}
	got, err := os.ReadFile(paths[0])
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("cached content=%q err=%v", got, err)
	}

	// 新 cache 实例（模拟重启）直接复用磁盘上的文件
	again := newRefAudioCache(cache.dir)
	if _, err := again.Path(ctx, fs, "reference-audio/1_a.wav"); err != nil {
		t.Fatal(err)
	}
	if n := atomic.LoadInt32(&fs.opens); n != 1 {
		t.Fatalf("opens after restart=%d, want 1", n)
	}

	if _, err := cache.Path(ctx, fs, "reference-audio/missing.wav"); err == nil {
		t.Fatal("missing key should fail")
	}
}
