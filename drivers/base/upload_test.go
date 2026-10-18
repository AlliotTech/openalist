package base

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type uploadRoundTripper func(*http.Request) (*http.Response, error)

func (f uploadRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestUploadChunkRetriesEOFAndReplaysChunk(t *testing.T) {
	var attempts int
	client := &http.Client{Transport: uploadRoundTripper(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, io.EOF
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if string(body) != "chunk" {
			return nil, errors.New("request body was not replayed")
		}
		if got := req.Header.Get("Content-Range"); got != "bytes 5-9/10" {
			return nil, errors.New("unexpected content range: " + got)
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}

	wait := func(context.Context, time.Duration) error { return nil }
	if err := uploadChunk(context.Background(), client, "https://upload.invalid/session", []byte("chunk"), 5, 10, wait); err != nil {
		t.Fatalf("uploadChunk() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("upload attempts = %d, want 2", attempts)
	}
}

func TestUploadChunkRetriesTransientStatus(t *testing.T) {
	var attempts int
	client := &http.Client{Transport: uploadRoundTripper(func(*http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("temporary"))}, nil
		}
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}

	wait := func(context.Context, time.Duration) error { return nil }
	if err := uploadChunk(context.Background(), client, "https://upload.invalid/session", []byte("chunk"), 0, 5, wait); err != nil {
		t.Fatalf("uploadChunk() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("upload attempts = %d, want 3", attempts)
	}
}

func TestUploadChunkStopsAfterTransientFailures(t *testing.T) {
	var attempts int
	client := &http.Client{Transport: uploadRoundTripper(func(*http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("busy"))}, nil
	})}

	wait := func(context.Context, time.Duration) error { return nil }
	err := uploadChunk(context.Background(), client, "https://upload.invalid/session", []byte("chunk"), 0, 5, wait)
	if err == nil || err.Error() != "busy" {
		t.Fatalf("uploadChunk() error = %v, want busy", err)
	}
	if attempts != uploadChunkAttempts {
		t.Fatalf("upload attempts = %d, want %d", attempts, uploadChunkAttempts)
	}
}
func TestUploadChunkDoesNotRetryPermanentStatus(t *testing.T) {
	var attempts int
	client := &http.Client{Transport: uploadRoundTripper(func(*http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("expired"))}, nil
	})}

	wait := func(context.Context, time.Duration) error { return nil }
	err := uploadChunk(context.Background(), client, "https://upload.invalid/session", []byte("chunk"), 0, 5, wait)
	if err == nil || err.Error() != "expired" {
		t.Fatalf("uploadChunk() error = %v, want expired", err)
	}
	if attempts != 1 {
		t.Fatalf("upload attempts = %d, want 1", attempts)
	}
}
