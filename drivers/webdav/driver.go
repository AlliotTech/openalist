package webdav

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/AlliotTech/openalist/drivers/base"
	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/errs"
	"github.com/AlliotTech/openalist/internal/model"
	"github.com/AlliotTech/openalist/pkg/cron"
	"github.com/AlliotTech/openalist/pkg/gowebdav"
	"github.com/AlliotTech/openalist/pkg/utils"
)

type WebDav struct {
	model.Storage
	Addition
	client *gowebdav.Client
	cron   *cron.Cron
}

func (d *WebDav) Config() driver.Config {
	return config
}

func (d *WebDav) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *WebDav) Init(ctx context.Context) error {
	err := d.setClient()
	if err == nil {
		d.cron = cron.NewCron(time.Hour * 12)
		d.cron.Do(func() {
			_ = d.setClient()
		})
	}
	return err
}

func (d *WebDav) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *WebDav) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.client.ReadDir(dir.GetPath())
	if err != nil {
		return nil, mapWebDAVError(err)
	}
	return utils.SliceConvert(files, func(src os.FileInfo) (model.Obj, error) {
		return &model.Object{
			Path:     path.Join(dir.GetPath(), src.Name()),
			Name:     src.Name(),
			Size:     src.Size(),
			Modified: src.ModTime(),
			IsFolder: src.IsDir(),
		}, nil
	})
}

func (d *WebDav) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	url, header, err := d.client.Link(file.GetPath())
	if err != nil {
		return nil, err
	}
	if args.Redirect {
		url, header, err = resolveWebDAVRedirect(ctx, url, header)
		if err != nil {
			return nil, err
		}
	}
	return &model.Link{
		URL:    url,
		Header: header,
	}, nil
}

func resolveWebDAVRedirect(ctx context.Context, rawURL string, header http.Header) (string, http.Header, error) {
	req := base.NoRedirectClient.R().SetContext(ctx).SetDoNotParseResponse(true)
	req.Header = header.Clone()
	res, err := req.Get(rawURL)
	if err != nil {
		return "", nil, err
	}
	if res.RawResponse != nil && res.RawResponse.Body != nil {
		defer res.RawResponse.Body.Close()
	}
	if res.StatusCode() >= http.StatusOK && res.StatusCode() < http.StatusMultipleChoices {
		return rawURL, header.Clone(), nil
	}
	switch res.StatusCode() {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return "", nil, fmt.Errorf("redirect failed, status: %d", res.StatusCode())
	}
	location := res.Header().Get("Location")
	if location == "" {
		return "", nil, fmt.Errorf("redirect failed: location is empty")
	}
	origin, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}
	target, err := origin.Parse(location)
	if err != nil {
		return "", nil, err
	}
	resultHeader := header.Clone()
	if !sameOrigin(origin, target) {
		resultHeader.Del("Authorization")
		resultHeader.Del("Cookie")
		resultHeader.Del("Proxy-Authorization")
	}
	return target.String(), resultHeader, nil
}

func sameOrigin(a, b *url.URL) bool {
	return a.Scheme == b.Scheme && a.Host == b.Host
}

func (d *WebDav) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return d.client.MkdirAll(path.Join(parentDir.GetPath(), dirName), 0644)
}

func (d *WebDav) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Rename(getPath(srcObj), path.Join(dstDir.GetPath(), srcObj.GetName()), true)
}

func (d *WebDav) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return d.client.Rename(getPath(srcObj), path.Join(path.Dir(srcObj.GetPath()), newName), true)
}

func (d *WebDav) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Copy(getPath(srcObj), path.Join(dstDir.GetPath(), srcObj.GetName()), true)
}

func (d *WebDav) Remove(ctx context.Context, obj model.Obj) error {
	return d.client.RemoveAll(getPath(obj))
}

func (d *WebDav) Put(ctx context.Context, dstDir model.Obj, s model.FileStreamer, up driver.UpdateProgress) error {
	callback := func(r *http.Request) {
		r.Header.Set("Content-Type", s.GetMimetype())
		r.ContentLength = s.GetSize()
	}
	reader := driver.NewLimitedUploadStream(ctx, &driver.ReaderUpdatingProgress{
		Reader:         s,
		UpdateProgress: up,
	})
	err := d.client.WriteStream(path.Join(dstDir.GetPath(), s.GetName()), reader, 0644, callback)
	return err
}

func (d *WebDav) Get(ctx context.Context, objPath string) (model.Obj, error) {
	objPath = path.Join(d.GetRootPath(), objPath)
	info, err := d.client.Stat(objPath)
	if err != nil {
		return nil, mapWebDAVError(err)
	}
	return &model.Object{
		Name:     info.Name(),
		Size:     info.Size(),
		Modified: info.ModTime(),
		IsFolder: info.IsDir(),
		Path:     objPath,
	}, nil
}

func mapWebDAVError(err error) error {
	if gowebdav.IsErrNotFound(err) {
		return errs.ObjectNotFound
	}
	return err
}

var _ driver.Driver = (*WebDav)(nil)
var _ driver.Getter = (*WebDav)(nil)
