package v3_45_14

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/AlliotTech/openalist/cmd/flags"
	"github.com/AlliotTech/openalist/internal/conf"
	"github.com/AlliotTech/openalist/pkg/utils"
)

func ResetSkipTLSVerify() {
	if conf.Conf == nil || !conf.Conf.TlsInsecureSkipVerify || !strings.HasPrefix(conf.Version, "v") {
		return
	}

	conf.Conf.TlsInsecureSkipVerify = false
	confBody, err := utils.Json.MarshalIndent(conf.Conf, "", "  ")
	if err != nil {
		utils.Log.Errorf("[ResetSkipTLSVerify] failed to marshal config: %+v", err)
		return
	}
	configPath := filepath.Join(flags.DataDir, "config.json")
	if err := os.WriteFile(configPath, confBody, 0o600); err != nil {
		utils.Log.Errorf("[ResetSkipTLSVerify] failed to rewrite config: %+v", err)
		return
	}
	utils.Log.Infof("[ResetSkipTLSVerify] set tls_insecure_skip_verify to false")
}
