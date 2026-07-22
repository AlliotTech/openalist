package pikpak

import "testing"

func TestRewriteDownloadURL(t *testing.T) {
	tests := []struct {
		name           string
		downloadDomain string
		customDomain   string
		input          string
		want           string
	}{
		{"original", "original", "", "https://dl-a.mypikpak.com/file?sig=1", "https://dl-a.mypikpak.com/file?sig=1"},
		{"selected domain", "mypikpak_net", "", "https://dl-a.mypikpak.com/file?sig=1", "https://dl-a.mypikpak.net/file?sig=1"},
		{"custom override", "mypikpak_net", "example.net", "https://vip.mypikpak.com/file", "https://vip.example.net/file"},
		{"custom URL normalized", "original", "https://example.net/", "https://vip.mypikpak.com/file", "https://vip.example.net/file"},
		{"custom original", "mypikpak_net", "original", "https://vip.mypikpak.com/file", "https://vip.mypikpak.net/file"},
		{"bare host", "pikpak_me", "", "https://mypikpak.net/file", "https://pikpak.me/file"},
		{"unknown host", "mypikpak_net", "", "https://cdn.example.com/file", "https://cdn.example.com/file"},
		{"invalid URL", "mypikpak_net", "", "://bad", "://bad"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &PikPak{Addition: Addition{
				DownloadDomain:       test.downloadDomain,
				CustomDownloadDomain: test.customDomain,
			}}
			if got := d.rewriteDownloadURL(test.input); got != test.want {
				t.Fatalf("rewriteDownloadURL(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestAPIAndUserURL(t *testing.T) {
	tests := []struct {
		apiDomain    string
		customDomain string
		wantAPI      string
		wantUser     string
	}{
		{"", "", "https://api-drive.mypikpak.net/drive/v1/files", "https://user.mypikpak.net/v1/auth/token"},
		{"mypikpak_com", "", "https://api-drive.mypikpak.com/drive/v1/files", "https://user.mypikpak.com/v1/auth/token"},
		{"pikpak_me", "", "https://api-drive.pikpak.me/drive/v1/files", "https://user.pikpak.me/v1/auth/token"},
		{"mypikpak_net", "example.org", "https://api-drive.example.org/drive/v1/files", "https://user.example.org/v1/auth/token"},
		{"mypikpak_net", "https://EXAMPLE.org/", "https://api-drive.example.org/drive/v1/files", "https://user.example.org/v1/auth/token"},
	}
	for _, test := range tests {
		d := &PikPak{Addition: Addition{APIDomain: test.apiDomain, CustomAPIDomain: test.customDomain}}
		if got := d.apiURL("/drive/v1/files"); got != test.wantAPI {
			t.Fatalf("apiURL() = %q, want %q", got, test.wantAPI)
		}
		if got := d.userURL("/v1/auth/token"); got != test.wantUser {
			t.Fatalf("userURL() = %q, want %q", got, test.wantUser)
		}
	}
}
