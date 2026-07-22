package github_releases

import (
	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/op"
)

type Addition struct {
	driver.RootID
	RepoStructure  string `json:"repo_structure" type:"text" required:"true" default:"alistGo/alist" help:"structure:[path:]org/repo"`
	ShowReadme     bool   `json:"show_readme" type:"bool" default:"true" help:"show README、LICENSE file"`
	Token          string `json:"token" type:"string" required:"false" help:"GitHub token, if you want to access private repositories or increase the rate limit"`
	ShowSourceCode bool   `json:"show_source_code" type:"bool" default:"false" help:"show Source code (zip/tar.gz)"`
	ShowAllVersion bool   `json:"show_all_version" type:"bool" default:"false" help:"show all versions"`
	PerPage        int    `json:"per_page" type:"number" default:"30" help:"releases per page (max 100), only works when show all versions"`
	MaxPage        int    `json:"max_page" type:"number" default:"0" help:"max pages to fetch (0 = unlimited), only works when show all versions"`
	GitHubProxy    string `json:"gh_proxy" type:"string" default:"" help:"GitHub proxy, e.g. https://ghproxy.net/github.com or https://gh-proxy.com/github.com "`
}

var config = driver.Config{
	Name:              "GitHub Releases",
	LocalSort:         false,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &GithubReleases{}
	})
}
