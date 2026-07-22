package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlliotTech/openalist/internal/model"
)

func TestRemoveDeletesThumbnailCache(t *testing.T) {
	root := t.TempDir()
	cacheDir := t.TempDir()
	filePath := filepath.Join(root, "image.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	d := &Local{Addition: Addition{ThumbCacheFolder: cacheDir, RecycleBinPath: "delete permanently"}}
	cachePath := d.thumbCachePath(filePath)
	if err := os.WriteFile(cachePath, []byte("thumb"), 0o600); err != nil {
		t.Fatal(err)
	}
	obj := &model.Object{Path: filePath, Name: "image.jpg"}
	if err := d.Remove(context.Background(), obj); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("thumbnail cache still exists: %v", err)
	}
}

func TestThumbCachePathDisabled(t *testing.T) {
	if got := (&Local{}).thumbCachePath("/tmp/image.jpg"); got != "" {
		t.Fatalf("thumbCachePath = %q, want empty", got)
	}
}
