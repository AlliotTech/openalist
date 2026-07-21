package base

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3UploadArgs struct {
	Endpoint           string
	Region             string
	AccessKeyID        string
	SecretAccessKey    string
	SessionToken       string
	Bucket             string
	Key                string
	ContentType        string
	ContentDisposition string
	Expires            *time.Time
	Body               io.Reader
	Size               int64
	UsePathStyle       bool
	Concurrency        int
}

func UploadToS3(ctx context.Context, args S3UploadArgs) error {
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(args.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			args.AccessKeyID,
			args.SecretAccessKey,
			args.SessionToken,
		)),
	}
	if HttpClient != nil {
		options = append(options, awsconfig.WithHTTPClient(HttpClient))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return err
	}

	client := awss3.NewFromConfig(cfg, func(options *awss3.Options) {
		if args.Endpoint != "" {
			options.BaseEndpoint = aws.String(normalizeS3Endpoint(args.Endpoint))
		}
		options.UsePathStyle = args.UsePathStyle
	})
	uploader := manager.NewUploader(client, func(uploader *manager.Uploader) {
		if args.Concurrency > 0 {
			uploader.Concurrency = args.Concurrency
		}
		if args.Size > int64(manager.MaxUploadParts)*manager.DefaultUploadPartSize {
			uploader.PartSize = args.Size/int64(manager.MaxUploadParts-1) + 1
		}
	})

	input := &awss3.PutObjectInput{
		Bucket:  aws.String(args.Bucket),
		Key:     aws.String(args.Key),
		Body:    args.Body,
		Expires: args.Expires,
	}
	if args.ContentType != "" {
		input.ContentType = aws.String(args.ContentType)
	}
	if args.ContentDisposition != "" {
		input.ContentDisposition = aws.String(args.ContentDisposition)
	}
	_, err = uploader.Upload(ctx, input)
	return err
}

func normalizeS3Endpoint(endpoint string) string {
	if strings.Contains(endpoint, "://") {
		return endpoint
	}
	return "https://" + endpoint
}
