package mediatrack

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlliotTech/openalist/drivers/base"
	"github.com/AlliotTech/openalist/internal/conf"
)

func TestRequestAddsDeviceFingerprint(t *testing.T) {
	if conf.Conf == nil {
		conf.Conf = conf.DefaultConfig()
	}
	base.InitClient()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Device-Fingerprint"); got != "device-fingerprint" {
			t.Errorf("X-Device-Fingerprint = %q, want %q", got, "device-fingerprint")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"SUCCESS"}`))
	}))
	defer server.Close()

	d := &MediaTrack{Addition: Addition{
		AccessToken:       "access-token",
		DeviceFingerprint: "device-fingerprint",
	}}
	if _, err := d.request(server.URL, http.MethodGet, nil, nil); err != nil {
		t.Fatalf("request failed: %v", err)
	}
}
