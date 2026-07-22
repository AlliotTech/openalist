package s3

import (
	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/op"
)

type Addition struct {
	driver.RootPath
	Bucket                   string `json:"bucket" required:"true"`
	Endpoint                 string `json:"endpoint" required:"true"`
	Region                   string `json:"region"`
	AccessKeyID              string `json:"access_key_id" required:"true"`
	SecretAccessKey          string `json:"secret_access_key" required:"true"`
	SessionToken             string `json:"session_token"`
	CustomHost               string `json:"custom_host"`
	EnableCustomHostPresign  bool   `json:"enable_custom_host_presign"`
	SignURLExpire            int    `json:"sign_url_expire" type:"number" default:"4"`
	Placeholder              string `json:"placeholder"`
	ForcePathStyle           bool   `json:"force_path_style"`
	ListObjectVersion        string `json:"list_object_version" type:"select" options:"v1,v2" default:"v1"`
	RemoveBucket             bool   `json:"remove_bucket" help:"Remove bucket name from path when using custom host."`
	AddFilenameToDisposition bool   `json:"add_filename_to_disposition" help:"Add filename to Content-Disposition header."`
	StorageClass             string `json:"storage_class" type:"select" options:",standard,reduced_redundancy,standard_ia,onezone_ia,intelligent_tiering,glacier,glacier_ir,deep_archive,outposts,snow,express_onezone,archive" help:"Storage class for new objects; supported values depend on the S3 provider."`
	UserAgent                string `json:"user_agent" help:"Custom User-Agent for S3 requests."`
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &S3{
			config: driver.Config{
				Name:        "S3",
				DefaultRoot: "/",
				LocalSort:   true,
				CheckStatus: true,
			},
		}
	})
	op.RegisterDriver(func() driver.Driver {
		return &S3{
			config: driver.Config{
				Name:        "Doge",
				DefaultRoot: "/",
				LocalSort:   true,
				CheckStatus: true,
			},
		}
	})
}
