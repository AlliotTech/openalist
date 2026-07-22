package template

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/AlliotTech/openalist/internal/model"
	streampkg "github.com/AlliotTech/openalist/internal/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareWopanUploadFile(t *testing.T) {
	content := bytes.Repeat([]byte("a"), 1024)
	stream := &streampkg.FileStream{
		Ctx:    context.Background(),
		Obj:    &model.Object{Name: "test.bin", Size: int64(len(content))},
		Reader: bytes.NewReader(content),
	}
	var progress []float64

	file, cleanup, err := prepareWopanUploadFile(context.Background(), stream, func(value float64) {
		progress = append(progress, value)
	})
	require.NoError(t, err)
	defer cleanup()

	got, err := io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, content, got)
	require.NotEmpty(t, progress)
	assert.Equal(t, wopanPrepareProgress, progress[len(progress)-1])
}

func TestPrepareWopanUploadFileWithoutProgressCallback(t *testing.T) {
	content := bytes.Repeat([]byte("b"), 1024)
	stream := &streampkg.FileStream{
		Ctx:    context.Background(),
		Obj:    &model.Object{Name: "test.bin", Size: int64(len(content))},
		Reader: bytes.NewReader(content),
	}

	file, cleanup, err := prepareWopanUploadFile(context.Background(), stream, nil)
	require.NoError(t, err)
	defer cleanup()
	got, err := io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}
