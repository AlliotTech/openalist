package tool

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlliotTech/openalist/internal/model"
)

type testArchiveReader struct {
	files []SubFile
}

func (r testArchiveReader) Files() []SubFile {
	return r.files
}

type testArchiveFile struct {
	archivePath string
	name        string
	content     string
}

func (f testArchiveFile) Name() string {
	return f.archivePath
}

func (f testArchiveFile) FileInfo() fs.FileInfo {
	return testFileInfo{name: f.name, size: int64(len(f.content))}
}

func (f testArchiveFile) Open() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.content)), nil
}

type testFileInfo struct {
	name string
	size int64
}

func (f testFileInfo) Name() string       { return f.name }
func (f testFileInfo) Size() int64        { return f.size }
func (f testFileInfo) Mode() fs.FileMode  { return 0o600 }
func (f testFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f testFileInfo) IsDir() bool        { return false }
func (f testFileInfo) Sys() any           { return nil }

func TestDecompressFromFolderTraversalRejectsZipSlip(t *testing.T) {
	root := t.TempDir()
	outputPath := filepath.Join(root, "extract")
	if err := os.Mkdir(outputPath, 0o700); err != nil {
		t.Fatalf("create extraction directory: %v", err)
	}
	outsidePath := filepath.Join(root, "outside.txt")
	reader := testArchiveReader{files: []SubFile{
		testArchiveFile{
			archivePath: "../outside.txt",
			name:        "outside.txt",
			content:     "escaped",
		},
	}}

	err := DecompressFromFolderTraversal(reader, outputPath, model.ArchiveInnerArgs{
		InnerPath: "/",
	}, func(float64) {})
	if err == nil {
		t.Fatal("DecompressFromFolderTraversal succeeded for a path outside the extraction directory")
	}
	if _, statErr := os.Stat(outsidePath); !os.IsNotExist(statErr) {
		t.Fatalf("archive wrote outside extraction directory: %s", outsidePath)
	}
}
