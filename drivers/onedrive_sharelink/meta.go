package onedrive_sharelink

import (
	"net/http"

	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/op"
)

type Addition struct {
	driver.RootPath
	ShareLinkURL       string `json:"url" required:"true"`
	ShareLinkPassword  string `json:"password"`
	IsSharepoint       bool
	downloadLinkPrefix string
	Headers            http.Header
	HeaderTime         int64
}

var config = driver.Config{
	Name:        "Onedrive Sharelink",
	OnlyProxy:   true,
	NoUpload:    true,
	DefaultRoot: "/",
	CheckStatus: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &OnedriveSharelink{}
	})
}
