package s3

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/AlliotTech/openalist/internal/model"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomObjectURL(t *testing.T) {
	tests := []struct {
		name   string
		driver S3
		key    string
		want   string
	}{
		{
			name:   "host without scheme",
			driver: S3{Addition: Addition{CustomHost: "cdn.example.com"}},
			key:    "dir/file name.txt",
			want:   "https://cdn.example.com/dir/file%20name.txt",
		},
		{
			name: "path style keeps bucket",
			driver: S3{Addition: Addition{
				CustomHost:     "https://storage.example.com/base",
				Bucket:         "bucket",
				ForcePathStyle: true,
			}},
			key:  "dir/file.txt",
			want: "https://storage.example.com/base/bucket/dir/file.txt",
		},
		{
			name: "remove bucket",
			driver: S3{Addition: Addition{
				CustomHost:     "https://storage.example.com",
				Bucket:         "bucket",
				ForcePathStyle: true,
				RemoveBucket:   true,
			}},
			key:  "dir/file.txt",
			want: "https://storage.example.com/dir/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.driver.customObjectURL(tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithUserAgent(t *testing.T) {
	stack := middleware.NewStack("user-agent", smithyhttp.NewStackRequest)
	require.NoError(t, withUserAgent("client/test")(stack))
	handler := middleware.DecorateHandler(middleware.HandlerFunc(func(_ context.Context, input interface{}) (interface{}, middleware.Metadata, error) {
		req := input.(*smithyhttp.Request)
		assert.Equal(t, "client/test", req.Header.Get("User-Agent"))
		return nil, middleware.Metadata{}, nil
	}), stack)
	_, _, err := handler.Handle(context.Background(), &smithyhttp.Request{Request: &http.Request{Header: make(http.Header)}})
	require.NoError(t, err)
}

func TestCustomHostPresign(t *testing.T) {
	tests := []struct {
		name           string
		forcePathStyle bool
		removeBucket   bool
		wantHost       string
		wantPath       string
	}{
		{
			name:     "virtual host style keeps custom host",
			wantHost: "cdn.example.com",
			wantPath: "/base/dir/file.txt",
		},
		{
			name:           "path style keeps bucket in path",
			forcePathStyle: true,
			wantHost:       "cdn.example.com",
			wantPath:       "/base/bucket/dir/file.txt",
		},
		{
			name:           "remove bucket keeps custom host",
			forcePathStyle: true,
			removeBucket:   true,
			wantHost:       "cdn.example.com",
			wantPath:       "/base/dir/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := S3{Addition: Addition{
				AccessKeyID:             "access-key",
				SecretAccessKey:         "secret-key",
				Region:                  "us-east-1",
				Bucket:                  "bucket",
				CustomHost:              "https://cdn.example.com/base",
				EnableCustomHostPresign: true,
				ForcePathStyle:          tt.forcePathStyle,
				RemoveBucket:            tt.removeBucket,
				SignURLExpire:           1,
			}}
			require.NoError(t, d.Init(context.Background()))

			link, err := d.Link(context.Background(), &model.Object{Path: "/dir/file.txt"}, model.LinkArgs{})
			require.NoError(t, err)
			parsed, err := url.Parse(link.URL)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, parsed.Host)
			assert.Equal(t, tt.wantPath, parsed.Path)
			assert.NotEmpty(t, parsed.Query().Get("X-Amz-Signature"))
		})
	}
}
