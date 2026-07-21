package net

import (
	"errors"
	"testing"
	"time"

	"github.com/AlliotTech/openalist/internal/conf"
)

func TestNewHTTPClientBlocksCloudMetadataEndpoint(t *testing.T) {
	originalConfig := conf.Conf
	conf.Conf = conf.DefaultConfig()
	defer func() { conf.Conf = originalConfig }()

	endpoints := []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://100.100.100.200/latest/meta-data/",
		"http://[fd00:ec2::254]/latest/meta-data/",
	}
	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			client := NewHttpClient()
			client.Timeout = 250 * time.Millisecond
			_, err := client.Get(endpoint)
			if !errors.Is(err, ErrCloudMetadataEndpoint) {
				t.Fatalf("request error = %v, want %v", err, ErrCloudMetadataEndpoint)
			}
		})
	}
}
