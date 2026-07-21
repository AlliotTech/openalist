package base

import "testing"

func TestNormalizeS3Endpoint(t *testing.T) {
	tests := map[string]string{
		"cos.example.com":         "https://cos.example.com",
		"https://cos.example.com": "https://cos.example.com",
		"http://127.0.0.1:9000":   "http://127.0.0.1:9000",
	}
	for input, expected := range tests {
		if actual := normalizeS3Endpoint(input); actual != expected {
			t.Fatalf("normalizeS3Endpoint(%q) = %q, want %q", input, actual, expected)
		}
	}
}
