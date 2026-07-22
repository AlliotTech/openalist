package s3

import (
	"context"
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/AlliotTech/openalist/drivers/base"
	"github.com/AlliotTech/openalist/internal/model"
	"github.com/AlliotTech/openalist/internal/op"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	log "github.com/sirupsen/logrus"
)

// do others that not defined in Driver interface

func (d *S3) initSession(ctx context.Context) error {
	accessKeyID, secretAccessKey, sessionToken := d.AccessKeyID, d.SecretAccessKey, d.SessionToken
	if d.config.Name == "Doge" {
		credentialsTmp, err := getCredentials(d.AccessKeyID, d.SecretAccessKey)
		if err != nil {
			return err
		}
		accessKeyID, secretAccessKey, sessionToken = credentialsTmp.AccessKeyId, credentialsTmp.SecretAccessKey, credentialsTmp.SessionToken
	}
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(d.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken)),
	}
	if strings.TrimSpace(d.UserAgent) != "" {
		options = append(options, awsconfig.WithAPIOptions([]func(*middleware.Stack) error{withUserAgent(d.UserAgent)}))
	}
	if base.HttpClient != nil {
		options = append(options, awsconfig.WithHTTPClient(base.HttpClient))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return err
	}
	d.awsConfig = cfg
	d.client = d.newClient(d.Endpoint)
	d.linkClient = d.client
	if d.CustomHost != "" && d.EnableCustomHostPresign {
		d.linkClient, err = d.newCustomHostClient()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *S3) newClient(endpoint string) *awss3.Client {
	return awss3.NewFromConfig(d.awsConfig, func(options *awss3.Options) {
		if endpoint != "" {
			options.BaseEndpoint = aws.String(normalizeEndpoint(endpoint))
		}
		options.UsePathStyle = d.ForcePathStyle
	})
}

type customHostEndpointResolver struct {
	endpoint url.URL
}

func (r customHostEndpointResolver) ResolveEndpoint(
	_ context.Context,
	params awss3.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	endpoint := r.endpoint
	if aws.ToBool(params.ForcePathStyle) && params.Bucket != nil {
		endpoint = *endpoint.JoinPath(*params.Bucket)
	}
	return smithyendpoints.Endpoint{URI: endpoint}, nil
}

func (d *S3) newCustomHostClient() (*awss3.Client, error) {
	endpoint, err := url.Parse(normalizeEndpoint(d.CustomHost))
	if err != nil {
		return nil, err
	}
	client := awss3.NewFromConfig(d.awsConfig, func(options *awss3.Options) {
		options.EndpointResolverV2 = customHostEndpointResolver{endpoint: *endpoint}
		options.UsePathStyle = d.ForcePathStyle
	})
	return client, nil
}

func normalizeEndpoint(endpoint string) string {
	if strings.Contains(endpoint, "://") {
		return endpoint
	}
	return "https://" + endpoint
}

func withUserAgent(userAgent string) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		if strings.TrimSpace(userAgent) == "" {
			return nil
		}
		return stack.Build.Add(middleware.BuildMiddlewareFunc("s3-custom-user-agent", func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (middleware.BuildOutput, middleware.Metadata, error) {
			if req, ok := in.Request.(*smithyhttp.Request); ok {
				req.Header.Set("User-Agent", userAgent)
			}
			return next.HandleBuild(ctx, in)
		}), middleware.After)
	}
}

func (d *S3) customObjectURL(key string) (string, error) {
	endpoint, err := url.Parse(normalizeEndpoint(d.CustomHost))
	if err != nil {
		return "", err
	}
	objectPath := strings.TrimPrefix(key, "/")
	if d.ForcePathStyle && !d.RemoveBucket {
		objectPath = path.Join(d.Bucket, objectPath)
	}
	endpoint.Path = strings.TrimSuffix(endpoint.Path, "/") + "/" + objectPath
	return endpoint.String(), nil
}

func getKey(path string, dir bool) string {
	path = strings.TrimPrefix(path, "/")
	if path != "" && dir {
		path += "/"
	}
	return path
}

var defaultPlaceholderName = ".alist"

func getPlaceholderName(placeholder string) string {
	if placeholder == "" {
		return defaultPlaceholderName
	}
	return placeholder
}

func (d *S3) listV1(ctx context.Context, prefix string, args model.ListArgs) ([]model.Obj, error) {
	prefix = getKey(prefix, true)
	log.Debugf("list: %s", prefix)
	files := make([]model.Obj, 0)
	marker := ""
	for {
		input := &awss3.ListObjectsInput{
			Bucket:    aws.String(d.Bucket),
			Marker:    aws.String(marker),
			Prefix:    aws.String(prefix),
			Delimiter: aws.String("/"),
		}
		listObjectsResult, err := d.client.ListObjects(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, object := range listObjectsResult.CommonPrefixes {
			name := path.Base(strings.Trim(*object.Prefix, "/"))
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Modified: d.Modified,
				IsFolder: true,
			}
			files = append(files, &file)
		}
		for _, object := range listObjectsResult.Contents {
			name := path.Base(*object.Key)
			if !args.S3ShowPlaceholder && (name == getPlaceholderName(d.Placeholder) || name == d.Placeholder) {
				continue
			}
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Size:     *object.Size,
				Modified: *object.LastModified,
			}
			files = append(files, &file)
		}
		if listObjectsResult.IsTruncated == nil {
			return nil, errors.New("IsTruncated nil")
		}
		if *listObjectsResult.IsTruncated {
			marker = *listObjectsResult.NextMarker
		} else {
			break
		}
	}
	return files, nil
}

func (d *S3) listV2(ctx context.Context, prefix string, args model.ListArgs) ([]model.Obj, error) {
	prefix = getKey(prefix, true)
	files := make([]model.Obj, 0)
	var continuationToken, startAfter *string
	for {
		input := &awss3.ListObjectsV2Input{
			Bucket:            aws.String(d.Bucket),
			ContinuationToken: continuationToken,
			Prefix:            aws.String(prefix),
			Delimiter:         aws.String("/"),
			StartAfter:        startAfter,
		}
		listObjectsResult, err := d.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, err
		}
		log.Debugf("resp: %+v", listObjectsResult)
		for _, object := range listObjectsResult.CommonPrefixes {
			name := path.Base(strings.Trim(*object.Prefix, "/"))
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Modified: d.Modified,
				IsFolder: true,
			}
			files = append(files, &file)
		}
		for _, object := range listObjectsResult.Contents {
			if strings.HasSuffix(*object.Key, "/") {
				continue
			}
			name := path.Base(*object.Key)
			if !args.S3ShowPlaceholder && (name == getPlaceholderName(d.Placeholder) || name == d.Placeholder) {
				continue
			}
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Size:     *object.Size,
				Modified: *object.LastModified,
			}
			files = append(files, &file)
		}
		if listObjectsResult.IsTruncated == nil || !*listObjectsResult.IsTruncated {
			break
		}
		if listObjectsResult.NextContinuationToken != nil {
			continuationToken = listObjectsResult.NextContinuationToken
			continue
		}
		if len(listObjectsResult.Contents) == 0 {
			break
		}
		startAfter = listObjectsResult.Contents[len(listObjectsResult.Contents)-1].Key
	}
	return files, nil
}

func (d *S3) copy(ctx context.Context, src string, dst string, isDir bool) error {
	if isDir {
		return d.copyDir(ctx, src, dst)
	}
	return d.copyFile(ctx, src, dst)
}

func (d *S3) copyFile(ctx context.Context, src string, dst string) error {
	srcKey := getKey(src, false)
	dstKey := getKey(dst, false)
	input := &awss3.CopyObjectInput{
		Bucket:     aws.String(d.Bucket),
		CopySource: aws.String(url.PathEscape(d.Bucket + "/" + srcKey)),
		Key:        aws.String(dstKey),
	}
	if storageClass := d.resolveStorageClass(); storageClass != "" {
		input.StorageClass = types.StorageClass(storageClass)
	}
	_, err := d.client.CopyObject(ctx, input)
	return err
}

func (d *S3) copyDir(ctx context.Context, src string, dst string) error {
	objs, err := op.List(ctx, d, src, model.ListArgs{S3ShowPlaceholder: true})
	if err != nil {
		return err
	}
	for _, obj := range objs {
		cSrc := path.Join(src, obj.GetName())
		cDst := path.Join(dst, obj.GetName())
		if obj.IsDir() {
			err = d.copyDir(ctx, cSrc, cDst)
		} else {
			err = d.copyFile(ctx, cSrc, cDst)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *S3) removeDir(ctx context.Context, src string) error {
	objs, err := op.List(ctx, d, src, model.ListArgs{})
	if err != nil {
		return err
	}
	for _, obj := range objs {
		cSrc := path.Join(src, obj.GetName())
		if obj.IsDir() {
			err = d.removeDir(ctx, cSrc)
		} else {
			err = d.removeFile(ctx, cSrc)
		}
		if err != nil {
			return err
		}
	}
	_ = d.removeFile(ctx, path.Join(src, getPlaceholderName(d.Placeholder)))
	_ = d.removeFile(ctx, path.Join(src, d.Placeholder))
	return nil
}

func (d *S3) removeFile(ctx context.Context, src string) error {
	key := getKey(src, false)
	input := &awss3.DeleteObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(key),
	}
	_, err := d.client.DeleteObject(ctx, input)
	return err
}
