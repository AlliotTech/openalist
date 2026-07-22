package _115

import (
	"net/http"
	"testing"
)

func TestWithDownloadCookie(t *testing.T) {
	tests := []struct {
		name       string
		header     http.Header
		cookie     string
		wantCookie string
	}{
		{name: "adds cookie to nil headers", cookie: "UID=u;CID=c", wantCookie: "UID=u;CID=c"},
		{name: "preserves sdk cookie", header: http.Header{"Cookie": []string{"sdk-cookie"}}, cookie: "driver-cookie", wantCookie: "sdk-cookie"},
		{name: "does not add empty cookie", cookie: "", wantCookie: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withDownloadCookie(tt.header, tt.cookie)
			if got.Get("Cookie") != tt.wantCookie {
				t.Fatalf("cookie = %q, want %q", got.Get("Cookie"), tt.wantCookie)
			}
		})
	}
}
