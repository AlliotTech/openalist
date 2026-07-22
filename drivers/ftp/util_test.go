package ftp

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestFileReaderSeek(t *testing.T) {
	r := NewFileReader(nil, "/file", 100)
	if got, err := r.Seek(-10, io.SeekEnd); err != nil || got != 90 {
		t.Fatalf("Seek(-10, SeekEnd) = (%d, %v), want (90, nil)", got, err)
	}
	if got, err := r.Seek(-90, io.SeekCurrent); err != nil || got != 0 {
		t.Fatalf("Seek(-90, SeekCurrent) = (%d, %v), want (0, nil)", got, err)
	}
	if got, err := r.Seek(-1, io.SeekStart); err != os.ErrInvalid || got != 0 {
		t.Fatalf("Seek(-1, SeekStart) = (%d, %v), want (0, %v)", got, err, os.ErrInvalid)
	}
}

func TestFileReaderReadAtRejectsNegativeOffset(t *testing.T) {
	r := NewFileReader(nil, "/file", 1)
	if n, err := r.ReadAt(make([]byte, 1), -1); n != 0 || err != os.ErrInvalid {
		t.Fatalf("ReadAt(-1) = (%d, %v), want (0, %v)", n, err, os.ErrInvalid)
	}
}

func TestFTPServerName(t *testing.T) {
	tests := map[string]string{
		"ftp.example.com:21": "ftp.example.com",
		"ftp.example.com":    "ftp.example.com",
		"[2001:db8::1]:990":  "2001:db8::1",
	}
	for address, want := range tests {
		if got := ftpServerName(address); got != want {
			t.Fatalf("ftpServerName(%q) = %q, want %q", address, got, want)
		}
	}
}

func TestLoginContextWithoutTLSDoesNotRequireNetwork(t *testing.T) {
	d := &FTP{Addition: Addition{Address: "127.0.0.1:1", TLSMode: "None"}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := d.login(ctx); err == nil {
		t.Fatal("login() error = nil, want canceled dial error")
	}
}
