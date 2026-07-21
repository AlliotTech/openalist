package http

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlliotTech/openalist/internal/model"
	internalnet "github.com/AlliotTech/openalist/internal/net"
	"github.com/AlliotTech/openalist/internal/offline_download/tool"
	"github.com/AlliotTech/openalist/pkg/utils"
)

type SimpleHttp struct {
	client *http.Client
}

func (s SimpleHttp) Name() string {
	return "SimpleHttp"
}

func (s SimpleHttp) Items() []model.SettingItem {
	return nil
}

func (s SimpleHttp) Init() (string, error) {
	return "ok", nil
}

func (s SimpleHttp) IsReady() bool {
	return true
}

func (s SimpleHttp) AddURL(args *tool.AddUrlArgs) (string, error) {
	panic("should not be called")
}

func (s SimpleHttp) Remove(task *tool.DownloadTask) error {
	panic("should not be called")
}

func (s SimpleHttp) Status(task *tool.DownloadTask) (*tool.Status, error) {
	panic("should not be called")
}

func (s SimpleHttp) Run(task *tool.DownloadTask) error {
	u := task.Url
	// parse url
	_u, err := url.Parse(u)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(task.Ctx(), http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	client := s.client
	if client == nil {
		client = internalnet.HttpClient()
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http status code %d", resp.StatusCode)
	}
	// If Path is empty, use Hostname; otherwise, filePath euqals TempDir which causes os.Create to fail
	urlPath := _u.Path
	if urlPath == "" {
		urlPath = strings.ReplaceAll(_u.Host, ".", "_")
	}
	filename, err := parseFilenameFromContentDisposition(resp.Header.Get("Content-Disposition"))
	if err != nil {
		filename, err = sanitizeFilename(urlPath)
	}
	if err != nil {
		filename = strings.ReplaceAll(_u.Host, ":", "_")
		filename, err = sanitizeFilename(filename)
		if err != nil {
			return err
		}
	}
	// save to temp dir
	if err := os.MkdirAll(task.TempDir, os.ModePerm); err != nil {
		return err
	}
	filePath := filepath.Join(task.TempDir, filename)
	cleanTempDir := filepath.Clean(task.TempDir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(filePath)+string(filepath.Separator), cleanTempDir) {
		return fmt.Errorf("filename illegal")
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fileSize := resp.ContentLength
	task.SetTotalBytes(fileSize)
	err = utils.CopyWithCtx(task.Ctx(), file, resp.Body, fileSize, task.SetProgress)
	return err
}

func init() {
	tool.Tools.Add(&SimpleHttp{})
}
