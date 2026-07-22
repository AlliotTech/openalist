package sftp

import (
	"testing"
	"time"
)

func TestConnectTimeout(t *testing.T) {
	if got := (&SFTP{}).connectTimeout(); got != 10*time.Second {
		t.Fatalf("default timeout = %s, want 10s", got)
	}
	if got := (&SFTP{Addition: Addition{ConnectTimeout: 25}}).connectTimeout(); got != 25*time.Second {
		t.Fatalf("configured timeout = %s, want 25s", got)
	}
}
