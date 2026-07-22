package webdav

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AlliotTech/openalist/drivers/base"
	"github.com/AlliotTech/openalist/internal/errs"
	"github.com/AlliotTech/openalist/pkg/gowebdav"
	"github.com/go-resty/resty/v2"
)

func TestMapWebDAVError(t *testing.T) {
	err := &os.PathError{Op: "stat", Path: "/missing", Err: gowebdav.StatusError{Status: http.StatusNotFound}}
	if got := mapWebDAVError(err); got != errs.ObjectNotFound {
		t.Fatalf("mapped error = %v, want object not found", got)
	}
}

func TestResolveWebDAVRedirectStripsCredentialsAcrossOrigins(t *testing.T) {
	initNoRedirectClientForTest()
	other := httptest.NewServer(http.NotFoundHandler())
	defer other.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", other.URL+"/file")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	header := http.Header{"Authorization": {"Basic secret"}, "Cookie": {"session=secret"}}
	gotURL, gotHeader, err := resolveWebDAVRedirect(context.Background(), server.URL+"/start", header)
	if err != nil {
		t.Fatalf("resolveWebDAVRedirect returned error: %v", err)
	}
	if gotURL != other.URL+"/file" {
		t.Fatalf("redirect URL = %q, want %q", gotURL, other.URL+"/file")
	}
	if gotHeader.Get("Authorization") != "" || gotHeader.Get("Cookie") != "" {
		t.Fatalf("credentials leaked across redirect: %#v", gotHeader)
	}
}

func TestResolveWebDAVRedirectKeepsCredentialsSameOrigin(t *testing.T) {
	initNoRedirectClientForTest()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/file")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer server.Close()

	header := http.Header{"Authorization": {"Basic secret"}}
	_, gotHeader, err := resolveWebDAVRedirect(context.Background(), server.URL+"/start", header)
	if err != nil {
		t.Fatalf("resolveWebDAVRedirect returned error: %v", err)
	}
	if gotHeader.Get("Authorization") != "Basic secret" {
		t.Fatalf("same-origin credentials were removed: %#v", gotHeader)
	}
}

func initNoRedirectClientForTest() {
	base.NoRedirectClient = resty.New().SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))
}
