package ftp

import (
	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/op"
	"github.com/axgle/mahonia"
)

func encode(str string, encoding string) string {
	if encoding == "" {
		return str
	}
	encoder := mahonia.NewEncoder(encoding)
	return encoder.ConvertString(str)
}

func decode(str string, encoding string) string {
	if encoding == "" {
		return str
	}
	decoder := mahonia.NewDecoder(encoding)
	return decoder.ConvertString(str)
}

type Addition struct {
	Address               string `json:"address" required:"true"`
	Encoding              string `json:"encoding" required:"false"`
	Username              string `json:"username" required:"true"`
	Password              string `json:"password" required:"true"`
	TLSMode               string `json:"tls_mode" type:"select" options:"None,Explicit,Implicit" default:"None" help:"Explicit enables STARTTLS on the control connection; Implicit starts TLS immediately."`
	TLSInsecureSkipVerify bool   `json:"tls_insecure_skip_verify" default:"false" help:"Allow insecure TLS connections with untrusted certificates."`
	driver.RootPath
}

var config = driver.Config{
	Name:        "FTP",
	LocalSort:   true,
	OnlyLocal:   true,
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &FTP{}
	})
}
