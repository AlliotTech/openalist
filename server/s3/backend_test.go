package s3

import (
	"sync"
	"testing"
	"time"
)

func TestObjectMetadataDoesNotForceDownload(t *testing.T) {
	b := &s3Backend{meta: new(sync.Map)}
	metadata := b.objectMetadata("/bucket/preview.pdf", time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC))

	if disposition, ok := metadata["Content-Disposition"]; ok {
		t.Fatalf("unexpected default Content-Disposition %q", disposition)
	}
	if got := metadata["Content-Type"]; got != "application/pdf" {
		t.Fatalf("Content-Type = %q, want application/pdf", got)
	}
}

func TestObjectMetadataKeepsStoredContentDisposition(t *testing.T) {
	b := &s3Backend{meta: new(sync.Map)}
	b.meta.Store("/bucket/download.bin", map[string]string{
		"Content-Disposition": `attachment; filename="download.bin"`,
	})

	metadata := b.objectMetadata("/bucket/download.bin", time.Time{})
	if got := metadata["Content-Disposition"]; got != `attachment; filename="download.bin"` {
		t.Fatalf("Content-Disposition = %q, want stored value", got)
	}
}
