package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFilePathAcceptsAbsoluteRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Rock & Roll! $demo.mp4")
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := sanitizeFilePath(path)
	if err != nil {
		t.Fatalf("sanitizeFilePath returned error: %v", err)
	}
	if got != path {
		t.Fatalf("sanitizeFilePath = %q, want %q", got, path)
	}
}

func TestSanitizeFilePathAcceptsSymlinkToRegularFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.mp4")
	if err := os.WriteFile(target, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "video.mp4")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink is not available: %v", err)
	}

	got, err := sanitizeFilePath(link)
	if err != nil {
		t.Fatalf("sanitizeFilePath returned error: %v", err)
	}
	if got != link {
		t.Fatalf("sanitizeFilePath = %q, want symlink path %q", got, link)
	}
}

func TestSanitizeFilePathRejectsUnsafeInputs(t *testing.T) {
	regularFile := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(regularFile, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "relative", path: "video.mp4", want: "absolute"},
		{name: "missing", path: filepath.Join(t.TempDir(), "missing.mp4"), want: "accessible"},
		{name: "directory", path: directory, want: "regular file"},
		{name: "newline", path: regularFile + "\ncopy", want: "invalid characters"},
		{name: "carriage return", path: regularFile + "\rcopy", want: "invalid characters"},
		{name: "nul", path: regularFile + "\x00copy", want: "invalid characters"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := sanitizeFilePath(test.path)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("sanitizeFilePath(%q) error = %v, want %q", test.path, err, test.want)
			}
		})
	}
}

func TestGetSnapshotRejectsInvalidPathBeforeRunningFFprobe(t *testing.T) {
	_, err := (&Local{}).GetSnapshot("relative-video.mp4")
	if err == nil || !strings.Contains(err.Error(), "invalid video path") {
		t.Fatalf("GetSnapshot error = %v, want invalid video path", err)
	}
}
