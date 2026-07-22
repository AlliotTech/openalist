package ftp

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jlaffaye/ftp"
)

// do others that not defined in Driver interface

func (d *FTP) login(ctx context.Context) error {
	if d.conn != nil {
		_, err := d.conn.CurrentDir()
		if err == nil {
			return nil
		}
		_ = d.conn.Quit()
		d.conn = nil
	}
	opts := []ftp.DialOption{
		ftp.DialWithShutTimeout(10 * time.Second),
		ftp.DialWithContext(ctx),
	}
	if tlsMode := strings.ToLower(strings.TrimSpace(d.TLSMode)); tlsMode == "explicit" || tlsMode == "implicit" {
		tlsConfig := &tls.Config{ServerName: ftpServerName(d.Address), InsecureSkipVerify: d.TLSInsecureSkipVerify}
		if tlsMode == "implicit" {
			opts = append(opts, ftp.DialWithTLS(tlsConfig))
		} else {
			opts = append(opts, ftp.DialWithExplicitTLS(tlsConfig))
		}
	}
	conn, err := ftp.Dial(d.Address, opts...)
	if err != nil {
		return err
	}
	err = conn.Login(d.Username, d.Password)
	if err != nil {
		_ = conn.Quit()
		return err
	}
	d.conn = conn
	return nil
}

func ftpServerName(address string) string {
	if host, _, err := net.SplitHostPort(address); err == nil {
		return host
	}
	return address
}

// FileReader An FTP file reader that implements io.MFile for seeking.
type FileReader struct {
	conn         *ftp.ServerConn
	resp         *ftp.Response
	offset       atomic.Int64
	readAtOffset int64
	mu           sync.Mutex
	path         string
	size         int64
}

func NewFileReader(conn *ftp.ServerConn, path string, size int64) *FileReader {
	return &FileReader{
		conn: conn,
		path: path,
		size: size,
	}
}

func (r *FileReader) Read(buf []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	off := r.offset.Load()
	n, err = r.readAtLocked(buf, off)
	r.offset.Add(int64(n))
	return
}

func (r *FileReader) ReadAt(buf []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, os.ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.readAtLocked(buf, off)
}

func (r *FileReader) readAtLocked(buf []byte, off int64) (n int, err error) {
	if r.resp != nil && off != r.readAtOffset {
		_ = r.resp.Close()
		r.resp = nil
	}

	if r.resp == nil {
		r.resp, err = r.conn.RetrFrom(r.path, uint64(off))
		r.readAtOffset = off
		if err != nil {
			return 0, err
		}
	}

	n, err = r.resp.Read(buf)
	r.readAtOffset += int64(n)
	return
}

func (r *FileReader) Seek(offset int64, whence int) (int64, error) {
	oldOffset := r.offset.Load()
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = oldOffset + offset
	case io.SeekEnd:
		newOffset = r.size + offset
	default:
		return -1, os.ErrInvalid
	}

	if newOffset < 0 {
		// offset out of range
		return oldOffset, os.ErrInvalid
	}
	if newOffset == oldOffset {
		// offset not changed, so return directly
		return oldOffset, nil
	}
	r.offset.Store(newOffset)
	return newOffset, nil
}

func (r *FileReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resp != nil {
		err := r.resp.Close()
		r.resp = nil
		return err
	}
	return nil
}
