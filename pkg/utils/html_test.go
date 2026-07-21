package utils

import "testing"

func TestSanitizeHTMLRemovesExecutableMarkup(t *testing.T) {
	got := SanitizeHTML(`<script>alert(1)</script><img src=x onerror=alert(2)>storage unavailable`)
	if got != "storage unavailable" {
		t.Fatalf("SanitizeHTML result = %q, want %q", got, "storage unavailable")
	}
}
