package v3_45_14

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlliotTech/openalist/cmd/flags"
	"github.com/AlliotTech/openalist/internal/conf"
	"github.com/AlliotTech/openalist/pkg/utils"
)

func TestResetSkipTLSVerifyUpdatesExistingConfig(t *testing.T) {
	originalDataDir := flags.DataDir
	originalConfig := conf.Conf
	originalVersion := conf.Version
	defer func() {
		flags.DataDir = originalDataDir
		conf.Conf = originalConfig
		conf.Version = originalVersion
	}()

	flags.DataDir = t.TempDir()
	conf.Version = "v3.45.14"
	conf.Conf = conf.DefaultConfig()
	conf.Conf.TlsInsecureSkipVerify = true
	configPath := filepath.Join(flags.DataDir, "config.json")
	if !utils.WriteJsonToFile(configPath, conf.Conf) {
		t.Fatal("write test config")
	}

	ResetSkipTLSVerify()

	if conf.Conf.TlsInsecureSkipVerify {
		t.Fatal("ResetSkipTLSVerify did not update the in-memory config")
	}
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read rewritten config: %v", err)
	}
	var persisted conf.Config
	if err := utils.Json.Unmarshal(configBytes, &persisted); err != nil {
		t.Fatalf("decode rewritten config: %v", err)
	}
	if persisted.TlsInsecureSkipVerify {
		t.Fatal("ResetSkipTLSVerify did not update the persisted config")
	}
}

func TestResetSkipTLSVerifySkipsDevelopmentBuild(t *testing.T) {
	originalConfig := conf.Conf
	originalVersion := conf.Version
	defer func() {
		conf.Conf = originalConfig
		conf.Version = originalVersion
	}()

	conf.Version = "dev"
	conf.Conf = conf.DefaultConfig()
	conf.Conf.TlsInsecureSkipVerify = true

	ResetSkipTLSVerify()

	if !conf.Conf.TlsInsecureSkipVerify {
		t.Fatal("ResetSkipTLSVerify changed development-build configuration")
	}
}
