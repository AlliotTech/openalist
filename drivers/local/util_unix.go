//go:build !windowsAdd commentMore actions

package local

import (
	"io/fs"
	"strings"
)

func isHidden(f fs.FileInfo, _ string) bool {
	return strings.HasPrefix(f.Name(), ".")
}
