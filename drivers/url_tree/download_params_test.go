package url_tree

import (
	"context"
	"testing"

	"github.com/AlliotTech/openalist/internal/model"
)

func TestDownloadParams(t *testing.T) {
	d := &Urls{Addition: Addition{
		UrlStructure:   "测试 & 文件.mp4:https://cdn.example.com/file?token=abc#play",
		DownloadParams: "attname={{.Name}}&path={{.Path}}&size={{.Size}}",
	}}
	if err := d.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	file, err := d.Get(context.Background(), "/测试 & 文件.mp4")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	preview, err := d.Link(context.Background(), file, model.LinkArgs{Type: "preview"})
	if err != nil {
		t.Fatalf("preview Link() error = %v", err)
	}
	if want := "https://cdn.example.com/file?token=abc#play"; preview.URL != want {
		t.Fatalf("preview URL = %q, want %q", preview.URL, want)
	}

	download, err := d.Link(context.Background(), file, model.LinkArgs{})
	if err != nil {
		t.Fatalf("download Link() error = %v", err)
	}
	want := "https://cdn.example.com/file?token=abc&attname=%E6%B5%8B%E8%AF%95+%26+%E6%96%87%E4%BB%B6.mp4&path=%2F%E6%B5%8B%E8%AF%95+%26+%E6%96%87%E4%BB%B6.mp4&size=0#play"
	if download.URL != want {
		t.Fatalf("download URL = %q, want %q", download.URL, want)
	}
}

func TestDownloadParamsValidation(t *testing.T) {
	d := &Urls{Addition: Addition{UrlStructure: "file:https://example.com/file", DownloadParams: "name={{"}}
	if err := d.Init(context.Background()); err == nil {
		t.Fatal("Init() error = nil, want invalid template error")
	}

	d = &Urls{Addition: Addition{UrlStructure: "file:https://example.com/file", DownloadParams: "bad=%zz"}}
	if err := d.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	file, err := d.Get(context.Background(), "/file")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if _, err = d.Link(context.Background(), file, model.LinkArgs{}); err == nil {
		t.Fatal("Link() error = nil, want invalid query error")
	}
}
