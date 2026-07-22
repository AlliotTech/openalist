package _189

import "testing"

func TestSanitizeName(t *testing.T) {
	d := &Cloud189{Addition: Addition{StripEmoji: true}}
	if got := d.sanitizeName("报告😀.txt"); got != "报告.txt" {
		t.Fatalf("sanitizeName = %q, want %q", got, "报告.txt")
	}
	if got := d.sanitizeName("😀.txt"); got != "file.txt" {
		t.Fatalf("sanitizeName = %q, want fallback filename", got)
	}
	if got := (&Cloud189{}).sanitizeName("报告😀.txt"); got != "报告😀.txt" {
		t.Fatalf("sanitization should be disabled by default, got %q", got)
	}
}
