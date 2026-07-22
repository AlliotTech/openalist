package pikpak

import (
	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/op"
)

type Addition struct {
	driver.RootID
	Username             string `json:"username" required:"true"`
	Password             string `json:"password" required:"true"`
	Platform             string `json:"platform" required:"true" default:"web" type:"select" options:"android,web,pc"`
	RefreshToken         string `json:"refresh_token" required:"true" default:""`
	CaptchaToken         string `json:"captcha_token" default:""`
	DeviceID             string `json:"device_id"  required:"false" default:""`
	DisableMediaLink     bool   `json:"disable_media_link" default:"true"`
	APIDomain            string `json:"api_domain" type:"select" options:"mypikpak_net,mypikpak_com,pikpak_me,pikpakdrive_com" default:"mypikpak_net"`
	CustomAPIDomain      string `json:"custom_api_domain" help:"Custom base domain for API requests. It overrides api_domain."`
	DownloadDomain       string `json:"download_domain" type:"select" options:"original,mypikpak_net,mypikpak_com,pikpak_me,pikpakdrive_com" default:"original"`
	CustomDownloadDomain string `json:"custom_download_domain" help:"Custom base domain for download links. It overrides download_domain."`
}

var config = driver.Config{
	Name:        "PikPak",
	LocalSort:   true,
	DefaultRoot: "",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &PikPak{}
	})
}
