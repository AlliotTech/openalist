package http

import (
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlliotTech/openalist/internal/offline_download/tool"
)

func TestSimpleHTTPRejectsTraversalFilename(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="../outside.txt"`)
		_, _ = w.Write([]byte("escaped"))
	}))
	defer server.Close()

	root := t.TempDir()
	tempDir := filepath.Join(root, "download")
	task := &tool.DownloadTask{
		Url:     server.URL,
		TempDir: tempDir,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	task.SetCtx(ctx)
	task.SetCancelFunc(cancel)

	if err := (SimpleHttp{client: server.Client()}).Run(task); err != nil {
		t.Fatalf("SimpleHttp.Run returned an unexpected error: %v", err)
	}
	outsidePath := filepath.Join(root, "outside.txt")
	if _, statErr := os.Stat(outsidePath); !os.IsNotExist(statErr) {
		t.Fatalf("download wrote outside temporary directory: %s", outsidePath)
	}
}
