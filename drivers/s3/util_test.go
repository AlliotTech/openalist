package s3

import (
	"testing"

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
