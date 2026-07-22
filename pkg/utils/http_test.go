package utils

import (
	"strings"
	"testing"
)

func TestGenerateContentDispositionUsesReadableFallback(t *testing.T) {
	got := GenerateContentDisposition("报告 100%.txt")
	want := `attachment; filename="报告 100%.txt"; filename*=utf-8''%E6%8A%A5%E5%91%8A%20100%25.txt`
	if got != want {
		t.Fatalf("Content-Disposition = %q, want %q", got, want)
	}
	if strings.Contains(got, `filename="%E6`) {
		t.Fatalf("legacy filename fallback is incorrectly percent-encoded: %q", got)
	}
}

func TestGenerateContentDispositionEscapesQuotedFilename(t *testing.T) {
	got := GenerateContentDisposition("a\"b\\c\r\n.txt")
	want := `attachment; filename="a\"b\\c__.txt"; filename*=utf-8''a%22b%5Cc__.txt`
	if got != want {
		t.Fatalf("Content-Disposition = %q, want %q", got, want)
	}
}
