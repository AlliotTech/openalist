package net

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AlliotTech/openalist/internal/conf"
)

func TestHTTPClientProxyDoesNotResolveTargetLocally(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Hostname() != "target-only-resolvable-by-proxy.invalid" {
			t.Fatalf("proxy received target %q", r.URL.Host)
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer proxy.Close()
	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	transport := &http.Transport{
		Proxy: func(*http.Request) (*url.URL, error) { return proxyURL, nil },
	}
	transport.DialContext = safeDialContext
	client := &http.Client{Transport: &safeTransport{base: transport}, Timeout: time.Second}
	resp, err := client.Get("http://target-only-resolvable-by-proxy.invalid/file")
	if err != nil {
		t.Fatalf("proxied request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("proxy response = %q", body)
	}
}

func TestSafeDialContextRejectsMetadataIP(t *testing.T) {
	_, err := safeDialContext(t.Context(), "tcp", fmt.Sprintf("%s:80", "169.254.169.254"))
	if !errors.Is(err, ErrCloudMetadataEndpoint) {
		t.Fatalf("error = %v, want %v", err, ErrCloudMetadataEndpoint)
	}
}

func TestSafeTransportRejectsMetadataIPBeforeProxy(t *testing.T) {
	proxyCalled := false
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		proxyCalled = true
		return nil, errors.New("proxy should not be called")
	})
	client := &http.Client{Transport: &safeTransport{base: transport}}
	_, err := client.Get("http://169.254.169.254/latest/meta-data/")
	if !errors.Is(err, ErrCloudMetadataEndpoint) {
		t.Fatalf("error = %v, want %v", err, ErrCloudMetadataEndpoint)
	}
	if proxyCalled {
		t.Fatal("base transport was called for a metadata IP")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewHTTPClientBlocksCloudMetadataEndpoint(t *testing.T) {
	originalConfig := conf.Conf
	conf.Conf = conf.DefaultConfig()
	defer func() { conf.Conf = originalConfig }()

	endpoints := []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://100.100.100.200/latest/meta-data/",
		"http://[fd00:ec2::254]/latest/meta-data/",
	}
	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			client := NewHttpClient()
			client.Timeout = 250 * time.Millisecond
			_, err := client.Get(endpoint)
			if !errors.Is(err, ErrCloudMetadataEndpoint) {
				t.Fatalf("request error = %v, want %v", err, ErrCloudMetadataEndpoint)
			}
		})
	}
}
