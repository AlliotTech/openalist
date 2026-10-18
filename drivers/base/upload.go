package base

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/Xhofe/go-cache"
)

const uploadChunkAttempts = 3

// UploadChunk sends one resumable-upload chunk and retries transient failures.
// The caller must keep data unchanged so each retry can replay the same range.
func UploadChunk(ctx context.Context, client *http.Client, uploadURL string, data []byte, start, total int64) error {
	return uploadChunk(ctx, client, uploadURL, data, start, total, waitUploadRetry)
}

func uploadChunk(ctx context.Context, client *http.Client, uploadURL string, data []byte, start, total int64, wait func(context.Context, time.Duration) error) error {
	if client == nil {
		return errors.New("upload client is nil")
	}
	end := start + int64(len(data)) - 1
	for attempt := 1; attempt <= uploadChunkAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, driver.NewLimitedUploadStream(ctx, bytes.NewReader(data)))
		if err != nil {
			return err
		}
		req.ContentLength = int64(len(data))
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, total))
		res, err := client.Do(req)
		if err != nil {
			if !isRetryableUploadError(err) || attempt == uploadChunkAttempts {
				return err
			}
			if err := wait(ctx, uploadRetryDelay(attempt)); err != nil {
				return err
			}
			continue
		}

		retryableStatus := res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500 && res.StatusCode <= 504
		if retryableStatus && attempt < uploadChunkAttempts {
			res.Body.Close()
			if err := wait(ctx, uploadRetryDelay(attempt)); err != nil {
				return err
			}
			continue
		}
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusAccepted {
			data, readErr := io.ReadAll(res.Body)
			res.Body.Close()
			if readErr != nil {
				return readErr
			}
			if len(data) == 0 {
				return fmt.Errorf("upload chunk failed with HTTP status %d", res.StatusCode)
			}
			return errors.New(string(data))
		}
		res.Body.Close()
		return nil
	}
	return errors.New("upload chunk retry loop exhausted")
}

func isRetryableUploadError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary())
}

func uploadRetryDelay(attempt int) time.Duration {
	return time.Duration(1<<attempt) * time.Second
}

func waitUploadRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// storage upload progress, for upload recovery
var UploadStateCache = cache.NewMemCache(cache.WithShards[any](32))

// Save upload progress for 20 minutes
func SaveUploadProgress(driver driver.Driver, state any, keys ...string) bool {
	return UploadStateCache.Set(
		fmt.Sprint(driver.Config().Name, "-upload-", strings.Join(keys, "-")),
		state,
		cache.WithEx[any](time.Minute*20))
}

// An upload progress can only be made by one process alone,
// so here you need to get it and then delete it.
func GetUploadProgress[T any](driver driver.Driver, keys ...string) (state T, ok bool) {
	v, ok := UploadStateCache.GetDel(fmt.Sprint(driver.Config().Name, "-upload-", strings.Join(keys, "-")))
	if ok {
		state, ok = v.(T)
	}
	return
}
