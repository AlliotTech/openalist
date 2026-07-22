package s3

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestNormalizeStorageClass(t *testing.T) {
	tests := map[string]string{
		"":                    "",
		" standard-ia ":       "STANDARD_IA",
		"intelligent_tiering": "INTELLIGENT_TIERING",
		"DEEP_ARCHIVE":        "DEEP_ARCHIVE",
		"archive":             "ARCHIVE",
	}
	for input, want := range tests {
		if got := normalizeStorageClass(input); got != want {
			t.Fatalf("normalizeStorageClass(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeRestoreTier(t *testing.T) {
	tests := map[string]types.Tier{
		"":          "",
		"default":   "",
		" BULK ":    types.TierBulk,
		"standard":  types.TierStandard,
		"expedited": types.TierExpedited,
		"custom":    "custom",
	}
	for input, want := range tests {
		if got := normalizeRestoreTier(input); got != want {
			t.Fatalf("normalizeRestoreTier(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseRestoreHeader(t *testing.T) {
	if got := parseRestoreHeader(nil); got != nil {
		t.Fatalf("parseRestoreHeader(nil) = %#v, want nil", got)
	}
	empty := "  "
	if got := parseRestoreHeader(&empty); got != nil {
		t.Fatalf("parseRestoreHeader(empty) = %#v, want nil", got)
	}

	header := `ongoing-request="false", expiry-date="Fri, 21 Dec 2012 00:00:00 GMT"`
	got := parseRestoreHeader(&header)
	if got == nil || got.Ongoing {
		t.Fatalf("parseRestoreHeader(%q) = %#v", header, got)
	}
	wantExpiry := time.Date(2012, 12, 21, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	if got.Expiry != wantExpiry || got.Raw != header {
		t.Fatalf("parseRestoreHeader(%q) = %#v, want expiry %q", header, got, wantExpiry)
	}

	ongoing := `ongoing-request="true"`
	if got := parseRestoreHeader(&ongoing); got == nil || !got.Ongoing || got.Expiry != "" {
		t.Fatalf("parseRestoreHeader(%q) = %#v", ongoing, got)
	}
}

func TestResolveStorageClass(t *testing.T) {
	d := &S3{Addition: Addition{StorageClass: " glacier-ir "}}
	if got := d.resolveStorageClass(); got != "GLACIER_IR" {
		t.Fatalf("resolveStorageClass() = %q, want GLACIER_IR", got)
	}
}
