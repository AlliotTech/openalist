package utils

import (
	"errors"
	"testing"

	"github.com/AlliotTech/openalist/internal/errs"
)

func TestEncodePath(t *testing.T) {
	t.Log(EncodePath("http://localhost:5244/d/123#.png"))
}

func TestFixAndCleanPath(t *testing.T) {
	datas := map[string]string{
		"":                          "/",
		".././":                     "/",
		"../../.../":                "/...",
		"x//\\y/":                   "/x/y",
		".././.x/.y/.//..x../..y..": "/.x/.y/..x../..y..",
	}
	for key, value := range datas {
		if FixAndCleanPath(key) != value {
			t.Logf("raw %s fix fail", key)
		}
	}
}

func TestJoinBasePathRejectsTraversal(t *testing.T) {
	tests := []string{
		"..",
		"../secret",
		"folder/../secret",
		`folder\..\secret`,
		"../../file..txt",
	}

	for _, reqPath := range tests {
		t.Run(reqPath, func(t *testing.T) {
			_, err := JoinBasePath("/home/user", reqPath)
			if !errors.Is(err, errs.RelativePath) {
				t.Fatalf("JoinBasePath(%q) error = %v, want %v", reqPath, err, errs.RelativePath)
			}
		})
	}
}

func TestJoinBasePathAllowsDotsInFileName(t *testing.T) {
	got, err := JoinBasePath("/home/user", "reports/file..txt")
	if err != nil {
		t.Fatalf("JoinBasePath returned unexpected error: %v", err)
	}
	if got != "/home/user/reports/file..txt" {
		t.Fatalf("JoinBasePath result = %q, want %q", got, "/home/user/reports/file..txt")
	}
}
