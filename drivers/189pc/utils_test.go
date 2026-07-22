package _189pc

import "testing"

func TestSanitizeName(t *testing.T) {
	y := &Cloud189PC{Addition: Addition{StripEmoji: true}}
	if got := y.sanitizeName("报告😀.txt"); got != "报告.txt" {
		t.Fatalf("sanitizeName = %q, want %q", got, "报告.txt")
	}
	if got := y.sanitizeName("😀.txt"); got != "file.txt" {
		t.Fatalf("sanitizeName = %q, want fallback filename", got)
	}
	if got := (&Cloud189PC{}).sanitizeName("报告😀.txt"); got != "报告😀.txt" {
		t.Fatalf("sanitization should be disabled by default, got %q", got)
	}
}
