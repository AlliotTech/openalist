package _189pc

import (
	"net/url"
	"testing"
)

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

func TestParseLoginParamURL(t *testing.T) {
	loginURL, err := url.Parse("https://open.e.189.cn/api/logbox/separate/web/index.html?appId=9317140619&lt=lt-value&reqId=req-value&encryptUrl=key")
	if err != nil {
		t.Fatal(err)
	}
	param, err := parseLoginParamURL(loginURL)
	if err != nil {
		t.Fatal(err)
	}
	if param.Lt != "lt-value" || param.ReqId != "req-value" || param.Referer != loginURL.String() {
		t.Fatalf("parsed redirect params = %+v", param)
	}
}

func TestParseLoginParamPage(t *testing.T) {
	param, err := parseLoginParamPage(`
lt = "lt-value"
reqId = "req-value"
paramId = "param-value"
<input name='captchaToken' value='captcha-value'>`)
	if err != nil {
		t.Fatal(err)
	}
	if param.Lt != "lt-value" || param.ReqId != "req-value" || param.ParamId != "param-value" || param.CaptchaToken != "captcha-value" {
		t.Fatalf("parsed legacy params = %+v", param)
	}
}

func TestParseTime(t *testing.T) {
	cases := []string{
		"2026-10-17 16:44:12",
		"Oct 17, 2026 4:44:12 PM",
		"Oct 17, 2026, 4:44:12\u202fPM +08",
	}
	for _, value := range cases {
		if _, err := parseTime(value); err != nil {
			t.Fatalf("parseTime(%q): %v", value, err)
		}
	}
}
