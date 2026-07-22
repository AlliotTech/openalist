package _139

import "testing"

func TestPickThumbnailPrefersLargeVariant(t *testing.T) {
	d := &Yun139{Addition: Addition{UseLargeThumbnail: true}}
	if got := d.pickThumbnail("small", "large"); got != "large" {
		t.Fatalf("thumbnail = %q, want large variant", got)
	}
	if got := d.pickThumbnail("small", ""); got != "small" {
		t.Fatalf("thumbnail = %q, want regular variant when large is empty", got)
	}
}

func TestPickThumbnailKeepsRegularVariantWhenDisabled(t *testing.T) {
	d := &Yun139{}
	if got := d.pickThumbnail("small", "large"); got != "small" {
		t.Fatalf("thumbnail = %q, want regular variant", got)
	}
}

func TestEnsurePersonalCloudHostUsesExistingHost(t *testing.T) {
	d := &Yun139{PersonalCloudHost: "https://personal.example/"}
	if err := d.ensurePersonalCloudHost(); err != nil {
		t.Fatalf("ensurePersonalCloudHost returned error: %v", err)
	}
}

func TestEnsurePersonalCloudHostRequiresAuthorization(t *testing.T) {
	d := &Yun139{}
	if err := d.ensurePersonalCloudHost(); err == nil {
		t.Fatal("ensurePersonalCloudHost accepted missing authorization")
	}
}
