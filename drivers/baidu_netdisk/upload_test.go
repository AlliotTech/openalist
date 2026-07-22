package baidu_netdisk

import "testing"

func TestSelectUploadURLPrefersPrimaryServer(t *testing.T) {
	resp := UploadServerResp{}
	resp.Servers = append(resp.Servers, struct {
		Server string `json:"server"`
	}{Server: "https://upload.example/"})
	resp.BakServers = append(resp.BakServers, struct {
		Server string `json:"server"`
	}{Server: "https://backup.example"})

	got, err := selectUploadURL(resp)
	if err != nil {
		t.Fatalf("selectUploadURL returned error: %v", err)
	}
	if got != "https://upload.example" {
		t.Fatalf("upload URL = %q, want primary URL without trailing slash", got)
	}
}

func TestSelectUploadURLFallsBackToBackup(t *testing.T) {
	resp := UploadServerResp{}
	resp.BakServers = append(resp.BakServers, struct {
		Server string `json:"server"`
	}{Server: "https://backup.example"})

	got, err := selectUploadURL(resp)
	if err != nil {
		t.Fatalf("selectUploadURL returned error: %v", err)
	}
	if got != "https://backup.example" {
		t.Fatalf("upload URL = %q, want backup URL", got)
	}
}

func TestSelectUploadURLRejectsInvalidAddress(t *testing.T) {
	resp := UploadServerResp{}
	resp.Servers = append(resp.Servers, struct {
		Server string `json:"server"`
	}{Server: "file:///tmp/upload"})

	if _, err := selectUploadURL(resp); err == nil {
		t.Fatal("selectUploadURL accepted unsupported URL scheme")
	}
}

func TestUploadIDExpiredResponse(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{`{"error":"uploadid expired"}`, true},
		{`{"errno":-1,"error":"invalid uploadid"}`, true},
		{`{"errno":0}`, false},
	} {
		if got := isUploadIDExpiredResponse(tc.body); got != tc.want {
			t.Errorf("isUploadIDExpiredResponse(%q) = %v, want %v", tc.body, got, tc.want)
		}
	}
}

func TestGetUploadURLUsesConfiguredFallbackWhenDisabled(t *testing.T) {
	d := &BaiduNetdisk{Addition: Addition{UploadAPI: "https://fallback.example", UseDynamicUploadAPI: false}}
	if got := d.getUploadURL("/file.bin", "upload-id"); got != "https://fallback.example" {
		t.Fatalf("upload URL = %q, want configured fallback", got)
	}
}
