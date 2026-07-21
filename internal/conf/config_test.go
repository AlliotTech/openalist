package conf

import "testing"

func TestDefaultConfigVerifiesTLSCertificates(t *testing.T) {
	config := DefaultConfig()
	if config.TlsInsecureSkipVerify {
		t.Fatal("DefaultConfig enables insecure TLS certificate skipping")
	}
}
