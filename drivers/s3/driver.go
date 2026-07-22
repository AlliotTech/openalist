package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	stdpath "path"
	"strings"
	"time"

	"github.com/AlliotTech/openalist/internal/driver"
	"github.com/AlliotTech/openalist/internal/model"
	"github.com/AlliotTech/openalist/internal/stream"
	"github.com/AlliotTech/openalist/pkg/cron"
	"github.com/AlliotTech/openalist/server/common"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
)

type S3 struct {
	model.Storage
	Addition
	awsConfig  aws.Config
	client     *awss3.Client
	linkClient *awss3.Client

	config driver.Config
	cron   *cron.Cron
}

func (d *S3) Config() driver.Config {
	return d.config
}

func (d *S3) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *S3) Init(ctx context.Context) error {
	if d.Region == "" {
		d.Region = "alist"
	}
	if d.config.Name == "Doge" {
		// 多吉云每次临时生成的秘钥有效期为 2h，所以这里设置为 118 分钟重新生成一次
		d.cron = cron.NewCron(time.Minute * 118)
		d.cron.Do(func() {
			err := d.initSession(context.Background())
			if err != nil {
				log.Errorln("Doge init session error:", err)
			}
		})
	}
	err := d.initSession(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *S3) Drop(ctx context.Context) error {
	if d.cron != nil {
		d.cron.Stop()
	}
	return nil
}

func (d *S3) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if d.ListObjectVersion == "v2" {
		return d.listV2(ctx, dir.GetPath(), args)
	}
	return d.listV1(ctx, dir.GetPath(), args)
}

func (d *S3) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	path := getKey(file.GetPath(), false)
	filename := stdpath.Base(path)
	disposition := fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename))
	if d.AddFilenameToDisposition {
		disposition = fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, url.PathEscape(filename))
	}
	input := &awss3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(path),
		//ResponseContentDisposition: &disposition,
	}
	if d.CustomHost == "" {
		input.ResponseContentDisposition = aws.String(disposition)
	}
	var link model.Link
	expires := time.Hour * time.Duration(d.SignURLExpire)
	if d.CustomHost != "" {
		if d.EnableCustomHostPresign {
			result, err := awss3.NewPresignClient(d.linkClient).PresignGetObject(ctx, input, func(options *awss3.PresignOptions) {
				options.Expires = expires
			})
			if err != nil {
				return nil, err
			}
			link.URL = result.URL
		} else {
			var err error
			link.URL, err = d.customObjectURL(path)
			if err != nil {
				return nil, err
			}
		}
		if d.RemoveBucket {
			link.URL = strings.Replace(link.URL, "/"+d.Bucket, "", 1)
		}
	} else {
		result, err := awss3.NewPresignClient(d.linkClient).PresignGetObject(ctx, input, func(options *awss3.PresignOptions) {
			options.Expires = expires
		})
		if err != nil {
			return nil, err
		}
		link.URL = result.URL
		if common.ShouldProxy(d, filename) {
			link.Header = result.SignedHeader
		}
	}
	return &link, nil
}

func (d *S3) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return d.Put(ctx, &model.Object{
		Path: stdpath.Join(parentDir.GetPath(), dirName),
	}, &stream.FileStream{
		Obj: &model.Object{
			Name:     getPlaceholderName(d.Placeholder),
			Modified: time.Now(),
		},
		Reader:   io.NopCloser(bytes.NewReader([]byte{})),
		Mimetype: "application/octet-stream",
	}, func(float64) {})
}

func (d *S3) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	err := d.Copy(ctx, srcObj, dstDir)
	if err != nil {
		return err
	}
	return d.Remove(ctx, srcObj)
}

func (d *S3) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	err := d.copy(ctx, srcObj.GetPath(), stdpath.Join(stdpath.Dir(srcObj.GetPath()), newName), srcObj.IsDir())
	if err != nil {
		return err
	}
	return d.Remove(ctx, srcObj)
}

func (d *S3) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.copy(ctx, srcObj.GetPath(), stdpath.Join(dstDir.GetPath(), srcObj.GetName()), srcObj.IsDir())
}

func (d *S3) Remove(ctx context.Context, obj model.Obj) error {
	if obj.IsDir() {
		return d.removeDir(ctx, obj.GetPath())
	}
	return d.removeFile(ctx, obj.GetPath())
}

func (d *S3) Put(ctx context.Context, dstDir model.Obj, s model.FileStreamer, up driver.UpdateProgress) error {
	key := getKey(stdpath.Join(dstDir.GetPath(), s.GetName()), false)
	log.Debugln("key:", key)
	uploader := manager.NewUploader(d.client, func(uploader *manager.Uploader) {
		if s.GetSize() > int64(manager.MaxUploadParts)*manager.DefaultUploadPartSize {
			uploader.PartSize = s.GetSize()/int64(manager.MaxUploadParts-1) + 1
		}
	})
	input := &awss3.PutObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(key),
		Body: driver.NewLimitedUploadStream(ctx, &driver.ReaderUpdatingProgress{
			Reader:         s,
			UpdateProgress: up,
		}),
		ContentType: aws.String(s.GetMimetype()),
	}
	if storageClass := d.resolveStorageClass(); storageClass != "" {
		input.StorageClass = types.StorageClass(storageClass)
	}
	_, err := uploader.Upload(ctx, input)
	return err
}

func (d *S3) resolveStorageClass() string {
	value := strings.TrimSpace(strings.ReplaceAll(d.StorageClass, "-", "_"))
	if value == "" {
		return ""
	}
	if normalized, ok := storageClassLookup[strings.ToLower(value)]; ok {
		return normalized
	}
	return strings.ToUpper(value)
}

var storageClassLookup = map[string]string{
	"standard":            "STANDARD",
	"reduced_redundancy":  "REDUCED_REDUNDANCY",
	"standard_ia":         "STANDARD_IA",
	"onezone_ia":          "ONEZONE_IA",
	"intelligent_tiering": "INTELLIGENT_TIERING",
	"glacier":             "GLACIER",
	"glacier_ir":          "GLACIER_IR",
	"deep_archive":        "DEEP_ARCHIVE",
	"outposts":            "OUTPOSTS",
	"snow":                "SNOW",
	"express_onezone":     "EXPRESS_ONEZONE",
	"archive":             "ARCHIVE",
}

var (
	_ driver.Driver = (*S3)(nil)
	_ driver.Other  = (*S3)(nil)
)
