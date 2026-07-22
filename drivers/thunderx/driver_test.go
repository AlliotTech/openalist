package thunderx

import (
	"errors"
	"net/http"
	"testing"

	"github.com/AlliotTech/openalist/internal/errs"
)

func TestRequestReturnsErrorWhenTokenIsMissing(t *testing.T) {
	xc := &XunLeiXCommon{}
	_, err := xc.Request("", http.MethodGet, nil, nil)
	if !errors.Is(err, errs.EmptyToken) {
		t.Fatalf("expected EmptyToken, got %v", err)
	}
}
